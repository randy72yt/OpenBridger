package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
)

// Upstream pricing cannot be reconstructed from the published price list. Real
// traffic showed the same model settling through a private group whose ratio is
// absent from /api/pricing, cache tokens discounted by a per-request ratio, a
// night-time factor, tiered expressions and a currency divisor. The calibration
// below sidesteps all of it: it divides what the provider actually charged by
// what our own offer table would have charged for the very same tokens. The
// quotient absorbs every factor at once, so base cost × quotient reproduces the
// invoice regardless of how the upstream bills internally.
const (
	defaultUpstreamUsageLogPath = "/api/log/token"

	pricingCalibrationPageSize      = 100
	pricingCalibrationMaxPages      = 40
	pricingCalibrationMaxRecords    = 3000
	pricingCalibrationResponseLimit = 32 << 20
	// DefaultQuotaPerUnit is the new-api convention: one US dollar is worth
	// 500000 quota units. A provider that publishes a different snapshot in the
	// per-request billing detail still lands on the same quotient because both
	// the billed and the theoretical side use the same unit.
	pricingCalibrationDefaultQuotaPerUnit = 500000.0

	pricingCalibrationDefaultWindowHours   = 72
	pricingCalibrationDefaultMinSamples    = 5
	pricingCalibrationDefaultMaxDispersion = 15.0
	pricingCalibrationDefaultStaleHours    = 96
)

// PricingCalibrationModelResult is one model's calibration outcome.
type PricingCalibrationModelResult struct {
	PublicModel    string  `json:"public_model"`
	UpstreamModel  string  `json:"upstream_model"`
	EffectiveRatio float64 `json:"effective_ratio"`
	PolicyRatio    float64 `json:"policy_ratio"`
	ObservedGroup  string  `json:"observed_group"`
	Samples        int     `json:"samples"`
	BilledUSD      float64 `json:"billed_usd"`
	TheoryUSD      float64 `json:"theory_usd"`
	DispersionPct  float64 `json:"dispersion_pct"`
	Confidence     float64 `json:"confidence"`
	Applied        bool    `json:"applied"`
	Reason         string  `json:"reason"`
}

// PricingCalibrationChannelResult is one channel's calibration outcome.
type PricingCalibrationChannelResult struct {
	ChannelID   int    `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Models      int    `json:"models"`
	Applied     int    `json:"applied"`
	Records     int    `json:"records"`
	Error       string `json:"error,omitempty"`
}

// PricingCalibrationResult summarizes one calibration run.
type PricingCalibrationResult struct {
	Trigger     string                            `json:"trigger"`
	Channels    []PricingCalibrationChannelResult `json:"channels"`
	Applied     int                               `json:"applied"`
	Skipped     int                               `json:"skipped"`
	WindowStart int64                             `json:"window_start"`
	WindowEnd   int64                             `json:"window_end"`
}

// upstreamUsageLogRecord is one consumption entry from the provider's log.
// Quota is what the provider actually charged; the token columns let us rebuild
// what our own price table would have charged for the same request.
type upstreamUsageLogRecord struct {
	ID               int64   `json:"id"`
	CreatedAt        int64   `json:"created_at"`
	ModelName        string  `json:"model_name"`
	Quota            float64 `json:"quota"`
	PromptTokens     float64 `json:"prompt_tokens"`
	CompletionTokens float64 `json:"completion_tokens"`
	CacheReadTokens  float64 `json:"cache_read_tokens"`
	Group            string  `json:"group"`
}

type upstreamUsageLogResponse struct {
	Data    []upstreamUsageLogRecord `json:"data"`
	Message string                   `json:"message"`
	Success bool                     `json:"success"`
}

// calibrationSample is one request reduced to the two numbers that matter.
type calibrationSample struct {
	PublicModel   string
	UpstreamModel string
	Group         string
	BilledUSD     float64
	TheoryUSD     float64
}

type calibrationAggregate struct {
	PublicModel   string
	UpstreamModel string
	Samples       int
	BilledUSD     float64
	TheoryUSD     float64
	MedianRatio   float64
	DispersionPct float64
	Group         string
	Confidence    float64
	Applied       bool
	Reason        string
}

func pricingCalibrationEnabled() bool {
	raw := strings.TrimSpace(os.Getenv("PRICING_CALIBRATION_ENABLED"))
	if raw == "" {
		return true
	}
	enabled, err := strconv.ParseBool(raw)
	return err == nil && enabled
}

func pricingCalibrationWindowSeconds() int64 {
	hours := common.GetEnvOrDefault("PRICING_CALIBRATION_WINDOW_HOURS", pricingCalibrationDefaultWindowHours)
	if hours < 1 {
		hours = pricingCalibrationDefaultWindowHours
	}
	if hours > 24*30 {
		hours = 24 * 30
	}
	return int64(hours) * 3600
}

func pricingCalibrationMinSamples() int {
	samples := common.GetEnvOrDefault("PRICING_CALIBRATION_MIN_SAMPLES", pricingCalibrationDefaultMinSamples)
	if samples < 1 {
		samples = 1
	}
	if samples > 500 {
		samples = 500
	}
	return samples
}

// pricingCalibrationMaxDispersionPct accepts a fractional value, so a provider
// that mixes several billing tiers can be tightened below the default.
func pricingCalibrationMaxDispersionPct() float64 {
	value := pricingCalibrationDefaultMaxDispersion
	if raw := strings.TrimSpace(os.Getenv("PRICING_CALIBRATION_MAX_DISPERSION_PCT")); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil && parsed > 0 {
			value = parsed
		}
	}
	if value > 100 {
		value = 100
	}
	return value
}

// pricingCalibrationStaleSeconds bounds how long a factor stays trustworthy.
// Past this age a calibration is ignored and the group-ratio policy takes over
// again, so a silent provider outage cannot pin prices to an old observation.
func pricingCalibrationStaleSeconds() int64 {
	hours := common.GetEnvOrDefault("PRICING_CALIBRATION_STALE_HOURS", pricingCalibrationDefaultStaleHours)
	if hours < 1 {
		hours = pricingCalibrationDefaultStaleHours
	}
	return int64(hours) * 3600
}

func pricingCalibrationQuotaPerUnit() float64 {
	if raw := strings.TrimSpace(os.Getenv("PRICING_CALIBRATION_QUOTA_PER_UNIT")); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil && parsed > 0 {
			return parsed
		}
	}
	return pricingCalibrationDefaultQuotaPerUnit
}

func upstreamUsageLogEndpoint(channel *model.Channel, baseURL *url.URL) (*url.URL, error) {
	endpoint := strings.TrimSpace(os.Getenv("PRICING_CALIBRATION_LOG_PATH"))
	if endpoint == "" {
		endpoint = defaultUpstreamUsageLogPath
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid usage log endpoint %q: %w", endpoint, err)
	}
	if parsed.IsAbs() {
		return parsed, nil
	}
	if !strings.HasPrefix(endpoint, "/") {
		return nil, fmt.Errorf("usage log endpoint %q must be absolute or start with /", endpoint)
	}
	resolved := *baseURL
	resolved.Path = endpoint
	resolved.RawQuery = ""
	resolved.Fragment = ""
	return &resolved, nil
}

// fetchUpstreamUsageLogs walks the provider's consumption log backwards until it
// reaches the window start. The log is what makes calibration possible at all:
// it is the only endpoint that reports what was actually charged.
func fetchUpstreamUsageLogs(ctx context.Context, channel *model.Channel, since int64) ([]upstreamUsageLogRecord, error) {
	baseURL, err := url.Parse(strings.TrimSpace(channel.GetBaseURL()))
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, errors.New("channel has no valid upstream base URL")
	}
	endpoint, err := upstreamUsageLogEndpoint(channel, baseURL)
	if err != nil {
		return nil, err
	}
	key, _, keyErr := channel.GetNextEnabledKey()
	if keyErr != nil || strings.TrimSpace(key) == "" {
		return nil, errors.New("channel has no enabled API key")
	}
	client := &http.Client{Timeout: 30 * time.Second}
	records := make([]upstreamUsageLogRecord, 0, pricingCalibrationPageSize*pricingCalibrationMaxPages/4)
	for page := 1; page <= pricingCalibrationMaxPages; page++ {
		if ctx != nil && ctx.Err() != nil {
			return records, ctx.Err()
		}
		query := endpoint.Query()
		query.Set("p", strconv.Itoa(page))
		query.Set("page_size", strconv.Itoa(pricingCalibrationPageSize))
		query.Set("type", "2")
		pageURL := *endpoint
		pageURL.RawQuery = query.Encode()
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, pageURL.String(), nil)
		if reqErr != nil {
			return records, reqErr
		}
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(key))
		req.Header.Set("Accept", "application/json")
		resp, doErr := client.Do(req)
		if doErr != nil {
			return records, doErr
		}
		var payload upstreamUsageLogResponse
		decodeErr := common.DecodeJson(io.LimitReader(resp.Body, pricingCalibrationResponseLimit), &payload)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return records, fmt.Errorf("upstream usage log returned %s", resp.Status)
		}
		if decodeErr != nil {
			return records, decodeErr
		}
		if !payload.Success && payload.Message != "" {
			return records, errors.New(payload.Message)
		}
		if len(payload.Data) == 0 {
			return records, nil
		}
		oldest := int64(0)
		for _, record := range payload.Data {
			if oldest == 0 || record.CreatedAt < oldest {
				oldest = record.CreatedAt
			}
			if since > 0 && record.CreatedAt > 0 && record.CreatedAt < since {
				continue
			}
			records = append(records, record)
		}
		// The page size is advisory: new-api clamps it and returned 142 rows
		// whatever we asked for, so "fewer rows than requested" cannot mean the
		// last page. Only an empty page or passing the window ends the walk.
		if len(records) >= pricingCalibrationMaxRecords {
			return records, nil
		}
		if since > 0 && oldest > 0 && oldest < since {
			return records, nil
		}
	}
	return records, nil
}

// buildCalibrationSamples turns raw log rows into (billed, theory) pairs. The
// theory side must use exactly the same unit prices the pricing path uses, which
// is why it reads the offer rather than the pricing table.
func buildCalibrationSamples(records []upstreamUsageLogRecord, offers map[string]model.UpstreamModelOffer, quotaPerUnit float64) ([]calibrationSample, int) {
	if quotaPerUnit <= 0 {
		quotaPerUnit = pricingCalibrationDefaultQuotaPerUnit
	}
	samples := make([]calibrationSample, 0, len(records))
	skipped := 0
	for _, record := range records {
		if record.Quota <= 0 || record.ModelName == "" {
			skipped++
			continue
		}
		offer, ok := offers[strings.ToLower(strings.TrimSpace(record.ModelName))]
		if !ok {
			skipped++
			continue
		}
		cacheCost := offer.CacheReadCost
		if cacheCost <= 0 {
			// A provider that publishes no cache price still bills cache reads;
			// treating them as full-price inputs keeps the theory side from
			// understating cost and the factor from drifting high.
			cacheCost = offer.InputCost
		}
		cached := math.Max(record.CacheReadTokens, 0)
		billableInput := math.Max(record.PromptTokens-cached, 0)
		theory := (billableInput*offer.InputCost + cached*cacheCost + math.Max(record.CompletionTokens, 0)*offer.OutputCost) / 1e6
		if theory <= 1e-12 || math.IsNaN(theory) || math.IsInf(theory, 0) {
			skipped++
			continue
		}
		samples = append(samples, calibrationSample{
			PublicModel:   offer.PublicModel,
			UpstreamModel: record.ModelName,
			Group:         strings.TrimSpace(record.Group),
			BilledUSD:     record.Quota / quotaPerUnit,
			TheoryUSD:     theory,
		})
	}
	return samples, skipped
}

// aggregateCalibration collapses samples per model. The median is used rather
// than the mean so a single mis-priced request cannot move the factor, and the
// median absolute deviation tells us whether traffic is mixing billing modes.
func aggregateCalibration(samples []calibrationSample, minSamples int, maxDispersionPct float64) []calibrationAggregate {
	grouped := make(map[string][]calibrationSample)
	order := make([]string, 0)
	for _, sample := range samples {
		key := strings.ToLower(sample.PublicModel)
		if _, seen := grouped[key]; !seen {
			order = append(order, key)
		}
		grouped[key] = append(grouped[key], sample)
	}
	sort.Strings(order)
	out := make([]calibrationAggregate, 0, len(order))
	for _, key := range order {
		rows := grouped[key]
		ratios := make([]float64, 0, len(rows))
		billed := 0.0
		theory := 0.0
		groups := make(map[string]int)
		upstreamModel := ""
		for _, row := range rows {
			ratios = append(ratios, row.BilledUSD/row.TheoryUSD)
			billed += row.BilledUSD
			theory += row.TheoryUSD
			if row.Group != "" {
				groups[row.Group]++
			}
			if upstreamModel == "" {
				upstreamModel = row.UpstreamModel
			}
		}
		sort.Float64s(ratios)
		median := medianFloat(ratios)
		deviations := make([]float64, 0, len(ratios))
		for _, ratio := range ratios {
			deviations = append(deviations, math.Abs(ratio-median))
		}
		dispersion := 0.0
		if median > 0 {
			dispersion = medianFloat(deviations) / median * 100
		}
		aggregate := calibrationAggregate{
			PublicModel:   rows[0].PublicModel,
			UpstreamModel: upstreamModel,
			Samples:       len(rows),
			BilledUSD:     billed,
			TheoryUSD:     theory,
			MedianRatio:   median,
			DispersionPct: dispersion,
			Group:         dominantKey(groups),
		}
		switch {
		case median <= 0 || math.IsNaN(median) || math.IsInf(median, 0):
			aggregate.Reason = "测得倍率无效"
		case len(rows) < minSamples:
			aggregate.Reason = fmt.Sprintf("样本不足（%d < %d）", len(rows), minSamples)
		case dispersion > maxDispersionPct:
			aggregate.Reason = fmt.Sprintf("离散度过高（%.1f%% > %.1f%%），流量可能混用多种计费档", dispersion, maxDispersionPct)
		default:
			aggregate.Applied = true
			aggregate.Confidence = calibrationConfidence(len(rows), dispersion)
		}
		out = append(out, aggregate)
	}
	return out
}

func medianFloat(sorted []float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}

func dominantKey(counts map[string]int) string {
	best := ""
	bestCount := 0
	for key, count := range counts {
		if count > bestCount || (count == bestCount && key < best) {
			best, bestCount = key, count
		}
	}
	return best
}

func calibrationConfidence(samples int, dispersionPct float64) float64 {
	sampleScore := math.Min(1, float64(samples)/30)
	dispersionScore := 0.2
	switch {
	case dispersionPct <= 5:
		dispersionScore = 1
	case dispersionPct <= 15:
		dispersionScore = 0.6
	}
	confidence := math.Round(sampleScore*dispersionScore*10000) / 10000
	if confidence > 1 {
		confidence = 1
	}
	return confidence
}

// calibrationOffersByUpstreamModel indexes the current offers by upstream model
// name, because the provider log speaks upstream names only.
func calibrationOffersByUpstreamModel(channelID int, now int64) (map[string]model.UpstreamModelOffer, error) {
	rows, err := model.ListCurrentUpstreamModelOffers(channelID, now)
	if err != nil {
		return nil, err
	}
	out := make(map[string]model.UpstreamModelOffer, len(rows))
	for _, row := range rows {
		key := strings.ToLower(strings.TrimSpace(row.UpstreamModel))
		if key == "" {
			continue
		}
		if existing, ok := out[key]; ok && existing.CollectedAt >= row.CollectedAt {
			continue
		}
		out[key] = row
	}
	return out, nil
}

// RunPricingCalibration derives the effective cost factor for every model that
// has both a current offer and recent upstream traffic.
func RunPricingCalibration(ctx context.Context, channelIDs []int, trigger string) (PricingCalibrationResult, error) {
	now := common.GetTimestamp()
	window := pricingCalibrationWindowSeconds()
	result := PricingCalibrationResult{
		Trigger:     trigger,
		Channels:    make([]PricingCalibrationChannelResult, 0),
		WindowStart: now - window,
		WindowEnd:   now,
	}
	if !pricingCalibrationEnabled() {
		return result, errors.New("pricing calibration is disabled")
	}
	channels, err := calibrationChannels(channelIDs)
	if err != nil {
		return result, err
	}
	if len(channels) == 0 {
		return result, errors.New("no channel has current upstream offers to calibrate")
	}
	minSamples := pricingCalibrationMinSamples()
	maxDispersion := pricingCalibrationMaxDispersionPct()
	quotaPerUnit := pricingCalibrationQuotaPerUnit()
	failures := make([]string, 0)
	for i := range channels {
		channel := channels[i]
		channelResult := PricingCalibrationChannelResult{ChannelID: channel.Id, ChannelName: channel.Name}
		offers, offerErr := calibrationOffersByUpstreamModel(channel.Id, now)
		if offerErr != nil {
			channelResult.Error = offerErr.Error()
			result.Channels = append(result.Channels, channelResult)
			failures = append(failures, fmt.Sprintf("channel %d: %s", channel.Id, offerErr.Error()))
			continue
		}
		records, logErr := fetchUpstreamUsageLogs(ctx, &channel, result.WindowStart)
		if logErr != nil {
			channelResult.Error = logErr.Error()
			result.Channels = append(result.Channels, channelResult)
			failures = append(failures, fmt.Sprintf("channel %d: %s", channel.Id, logErr.Error()))
			continue
		}
		channelResult.Records = len(records)
		samples, _ := buildCalibrationSamples(records, offers, quotaPerUnit)
		aggregates := aggregateCalibration(samples, minSamples, maxDispersion)
		channelResult.Models = len(aggregates)
		for _, aggregate := range aggregates {
			policyRatio := 0.0
			if offer, ok := offers[strings.ToLower(aggregate.UpstreamModel)]; ok && offer.UpstreamGroupRatio != nil {
				policyRatio = *offer.UpstreamGroupRatio
			}
			applied := aggregate.Applied
			reason := aggregate.Reason
			if applied {
				if existing, findErr := model.FindUpstreamCostCalibration(channel.Id, aggregate.PublicModel); findErr == nil &&
					existing != nil && existing.Source == model.PricingCalibrationSourceManual {
					// A human pin outranks the automatic loop until it is cleared.
					applied = false
					reason = "人工校准优先，自动值仅作参考"
				}
			}
			row := &model.UpstreamCostCalibration{
				ChannelID:      channel.Id,
				PublicModel:    aggregate.PublicModel,
				UpstreamModel:  aggregate.UpstreamModel,
				EffectiveRatio: roundRatioValue(aggregate.MedianRatio),
				PolicyRatio:    roundRatioValue(policyRatio),
				ObservedGroup:  aggregate.Group,
				Source:         model.PricingCalibrationSourceUpstreamLog,
				Samples:        aggregate.Samples,
				BilledUSD:      roundRatioValue(aggregate.BilledUSD),
				TheoryUSD:      roundRatioValue(aggregate.TheoryUSD),
				DispersionPct:  math.Round(aggregate.DispersionPct*100) / 100,
				Confidence:     aggregate.Confidence,
				WindowStart:    result.WindowStart,
				WindowEnd:      result.WindowEnd,
				Applied:        applied,
				Reason:         reason,
			}
			if existing, findErr := model.FindUpstreamCostCalibration(channel.Id, aggregate.PublicModel); findErr == nil &&
				existing != nil && existing.Source == model.PricingCalibrationSourceManual {
				continue
			}
			if saveErr := model.UpsertUpstreamCostCalibration(row); saveErr != nil {
				failures = append(failures, fmt.Sprintf("channel %d model %s: %s", channel.Id, aggregate.PublicModel, saveErr.Error()))
				continue
			}
			if applied {
				channelResult.Applied++
				result.Applied++
			} else {
				result.Skipped++
			}
		}
		result.Channels = append(result.Channels, channelResult)
	}
	if len(failures) > 0 {
		return result, errors.New(strings.Join(failures, "; "))
	}
	return result, nil
}

// calibrationChannels restricts calibration to channels that own offers. A
// channel without offers has no base cost to calibrate against.
func calibrationChannels(channelIDs []int) ([]model.Channel, error) {
	ids := channelIDs
	if len(ids) == 0 {
		derived, err := model.ListChannelsWithPricingOffers()
		if err != nil {
			return nil, err
		}
		ids = derived
	}
	if len(ids) == 0 {
		return nil, nil
	}
	channels := make([]model.Channel, 0, len(ids))
	for _, id := range ids {
		channel, err := model.GetChannelById(id, true)
		if err != nil {
			return nil, err
		}
		if channel == nil || channel.Status != common.ChannelStatusEnabled {
			continue
		}
		channels = append(channels, *channel)
	}
	return channels, nil
}

func roundRatioValue(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return math.Round(value*1e8) / 1e8
}

// PricingCalibrationRatio resolves the multiplier the pricing path should use
// for one model. It returns false when no fresh, applied calibration exists, so
// the caller falls back to the group-ratio policy.
func PricingCalibrationRatio(calibrations map[string]model.UpstreamCostCalibration, channelID int, publicModel string, now int64) (float64, bool) {
	if len(calibrations) == 0 {
		return 0, false
	}
	row, ok := calibrations[model.UpstreamCostCalibrationKey(channelID, publicModel)]
	if !ok {
		return 0, false
	}
	if !row.IsUsable(now, pricingCalibrationStaleSeconds()) {
		return 0, false
	}
	return row.EffectiveRatio, true
}

type pricingCalibrationHandler struct{}

func (pricingCalibrationHandler) Type() string { return model.SystemTaskTypePricingCalibration }
func (pricingCalibrationHandler) Enabled() bool {
	return strings.EqualFold(os.Getenv("PRICING_CONTROL_ENABLED"), "true") && pricingCalibrationEnabled()
}
func (pricingCalibrationHandler) Interval() time.Duration {
	minutes := common.GetEnvOrDefault("PRICING_CALIBRATION_INTERVAL_MINUTES", 720)
	if minutes < 15 {
		minutes = 15
	}
	return time.Duration(minutes) * time.Minute
}
func (pricingCalibrationHandler) NewPayload() any { return map[string]any{"scheduled": true} }
func (pricingCalibrationHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	result, err := RunPricingCalibration(ctx, nil, "scheduled")
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("pricing calibration failed: %s", err.Error()))
		return
	}
	logger.LogInfo(ctx, fmt.Sprintf("pricing calibration applied=%d skipped=%d channels=%d",
		result.Applied, result.Skipped, len(result.Channels)))
}

func init() { RegisterSystemTaskHandler(pricingCalibrationHandler{}) }

// PricingCalibrationSummary is what the console shows above the model list:
// how much of the catalogue is measured, how old the measurements are, and how
// far the fallback policy sits from reality. Overstatement is the headline
// number: a value of 3 means the published group ratios would price the cost
// three times higher than the provider actually invoices.
func PricingCalibrationSummary() (map[string]any, error) {
	rows, err := model.ListUpstreamCostCalibrations()
	if err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	staleSeconds := pricingCalibrationStaleSeconds()
	applied, usable, stale := 0, 0, 0
	lastUpdated := int64(0)
	policySum, calibrationSum := 0.0, 0.0
	groups := make(map[string]int)
	for _, row := range rows {
		if row.UpdatedAt > lastUpdated {
			lastUpdated = row.UpdatedAt
		}
		if row.Applied {
			applied++
		}
		if row.IsUsable(now, staleSeconds) {
			usable++
			policySum += row.PolicyRatio
			calibrationSum += row.EffectiveRatio
		} else if row.Applied {
			stale++
		}
		if row.ObservedGroup != "" {
			groups[row.ObservedGroup]++
		}
	}
	overstatement := 0.0
	if calibrationSum > 0 {
		overstatement = math.Round(policySum/calibrationSum*100) / 100
	}
	observed := make([]map[string]any, 0, len(groups))
	for group, count := range groups {
		observed = append(observed, map[string]any{"group": group, "models": count})
	}
	sort.Slice(observed, func(i, j int) bool {
		return observed[i]["models"].(int) > observed[j]["models"].(int)
	})
	return map[string]any{
		"models":               len(rows),
		"applied":              applied,
		"usable":               usable,
		"stale":                stale,
		"last_calibrated_at":   lastUpdated,
		"policy_overstatement": overstatement,
		"observed_groups":      observed,
		"stale_hours":          staleSeconds / 3600,
	}, nil
}

// PinUpstreamCostCalibration freezes a factor that the automatic loop cannot
// derive, typically a low-traffic model with no recent requests. A pinned row
// outranks the automatic loop until it is cleared with ratio <= 0.
func PinUpstreamCostCalibration(channelID int, publicModel string, ratio float64, reason string) error {
	publicModel = strings.TrimSpace(publicModel)
	if channelID <= 0 || publicModel == "" {
		return errors.New("channel and model are required")
	}
	if ratio <= 0 || math.IsNaN(ratio) || math.IsInf(ratio, 0) {
		return errors.New("ratio must be a positive number")
	}
	now := common.GetTimestamp()
	return model.UpsertUpstreamCostCalibration(&model.UpstreamCostCalibration{
		ChannelID:      channelID,
		PublicModel:    publicModel,
		EffectiveRatio: ratio,
		Source:         model.PricingCalibrationSourceManual,
		Samples:        0,
		Confidence:     1,
		WindowStart:    now,
		WindowEnd:      now,
		Applied:        true,
		Reason:         strings.TrimSpace(reason),
	})
}

// ClearPinnedUpstreamCostCalibration removes a manual pin so the automatic loop
// takes over again on the next run.
func ClearPinnedUpstreamCostCalibration(channelID int, publicModel string) error {
	if channelID <= 0 || strings.TrimSpace(publicModel) == "" {
		return errors.New("channel and model are required")
	}
	return model.DeleteUpstreamCostCalibration(channelID, publicModel, model.PricingCalibrationSourceManual)
}
