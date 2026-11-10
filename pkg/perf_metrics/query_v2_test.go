package perfmetrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/perf_metrics_setting"
)

// TestModelSummaryV1JSONSnapshot locks the public v1 response shape: a summary
// built without v2 data must marshal byte-identically to the legacy format,
// with none of the v2-only keys present.
func TestModelSummaryV1JSONSnapshot(t *testing.T) {
	summary := ModelSummary{
		ModelName:           "gpt-5.6-luna",
		AvgLatencyMs:        812,
		SuccessRate:         98.68,
		AvgTps:              42.51,
		RecentSuccessSeries: []SuccessRatePoint{{Ts: 1794300000, SuccessRate: 100}, {Ts: 1794303600, SuccessRate: 97.5}},
		RequestCount:        76,
	}
	data, err := common.Marshal(summary)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"model_name": "gpt-5.6-luna",
		"avg_latency_ms": 812,
		"success_rate": 98.68,
		"avg_tps": 42.51,
		"recent_success_series": [
			{"ts": 1794300000, "success_rate": 100},
			{"ts": 1794303600, "success_rate": 97.5}
		]
	}`, string(data))
	// Byte-exact, not just key-equal: struct field order is stable.
	assert.Equal(t, `{"model_name":"gpt-5.6-luna","avg_latency_ms":812,"success_rate":98.68,"avg_tps":42.51,"recent_success_series":[{"ts":1794300000,"success_rate":100},{"ts":1794303600,"success_rate":97.5}]}`, string(data))
	for _, forbidden := range []string{"avg_ttft_ms", "availability_24h", "availability_7d", "sample_count", "probe_share", "insufficient", "channel", "base_url", "key"} {
		assert.NotContains(t, string(data), forbidden)
	}
}

func seedBucket(t *testing.T, modelName string, source string, age time.Duration, value counters) {
	t.Helper()
	ts := time.Now().Add(-age).Unix()
	ts -= ts % 3600
	if source == SourceReal {
		require.NoError(t, model.UpsertPerfMetric(&model.PerfMetric{
			ModelName: modelName, Group: "default", BucketTs: ts,
			RequestCount: value.requestCount, SuccessCount: value.successCount,
			TotalLatencyMs: value.totalLatencyMs, TtftSumMs: value.ttftSumMs, TtftCount: value.ttftCount,
			OutputTokens: value.outputTokens, GenerationMs: value.generationMs,
		}))
		return
	}
	require.NoError(t, model.UpsertPerfMetricAux(&model.PerfMetricAux{
		ModelName: modelName, Group: "default", Source: source, BucketTs: ts,
		RequestCount: value.requestCount, SuccessCount: value.successCount,
		TotalLatencyMs: value.totalLatencyMs, TtftSumMs: value.ttftSumMs, TtftCount: value.ttftCount,
		OutputTokens: value.outputTokens, GenerationMs: value.generationMs,
	}))
}

func TestQuerySummaryAllV2MergesRealAndProbe(t *testing.T) {
	setupPerfTestDB(t)
	resetHotBuckets(t)

	seedBucket(t, "m-v2", SourceReal, 30*time.Minute, counters{requestCount: 4, successCount: 3, totalLatencyMs: 4000, ttftSumMs: 600, ttftCount: 3})
	seedBucket(t, "m-v2", SourceProbe, 30*time.Minute, counters{requestCount: 6, successCount: 6, totalLatencyMs: 3000, ttftSumMs: 900, ttftCount: 6})
	seedBucket(t, "m-v2", SourceInternal, 30*time.Minute, counters{requestCount: 50, successCount: 0})
	seedBucket(t, "m-v2", SourceClientError, 30*time.Minute, counters{requestCount: 99, successCount: 0})

	result, err := QuerySummaryAllV2(24, 24, nil, nil)
	require.NoError(t, err)
	require.Len(t, result.Models, 1)
	summary := result.Models[0]

	assert.Equal(t, "m-v2", summary.ModelName)
	assert.Equal(t, int64(10), summary.SampleCount)    // real 4 + probe 6
	assert.Equal(t, int64(10), summary.RequestCount)   // internal/client_error excluded
	assert.InDelta(t, 90.0, summary.SuccessRate, 0.01) // (3+6)/10
	require.NotNil(t, summary.Availability24h)
	assert.InDelta(t, 90.0, *summary.Availability24h, 0.01)
	require.NotNil(t, summary.Availability7d)
	assert.InDelta(t, 90.0, *summary.Availability7d, 0.01)
	require.NotNil(t, summary.AvgTtftMs)
	assert.Equal(t, int64(166), *summary.AvgTtftMs) // (600+900)/(3+6) = 166 (integer avg)
	assert.InDelta(t, 0.6, summary.ProbeShare, 0.001)
	assert.False(t, summary.Insufficient) // 10 >= default min 5
	assert.Len(t, summary.RecentSuccessSeries, 1)
}

func TestQuerySummaryAllV2AllFailedAvailabilityIsZeroNotOmitted(t *testing.T) {
	setupPerfTestDB(t)
	resetHotBuckets(t)

	seedBucket(t, "m-down", SourceReal, 30*time.Minute, counters{requestCount: 7, successCount: 0})
	seedBucket(t, "m-down", SourceProbe, 30*time.Minute, counters{requestCount: 3, successCount: 0})

	result, err := QuerySummaryAllV2(24, 24, nil, nil)
	require.NoError(t, err)
	require.Len(t, result.Models, 1)
	summary := result.Models[0]
	require.NotNil(t, summary.Availability24h)
	assert.InDelta(t, 0.0, *summary.Availability24h, 0.001)
	assert.False(t, summary.Insufficient) // 10 samples >= default min 5
}

func TestQuerySummaryAllV2BarsWindowAndInsufficient(t *testing.T) {
	setupPerfTestDB(t)
	resetHotBuckets(t)
	withPerfSetting(t, func(s *perf_metrics_setting.PerfMetricsSetting) {
		s.MinPublicSamples = 100
	})

	// Sample inside the last 24h but outside a 1-bar series window.
	seedBucket(t, "m-old-series", SourceReal, 70*time.Hour, counters{requestCount: 5, successCount: 5})
	// Fresh probe sample, within every window.
	seedBucket(t, "m-old-series", SourceProbe, 30*time.Minute, counters{requestCount: 1, successCount: 1})

	result, err := QuerySummaryAllV2(24, 24, nil, nil)
	require.NoError(t, err)
	require.Len(t, result.Models, 1)
	summary := result.Models[0]
	// The 70h-old bucket is beyond both windows: it must not feed the series
	// (1 point) nor the 24h sample count (1 probe sample only).
	assert.Len(t, summary.RecentSuccessSeries, 1)
	assert.Equal(t, int64(1), summary.SampleCount)
	assert.True(t, summary.Insufficient) // 1 < 100
}

func TestQueryAdminSummarySourceSplit(t *testing.T) {
	db := setupPerfTestDB(t)
	resetHotBuckets(t)
	origLogDB := model.LOG_DB
	model.LOG_DB = db
	t.Cleanup(func() { model.LOG_DB = origLogDB })
	require.NoError(t, db.AutoMigrate(&model.Log{}))

	seedBucket(t, "m-admin", SourceReal, 30*time.Minute, counters{requestCount: 4, successCount: 3, totalLatencyMs: 4000, ttftSumMs: 600, ttftCount: 3})
	seedBucket(t, "m-admin", SourceProbe, 30*time.Minute, counters{requestCount: 6, successCount: 6, totalLatencyMs: 3000})
	seedBucket(t, "m-admin", SourceInternal, 30*time.Minute, counters{requestCount: 50, successCount: 50})
	seedBucket(t, "m-admin", SourceClientError, 30*time.Minute, counters{requestCount: 99, successCount: 0})

	require.NoError(t, db.Create(&model.Log{
		ModelName: "m-admin", Type: model.LogTypeError, CreatedAt: time.Now().Add(-10 * time.Minute).Unix(),
		Other: `{"error_code":"channel:invalid_key","status_code":401}`,
	}).Error)
	require.NoError(t, db.Create(&model.Log{
		ModelName: "m-admin", Type: model.LogTypeError, CreatedAt: time.Now().Add(-20 * time.Minute).Unix(),
		Other: `{"error_code":"","status_code":429}`,
	}).Error)
	require.NoError(t, db.Create(&model.Log{
		ModelName: "m-admin", Type: model.LogTypeConsume, CreatedAt: time.Now().Add(-5 * time.Minute).Unix(),
		Other: `{"error_code":"must_be_ignored"}`,
	}).Error)

	summary, err := QueryAdminSummary(24, nil)
	require.NoError(t, err)
	require.Len(t, summary.Models, 1)
	entry := summary.Models[0]

	assert.Equal(t, int64(4), entry.Real.Requests)
	assert.Equal(t, int64(3), entry.Real.Success)
	assert.Equal(t, int64(6), entry.Probe.Requests)
	assert.Equal(t, int64(50), entry.Internal.Requests)
	assert.Equal(t, int64(99), entry.ClientError.Requests)

	// Public v1 projection is real-only; v2 merges real+probe.
	assert.Equal(t, int64(4), entry.PublicV1.RequestCount)
	assert.Equal(t, int64(10), entry.PublicV2.SampleCount)
	assert.InDelta(t, 75.0, entry.PublicV1.SuccessRate, 0.01)
	assert.InDelta(t, 90.0, entry.PublicV2.SuccessRate, 0.01)
	assert.InDelta(t, 90.0, entry.Availability24h, 0.01)
	assert.Len(t, entry.Series, 1)

	require.Len(t, entry.ErrorCodes, 2)
	assert.Equal(t, int64(1), entry.ErrorCodes["channel:invalid_key"])
	assert.Equal(t, int64(1), entry.ErrorCodes["status:429"])

	// Admin JSON must not leak channel identity fields either.
	data, err := common.Marshal(entry)
	require.NoError(t, err)
	for _, forbidden := range []string{"base_url", "channel_name", `"key"`} {
		assert.NotContains(t, string(data), forbidden)
	}
}
