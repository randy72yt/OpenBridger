package perf_metrics_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/QuantumNous/new-api/setting/config"
)

// legacyPerfMetricsSetting mirrors the pre-aux option set. Loading the option
// rows persisted by the new version into it must not error (rollback safety:
// the old binary ignores keys it has no field for).
type legacyPerfMetricsSetting struct {
	Enabled       bool   `json:"enabled"`
	FlushInterval int    `json:"flush_interval"`
	BucketTime    string `json:"bucket_time"`
	RetentionDays int    `json:"retention_days"`
}

func newOptionRows() map[string]string {
	return map[string]string{
		"perf_metrics_setting.enabled":                "true",
		"perf_metrics_setting.flush_interval":         "7",
		"perf_metrics_setting.bucket_time":            "5min",
		"perf_metrics_setting.retention_days":         "90",
		"perf_metrics_setting.public_v2":              "true",
		"perf_metrics_setting.min_public_samples":     "5",
		"perf_metrics_setting.internal_token_names":   "yuntao-main,ob-probe",
		"perf_metrics_setting.probe_token_name":       "ob-probe",
		"perf_metrics_setting.probe_enabled":          "true",
		"perf_metrics_setting.probe_scan_minutes":     "15",
		"perf_metrics_setting.probe_max_per_run":      "4",
		"perf_metrics_setting.probe_models":           "gpt-x cheap max_tokens 16",
		"perf_metrics_setting.some_future_option_key": "unknown-to-this-binary",
	}
}

func TestLoadFromDBLegacyStructIgnoresNewKeys(t *testing.T) {
	cm := config.NewConfigManager()
	legacy := &legacyPerfMetricsSetting{Enabled: true, FlushInterval: 5, BucketTime: "hour", RetentionDays: 0}
	cm.Register("perf_metrics_setting", legacy)

	require.NoError(t, cm.LoadFromDB(newOptionRows()))
	assert.True(t, legacy.Enabled)
	assert.Equal(t, 7, legacy.FlushInterval)
	assert.Equal(t, "5min", legacy.BucketTime)
	assert.Equal(t, 90, legacy.RetentionDays)
}

func TestLoadFromDBNewStructConsumesNewKeys(t *testing.T) {
	cm := config.NewConfigManager()
	setting := &PerfMetricsSetting{}
	cm.Register("perf_metrics_setting", setting)

	require.NoError(t, cm.LoadFromDB(newOptionRows()))
	assert.True(t, setting.Enabled)
	assert.True(t, setting.PublicV2)
	assert.True(t, setting.ProbeEnabled)
	assert.Equal(t, 15, setting.ProbeScanMinutes)
	assert.Equal(t, 4, setting.ProbeMaxPerRun)
	assert.Equal(t, "yuntao-main,ob-probe", setting.InternalTokenNames)
	assert.Equal(t, "gpt-x cheap max_tokens 16", setting.ProbeModels)
}

func TestInternalTokenSet(t *testing.T) {
	setting := PerfMetricsSetting{InternalTokenNames: " yuntao-main , ob-probe ,,"}
	set := setting.InternalTokenSet()
	assert.Len(t, set, 2)
	_, hasInternal := set["yuntao-main"]
	assert.True(t, hasInternal)
	_, hasProbe := set["ob-probe"]
	assert.True(t, hasProbe)
}
