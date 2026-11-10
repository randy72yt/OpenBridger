package perf_metrics_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

type PerfMetricsSetting struct {
	Enabled       bool   `json:"enabled"`
	FlushInterval int    `json:"flush_interval"`
	BucketTime    string `json:"bucket_time"`
	RetentionDays int    `json:"retention_days"`

	// PublicV2 gates the enriched public summary response. While false the
	// public endpoint keeps the exact legacy response shape and only reports
	// real traffic.
	PublicV2         bool `json:"public_v2"`
	MinPublicSamples int  `json:"min_public_samples"`

	// Source classification: requests from these tokens (and the playground)
	// are excluded from public numbers.
	InternalTokenNames string `json:"internal_token_names"`
	ProbeTokenName     string `json:"probe_token_name"`

	// Model probe scheduler. Disabled by default; when enabled the gateway
	// probes on-sale chat models that lack recent real traffic.
	ProbeEnabled              bool   `json:"probe_enabled"`
	ProbeScanMinutes          int    `json:"probe_scan_minutes"`
	ProbeMaxPerRun            int    `json:"probe_max_per_run"`
	ProbeDailyBudgetQuota     int    `json:"probe_daily_budget_quota"`
	ProbeRealTrafficThreshold int    `json:"probe_real_traffic_threshold"`
	ProbeModels               string `json:"probe_models"`
}

// defaultProbeModels seeds probe_models with the suggested tier assignment for
// the current on-sale catalog (plan section 4, phase 2). Lines for models that
// are no longer enabled are ignored, so this list can outlive catalog churn.
const defaultProbeModels = `# <model> <tier:cheap|medium|expensive|off> [token_param:max_tokens|max_completion_tokens|none] [limit] [endpoint:chat|responses]
gpt-5.6-luna            cheap     max_completion_tokens 16
gpt-6-luna              cheap     max_completion_tokens 16
gemini-2.5-flash-lite   cheap     max_tokens 16
gemini-3.1-flash-lite   cheap     max_tokens 16
gemini-2.5-flash        cheap     max_tokens 16
gemini-3-flash          cheap     max_tokens 16
gemini-3.5-flash        cheap     max_tokens 16
gemini-3.6-flash        cheap     max_tokens 16
gemini-3.7-flash        cheap     max_tokens 16
gemini-3.8-flash        cheap     max_tokens 16
deepseek-v4.1-flash     cheap     max_tokens 16
glm-5.2                 cheap     max_tokens 16
glm-5.3                 cheap     max_tokens 16
grok-4.5                cheap     max_tokens 16
grok-4.6                cheap     max_tokens 16
grok-4.7                cheap     max_tokens 16
claude-haiku-4-5        medium    max_tokens 16
claude-sonnet-4-6       medium    max_tokens 16
claude-sonnet-5         medium    max_tokens 16
claude-sonnet-5-5       medium    max_tokens 16
gpt-5.6                 medium    max_completion_tokens 16
gpt-5.6-terra           medium    max_completion_tokens 16
gpt-5.3-codex-spark     medium    max_completion_tokens 16
deepseek-v4-pro         medium    max_tokens 16
kimi-k3                 medium    max_tokens 16
gemini-2.5-pro          medium    max_tokens 16
gemini-3-pro            medium    max_tokens 16
gemini-3.1-pro          medium    max_tokens 16
gemini-3.1-pro-preview  medium    max_tokens 16
gemini-4-argon          medium    max_tokens 16
claude-opus-4-6         off
claude-opus-4-7         off
claude-opus-4-8         off
claude-opus-5           off
claude-opus-5-5         off
claude-fable-5          off
claude-fable-5-1        off
gpt-5.6-sol             off
gpt-6-sol               off
gpt-6.1-sol             off
gpt-6-astra             off
`

var perfMetricsSetting = PerfMetricsSetting{
	Enabled:       true,
	FlushInterval: 5,
	BucketTime:    "hour",
	// 90 days of history covers the 7-day availability windows with slack;
	// an existing perf_metrics_setting.retention_days option row still wins.
	RetentionDays: 90,

	PublicV2:         false,
	MinPublicSamples: 5,

	InternalTokenNames: "yuntao-main,ob-probe",
	ProbeTokenName:     "ob-probe",

	ProbeEnabled:              false,
	ProbeScanMinutes:          10,
	ProbeMaxPerRun:            6,
	ProbeDailyBudgetQuota:     500000,
	ProbeRealTrafficThreshold: 3,
	ProbeModels:               defaultProbeModels,
}

func init() {
	config.GlobalConfig.Register("perf_metrics_setting", &perfMetricsSetting)
}

func GetSetting() PerfMetricsSetting {
	return perfMetricsSetting
}

// InternalTokenSet parses InternalTokenNames into a lookup set of trimmed,
// non-empty token names.
func (s PerfMetricsSetting) InternalTokenSet() map[string]struct{} {
	set := make(map[string]struct{})
	for _, name := range strings.Split(s.InternalTokenNames, ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			set[name] = struct{}{}
		}
	}
	return set
}

func GetBucketSeconds() int64 {
	switch perfMetricsSetting.BucketTime {
	case "minute":
		return 60
	case "5min":
		return 300
	case "hour":
		return 3600
	default:
		return 3600
	}
}

func GetFlushIntervalMinutes() int {
	if perfMetricsSetting.FlushInterval < 1 {
		return 1
	}
	return perfMetricsSetting.FlushInterval
}

func GetProbeScanIntervalMinutes() int {
	if perfMetricsSetting.ProbeScanMinutes < 1 {
		return 10
	}
	return perfMetricsSetting.ProbeScanMinutes
}
