package perfmetrics

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/setting/perf_metrics_setting"
)

func setupPerfTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.PerfMetric{}, &model.PerfMetricAux{}))
	origDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = origDB })
	return db
}

func resetHotBuckets(t *testing.T) {
	t.Helper()
	drainHotBuckets()
	t.Cleanup(drainHotBuckets)
}

func drainHotBuckets() {
	hotBuckets.Range(func(key, _ any) bool {
		hotBuckets.Delete(key)
		return true
	})
}

func stubTokenNames(t *testing.T, names map[int]string) {
	t.Helper()
	orig := tokenNameResolver
	tokenNameResolver = func(id int) string { return names[id] }
	t.Cleanup(func() { tokenNameResolver = orig })
}

func hotBucketSources() map[string]bool {
	sources := map[string]bool{}
	hotBuckets.Range(func(key, _ any) bool {
		sources[key.(bucketKey).source] = true
		return true
	})
	return sources
}

func TestQuerySummaryAllOnlyCountsRealSamples(t *testing.T) {
	setupPerfTestDB(t)
	resetHotBuckets(t)

	Record(Sample{Model: "m1", Group: "default", LatencyMs: 100, Success: true, OutputTokens: 10, GenerationMs: 500, Source: SourceReal})
	Record(Sample{Model: "m1", Group: "default", LatencyMs: 200, Success: false, Source: SourceReal})
	Record(Sample{Model: "m1", Group: "default", LatencyMs: 50, Success: true, Source: SourceInternal})
	Record(Sample{Model: "m1", Group: "default", LatencyMs: 50, Success: false, Source: SourceProbe})
	Record(Sample{Model: "m1", Group: "default", LatencyMs: 50, Success: false, Source: SourceClientError})

	result, err := QuerySummaryAll(24, nil)
	require.NoError(t, err)
	require.Len(t, result.Models, 1)
	assert.Equal(t, "m1", result.Models[0].ModelName)
	assert.Equal(t, int64(2), result.Models[0].RequestCount)
	assert.InDelta(t, 50.0, result.Models[0].SuccessRate, 0.01)
}

func TestRecordRelayOutcomeRoutesSources(t *testing.T) {
	setupPerfTestDB(t)
	resetHotBuckets(t)
	stubTokenNames(t, map[int]string{7: "yuntao-main", 9: "ob-probe"})

	mkInfo := func(tokenId int, playground bool) *relaycommon.RelayInfo {
		return &relaycommon.RelayInfo{
			TokenId:         tokenId,
			IsPlayground:    playground,
			OriginModelName: "m1",
			UsingGroup:      "default",
			StartTime:       time.Now(),
		}
	}

	RecordRelayOutcome(mkInfo(7, false), OutcomeSuccess, 3, "")                               // internal token
	RecordRelayOutcome(mkInfo(9, false), OutcomeSuccess, 3, "")                               // probe token
	RecordRelayOutcome(mkInfo(12, true), OutcomeSuccess, 3, "")                               // playground
	RecordRelayOutcome(mkInfo(11, false), OutcomeClientExcluded, 0, "invalid_request")        // real + excluded -> client_error
	RecordRelayOutcome(mkInfo(11, false), OutcomeSuccess, 3, "")                              // real
	RecordRelayOutcome(mkInfo(7, false), OutcomeClientExcluded, 0, "insufficient_user_quota") // internal + excluded stays internal

	assert.Equal(t, map[string]bool{
		SourceReal:        true,
		SourceInternal:    true,
		SourceProbe:       true,
		SourceClientError: true,
	}, hotBucketSources())

	result, err := QuerySummaryAll(24, nil)
	require.NoError(t, err)
	require.Len(t, result.Models, 1)
	assert.Equal(t, int64(1), result.Models[0].RequestCount)
	assert.InDelta(t, 100.0, result.Models[0].SuccessRate, 0.01)
}

func TestQueryDetailExcludesNonRealHotBuckets(t *testing.T) {
	setupPerfTestDB(t)
	resetHotBuckets(t)

	Record(Sample{Model: "m1", Group: "default", LatencyMs: 100, Success: true, Source: SourceReal})
	Record(Sample{Model: "m1", Group: "default", LatencyMs: 100, Success: false, Source: SourceProbe})
	Record(Sample{Model: "m1", Group: "default", LatencyMs: 100, Success: false, Source: SourceProbe})

	// No group filter: commonGroupCol is only initialized by model.InitDB,
	// which unit tests do not run.
	result, err := Query(QueryParams{Model: "m1", Hours: 24})
	require.NoError(t, err)
	require.Len(t, result.Groups, 1)
	// Only the single real request counts; probe failures must not drag the rate down.
	assert.InDelta(t, 100.0, result.Groups[0].SuccessRate, 0.01)
}

func TestFlushRoutesAndAccumulatesBySource(t *testing.T) {
	db := setupPerfTestDB(t)
	resetHotBuckets(t)

	past := bucketStart(time.Now().Add(-2 * time.Hour).Unix())
	storeHotBucket := func(source string) {
		bucket := &atomicBucket{}
		bucket.add(Sample{Model: "m1", Group: "default", LatencyMs: 100, Success: source == SourceReal, Source: source})
		hotBuckets.Store(bucketKey{model: "m1", group: "default", source: source, bucketTs: past}, bucket)
	}

	storeHotBucket(SourceReal)
	storeHotBucket(SourceProbe)
	flushCompletedBuckets()

	var perfRows []model.PerfMetric
	require.NoError(t, db.Find(&perfRows).Error)
	require.Len(t, perfRows, 1)
	assert.Equal(t, "m1", perfRows[0].ModelName)
	assert.Equal(t, int64(1), perfRows[0].RequestCount)
	assert.Equal(t, int64(1), perfRows[0].SuccessCount)

	var auxRows []model.PerfMetricAux
	require.NoError(t, db.Find(&auxRows).Error)
	require.Len(t, auxRows, 1)
	assert.Equal(t, SourceProbe, auxRows[0].Source)
	assert.Equal(t, int64(1), auxRows[0].RequestCount)
	assert.Equal(t, int64(0), auxRows[0].SuccessCount)

	// Flushing the same bucket again accumulates instead of duplicating rows.
	storeHotBucket(SourceProbe)
	flushCompletedBuckets()
	require.NoError(t, db.Find(&auxRows).Error)
	require.Len(t, auxRows, 1)
	assert.Equal(t, int64(2), auxRows[0].RequestCount)
}

func TestCleanupExpiredMetricsDeletesBothTables(t *testing.T) {
	db := setupPerfTestDB(t)
	lastMetricsCleanup = 0
	t.Cleanup(func() { lastMetricsCleanup = 0 })

	oldTs := time.Now().Add(-48 * time.Hour).Unix()
	freshTs := time.Now().Add(-1 * time.Hour).Unix()
	require.NoError(t, db.Create(&model.PerfMetric{ModelName: "old", Group: "default", BucketTs: oldTs, RequestCount: 1, SuccessCount: 1}).Error)
	require.NoError(t, db.Create(&model.PerfMetric{ModelName: "fresh", Group: "default", BucketTs: freshTs, RequestCount: 1, SuccessCount: 1}).Error)
	require.NoError(t, db.Create(&model.PerfMetricAux{ModelName: "old", Group: "default", Source: SourceProbe, BucketTs: oldTs, RequestCount: 1, SuccessCount: 0}).Error)
	require.NoError(t, db.Create(&model.PerfMetricAux{ModelName: "fresh", Group: "default", Source: SourceProbe, BucketTs: freshTs, RequestCount: 1, SuccessCount: 0}).Error)

	cleanupExpiredMetrics(1)

	var perfCount, auxCount int64
	require.NoError(t, db.Model(&model.PerfMetric{}).Where("model_name = ?", "old").Count(&perfCount).Error)
	assert.Zero(t, perfCount)
	require.NoError(t, db.Model(&model.PerfMetric{}).Where("model_name = ?", "fresh").Count(&perfCount).Error)
	assert.Equal(t, int64(1), perfCount)
	require.NoError(t, db.Model(&model.PerfMetricAux{}).Where("model_name = ?", "old").Count(&auxCount).Error)
	assert.Zero(t, auxCount)
	require.NoError(t, db.Model(&model.PerfMetricAux{}).Where("model_name = ?", "fresh").Count(&auxCount).Error)
	assert.Equal(t, int64(1), auxCount)
}

func TestCleanupExpiredMetricsRunsAtMostOncePerDay(t *testing.T) {
	db := setupPerfTestDB(t)
	lastMetricsCleanup = 0
	t.Cleanup(func() { lastMetricsCleanup = 0 })

	cleanupExpiredMetrics(1)
	require.NotZero(t, lastMetricsCleanup)

	// Rows that expired after the first sweep survive the throttled pass.
	oldTs := time.Now().Add(-48 * time.Hour).Unix()
	require.NoError(t, db.Create(&model.PerfMetric{ModelName: "late-old", Group: "default", BucketTs: oldTs, RequestCount: 1, SuccessCount: 1}).Error)
	cleanupExpiredMetrics(1)

	var count int64
	require.NoError(t, db.Model(&model.PerfMetric{}).Where("model_name = ?", "late-old").Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

// withPerfSetting mutates the live registered setting for one test and
// restores it afterwards.
func withPerfSetting(t *testing.T, mutate func(*perf_metrics_setting.PerfMetricsSetting)) {
	t.Helper()
	raw := config.GlobalConfig.Get("perf_metrics_setting")
	setting, ok := raw.(*perf_metrics_setting.PerfMetricsSetting)
	require.True(t, ok)
	orig := *setting
	mutate(setting)
	t.Cleanup(func() { *setting = orig })
}
