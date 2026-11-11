package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	perfmetrics "github.com/QuantumNous/new-api/pkg/perf_metrics"
	"github.com/QuantumNous/new-api/setting/perf_metrics_setting"
)

const (
	probeTierCheapInterval     = 30 * time.Minute
	probeTierMediumInterval    = 120 * time.Minute
	probeTierExpensiveInterval = 360 * time.Minute
	probeRequestTimeout        = 60 * time.Second
	probeInterRequestGap       = 2 * time.Second
	probeDefaultTokenLimit     = 16
)

// ProbeSpec is the resolved probing plan for one model: which endpoint and
// token-limit parameter to use, at which cadence.
type ProbeSpec struct {
	Model      string
	Tier       string // cheap|medium|expensive
	TokenParam string // max_tokens|max_completion_tokens|none (chat endpoint)
	Limit      int
	Endpoint   string // chat|responses
	Interval   time.Duration
}

// ProbeResult is one probe attempt. It deliberately carries no token or
// channel information.
type ProbeResult struct {
	Model     string `json:"model"`
	Ok        bool   `json:"ok"`
	Status    int    `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
}

// ProbeRoundResult summarizes one scheduler pass.
type ProbeRoundResult struct {
	Skipped bool          `json:"skipped"`
	Reason  string        `json:"reason,omitempty"`
	Probed  int           `json:"probed"`
	Failed  int           `json:"failed"`
	Results []ProbeResult `json:"results,omitempty"`
}

var (
	probeSkipNamePattern      = regexp.MustCompile(`(?i)image|seedance|video|tts|whisper|embedding|rerank`)
	probeExpensiveNamePattern = regexp.MustCompile(`(?i)opus|fable|sol|astra|-pro$|-pro-`)
	probeCheapNamePattern     = regexp.MustCompile(`(?i)flash-lite|flash|mini|luna|lite|glm|grok|deepseek-.*flash`)
)

// DefaultTokenParam picks the upstream token-limit parameter for a model.
// gpt-5/gpt-6/o-series reject max_tokens (and max_tokens:1 breaks several
// providers), so they must use max_completion_tokens.
func DefaultTokenParam(model string) string {
	lower := strings.ToLower(model)
	for _, prefix := range []string{"gpt-5", "gpt-6", "o1", "o3", "o4"} {
		if strings.HasPrefix(lower, prefix) {
			return "max_completion_tokens"
		}
	}
	return "max_tokens"
}

// DefaultTier infers the probe cadence tier from the model name. Inferred
// premium models (opus/fable/sol/astra/-pro) default to off: one opus-class
// probe can cost $0.1+.
func DefaultTier(model string) string {
	lower := strings.ToLower(model)
	if probeExpensiveNamePattern.MatchString(lower) {
		return "off"
	}
	if probeCheapNamePattern.MatchString(lower) {
		return "cheap"
	}
	return "medium"
}

func tierInterval(tier string) (time.Duration, bool) {
	switch tier {
	case "cheap":
		return probeTierCheapInterval, true
	case "medium":
		return probeTierMediumInterval, true
	case "expensive":
		return probeTierExpensiveInterval, true
	default:
		return 0, false
	}
}

type probeOverride struct {
	tier       string
	tokenParam string
	limit      int
	endpoint   string
}

// parseProbeModels parses the probe_models setting. Lines starting with '#'
// (including inline trailing comments) are ignored; format:
// `<model> <tier> [token_param] [limit] [endpoint:chat|responses]`.
func parseProbeModels(text string) map[string]*probeOverride {
	overrides := map[string]*probeOverride{}
	for _, rawLine := range strings.Split(text, "\n") {
		if idx := strings.Index(rawLine, "#"); idx >= 0 {
			rawLine = rawLine[:idx]
		}
		fields := strings.Fields(rawLine)
		if len(fields) < 2 {
			continue
		}
		override := &probeOverride{tier: fields[1]}
		for _, field := range fields[2:] {
			switch {
			case field == "max_tokens" || field == "max_completion_tokens" || field == "none":
				override.tokenParam = field
			case field == "endpoint:chat" || field == "endpoint:responses":
				override.endpoint = strings.TrimPrefix(field, "endpoint:")
			default:
				if limit, err := strconv.Atoi(field); err == nil && limit > 0 {
					override.limit = limit
				}
			}
		}
		overrides[fields[0]] = override
	}
	return overrides
}

// LoadProbeSpecs resolves the probing plan for all enabled models: skip
// non-chat media models, pick chat or responses from the pricing endpoint
// support, apply probe_models overrides, then name-based defaults.
func LoadProbeSpecs(enabledModels []string, pricing []model.Pricing, cfg perf_metrics_setting.PerfMetricsSetting) []ProbeSpec {
	supported := make(map[string][]constant.EndpointType, len(pricing))
	for _, entry := range pricing {
		supported[entry.ModelName] = entry.SupportedEndpointTypes
	}
	overrides := parseProbeModels(cfg.ProbeModels)

	specs := make([]ProbeSpec, 0)
	for _, name := range enabledModels {
		if probeSkipNamePattern.MatchString(strings.ToLower(name)) {
			continue
		}
		if spec, ok := buildProbeSpec(name, supported[name], overrides[name]); ok {
			specs = append(specs, spec)
		}
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Model < specs[j].Model })
	return specs
}

func buildProbeSpec(name string, endpointTypes []constant.EndpointType, override *probeOverride) (ProbeSpec, bool) {
	supportsChat, supportsResponses := false, false
	for _, endpointType := range endpointTypes {
		if endpointType == constant.EndpointTypeOpenAI {
			supportsChat = true
		}
		if endpointType == constant.EndpointTypeOpenAIResponse || endpointType == constant.EndpointTypeOpenAIResponseCompact {
			supportsResponses = true
		}
	}

	endpoint := ""
	if supportsChat {
		endpoint = "chat"
	} else if supportsResponses {
		endpoint = "responses"
	}
	tier := DefaultTier(name)
	tokenParam := DefaultTokenParam(name)
	limit := probeDefaultTokenLimit
	if override != nil {
		if override.tier != "" {
			tier = override.tier
		}
		if override.tokenParam != "" {
			tokenParam = override.tokenParam
		}
		if override.limit > 0 {
			limit = override.limit
		}
		if override.endpoint != "" {
			endpoint = override.endpoint
		}
	}
	if endpoint == "" || tier == "off" {
		return ProbeSpec{}, false
	}
	interval, ok := tierInterval(tier)
	if !ok {
		return ProbeSpec{}, false
	}
	return ProbeSpec{
		Model:      name,
		Tier:       tier,
		TokenParam: tokenParam,
		Limit:      limit,
		Endpoint:   endpoint,
		Interval:   interval,
	}, true
}

// probeBudgetZone is Asia/Shanghai as a fixed offset so the daily budget
// window does not depend on tzdata being installed.
var probeBudgetZone = time.FixedZone("Asia/Shanghai", 8*3600)

func startOfProbeBudgetDay() int64 {
	now := time.Now().In(probeBudgetZone)
	year, month, day := now.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, probeBudgetZone).Unix()
}

// ProbeBudgetLeft returns the remaining daily probe budget in quota, measured
// as today's consumed quota of the probe token (Asia/Shanghai day). It sums by
// token id: names are per-user, so a name-based sum would count traffic of
// unrelated same-name tokens.
func ProbeBudgetLeft(tokenId int, budgetQuota int) (int64, error) {
	if tokenId <= 0 {
		return 0, errors.New("probe token id is required")
	}
	if budgetQuota <= 0 {
		return 0, nil
	}
	var stat struct {
		Quota int64
	}
	err := model.LOG_DB.Table("logs").
		Select("COALESCE(sum(quota), 0) quota").
		Where("token_id = ? AND type = ? AND created_at >= ?", tokenId, model.LogTypeConsume, startOfProbeBudgetDay()).
		Scan(&stat).Error
	if err != nil {
		return 0, err
	}
	return int64(budgetQuota) - stat.Quota, nil
}

// ResolveProbeToken loads the enabled probe token owned by an admin/root
// user; a same-name token of a regular user is never picked. The returned
// error explains why probing cannot run and never exposes the token key.
func ResolveProbeToken(name string) (*model.Token, error) {
	if name == "" {
		return nil, errors.New("probe token name is empty")
	}
	token, err := model.GetAdminTokenByName(name)
	if err != nil {
		return nil, fmt.Errorf("admin-owned probe token %q not found", name)
	}
	if token.Status != common.TokenStatusEnabled {
		return nil, fmt.Errorf("probe token %q is disabled", name)
	}
	if token.Key == "" {
		return nil, fmt.Errorf("probe token %q has no key", name)
	}
	return token, nil
}

var probeLastRunFallback sync.Map // model -> unix ts, used when Redis is unavailable

func probeLastRunKey(model string) string {
	return "perf:probe:last:" + model
}

func getProbeLastRun(model string) int64 {
	if common.RedisEnabled && common.RDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		value, err := common.RDB.Get(ctx, probeLastRunKey(model)).Result()
		if err == nil {
			if ts, parseErr := strconv.ParseInt(value, 10, 64); parseErr == nil {
				return ts
			}
		}
		return 0
	}
	if cached, ok := probeLastRunFallback.Load(model); ok {
		return cached.(int64)
	}
	return 0
}

func setProbeLastRun(model string, ts int64) {
	if common.RedisEnabled && common.RDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = common.RDB.Set(ctx, probeLastRunKey(model), strconv.FormatInt(ts, 10), 14*24*time.Hour).Err()
		return
	}
	probeLastRunFallback.Store(model, ts)
}

// GetProbeLastRunAt exposes the last probe timestamp for the admin perf view.
func GetProbeLastRunAt(model string) int64 {
	return getProbeLastRun(model)
}

// RunModelProbeRound is the scheduled entry point: it checks the switch, the
// probe token and the daily budget, resolves specs, then probes the due
// models. Probes go through the normal relay chain, so billing, logging and
// perf sampling all apply with source=probe.
func RunModelProbeRound(ctx context.Context) ProbeRoundResult {
	cfg := perf_metrics_setting.GetSetting()
	if !cfg.ProbeEnabled {
		return ProbeRoundResult{Skipped: true, Reason: "probe disabled"}
	}

	probeToken, err := ResolveProbeToken(cfg.ProbeTokenName)
	if err != nil {
		common.SysError("model probe skipped: " + err.Error())
		return ProbeRoundResult{Skipped: true, Reason: "probe token unavailable"}
	}

	budgetLeft, err := ProbeBudgetLeft(probeToken.Id, cfg.ProbeDailyBudgetQuota)
	if err != nil {
		common.SysError("model probe skipped: failed to read probe budget: " + err.Error())
		return ProbeRoundResult{Skipped: true, Reason: "probe budget unavailable"}
	}

	return runProbeRound(ctx, probeRoundDeps{
		tokenKey:   probeToken.Key,
		specs:      LoadProbeSpecs(model.GetEnabledModels(), model.GetPricing(), cfg),
		budgetLeft: budgetLeft,
		threshold:  cfg.ProbeRealTrafficThreshold,
		maxPerRun:  cfg.ProbeMaxPerRun,
	})
}

// probeRoundDeps collects the injectable inputs of one probe pass so the
// scheduling logic stays testable without a database or live relay.
type probeRoundDeps struct {
	tokenKey   string
	specs      []ProbeSpec
	budgetLeft int64
	threshold  int
	maxPerRun  int
	now        func() time.Time
	send       func(ctx context.Context, spec ProbeSpec, tokenKey string) ProbeResult
}

func runProbeRound(ctx context.Context, deps probeRoundDeps) ProbeRoundResult {
	if deps.now == nil {
		deps.now = time.Now
	}
	if deps.send == nil {
		deps.send = probeOnce
	}

	if deps.budgetLeft <= 0 {
		return ProbeRoundResult{Skipped: true, Reason: "daily probe budget exhausted"}
	}

	now := deps.now().Unix()
	type dueSpec struct {
		spec    ProbeSpec
		lastRun int64
	}
	due := make([]dueSpec, 0)
	for _, spec := range deps.specs {
		lastRun := getProbeLastRun(spec.Model)
		if now-lastRun < int64(spec.Interval.Seconds()) {
			continue
		}
		if perfmetrics.RealRequestCountSince(spec.Model, now-3600) >= int64(deps.threshold) {
			continue
		}
		due = append(due, dueSpec{spec: spec, lastRun: lastRun})
	}
	// Oldest probe first so nothing starves when maxPerRun truncates the batch.
	sort.Slice(due, func(i, j int) bool { return due[i].lastRun < due[j].lastRun })
	if deps.maxPerRun >= 0 && len(due) > deps.maxPerRun {
		due = due[:deps.maxPerRun]
	}

	result := ProbeRoundResult{}
	for _, item := range due {
		select {
		case <-ctx.Done():
			return result
		default:
		}
		probeResult := deps.send(ctx, item.spec, deps.tokenKey)
		setProbeLastRun(item.spec.Model, deps.now().Unix())
		result.Probed++
		if !probeResult.Ok {
			result.Failed++
		}
		result.Results = append(result.Results, probeResult)
		select {
		case <-ctx.Done():
			return result
		case <-time.After(probeInterRequestGap):
		}
	}
	return result
}

var probeHTTPClient = &http.Client{Timeout: probeRequestTimeout}

// probeBaseURL targets the local relay listener; swappable in tests.
var probeBaseURL = func() string {
	return fmt.Sprintf("http://127.0.0.1:%d", *common.Port)
}

func probeOnce(ctx context.Context, spec ProbeSpec, tokenKey string) ProbeResult {
	start := time.Now()
	status, err := sendProbeRequest(ctx, spec, tokenKey)
	return ProbeResult{
		Model:     spec.Model,
		Ok:        err == nil && status >= 200 && status < 300,
		Status:    status,
		LatencyMs: time.Since(start).Milliseconds(),
	}
}

func sendProbeRequest(ctx context.Context, spec ProbeSpec, tokenKey string) (int, error) {
	payload, err := buildProbePayload(spec)
	if err != nil {
		return 0, err
	}

	path := "/v1/chat/completions"
	if spec.Endpoint == "responses" {
		path = "/v1/responses"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, probeBaseURL()+path, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := probeHTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	// Drain the SSE stream so the relay completes billing and perf sampling
	// before we hang up.
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func buildProbePayload(spec ProbeSpec) ([]byte, error) {
	if spec.Endpoint == "responses" {
		return common.Marshal(map[string]any{
			"model":             spec.Model,
			"input":             "ping",
			"stream":            true,
			"max_output_tokens": spec.Limit,
		})
	}
	payload := map[string]any{
		"model":    spec.Model,
		"messages": []map[string]string{{"role": "user", "content": "ping"}},
		"stream":   true,
	}
	if spec.Limit <= 0 {
		return common.Marshal(payload)
	}
	switch spec.TokenParam {
	case "none":
	case "max_completion_tokens":
		payload["max_completion_tokens"] = spec.Limit
	default:
		payload["max_tokens"] = spec.Limit
	}
	return common.Marshal(payload)
}
