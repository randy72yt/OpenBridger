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
	// The provider ignores page_size and returns 142 rows per page, so a 72h
	// window can sit hundreds of pages in. Probes double up to this ceiling.
	pricingCalibrationProbePageCeiling = 4096
	// Per-model walk bounds. One page carries 142 rows, which is far more than
	// the five samples a factor needs, so a second page is already the exception.
	pricingCalibrationModelPages = 4
	// Keeps a single very busy model from crowding the sample set and from
	// dominating the run time.
	pricingCalibrationModelRecords = 600
	// Upper bound on requests to a third-party provider per channel per run.
	pricingCalibrationWindowRequestBudget = 200
	// The log endpoint rate-limits by IP. One request per model turns a 28-model
	// catalogue into a burst that answers 429, so requests are spaced out and a
	// 429 is retried with backoff instead of aborting the pass.
	pricingCalibrationMaxRetries = 4	// DefaultQuotaPerUnit is the new-api convention: one US dollar is worth
	// 500000 quota units. A provider that publishes a different snapshot in the
	// per-request billing detail still lands on the same quotient because both
	// the billed and the theoretical side use the same unit.
	pricingCalibrationDefaultQuotaPerUnit = 500000.0

	pricingCalibrationDefaultWindowHours   = 72
	pricingCalibrationDefaultMinSamples    = 5
	pricingCalibrationDefaultMaxDispersion = 15.0
	pricingCalibrationDefaultStaleHours    = 96
)

// Pacing and backoff for the provider's log endpoint. Variables rather than
// constants so tests can shrink them; nothing in production writes to them.
var (
	pricingCalibrationModelGap      = 200 * time.Millisecond
	pricingCalibrationRetryBaseWait = 2 * time.Second
	pricingCalibrationMaxRetryWait  = 30 * time.Second
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
	ChannelID int    `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Models      int    `json:"models"`
	Applied     int    `json:"applied"`
	Records     int    `json:"records"`
	// FailedModels counts models whose log query errored and was skipped. A
	// partially blocked pass still produces usable factors for the rest, so this
	// is reported instead of aborting.
	FailedModels int    `json:"failed_models,omitempty"`
	Error        string `json:"error,omitempty"`
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

// fetchUpstreamUsageLogs walks the provider's consumption log and keeps only the
// entries inside the window. It is what makes calibration possible at all: this
// is the only endpoint that reports what was actually charged.
//
// The provider returns the log oldest-first, so the window sits at the tail of
// the log rather than the head. Walking forward from page 1 and stopping at the
// first old entry therefore reads the wrong end and finds almost nothing - which
// is exactly what happened on the first run (19 rows out of ~2000 in the window).
// An exponential search locates the first page that reaches the window, then the
// walk continues forward. The same code still works if a provider flips to
// newest-first: page 1 is already inside the window, and the "whole page is
// older than the window" rule ends the walk instead.
//
// This is the fallback path. It sees whatever the provider happens to have
// logged recently, which on a busy account is minutes of traffic covering a
// handful of models - on the first production run, 3060 rows that only matched
// 3 of 28 models. fetchChannelUsageLogs prefers the per-model window walk and
// only lands here when the provider ignores the window filter.
func fetchUpstreamUsageLogs(ctx context.Context, channel *model.Channel, since int64) ([]upstreamUsageLogRecord, error) {
	client, err := newUpstreamUsageLogClient(ctx, channel)
	if err != nil {
		return nil, err
	}
	return client.tailWalk(since)
}

// upstreamUsageLogClient holds one channel's log connection so both walkers
// send identical headers, decode identically and share a request budget: the
// provider is a third party and a calibration pass must not turn into a crawl.
type upstreamUsageLogClient struct {
	ctx      context.Context
	endpoint *url.URL
	key      string
	client   *http.Client
	requests int
}

func newUpstreamUsageLogClient(ctx context.Context, channel *model.Channel) (*upstreamUsageLogClient, error) {
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
	return &upstreamUsageLogClient{
		ctx:      ctx,
		endpoint: endpoint,
		key:      strings.TrimSpace(key),
		client:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// upstreamRateLimitError is the provider's 429, kept as its own type so the
// retry loop can tell "you are going too fast" apart from "something is broken".
type upstreamRateLimitError struct {
	status     string
	retryAfter time.Duration
}

func (e *upstreamRateLimitError) Error() string {
	if e.retryAfter > 0 {
		return fmt.Sprintf("upstream usage log rate-limited (%s, retry after %s)", e.status, e.retryAfter)
	}
	return fmt.Sprintf("upstream usage log rate-limited (%s)", e.status)
}

// parseRetryAfter accepts both forms RFC 7231 allows: delay-seconds and HTTP-date.
func parseRetryAfter(raw string) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(raw); err == nil {
		if wait := time.Until(when); wait > 0 {
			return wait
		}
	}
	return 0
}

// fetchPage asks for one page, retrying the provider's rate limit instead of
// failing the whole pass. A per-model walk sends one request per model, which a
// wide catalogue turns into exactly the burst the log endpoint throttles.
func (c *upstreamUsageLogClient) fetchPage(page int, since, until int64, modelName string) ([]upstreamUsageLogRecord, error) {
	wait := pricingCalibrationRetryBaseWait
	var lastErr error
	for attempt := 1; attempt <= pricingCalibrationMaxRetries; attempt++ {
		rows, retryAfter, err := c.fetchPageOnce(page, since, until, modelName)
		if err == nil {
			return rows, nil
		}
		lastErr = err
		var limited *upstreamRateLimitError
		if !errors.As(err, &limited) {
			return nil, err
		}
		if attempt == pricingCalibrationMaxRetries {
			break
		}
		pause := retryAfter
		if pause <= 0 {
			pause = wait
		}
		if pause > pricingCalibrationMaxRetryWait {
			pause = pricingCalibrationMaxRetryWait
		}
		if c.ctx != nil {
			select {
			case <-time.After(pause):
			case <-c.ctx.Done():
				return nil, c.ctx.Err()
			}
		} else {
			time.Sleep(pause)
		}
		if wait < pricingCalibrationMaxRetryWait {
			wait *= 2
		}
	}
	return nil, lastErr
}

// fetchPageOnce sends one request. since/until/modelName are sent to the
// provider, which filters server-side; callers still re-check the window
// client-side so a provider that ignores a filter cannot smuggle out-of-window
// rows in.
func (c *upstreamUsageLogClient) fetchPageOnce(page int, since, until int64, modelName string) ([]upstreamUsageLogRecord, time.Duration, error) {
	query := c.endpoint.Query()
	query.Set("p", strconv.Itoa(page))
	query.Set("page_size", strconv.Itoa(pricingCalibrationPageSize))
	query.Set("type", "2")
	if since > 0 {
		query.Set("start_timestamp", strconv.FormatInt(since, 10))
	}
	if until > 0 {
		query.Set("end_timestamp", strconv.FormatInt(until, 10))
	}
	if modelName != "" {
		query.Set("model_name", modelName)
	}
	pageURL := *c.endpoint
	pageURL.RawQuery = query.Encode()
	req, reqErr := http.NewRequestWithContext(c.ctx, http.MethodGet, pageURL.String(), nil)
	if reqErr != nil {
		return nil, 0, reqErr
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Accept", "application/json")
	c.requests++
	resp, doErr := c.client.Do(req)
	if doErr != nil {
		return nil, 0, doErr
	}
	var payload upstreamUsageLogResponse
	decodeErr := common.DecodeJson(io.LimitReader(resp.Body, pricingCalibrationResponseLimit), &payload)
	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return nil, retryAfter, &upstreamRateLimitError{status: resp.Status, retryAfter: retryAfter}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("upstream usage log returned %s", resp.Status)
	}
	if decodeErr != nil {
		return nil, 0, decodeErr
	}
	if !payload.Success && payload.Message != "" {
		return nil, 0, errors.New(payload.Message)
	}
	return payload.Data, 0, nil
}

func (c *upstreamUsageLogClient) tailWalk(since int64) ([]upstreamUsageLogRecord, error) {
	startPage := 1
	// tailMode marks an oldest-first log, where the pages before the window are
	// simply old and must be skipped rather than treated as the end of the walk.
	tailMode := false
	if since > 0 {
		first, probeErr := c.fetchPage(1, since, 0, "")
		if probeErr != nil {
			return nil, probeErr
		}
		if len(first) > 0 && newestCreatedAt(first) < since {
			tailMode = true
			// Oldest-first log: the window sits at the tail, so the last page has
			// to be located before we know where to start reading.
			lastPage := 1
			emptyAt := 0
			for probe := 2; probe <= pricingCalibrationProbePageCeiling; probe *= 2 {
				rows, err := c.fetchPage(probe, since, 0, "")
				if err != nil {
					return nil, err
				}
				if len(rows) == 0 {
					emptyAt = probe
					break
				}
				lastPage = probe
			}
			if emptyAt > 0 {
				lo, hi := lastPage, emptyAt
				for hi-lo > 1 {
					mid := (lo + hi) / 2
					rows, err := c.fetchPage(mid, since, 0, "")
					if err != nil {
						return nil, err
					}
					if len(rows) == 0 {
						hi = mid
					} else {
						lo = mid
					}
				}
				lastPage = lo
			}
			startPage = lastPage - pricingCalibrationMaxPages + 1
			if startPage < 1 {
				startPage = 1
			}
		}
	}

	records := make([]upstreamUsageLogRecord, 0, pricingCalibrationPageSize*4)
	for page := startPage; page < startPage+pricingCalibrationMaxPages; page++ {
		if c.ctx != nil && c.ctx.Err() != nil {
			return records, c.ctx.Err()
		}
		rows, pageErr := c.fetchPage(page, since, 0, "")
		if pageErr != nil {
			return records, pageErr
		}
		if len(rows) == 0 {
			return records, nil
		}
		inWindow := 0
		for _, record := range rows {
			if since > 0 && record.CreatedAt > 0 && record.CreatedAt < since {
				continue
			}
			inWindow++
			records = append(records, record)
		}
		if inWindow == 0 && !tailMode {
			// Newest-first provider: this page predates the window entirely.
			return records, nil
		}
		if len(records) >= pricingCalibrationMaxRecords {
			return records, nil
		}
	}
	return records, nil
}

// windowWalk reads the window one model at a time. A single global tail walk
// only ever sees the models that happen to be hot right now, which is why the
// first production pass priced 3 of 28 models; asking per model gives every
// model in the catalogue the same chance regardless of how much of the traffic
// it owns. The provider filters on start_timestamp/end_timestamp/model_name, so
// one page usually carries enough samples for a model.
//
// supported=false means the provider ignored the window filter (an unsupported
// fork, or a proxy that strips query parameters). The caller then falls back to
// the tail walk instead of reading the wrong end of the log.
func (c *upstreamUsageLogClient) windowWalk(models []string, since, until int64) ([]upstreamUsageLogRecord, bool, int, error) {
	records := make([]upstreamUsageLogRecord, 0, pricingCalibrationPageSize*len(models))
	probed := false
	failed := 0
	var lastErr error
	for index, name := range models {
		if c.requests >= pricingCalibrationWindowRequestBudget {
			break
		}
		if c.ctx != nil && c.ctx.Err() != nil {
			return records, true, failed, c.ctx.Err()
		}
		if index > 0 && pricingCalibrationModelGap > 0 {
			// Pace the walk. The log endpoint throttles bursts, and a 200ms gap is
			// far cheaper than a 429 plus its backoff.
			if c.ctx != nil {
				select {
				case <-time.After(pricingCalibrationModelGap):
				case <-c.ctx.Done():
					return records, true, failed, c.ctx.Err()
				}
			} else {
				time.Sleep(pricingCalibrationModelGap)
			}
		}
		kept := 0
		for page := 1; page <= pricingCalibrationModelPages; page++ {
			rows, err := c.fetchPage(page, since, until, name)
			if err != nil {
				// One blocked model must not throw away the factors every other
				// model already produced. Abort only when nothing has come back and
				// two models in a row failed, which means the endpoint itself is
				// unusable (bad key, wrong path, persistent throttle).
				failed++
				lastErr = err
				if len(records) == 0 && failed >= 2 {
					return records, true, failed, lastErr
				}
				break
			}
			if len(rows) == 0 {
				break
			}
			if !probed {
				probed = true
				if since > 0 && newestCreatedAt(rows) < since {
					// The filter was ignored: this is the oldest page of an
					// unfiltered log, not a page from inside the window.
					return nil, false, failed, nil
				}
			}
			for _, record := range rows {
				if since > 0 && record.CreatedAt > 0 && record.CreatedAt < since {
					continue
				}
				if until > 0 && record.CreatedAt > 0 && record.CreatedAt > until {
					continue
				}
				records = append(records, record)
				kept++
			}
			if kept >= pricingCalibrationModelRecords {
				break
			}
		}
	}
	return records, true, failed, nil
}

// fetchChannelUsageLogs prefers the per-model window walk and degrades to the
// tail walk when the provider ignores the window filter. failed counts models
// the provider refused (rate limit or transient error); the run keeps the rest.
func fetchChannelUsageLogs(ctx context.Context, channel *model.Channel, since, until int64, models []string) ([]upstreamUsageLogRecord, bool, int, error) {
	client, err := newUpstreamUsageLogClient(ctx, channel)
	if err != nil {
		return nil, false, 0, err
	}
	if len(models) > 0 {
		records, supported, failed, walkErr := client.windowWalk(models, since, until)
		if walkErr != nil {
			return records, supported, failed, walkErr
		}
		if supported {
			return records, true, failed, nil
		}
	}
	records, err := client.tailWalk(since)
	return records, false, 0, err
}

func newestCreatedAt(rows []upstreamUsageLogRecord) int64 {
	newest := int64(0)
	for _, row := range rows {
		if row.CreatedAt > newest {
			newest = row.CreatedAt
		}
	}
	return newest
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

// calibrationModelNames lists the upstream model names to ask the provider
// about, in a stable order so a run is reproducible and the request budget is
// spent on the same models every time.
func calibrationModelNames(offers map[string]model.UpstreamModelOffer) []string {
	names := make([]string, 0, len(offers))
	for _, offer := range offers {
		name := strings.TrimSpace(offer.UpstreamModel)
		if name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
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
		records, _, failedModels, logErr := fetchChannelUsageLogs(ctx, &channel, result.WindowStart, result.WindowEnd, calibrationModelNames(offers))
		if logErr != nil {
			channelResult.Error = logErr.Error()
			channelResult.FailedModels = failedModels
			result.Channels = append(result.Channels, channelResult)
			failures = append(failures, fmt.Sprintf("channel %d: %s", channel.Id, logErr.Error()))
			continue
		}
		channelResult.Records = len(records)
		channelResult.FailedModels = failedModels
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
		// The runner learns the outcome only through FinishSystemTask. Returning
		// without it left the row running until the 60s lease expired, so every
		// run - including the ones that worked - ended up as "task lease
		// expired" in the console and hid the real reason.
		if finishErr := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusFailed, result, err.Error()); finishErr != nil {
			logSystemTaskLockError(ctx, task, finishErr)
		}
		return
	}
	records := 0
	failedModels := 0
	for _, channel := range result.Channels {
		records += channel.Records
		failedModels += channel.FailedModels
	}
	// Records is logged because "applied=0" alone cannot tell an empty window
	// apart from a window that was never read. failed_models shows throttled
	// models that were skipped instead of aborting the pass.
	logger.LogInfo(ctx, fmt.Sprintf("pricing calibration applied=%d skipped=%d channels=%d records=%d failed_models=%d",
		result.Applied, result.Skipped, len(result.Channels), records, failedModels))
	if finishErr := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, result, ""); finishErr != nil {
		logSystemTaskLockError(ctx, task, finishErr)
	}
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
