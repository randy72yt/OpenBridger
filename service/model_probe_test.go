package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	perfmetrics "github.com/QuantumNous/new-api/pkg/perf_metrics"
	"github.com/QuantumNous/new-api/setting/perf_metrics_setting"
)

func TestDefaultTokenParam(t *testing.T) {
	cases := []struct {
		model string
		want  string
	}{
		{"gpt-5.6-luna", "max_completion_tokens"},
		{"GPT-6-SOL", "max_completion_tokens"},
		{"o3-mini", "max_completion_tokens"},
		{"o4-mini", "max_completion_tokens"},
		{"claude-haiku-4-5", "max_tokens"},
		{"gemini-3.1-flash-lite", "max_tokens"},
		{"glm-5.3", "max_tokens"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, DefaultTokenParam(tc.model), tc.model)
	}
}

func TestDefaultTier(t *testing.T) {
	cases := []struct {
		model string
		want  string
	}{
		{"claude-opus-5", "off"},
		{"claude-fable-5-1", "off"},
		{"gpt-5.6-sol", "off"},
		{"gpt-6-astra", "off"},
		{"gemini-3-pro", "off"}, // -pro$ infers premium, default off
		{"gpt-5.6-luna", "cheap"},
		{"gemini-3.1-flash-lite", "cheap"},
		{"deepseek-v4.1-flash", "cheap"},
		{"glm-5.3", "cheap"},
		{"grok-4.7", "cheap"},
		{"claude-haiku-4-5", "medium"},
		{"kimi-k3", "medium"},
		{"gpt-5.6", "medium"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, DefaultTier(tc.model), tc.model)
	}
}

func chatPricing(name string) model.Pricing {
	return model.Pricing{ModelName: name, SupportedEndpointTypes: []constant.EndpointType{constant.EndpointTypeOpenAI}}
}

func TestLoadProbeSpecs(t *testing.T) {
	cfg := perf_metrics_setting.PerfMetricsSetting{ProbeModels: `
# comment line is ignored
gemini-3-pro          medium    max_tokens 8
gpt-5.6-luna          cheap     max_completion_tokens 16
kimi-k3               off
claude-haiku-4-5      medium    max_tokens 16   # trailing comment
`}
	enabled := []string{
		"gpt-5.6-luna", "gemini-3-pro", "kimi-k3", "claude-haiku-4-5",
		"gpt-image-2", "doubao-seedance-2.0", "tts-1", "whisper-1",
		"embedding-3", "video-gen-x", "rerank-x",
		"codex-auto-review", "audio-only-model", "not-enabled-model-x",
	}
	pricing := []model.Pricing{
		chatPricing("gpt-5.6-luna"),
		chatPricing("gemini-3-pro"),
		chatPricing("kimi-k3"),
		chatPricing("claude-haiku-4-5"),
		chatPricing("gpt-image-2"),
		chatPricing("doubao-seedance-2.0"),
		chatPricing("tts-1"),
		chatPricing("whisper-1"),
		chatPricing("embedding-3"),
		chatPricing("video-gen-x"),
		chatPricing("rerank-x"),
		{ModelName: "codex-auto-review", SupportedEndpointTypes: []constant.EndpointType{constant.EndpointTypeOpenAIResponse}},
		{ModelName: "audio-only-model", SupportedEndpointTypes: []constant.EndpointType{constant.EndpointTypeGemini}},
	}

	specs := LoadProbeSpecs(enabled, pricing, cfg)
	byModel := map[string]ProbeSpec{}
	for _, spec := range specs {
		byModel[spec.Model] = spec
	}

	// Skip media models and models without chat/responses endpoints.
	for _, skipped := range []string{"gpt-image-2", "doubao-seedance-2.0", "tts-1", "whisper-1", "embedding-3", "video-gen-x", "rerank-x", "audio-only-model", "kimi-k3"} {
		_, exists := byModel[skipped]
		assert.False(t, exists, skipped)
	}

	// Responses-only models probe through /v1/responses.
	assert.Equal(t, "responses", byModel["codex-auto-review"].Endpoint)
	assert.Equal(t, probeTierMediumInterval, byModel["codex-auto-review"].Interval)

	// probe_models overrides beat name-based defaults.
	assert.Equal(t, "medium", byModel["gemini-3-pro"].Tier)
	assert.Equal(t, "max_tokens", byModel["gemini-3-pro"].TokenParam)
	assert.Equal(t, 8, byModel["gemini-3-pro"].Limit)
	assert.Equal(t, probeTierMediumInterval, byModel["gemini-3-pro"].Interval)
	assert.Equal(t, "max_completion_tokens", byModel["gpt-5.6-luna"].TokenParam)

	// Lines for models that are not enabled are ignored.
	assert.Len(t, specs, 4)
}

func TestRunProbeRoundBudgetExhausted(t *testing.T) {
	var calls atomic.Int64
	result := runProbeRound(context.Background(), probeRoundDeps{
		tokenKey:   "test-key",
		specs:      []ProbeSpec{{Model: "m1", Tier: "cheap", Interval: probeTierCheapInterval, Endpoint: "chat", Limit: 16, TokenParam: "max_tokens"}},
		budgetLeft: 0,
		threshold:  3,
		maxPerRun:  6,
		send: func(ctx context.Context, spec ProbeSpec, tokenKey string) ProbeResult {
			calls.Add(1)
			return ProbeResult{Model: spec.Model, Ok: true}
		},
	})
	assert.True(t, result.Skipped)
	assert.Zero(t, result.Probed)
	assert.Zero(t, calls.Load())
}

func TestRunProbeRoundSkipsModelsUntilDue(t *testing.T) {
	now := time.Unix(1794300000, 0)
	busyModel := "probe-busy-model-x"
	idleModel := "probe-idle-model-x"
	cooldownModel := "probe-cooldown-model-x"

	// Fresh real traffic: busy model must not be probed.
	for range 3 {
		perfmetrics.Record(perfmetrics.Sample{Model: busyModel, Group: "default", LatencyMs: 10, Success: true, Source: perfmetrics.SourceReal})
	}

	setProbeLastRun(cooldownModel, now.Add(-10*time.Minute).Unix()) // medium tier = 120min, still cooling
	setProbeLastRun(idleModel, now.Add(-300*time.Minute).Unix())    // long overdue

	var probed []string
	result := runProbeRound(context.Background(), probeRoundDeps{
		tokenKey:  "test-key",
		budgetLeft: 500000,
		threshold: 3,
		maxPerRun: 6,
		now:       func() time.Time { return now },
		send: func(ctx context.Context, spec ProbeSpec, tokenKey string) ProbeResult {
			probed = append(probed, spec.Model)
			return ProbeResult{Model: spec.Model, Ok: true}
		},
		specs: []ProbeSpec{
			{Model: busyModel, Tier: "cheap", Interval: probeTierCheapInterval, Endpoint: "chat"},
			{Model: cooldownModel, Tier: "medium", Interval: probeTierMediumInterval, Endpoint: "chat"},
			{Model: idleModel, Tier: "medium", Interval: probeTierMediumInterval, Endpoint: "chat"},
		},
	})

	assert.False(t, result.Skipped)
	assert.Equal(t, 1, result.Probed)
	assert.Equal(t, []string{idleModel}, probed)
	// The round marks probed models, so the next pass leaves them cooling down.
	assert.Equal(t, now.Unix(), getProbeLastRun(idleModel))
}

func TestRunProbeRoundRespectsMaxPerRun(t *testing.T) {
	now := time.Unix(1794300000, 0)
	var probed atomic.Int64
	result := runProbeRound(context.Background(), probeRoundDeps{
		tokenKey:   "test-key",
		budgetLeft: 500000,
		threshold:  3,
		maxPerRun:  2,
		now:        func() time.Time { return now },
		send: func(ctx context.Context, spec ProbeSpec, tokenKey string) ProbeResult {
			probed.Add(1)
			return ProbeResult{Model: spec.Model, Ok: false}
		},
		specs: []ProbeSpec{
			{Model: "probe-maxrun-a", Tier: "cheap", Interval: probeTierCheapInterval, Endpoint: "chat"},
			{Model: "probe-maxrun-b", Tier: "cheap", Interval: probeTierCheapInterval, Endpoint: "chat"},
			{Model: "probe-maxrun-c", Tier: "cheap", Interval: probeTierCheapInterval, Endpoint: "chat"},
		},
	})
	assert.Equal(t, 2, result.Probed)
	assert.Equal(t, 2, result.Failed)
	assert.Equal(t, int64(2), probed.Load())
}

func stubProbeTransport(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	origBase, origClient := probeBaseURL, probeHTTPClient
	probeBaseURL = func() string { return server.URL }
	probeHTTPClient = server.Client()
	t.Cleanup(func() {
		probeBaseURL, probeHTTPClient = origBase, origClient
		server.Close()
	})
	return server
}

func TestProbeOnceChatSSERequests(t *testing.T) {
	var seenAuth string
	var bodies []string
	stubProbeTransport(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		seenAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"pong\"}}]}\n\ndata: [DONE]\n\n")
	})

	gptSpec := ProbeSpec{Model: "gpt-5.6-luna", Tier: "cheap", TokenParam: "max_completion_tokens", Limit: 16, Endpoint: "chat"}
	claudeSpec := ProbeSpec{Model: "claude-haiku-4-5", Tier: "medium", TokenParam: "max_tokens", Limit: 16, Endpoint: "chat"}

	assert.True(t, probeOnce(context.Background(), gptSpec, "probe-secret-key").Ok)
	assert.True(t, probeOnce(context.Background(), claudeSpec, "probe-secret-key").Ok)

	require.Len(t, bodies, 2)
	assert.Contains(t, bodies[0], `"max_completion_tokens":16`)
	assert.Contains(t, bodies[0], `"stream":true`)
	assert.Contains(t, bodies[0], `"model":"gpt-5.6-luna"`)
	assert.NotContains(t, bodies[0], `"max_tokens"`)
	assert.Contains(t, bodies[1], `"max_tokens":16`)
	assert.NotContains(t, bodies[1], `"max_completion_tokens"`)
	assert.Equal(t, "Bearer probe-secret-key", seenAuth)
}

func TestProbeOnceResponsesEndpoint(t *testing.T) {
	var body string
	stubProbeTransport(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		if !strings.HasSuffix(r.URL.Path, "/v1/responses") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	})

	spec := ProbeSpec{Model: "codex-auto-review", Tier: "medium", TokenParam: "max_completion_tokens", Limit: 16, Endpoint: "responses"}
	result := probeOnce(context.Background(), spec, "probe-secret-key")
	assert.True(t, result.Ok)
	assert.Contains(t, body, `"max_output_tokens":16`)
	assert.Contains(t, body, `"input":"ping"`)
}

func TestProbeOnceUpstreamErrorAndTimeout(t *testing.T) {
	stubProbeTransport(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "upstream exploded")
	})
	spec := ProbeSpec{Model: "m1", Tier: "cheap", TokenParam: "max_tokens", Limit: 16, Endpoint: "chat"}
	result := probeOnce(context.Background(), spec, "probe-secret-key")
	assert.False(t, result.Ok)
	assert.Equal(t, http.StatusInternalServerError, result.Status)

	// Timeout: server stalls longer than the (shortened) client timeout.
	server := stubProbeTransport(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	origClient := probeHTTPClient
	probeHTTPClient = &http.Client{Timeout: 100 * time.Millisecond}
	t.Cleanup(func() { probeHTTPClient = origClient })
	_ = server
	result = probeOnce(context.Background(), spec, "probe-secret-key")
	assert.False(t, result.Ok)
	assert.Zero(t, result.Status)
}
