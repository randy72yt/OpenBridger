package perfmetrics

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/perf_metrics_setting"
)

// CounterView is one source's aggregate for the admin summary.
type CounterView struct {
	Requests     int64   `json:"requests"`
	Success      int64   `json:"success"`
	AvgLatencyMs int64   `json:"avg_latency_ms"`
	AvgTtftMs    int64   `json:"avg_ttft_ms"`
	AvgTps       float64 `json:"avg_tps"`
}

// AdminModelPerf is the admin-only per-model view: source split, the public
// v1/v2 projections, availability, recent error codes and probe info. It never
// carries channel names, upstream names, base URLs or keys; LastProbeAt /
// ProbeTier / ProbeTokenParam are filled by the controller from the probe
// scheduler.
type AdminModelPerf struct {
	ModelName       string           `json:"model_name"`
	Real            CounterView      `json:"real"`
	Probe           CounterView      `json:"probe"`
	Internal        CounterView      `json:"internal"`
	ClientError     CounterView      `json:"client_error"`
	PublicV1        ModelSummary     `json:"public_v1"`
	PublicV2        ModelSummary     `json:"public_v2"`
	Availability24h float64          `json:"availability_24h"`
	Availability7d  float64          `json:"availability_7d"`
	ErrorCodes      map[string]int64 `json:"error_codes"`
	Series          []BucketPoint    `json:"series"`
	LastProbeAt     int64            `json:"last_probe_at"`
	ProbeTier       string           `json:"probe_tier"`
	ProbeTokenParam string           `json:"probe_token_param"`
}

// AdminPerfSummary aggregates the admin view. Probe budget numbers are filled
// by the controller from the probe scheduler.
type AdminPerfSummary struct {
	Models               []AdminModelPerf `json:"models"`
	ProbeBudgetUsedQuota int64            `json:"probe_budget_used_quota"`
	ProbeBudgetQuota     int64            `json:"probe_budget_quota"`
	GeneratedAt          int64            `json:"generated_at"`
	// ErrorCodesTruncated marks that the error-log scan hit its row limit and
	// the code counts cover only part of the window.
	ErrorCodesTruncated bool `json:"error_codes_truncated"`
}

// sourceBuckets maps model -> source -> bucketTs -> counters.
type sourceBuckets map[string]map[string]map[int64]counters

func (merged sourceBuckets) add(modelName string, source string, bucketTs int64, value counters, allowedSources map[string]struct{}) {
	if value.requestCount == 0 {
		return
	}
	if _, ok := allowedSources[source]; !ok {
		return
	}
	if _, ok := merged[modelName]; !ok {
		merged[modelName] = map[string]map[int64]counters{}
	}
	if _, ok := merged[modelName][source]; !ok {
		merged[modelName][source] = map[int64]counters{}
	}
	merged[modelName][source][bucketTs] = addCounters(merged[modelName][source][bucketTs], value)
}

// collectSourceBuckets merges persisted real buckets, aux rows and hot buckets
// for the allowed sources within the window. Groups nil means all groups.
func collectSourceBuckets(startTs int64, endTs int64, groups []string, allowedSources map[string]struct{}) sourceBuckets {
	merged := sourceBuckets{}
	if rows, err := model.GetPerfMetricsSummaryBucketsAll(startTs, endTs, groups); err == nil {
		for _, row := range rows {
			merged.add(row.ModelName, SourceReal, row.BucketTs, counters{
				requestCount:   row.RequestCount,
				successCount:   row.SuccessCount,
				totalLatencyMs: row.TotalLatencyMs,
				ttftSumMs:      row.TtftSumMs,
				ttftCount:      row.TtftCount,
				outputTokens:   row.OutputTokens,
				generationMs:   row.GenerationMs,
			}, allowedSources)
		}
	} else {
		common.SysError("perf query: failed to load perf buckets: " + err.Error())
	}
	if auxRows, err := model.GetPerfMetricsAuxBucketsAll(startTs, endTs, groups); err == nil {
		for _, row := range auxRows {
			merged.add(row.ModelName, row.Source, row.BucketTs, counters{
				requestCount:   row.RequestCount,
				successCount:   row.SuccessCount,
				totalLatencyMs: row.TotalLatencyMs,
				ttftSumMs:      row.TtftSumMs,
				ttftCount:      row.TtftCount,
				outputTokens:   row.OutputTokens,
				generationMs:   row.GenerationMs,
			}, allowedSources)
		}
	} else {
		common.SysError("perf query: failed to load perf aux buckets: " + err.Error())
	}

	allowedGroups := allowedGroupSet(groups)
	hotBuckets.Range(func(key, value any) bool {
		k := key.(bucketKey)
		if k.bucketTs < startTs || k.bucketTs > endTs {
			return true
		}
		if allowedGroups != nil {
			if _, ok := allowedGroups[k.group]; !ok {
				return true
			}
		}
		merged.add(k.model, k.source, k.bucketTs, value.(*atomicBucket).snapshot(), allowedSources)
		return true
	})
	return merged
}

func sourceSet(sources ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(sources))
	for _, source := range sources {
		set[source] = struct{}{}
	}
	return set
}

func addCounters(left counters, right counters) counters {
	left.requestCount += right.requestCount
	left.successCount += right.successCount
	left.totalLatencyMs += right.totalLatencyMs
	left.ttftSumMs += right.ttftSumMs
	left.ttftCount += right.ttftCount
	left.outputTokens += right.outputTokens
	left.generationMs += right.generationMs
	return left
}

func sumSources(totals map[string]counters, sources ...string) counters {
	total := counters{}
	for _, source := range sources {
		total = addCounters(total, totals[source])
	}
	return total
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func counterView(total counters) CounterView {
	return CounterView{
		Requests:     total.requestCount,
		Success:      total.successCount,
		AvgLatencyMs: avg(total.totalLatencyMs, total.requestCount),
		AvgTtftMs:    avg(total.ttftSumMs, total.ttftCount),
		AvgTps:       round2(avgTps(total)),
	}
}

func floatPtr(value float64) *float64 {
	return &value
}

func int64Ptr(value int64) *int64 {
	return &value
}

// foldSources folds one model's per-source buckets at or after sinceTs into
// per-source totals and per-source hourly series.
func foldSources(bySource map[string]map[int64]counters, sinceTs int64) (map[string]counters, map[string]map[int64]counters) {
	totals := map[string]counters{}
	series := map[string]map[int64]counters{}
	for source, buckets := range bySource {
		for bucketTs, value := range buckets {
			if bucketTs < sinceTs {
				continue
			}
			totals[source] = addCounters(totals[source], value)
			if series[source] == nil {
				series[source] = map[int64]counters{}
			}
			series[source][bucketTs] = addCounters(series[source][bucketTs], value)
		}
	}
	return totals, series
}

// sumRealProbeFromBuckets folds real+probe buckets within [startTs, endTs].
func sumRealProbeFromBuckets(bySource map[string]map[int64]counters, startTs int64, endTs int64) counters {
	total := counters{}
	for _, source := range []string{SourceReal, SourceProbe} {
		for bucketTs, value := range bySource[source] {
			if bucketTs >= startTs && bucketTs <= endTs {
				total = addCounters(total, value)
			}
		}
	}
	return total
}

// mergeSeries merges several per-source series into one bucket map.
func mergeSeries(series map[string]map[int64]counters, sources ...string) map[int64]counters {
	merged := map[int64]counters{}
	for _, source := range sources {
		for bucketTs, value := range series[source] {
			merged[bucketTs] = addCounters(merged[bucketTs], value)
		}
	}
	return merged
}

// bucketSeries converts merged bucket counters into sorted BucketPoints.
func bucketSeries(buckets map[int64]counters) []BucketPoint {
	if len(buckets) == 0 {
		return nil
	}
	timestamps := make([]int64, 0, len(buckets))
	for ts := range buckets {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })
	points := make([]BucketPoint, 0, len(timestamps))
	for _, ts := range timestamps {
		points = append(points, bucketPoint(ts, buckets[ts]))
	}
	return points
}

// buildModelSummaryV2 computes the enriched summary for one model. Models with
// no real+probe samples in the visibility window (availStart/seriesStart,
// whichever is older) are dropped. success_rate / sample_count follow the
// availability window; recent_success_series follows the series window; 24h/7d
// availability are fixed windows.
func buildModelSummaryV2(name string, bySource map[string]map[int64]counters, availStart int64, seriesStart int64, endTs int64) (ModelSummary, bool) {
	visibleStart := availStart
	if seriesStart < visibleStart {
		visibleStart = seriesStart
	}
	if sumRealProbeFromBuckets(bySource, visibleStart, endTs).requestCount == 0 {
		return ModelSummary{}, false
	}

	availTotals, _ := foldSources(bySource, availStart)
	_, seriesBySource := foldSources(bySource, seriesStart)
	availability := sumSources(availTotals, SourceReal, SourceProbe)

	probeShare := 0.0
	if availability.requestCount > 0 {
		probeShare = float64(availTotals[SourceProbe].requestCount) / float64(availability.requestCount)
	}

	cfg := perf_metrics_setting.GetSetting()
	summary := ModelSummary{
		ModelName:       name,
		AvgLatencyMs:    avg(availability.totalLatencyMs, availability.requestCount),
		SuccessRate:     round2(successRate(availability)),
		AvgTps:          round2(avgTps(availability)),
		AvgTtftMs:       int64Ptr(avg(availability.ttftSumMs, availability.ttftCount)),
		Availability24h: floatPtr(round2(successRate(sumRealProbeFromBuckets(bySource, endTs-24*3600, endTs)))),
		Availability7d:  floatPtr(round2(successRate(sumRealProbeFromBuckets(bySource, endTs-7*24*3600, endTs)))),
		SampleCount:     availability.requestCount,
		ProbeShare:      probeShare,
		Insufficient:    availability.requestCount < int64(cfg.MinPublicSamples),
	}
	summary.RecentSuccessSeries = recentSuccessSeries(mergeSeries(seriesBySource, SourceReal, SourceProbe))
	summary.RequestCount = availability.requestCount
	return summary, true
}

// QuerySummaryAllV2 is the enriched public summary (public_v2=true): real+probe
// caliber, TTFT, fixed 24h/7d availability, sample counts and a status series
// covering the last bars hours. includeSources defaults to real+probe.
func QuerySummaryAllV2(hours int, bars int, groups []string, includeSources []string) (SummaryAllResult, error) {
	if hours <= 0 {
		hours = 24
	}
	if hours > 24*30 {
		hours = 24 * 30
	}
	if bars <= 0 {
		bars = 24
	}
	if bars > 72 {
		bars = 72
	}
	endTs := time.Now().Unix()
	seriesStart := endTs - int64(bars)*3600
	availStart := endTs - int64(hours)*3600
	weekStart := endTs - 7*24*3600
	windowStart := minInt64(seriesStart, availStart, weekStart)

	if len(includeSources) == 0 {
		includeSources = []string{SourceReal, SourceProbe}
	}
	merged := collectSourceBuckets(windowStart, endTs, groups, sourceSet(includeSources...))

	models := make([]ModelSummary, 0, len(merged))
	for name, bySource := range merged {
		if summary, ok := buildModelSummaryV2(name, bySource, availStart, seriesStart, endTs); ok {
			models = append(models, summary)
		}
	}
	sort.Slice(models, func(i, j int) bool {
		return models[i].RequestCount > models[j].RequestCount
	})
	return SummaryAllResult{Models: models}, nil
}

// QueryAdminSummary builds the admin-only view over the last hours (max 168):
// all four sources, public v1/v2 projections, fixed-window availability,
// error-code counts from error logs (type=5, parsed in Go, no SQL JSON
// functions), and an hourly real+probe series.
func QueryAdminSummary(hours int, groups []string) (AdminPerfSummary, error) {
	if hours <= 0 {
		hours = 24
	}
	if hours > 168 {
		hours = 168
	}
	endTs := time.Now().Unix()
	startTs := endTs - int64(hours)*3600
	weekStart := endTs - 7*24*3600
	dayStart := endTs - 24*3600

	merged := collectSourceBuckets(weekStart, endTs, groups, sourceSet(SourceReal, SourceProbe, SourceInternal, SourceClientError))

	models := make([]AdminModelPerf, 0, len(merged))
	for name, bySource := range merged {
		totals, seriesBySource := foldSources(bySource, startTs)
		realTotal := totals[SourceReal]
		publicV1 := ModelSummary{
			ModelName:           name,
			AvgLatencyMs:        avg(realTotal.totalLatencyMs, realTotal.requestCount),
			SuccessRate:         round2(successRate(realTotal)),
			AvgTps:              round2(avgTps(realTotal)),
			RecentSuccessSeries: recentSuccessSeries(seriesBySource[SourceReal]),
			RequestCount:        realTotal.requestCount,
		}
		v2Summary, ok := buildModelSummaryV2(name, bySource, startTs, startTs, endTs)
		if !ok {
			v2Summary = ModelSummary{ModelName: name}
		}

		models = append(models, AdminModelPerf{
			ModelName:       name,
			Real:            counterView(realTotal),
			Probe:           counterView(totals[SourceProbe]),
			Internal:        counterView(totals[SourceInternal]),
			ClientError:     counterView(totals[SourceClientError]),
			PublicV1:        publicV1,
			PublicV2:        v2Summary,
			Availability24h: round2(successRate(sumRealProbeFromBuckets(bySource, dayStart, endTs))),
			Availability7d:  round2(successRate(sumRealProbeFromBuckets(bySource, weekStart, endTs))),
			ErrorCodes:      map[string]int64{},
			Series:          bucketSeries(mergeSeries(seriesBySource, SourceReal, SourceProbe)),
		})
	}
	sort.Slice(models, func(i, j int) bool {
		return models[i].ModelName < models[j].ModelName
	})

	errorCodes, codesTruncated := collectErrorCodes(startTs)
	for i := range models {
		if codes, ok := errorCodes[models[i].ModelName]; ok {
			models[i].ErrorCodes = codes
		}
	}

	return AdminPerfSummary{
		Models:              models,
		GeneratedAt:         endTs,
		ErrorCodesTruncated: codesTruncated,
	}, nil
}

// errorCodeScanLimit caps how many error-log rows one admin query scans; past
// that the counts are marked truncated instead of paging through an unbounded
// error burst.
const errorCodeScanLimit = 20000

// collectErrorCodes counts error-log error codes per model within the window.
// logs.other is parsed in Go (common.UnmarshalJsonStr) so the query works on
// SQLite/MySQL/PostgreSQL without dialect JSON functions. The scan is row-
// capped; the bool reports whether the cap was hit.
func collectErrorCodes(startTs int64) (map[string]map[string]int64, bool) {
	codes := map[string]map[string]int64{}
	var rows []struct {
		ModelName string
		Other     string
	}
	if err := model.LOG_DB.Model(&model.Log{}).
		Select("model_name, other").
		Where("type = ? AND created_at >= ? AND model_name <> ''", model.LogTypeError, startTs).
		Limit(errorCodeScanLimit + 1).
		Find(&rows).Error; err != nil {
		common.SysError("perf admin: failed to load error logs: " + err.Error())
		return codes, false
	}
	truncated := len(rows) > errorCodeScanLimit
	if truncated {
		rows = rows[:errorCodeScanLimit]
	}
	for _, row := range rows {
		var other struct {
			ErrorCode  string `json:"error_code"`
			StatusCode int    `json:"status_code"`
		}
		if err := common.UnmarshalJsonStr(row.Other, &other); err != nil {
			continue
		}
		label := other.ErrorCode
		if label == "" {
			label = fmt.Sprintf("status:%d", other.StatusCode)
		}
		if codes[row.ModelName] == nil {
			codes[row.ModelName] = map[string]int64{}
		}
		codes[row.ModelName][label]++
	}
	return codes, truncated
}

func minInt64(values ...int64) int64 {
	result := values[0]
	for _, value := range values[1:] {
		if value < result {
			result = value
		}
	}
	return result
}
