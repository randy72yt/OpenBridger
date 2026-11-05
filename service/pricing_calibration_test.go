package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// The factor is measured against our own offer unit prices
// ---------------------------------------------------------------------------

func TestPricingCalibrationTheoryUsesOfferUnitPrices(t *testing.T) {
	offers := map[string]model.UpstreamModelOffer{
		"gpt-x": {ChannelID: 1, PublicModel: "public-x", InputCost: 2, OutputCost: 12, CacheReadCost: 0.2},
	}
	records := []upstreamUsageLogRecord{{
		ModelName: "gpt-x", Quota: 57,
		PromptTokens: 7626, CompletionTokens: 16, CacheReadTokens: 5662, Group: "codex-plus",
	}}
	samples, skipped := buildCalibrationSamples(records, offers, 500000)
	require.Equal(t, 0, skipped)
	require.Len(t, samples, 1)
	expectedTheory := (float64(7626-5662)*2 + 5662*0.2 + 16*12) / 1e6
	assert.InDelta(t, expectedTheory, samples[0].TheoryUSD, 1e-12)
	assert.InDelta(t, 57.0/500000, samples[0].BilledUSD, 1e-12)
	assert.Equal(t, "public-x", samples[0].PublicModel)
	assert.Equal(t, "codex-plus", samples[0].Group)
}

func TestPricingCalibrationCacheFallsBackToInputPrice(t *testing.T) {
	// A provider that publishes no cache price still bills cache reads. Using the
	// input price keeps the theoretical side from understating cost, which would
	// otherwise push the factor up.
	offers := map[string]model.UpstreamModelOffer{
		"gpt-x": {PublicModel: "public-x", InputCost: 2, OutputCost: 12, CacheReadCost: 0},
	}
	records := []upstreamUsageLogRecord{{ModelName: "gpt-x", Quota: 100, PromptTokens: 1000, CacheReadTokens: 1000}}
	samples, _ := buildCalibrationSamples(records, offers, 500000)
	require.Len(t, samples, 1)
	assert.InDelta(t, 1000*2.0/1e6, samples[0].TheoryUSD, 1e-12)
}

func TestPricingCalibrationSkipsUnusableRows(t *testing.T) {
	offers := map[string]model.UpstreamModelOffer{
		"gpt-x": {PublicModel: "public-x", InputCost: 2, OutputCost: 12},
	}
	records := []upstreamUsageLogRecord{
		{ModelName: "gpt-x", Quota: 0, PromptTokens: 100},         // failed request, no charge
		{ModelName: "unknown-model", Quota: 5, PromptTokens: 100}, // no offer to price against
		{ModelName: "gpt-x", Quota: 5},                            // no tokens at all
	}
	samples, skipped := buildCalibrationSamples(records, offers, 500000)
	assert.Empty(t, samples)
	assert.Equal(t, 3, skipped)
}

// ---------------------------------------------------------------------------
// Aggregation guards
// ---------------------------------------------------------------------------

func identicalSamples(model string, ratio float64, count int) []calibrationSample {
	samples := make([]calibrationSample, 0, count)
	for i := 0; i < count; i++ {
		samples = append(samples, calibrationSample{
			PublicModel: model, UpstreamModel: model, Group: "g",
			BilledUSD: ratio, TheoryUSD: 1,
		})
	}
	return samples
}

func TestPricingCalibrationRejectsThinSamples(t *testing.T) {
	aggregates := aggregateCalibration(identicalSamples("thin", 0.02, 3), 5, 15)
	require.Len(t, aggregates, 1)
	assert.False(t, aggregates[0].Applied)
	assert.Contains(t, aggregates[0].Reason, "样本不足")
}

func TestPricingCalibrationFlagsMixedBillingModes(t *testing.T) {
	// Traffic that alternates between the standard and the long-context tier
	// produces two factors, and one number cannot describe both.
	samples := append(identicalSamples("mixed", 0.02, 10), identicalSamples("mixed", 0.06, 10)...)
	aggregates := aggregateCalibration(samples, 5, 15)
	require.Len(t, aggregates, 1)
	assert.False(t, aggregates[0].Applied)
	assert.Contains(t, aggregates[0].Reason, "离散度过高")
	assert.Greater(t, aggregates[0].DispersionPct, 15.0)
}

func TestPricingCalibrationMedianIgnoresSingleOutlier(t *testing.T) {
	samples := identicalSamples("stable", 0.0365, 20)
	samples = append(samples, calibrationSample{PublicModel: "stable", UpstreamModel: "stable", Group: "g", BilledUSD: 3.65, TheoryUSD: 1})
	aggregates := aggregateCalibration(samples, 5, 15)
	require.Len(t, aggregates, 1)
	assert.True(t, aggregates[0].Applied)
	assert.InDelta(t, 0.0365, aggregates[0].MedianRatio, 1e-9, "one bad request must not move the factor")
	assert.Greater(t, aggregates[0].Confidence, 0.0)
}

func TestPricingCalibrationGroupsModelsIndependently(t *testing.T) {
	samples := append(identicalSamples("a", 0.02, 6), identicalSamples("b", 0.9, 6)...)
	aggregates := aggregateCalibration(samples, 5, 15)
	require.Len(t, aggregates, 2)
	assert.InDelta(t, 0.02, aggregates[0].MedianRatio, 1e-9)
	assert.InDelta(t, 0.9, aggregates[1].MedianRatio, 1e-9)
}

// ---------------------------------------------------------------------------
// Staleness: an old observation must never pin prices forever
// ---------------------------------------------------------------------------

func TestPricingCalibrationRatioRejectsStaleRows(t *testing.T) {
	now := common.GetTimestamp()
	row := model.UpstreamCostCalibration{
		ChannelID: 1, PublicModel: "m", EffectiveRatio: 0.5, Applied: true,
		WindowEnd: now - pricingCalibrationStaleSeconds() - 60,
	}
	rows := map[string]model.UpstreamCostCalibration{model.UpstreamCostCalibrationKey(1, "m"): row}
	_, ok := PricingCalibrationRatio(rows, 1, "m", now)
	assert.False(t, ok, "a stale calibration must fall back to the group-ratio policy")

	row.WindowEnd = now
	rows[model.UpstreamCostCalibrationKey(1, "m")] = row
	ratio, ok := PricingCalibrationRatio(rows, 1, "m", now)
	assert.True(t, ok)
	assert.InDelta(t, 0.5, ratio, 1e-9)
}

func TestPricingCalibrationRatioIgnoresUnappliedRows(t *testing.T) {
	now := common.GetTimestamp()
	row := model.UpstreamCostCalibration{
		ChannelID: 1, PublicModel: "m", EffectiveRatio: 0.5, Applied: false, WindowEnd: now,
	}
	rows := map[string]model.UpstreamCostCalibration{model.UpstreamCostCalibrationKey(1, "m"): row}
	_, ok := PricingCalibrationRatio(rows, 1, "m", now)
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// The pricing path must prefer calibration over the published group ratio
// ---------------------------------------------------------------------------

func TestPricingProposalPrefersCalibrationOverGroupRatio(t *testing.T) {
	truncate(t)
	now := common.GetTimestamp()
	// The published ratio says 1.8, real traffic said 0.0368 - a 49x difference.
	groupRatio := 1.8
	require.NoError(t, model.DB.Create(&model.UpstreamModelOffer{
		ChannelID: 7, UpstreamModel: "claude-x", PublicModel: "claude-x",
		InputCost: 1, OutputCost: 5, CacheReadCost: 0.1, UpstreamGroupRatio: &groupRatio,
		Currency: "USD", SourceType: "test", CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
	}).Error)
	require.NoError(t, model.UpsertUpstreamCostCalibration(&model.UpstreamCostCalibration{
		ChannelID: 7, PublicModel: "claude-x", EffectiveRatio: 0.0368, PolicyRatio: 1.8,
		ObservedGroup: "claude-kiro", Source: model.PricingCalibrationSourceUpstreamLog,
		Samples: 30, Applied: true, WindowEnd: now,
	}))
	require.NoError(t, model.UpsertModelPricePolicy(&model.ModelPricePolicy{
		PublicModel: "claude-x", ServiceTier: "default", PrimaryChannelID: 7,
		TargetMarginBPS: 0, MinimumMarginBPS: 0, OverheadBPS: 0, Enabled: true,
	}))

	calibrations, err := model.UpstreamCostCalibrationMap()
	require.NoError(t, err)
	created, _, err := recalculatePricingProposals(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, created)

	proposal, err := buildPriceProposalWithCalibration(loadTestPolicy(t, "claude-x", "default"), calibrations)
	require.NoError(t, err)
	assert.InDelta(t, 1*0.0368, proposal.ExpectedInputCost, 1e-9,
		"cost must follow the measured factor, not the published group ratio")
	assert.Contains(t, proposal.Reason, "cost_ratio_source=calibration")
	assert.Contains(t, proposal.Reason, "cost_ratio=0.03680000")
}

func TestPricingProposalFallsBackToGroupRatioWithoutCalibration(t *testing.T) {
	truncate(t)
	now := common.GetTimestamp()
	groupRatio := 0.0441
	require.NoError(t, model.DB.Create(&model.UpstreamModelOffer{
		ChannelID: 8, UpstreamModel: "gemini-x", PublicModel: "gemini-x",
		InputCost: 2, OutputCost: 12, CacheReadCost: 0.2, UpstreamGroupRatio: &groupRatio,
		Currency: "USD", SourceType: "test", CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
	}).Error)
	require.NoError(t, model.UpsertModelPricePolicy(&model.ModelPricePolicy{
		PublicModel: "gemini-x", ServiceTier: "default", PrimaryChannelID: 8,
		TargetMarginBPS: 0, MinimumMarginBPS: 0, OverheadBPS: 0, Enabled: true,
	}))
	proposal, err := buildPriceProposalWithCalibration(loadTestPolicy(t, "gemini-x", "default"), map[string]model.UpstreamCostCalibration{})
	require.NoError(t, err)
	assert.InDelta(t, 2*0.0441, proposal.ExpectedInputCost, 1e-9)
	assert.Contains(t, proposal.Reason, "cost_ratio_source=policy")
}

// A calibrated model can be priced even when the provider publishes no ratio,
// which is the normal case for private settlement groups.
func TestPricingProposalPricesWithoutPublishedRatioWhenCalibrated(t *testing.T) {
	truncate(t)
	now := common.GetTimestamp()
	require.NoError(t, model.DB.Create(&model.UpstreamModelOffer{
		ChannelID: 9, UpstreamModel: "private-x", PublicModel: "private-x",
		InputCost: 4, OutputCost: 16, CacheReadCost: 0.4,
		Currency: "USD", SourceType: "test", CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
	}).Error)
	require.NoError(t, model.UpsertModelPricePolicy(&model.ModelPricePolicy{
		PublicModel: "private-x", ServiceTier: "default", PrimaryChannelID: 9,
		TargetMarginBPS: 0, MinimumMarginBPS: 0, OverheadBPS: 0, Enabled: true,
	}))
	calibrations := map[string]model.UpstreamCostCalibration{
		model.UpstreamCostCalibrationKey(9, "private-x"): {
			ChannelID: 9, PublicModel: "private-x", EffectiveRatio: 0.02, Applied: true,
			Source: model.PricingCalibrationSourceUpstreamLog, WindowEnd: now,
		},
	}
	proposal, err := buildPriceProposalWithCalibration(loadTestPolicy(t, "private-x", "default"), calibrations)
	require.NoError(t, err)
	assert.InDelta(t, 4*0.02, proposal.ExpectedInputCost, 1e-9)
}

// A pinned manual factor must survive the automatic loop.
func TestPricingCalibrationKeepsManualPinsApplied(t *testing.T) {
	require.NoError(t, model.UpsertUpstreamCostCalibration(&model.UpstreamCostCalibration{
		ChannelID: 3, PublicModel: "pinned", EffectiveRatio: 0.5,
		Source: model.PricingCalibrationSourceManual, WindowEnd: 1,
	}))
	_, ok := PricingCalibrationRatio(
		mustCalibrationMap(t), 3, "pinned", PricingCalibrationTimestampForTest())
	assert.False(t, ok, "an expired pin still stops being used, but it is not overwritten")
	row, err := model.FindUpstreamCostCalibration(3, "pinned")
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, model.PricingCalibrationSourceManual, row.Source)
}

func mustCalibrationMap(t *testing.T) map[string]model.UpstreamCostCalibration {
	t.Helper()
	rows, err := model.UpstreamCostCalibrationMap()
	require.NoError(t, err)
	return rows
}

func PricingCalibrationTimestampForTest() int64 { return common.GetTimestamp() }

func loadTestPolicy(t *testing.T, publicModel, tier string) *model.ModelPricePolicy {
	t.Helper()
	var policy model.ModelPricePolicy
	require.NoError(t, model.DB.Where("public_model = ? AND service_tier = ?", publicModel, tier).First(&policy).Error)
	return &policy
}

// ---------------------------------------------------------------------------
// Log pagination: the provider returns the oldest entries first
// ---------------------------------------------------------------------------

// The first production run read page 1, saw an entry older than the window and
// stopped, so a 72h window yielded 19 rows instead of ~2000. The walk has to
// find the tail of an oldest-first log.
func usageLogServer(t *testing.T, count, perPage int, descending bool, base int64) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("p"))
		if page < 1 {
			page = 1
		}
		start := (page - 1) * perPage
		if start >= count {
			fmt.Fprint(w, `{"success":true,"data":[]}`)
			return
		}
		end := start + perPage
		if end > count {
			end = count
		}
		indexes := make([]int, 0, end-start)
		for i := start; i < end; i++ {
			if descending {
				indexes = append(indexes, count-1-i)
			} else {
				indexes = append(indexes, i)
			}
		}
		fmt.Fprint(w, `{"success":true,"data":[`)
		for i, idx := range indexes {
			if i > 0 {
				fmt.Fprint(w, ",")
			}
			fmt.Fprintf(w, `{"id":%d,"created_at":%d,"model_name":"m","quota":1,"prompt_tokens":10,"completion_tokens":1,"group":"g"}`,
				idx+1, base+int64(idx)*60)
		}
		fmt.Fprint(w, `]}`)
	}))
}

func calibrationTestChannel(serverURL string) *model.Channel {
	base := serverURL
	return &model.Channel{Id: 1, Key: "test-key", BaseURL: &base}
}

func TestFetchUpstreamUsageLogsFindsWindowAtTailOfOldestFirstLog(t *testing.T) {
	const count, perPage = 60, 5
	base := int64(1_800_000_000)
	server := usageLogServer(t, count, perPage, false, base)
	defer server.Close()
	// Window covers the newest 15 entries.
	since := base + 45*60
	records, err := fetchUpstreamUsageLogs(context.Background(), calibrationTestChannel(server.URL), since)
	require.NoError(t, err)
	assert.Len(t, records, 15, "an oldest-first log must be read from its tail")
	assert.GreaterOrEqual(t, records[0].CreatedAt, since)
}

func TestFetchUpstreamUsageLogsHandlesNewestFirstLog(t *testing.T) {
	const count, perPage = 60, 5
	base := int64(1_800_000_000)
	server := usageLogServer(t, count, perPage, true, base)
	defer server.Close()
	since := base + 45*60
	records, err := fetchUpstreamUsageLogs(context.Background(), calibrationTestChannel(server.URL), since)
	require.NoError(t, err)
	assert.Len(t, records, 15, "a newest-first log must stop once a page predates the window")
}

func TestFetchUpstreamUsageLogsReturnsEmptyWhenWindowPredatesLog(t *testing.T) {
	const count, perPage = 10, 5
	base := int64(1_800_000_000)
	server := usageLogServer(t, count, perPage, false, base)
	defer server.Close()
	records, err := fetchUpstreamUsageLogs(context.Background(), calibrationTestChannel(server.URL), base+100000)
	require.NoError(t, err)
	assert.Empty(t, records)
}

// ---------------------------------------------------------------------------
// Catalogue coverage: one windowed query per model
// ---------------------------------------------------------------------------

type usageLogSeed struct {
	model     string
	createdAt int64
}

type usageLogServerOptions struct {
	honorModel  bool
	honorWindow bool
	descending  bool
	perPage     int
	requests    *int
	// before runs for every request and can answer it itself (returning true).
	// Used to simulate the provider's 429 without duplicating the paging logic.
	before func(w http.ResponseWriter, model string) bool
}

// windowedUsageLogServer honours model_name and, when honorWindow is set, the
// start_timestamp/end_timestamp filters. Rows are returned oldest-first unless
// descending is set, matching how the provider orders its log.
func windowedUsageLogServer(t *testing.T, rows []usageLogSeed, opts usageLogServerOptions) *httptest.Server {
	t.Helper()
	perPage := opts.perPage
	if perPage <= 0 {
		perPage = 100
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if opts.requests != nil {
			*opts.requests++
		}
		query := r.URL.Query()
		model := query.Get("model_name")
		if opts.before != nil && opts.before(w, model) {
			return
		}
		start, _ := strconv.ParseInt(query.Get("start_timestamp"), 10, 64)
		end, _ := strconv.ParseInt(query.Get("end_timestamp"), 10, 64)
		matched := make([]usageLogSeed, 0, len(rows))
		for _, row := range rows {
			if opts.honorModel && model != "" && row.model != model {
				continue
			}
			if opts.honorWindow {
				if start > 0 && row.createdAt < start {
					continue
				}
				if end > 0 && row.createdAt > end {
					continue
				}
			}
			matched = append(matched, row)
		}
		page, _ := strconv.Atoi(query.Get("p"))
		if page < 1 {
			page = 1
		}
		begin := (page - 1) * perPage
		if begin >= len(matched) {
			fmt.Fprint(w, `{"success":true,"data":[]}`)
			return
		}
		stop := begin + perPage
		if stop > len(matched) {
			stop = len(matched)
		}
		fmt.Fprint(w, `{"success":true,"data":[`)
		for i := begin; i < stop; i++ {
			index := i
			if opts.descending {
				index = len(matched) - 1 - i
			}
			if i > begin {
				fmt.Fprint(w, ",")
			}
			fmt.Fprintf(w, `{"id":%d,"created_at":%d,"model_name":%q,"quota":1,"prompt_tokens":10,"completion_tokens":1,"group":"g"}`,
				index+1, matched[index].createdAt, matched[index].model)
		}
		fmt.Fprint(w, `]}`)
	}))
}

// seedCalibrationLog builds an old block of history plus a few recent rows per
// model, the shape that made the global tail walk price only the hot models.
func seedCalibrationLog(base int64, recentPerModel map[string]int) ([]usageLogSeed, int64) {
	rows := make([]usageLogSeed, 0)
	// History: 200 rows two days back, all outside the window.
	for i := 0; i < 200; i++ {
		rows = append(rows, usageLogSeed{model: "m-legacy", createdAt: base + int64(i)*60})
	}
	since := base + 2*86400
	models := make([]string, 0, len(recentPerModel))
	for model := range recentPerModel {
		models = append(models, model)
	}
	sort.Strings(models)
	for _, model := range models {
		for i := 0; i < recentPerModel[model]; i++ {
			rows = append(rows, usageLogSeed{model: model, createdAt: since + int64(i)*60})
		}
	}
	return rows, since
}

func TestFetchChannelUsageLogsAsksPerModelSoCoverageDoesNotDependOnTraffic(t *testing.T) {
	base := int64(1_800_000_000)
	rows, since := seedCalibrationLog(base, map[string]int{"m-a": 8, "m-b": 6, "m-c": 0})
	server := windowedUsageLogServer(t, rows, usageLogServerOptions{honorModel: true, honorWindow: true})
	defer server.Close()

	records, windowed, _, err := fetchChannelUsageLogs(context.Background(), calibrationTestChannel(server.URL), since, since+86400, []string{"m-a", "m-b", "m-c"})
	require.NoError(t, err)
	assert.True(t, windowed, "the provider honours the window filter, so no tail walk is needed")
	assert.Len(t, records, 14)
	seen := map[string]int{}
	for _, record := range records {
		assert.GreaterOrEqual(t, record.CreatedAt, since)
		seen[record.ModelName]++
	}
	assert.Equal(t, 8, seen["m-a"])
	assert.Equal(t, 6, seen["m-b"])
	assert.Zero(t, seen["m-legacy"], "history outside the window must not be counted")
}

// A provider that ignores both filters answers page 1 with the oldest rows of
// the whole log. Detecting that is what keeps calibration from silently pricing
// yesterday's traffic as today's.
func TestFetchChannelUsageLogsFallsBackWhenProviderIgnoresWindowFilter(t *testing.T) {
	base := int64(1_800_000_000)
	rows, since := seedCalibrationLog(base, map[string]int{"m-a": 12})
	server := windowedUsageLogServer(t, rows, usageLogServerOptions{perPage: 50})
	defer server.Close()

	records, windowed, _, err := fetchChannelUsageLogs(context.Background(), calibrationTestChannel(server.URL), since, since+86400, []string{"m-a"})
	require.NoError(t, err)
	assert.False(t, windowed, "an ignored filter must be reported so the operator can see it")
	require.NotEmpty(t, records, "the fallback tail walk still reads the window")
	for _, record := range records {
		assert.GreaterOrEqual(t, record.CreatedAt, since)
	}
}

// A provider that keeps model_name but drops the timestamps returns the whole
// history of that model. Page 1 is then inside the window, so the fallback is
// not triggered - correct results have to come from the client-side clip.
func TestFetchChannelUsageLogsClipsRowsWhenProviderIgnoresTimeFilterOnly(t *testing.T) {
	base := int64(1_800_000_000)
	rows, since := seedCalibrationLog(base, map[string]int{"m-a": 9})
	server := windowedUsageLogServer(t, rows, usageLogServerOptions{honorModel: true, perPage: 50})
	defer server.Close()

	records, _, _, err := fetchChannelUsageLogs(context.Background(), calibrationTestChannel(server.URL), since, since+86400, []string{"m-a"})
	require.NoError(t, err)
	require.NotEmpty(t, records)
	for _, record := range records {
		assert.GreaterOrEqual(t, record.CreatedAt, since)
		assert.Equal(t, "m-a", record.ModelName)
	}
	assert.LessOrEqual(t, len(records), pricingCalibrationModelRecords)
}

// The mirror case: a newest-first log that ignores every filter still answers
// page 1 from inside the window, so the window is not detected as unsupported.
// The client-side window check has to keep the result correct anyway.
func TestFetchChannelUsageLogsClipsRowsWhenNewestFirstProviderIgnoresFilter(t *testing.T) {
	base := int64(1_800_000_000)
	rows, since := seedCalibrationLog(base, map[string]int{"m-a": 9})
	server := windowedUsageLogServer(t, rows, usageLogServerOptions{descending: true, perPage: 50})
	defer server.Close()

	records, _, _, err := fetchChannelUsageLogs(context.Background(), calibrationTestChannel(server.URL), since, since+86400, []string{"m-a"})
	require.NoError(t, err)
	require.NotEmpty(t, records)
	for _, record := range records {
		assert.GreaterOrEqual(t, record.CreatedAt, since)
	}
	assert.LessOrEqual(t, len(records), pricingCalibrationModelRecords)
}

func TestWindowWalkStaysInsideRequestBudget(t *testing.T) {
	fastRetry(t)
	base := int64(1_800_000_000)
	perModel := make(map[string]int, 60)
	models := make([]string, 0, 60)
	for i := 0; i < 60; i++ {
		model := fmt.Sprintf("model-%02d", i)
		perModel[model] = 40
		models = append(models, model)
	}
	sort.Strings(models)
	rows, since := seedCalibrationLog(base, perModel)
	requests := 0
	server := windowedUsageLogServer(t, rows, usageLogServerOptions{honorModel: true, honorWindow: true, perPage: 10, requests: &requests})
	defer server.Close()

	_, windowed, _, err := fetchChannelUsageLogs(context.Background(), calibrationTestChannel(server.URL), since, since+86400, models)
	require.NoError(t, err)
	assert.True(t, windowed)
	assert.LessOrEqual(t, requests, pricingCalibrationWindowRequestBudget,
		"a wide catalogue must not turn one calibration run into a crawl")
}

// fastRetry shrinks the pacing and 429 backoff so these tests stay fast.
func fastRetry(t *testing.T) {
	t.Helper()
	previousGap := pricingCalibrationModelGap
	previousBase := pricingCalibrationRetryBaseWait
	previousMax := pricingCalibrationMaxRetryWait
	pricingCalibrationModelGap = 0
	pricingCalibrationRetryBaseWait = 2 * time.Millisecond
	pricingCalibrationMaxRetryWait = 10 * time.Millisecond
	t.Cleanup(func() {
		pricingCalibrationModelGap = previousGap
		pricingCalibrationRetryBaseWait = previousBase
		pricingCalibrationMaxRetryWait = previousMax
	})
}

// rateLimitHook answers 429 for the models in blocked while their budget lasts
// (blocked[model] = remaining 429s; -1 means forever).
func rateLimitHook(blocked map[string]int, retryAfterSeconds int) func(http.ResponseWriter, string) bool {
	var mu sync.Mutex
	return func(w http.ResponseWriter, model string) bool {
		mu.Lock()
		remaining, limited := blocked[model]
		if !limited || remaining == 0 {
			mu.Unlock()
			return false
		}
		if remaining > 0 {
			blocked[model] = remaining - 1
		}
		mu.Unlock()
		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"success":false,"message":"too many requests"}`)
		return true
	}
}

// A per-model walk sends one request per model, which is exactly the burst the
// log endpoint throttles. A single 429 must be waited out, not surfaced.
func TestFetchChannelUsageLogsRetriesRateLimit(t *testing.T) {
	fastRetry(t)
	base := int64(1_800_000_000)
	rows, since := seedCalibrationLog(base, map[string]int{"m-a": 6})
	server := windowedUsageLogServer(t, rows, usageLogServerOptions{
		honorModel:  true,
		honorWindow: true,
		before:      rateLimitHook(map[string]int{"m-a": 1}, 0),
	})
	defer server.Close()

	records, windowed, failed, err := fetchChannelUsageLogs(context.Background(), calibrationTestChannel(server.URL), since, since+86400, []string{"m-a"})
	require.NoError(t, err, "a retryable 429 must not fail the pass")
	assert.True(t, windowed)
	assert.Equal(t, 0, failed)
	assert.NotEmpty(t, records, "the retried page must be read")
}

// The first production run of the per-model walk died on one 429 and threw away
// every factor the other models had already produced. A blocked model may be
// skipped; the rest of the catalogue still gets calibrated.
func TestWindowWalkKeepsTheOtherModelsWhenOneIsRateLimited(t *testing.T) {
	fastRetry(t)
	base := int64(1_800_000_000)
	rows, since := seedCalibrationLog(base, map[string]int{"m-a": 6, "m-b": 6})
	server := windowedUsageLogServer(t, rows, usageLogServerOptions{
		honorModel:  true,
		honorWindow: true,
		before:      rateLimitHook(map[string]int{"m-a": -1}, 0),
	})
	defer server.Close()

	records, windowed, failed, err := fetchChannelUsageLogs(context.Background(), calibrationTestChannel(server.URL), since, since+86400, []string{"m-a", "m-b"})
	require.NoError(t, err, "one blocked model must not fail the whole pass")
	assert.True(t, windowed)
	assert.Equal(t, 1, failed)
	require.NotEmpty(t, records)
	for _, record := range records {
		assert.Equal(t, "m-b", record.ModelName, "only the blocked model may be missing")
	}
}

// The mirror case: when the endpoint itself is unusable every model fails, and
// reporting success with zero samples would look like "no traffic" instead of
// "we were throttled".
func TestWindowWalkReportsFailureWhenEveryModelIsRateLimited(t *testing.T) {
	fastRetry(t)
	base := int64(1_800_000_000)
	rows, since := seedCalibrationLog(base, map[string]int{"m-a": 4, "m-b": 4})
	server := windowedUsageLogServer(t, rows, usageLogServerOptions{
		honorModel:  true,
		honorWindow: true,
		before:      rateLimitHook(map[string]int{"m-a": -1, "m-b": -1}, 0),
	})
	defer server.Close()

	records, _, failed, err := fetchChannelUsageLogs(context.Background(), calibrationTestChannel(server.URL), since, since+86400, []string{"m-a", "m-b"})
	require.Error(t, err)
	assert.Empty(t, records)
	assert.GreaterOrEqual(t, failed, 2)
}

// The provider's real answer is "retry after 5m17s". Waiting that out inside a
// task would block the runner for the whole window and burn the next window's
// quota, so a long Retry-After must end the pass quickly with what it has.
func TestWindowWalkStopsQuicklyOnMinuteScaleRetryAfter(t *testing.T) {
	fastRetry(t)
	base := int64(1_800_000_000)
	rows, since := seedCalibrationLog(base, map[string]int{"m-a": 6, "m-b": 6})
	server := windowedUsageLogServer(t, rows, usageLogServerOptions{
		honorModel:  true,
		honorWindow: true,
		before:      rateLimitHook(map[string]int{"m-b": -1}, 317),
	})
	defer server.Close()

	started := time.Now()
	records, windowed, failed, err := fetchChannelUsageLogs(context.Background(), calibrationTestChannel(server.URL), since, since+86400, []string{"m-a", "m-b", "m-c"})
	elapsed := time.Since(started)

	require.NoError(t, err, "a throttle mid-pass must not fail the models already read")
	assert.True(t, windowed)
	assert.Equal(t, 1, failed)
	assert.Less(t, elapsed, 5*time.Second, "a 5-minute Retry-After must not be waited out")
	require.NotEmpty(t, records)
	for _, record := range records {
		assert.Equal(t, "m-a", record.ModelName)
	}
}
