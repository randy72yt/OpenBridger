package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	perfmetrics "github.com/QuantumNous/new-api/pkg/perf_metrics"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/setting/perf_metrics_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupPerfControllerTest(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.PerfMetric{}, &model.PerfMetricAux{}, &model.Log{}))
	origDB, origLogDB := model.DB, model.LOG_DB
	model.DB = db
	// InitLogDB points LOG_DB at DB and initializes the dialect column-name
	// globals (commonGroupCol etc.) that group-filtered queries rely on.
	require.NoError(t, model.InitLogDB())
	t.Cleanup(func() { model.DB, model.LOG_DB = origDB, origLogDB })

	// Drop cached summaries between tests so each case rebuilds.
	perfSummaryCache.Range(func(key, _ any) bool {
		perfSummaryCache.Delete(key)
		return true
	})
	t.Cleanup(func() {
		perfSummaryCache.Range(func(key, _ any) bool {
			perfSummaryCache.Delete(key)
			return true
		})
	})
}

func withPublicV2(t *testing.T, enabled bool) {
	t.Helper()
	raw := config.GlobalConfig.Get("perf_metrics_setting")
	setting, ok := raw.(*perf_metrics_setting.PerfMetricsSetting)
	require.True(t, ok)
	orig := *setting
	setting.PublicV2 = enabled
	t.Cleanup(func() { *setting = orig })
}

func perfSummaryRequest(t *testing.T, query string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/perf-metrics/summary", GetPerfMetricsSummary)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/perf-metrics/summary"+query, nil)
	router.ServeHTTP(w, req)
	return w
}

func TestGetPerfMetricsSummaryV1ShapeUnchanged(t *testing.T) {
	setupPerfControllerTest(t)
	withPublicV2(t, false)

	require.NoError(t, model.UpsertPerfMetric(&model.PerfMetric{
		ModelName: "m-ctl", Group: "default", BucketTs: nowHourTs(),
		RequestCount: 2, SuccessCount: 2, TotalLatencyMs: 900, GenerationMs: 1000, OutputTokens: 50,
	}))
	require.NoError(t, model.UpsertPerfMetricAux(&model.PerfMetricAux{
		ModelName: "m-ctl", Group: "default", Source: perfmetrics.SourceProbe, BucketTs: nowHourTs() - 3600,
		RequestCount: 7, SuccessCount: 0,
	}))

	w := perfSummaryRequest(t, "")
	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"model_name":"m-ctl"`)
	for _, forbidden := range []string{"availability_24h", "avg_ttft_ms", "sample_count", "probe_share", "insufficient", "channel", "base_url", "\"key\""} {
		assert.NotContains(t, body, forbidden)
	}
	// Probe failures live in the aux table and must not reach the v1 numbers.
	assert.Contains(t, body, `"success_rate":100`)
	assert.Equal(t, "public, max-age=60", w.Header().Get("Cache-Control"))
}

func TestGetPerfMetricsSummaryV2AddsFields(t *testing.T) {
	setupPerfControllerTest(t)
	withPublicV2(t, true)

	require.NoError(t, model.UpsertPerfMetric(&model.PerfMetric{
		ModelName: "m-ctl", Group: "default", BucketTs: nowHourTs(),
		RequestCount: 4, SuccessCount: 3, TotalLatencyMs: 900, GenerationMs: 1000, OutputTokens: 50, TtftSumMs: 300, TtftCount: 3,
	}))
	require.NoError(t, model.UpsertPerfMetricAux(&model.PerfMetricAux{
		ModelName: "m-ctl", Group: "default", Source: perfmetrics.SourceProbe, BucketTs: nowHourTs(),
		RequestCount: 6, SuccessCount: 6, TotalLatencyMs: 600, TtftSumMs: 600, TtftCount: 6,
	}))

	w := perfSummaryRequest(t, "?bars=48")
	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"availability_24h":90`)
	assert.Contains(t, body, `"sample_count":10`)
	assert.Contains(t, body, `"avg_ttft_ms":100`)
	assert.Contains(t, body, `"probe_share":0.6`)

	// bars is clamped to <= 72.
	w = perfSummaryRequest(t, "?bars=999")
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"model_name":"m-ctl"`)
}

func TestGetAdminPerfSummaryRequiresAdmin(t *testing.T) {
	setupPerfControllerTest(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/perf-metrics/admin/summary", middleware.AdminAuth(), GetAdminPerfSummary)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/perf-metrics/admin/summary", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func nowHourTs() int64 {
	return time.Now().Unix() / 3600 * 3600
}
