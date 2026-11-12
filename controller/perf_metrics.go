package controller

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/model"
	perfmetrics "github.com/QuantumNous/new-api/pkg/perf_metrics"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/perf_metrics_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

// perfSummaryCacheTTL bounds staleness of the public summary (and the admin
// 7d availability figures reuse the same idea); clients additionally get a
// Cache-Control header so the model square does not hammer the endpoint.
const perfSummaryCacheTTL = 60 * time.Second

// perfSummaryCacheMaxEntries caps the in-process cache size; the eviction
// pass enforces it so hostile query combinations cannot grow the map.
const perfSummaryCacheMaxEntries = 64

type perfCacheEntry struct {
	payload  any
	expireAt time.Time
}

var perfSummaryCache sync.Map

func cachedPerfPayload(key string, build func() (any, error)) (any, error) {
	if cached, ok := perfSummaryCache.Load(key); ok {
		entry := cached.(perfCacheEntry)
		if time.Now().Before(entry.expireAt) {
			return entry.payload, nil
		}
	}
	payload, err := build()
	if err != nil {
		return nil, err
	}
	storePerfCacheEntry(key, payload)
	return payload, nil
}

func storePerfCacheEntry(key string, payload any) {
	now := time.Now()
	perfSummaryCache.Store(key, perfCacheEntry{payload: payload, expireAt: now.Add(perfSummaryCacheTTL)})
	evictPerfCache(now)
}

// evictPerfCache sweeps expired entries and, if the map is still over the cap,
// drops the soonest-to-expire (oldest) entries until it is back at the cap.
func evictPerfCache(now time.Time) {
	type liveEntry struct {
		key      string
		expireAt time.Time
	}
	live := make([]liveEntry, 0, perfSummaryCacheMaxEntries+1)
	perfSummaryCache.Range(func(k, v any) bool {
		entry := v.(perfCacheEntry)
		if now.After(entry.expireAt) {
			perfSummaryCache.Delete(k)
			return true
		}
		live = append(live, liveEntry{key: k.(string), expireAt: entry.expireAt})
		return true
	})
	if len(live) <= perfSummaryCacheMaxEntries {
		return
	}
	sort.Slice(live, func(i, j int) bool { return live[i].expireAt.Before(live[j].expireAt) })
	for _, entry := range live[:len(live)-perfSummaryCacheMaxEntries] {
		perfSummaryCache.Delete(entry.key)
	}
}

// summaryCacheKey builds the cache key with sorted groups so request order
// cannot multiply entries.
func summaryCacheKey(prefix string, hours int, bars int, groups []string, publicV2 bool) string {
	sorted := append([]string(nil), groups...)
	sort.Strings(sorted)
	return fmt.Sprintf("%s:%d:%d:%s:%t", prefix, hours, bars, strings.Join(sorted, ","), publicV2)
}

// normalizeSummaryHours whitelists the supported windows; anything else
// (including junk) falls back to the default instead of producing a fresh
// cache key per value.
func normalizeSummaryHours(raw string) int {
	if raw == "168" {
		return 168
	}
	return 24
}

func normalizeSummaryBars(raw string) int {
	switch raw {
	case "48":
		return 48
	case "72":
		return 72
	default:
		return 24
	}
}

func activePricingGroups() []string {
	return append(lo.Keys(ratio_setting.GetGroupRatioCopy()), "auto")
}

func queryHours(c *gin.Context, fallback int, max int) int {
	hours := fallback
	if raw := c.Query("hours"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			hours = parsed
		}
	}
	if hours < 1 {
		hours = fallback
	}
	if hours > max {
		hours = max
	}
	return hours
}

func GetPerfMetricsSummary(c *gin.Context) {
	hours := normalizeSummaryHours(c.Query("hours"))
	publicV2 := perf_metrics_setting.GetSetting().PublicV2
	bars := 24
	if publicV2 {
		bars = normalizeSummaryBars(c.Query("bars"))
	}

	activeGroups := activePricingGroups()
	cacheKey := summaryCacheKey("summary", hours, bars, activeGroups, publicV2)
	payload, err := cachedPerfPayload(cacheKey, func() (any, error) {
		if publicV2 {
			return perfmetrics.QuerySummaryAllV2(hours, bars, activeGroups, []string{perfmetrics.SourceReal, perfmetrics.SourceProbe})
		}
		return perfmetrics.QuerySummaryAll(hours, activeGroups)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    payload,
	})
}

// GetAdminPerfSummary serves the admin-only full view (all four sources, probe
// budget, error codes). It is unaffected by public_v2. The aggregation result
// is cached for 60s like the public summary; live enrichment (probe budget,
// last probe run) runs on a per-request copy so the cached payload is never
// mutated concurrently.
func GetAdminPerfSummary(c *gin.Context) {
	hours := normalizeSummaryHours(c.Query("hours"))
	activeGroups := activePricingGroups()

	payload, err := cachedPerfPayload(summaryCacheKey("admin", hours, 0, activeGroups, false), func() (any, error) {
		return perfmetrics.QueryAdminSummary(hours, activeGroups)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	cachedSummary := payload.(perfmetrics.AdminPerfSummary)
	summary := cachedSummary
	summary.Models = append([]perfmetrics.AdminModelPerf(nil), cachedSummary.Models...)

	cfg := perf_metrics_setting.GetSetting()
	summary.ProbeBudgetQuota = int64(cfg.ProbeDailyBudgetQuota)
	if probeToken, err := service.ResolveProbeToken(cfg.ProbeTokenName); err == nil {
		if left, budgetErr := service.ProbeBudgetLeft(probeToken.Id, cfg.ProbeDailyBudgetQuota); budgetErr == nil {
			summary.ProbeBudgetUsedQuota = int64(cfg.ProbeDailyBudgetQuota) - left
		}
	}

	specByModel := map[string]service.ProbeSpec{}
	for _, spec := range service.LoadProbeSpecs(model.GetEnabledModels(), model.GetPricing(), cfg) {
		specByModel[spec.Model] = spec
	}
	seen := map[string]struct{}{}
	for i := range summary.Models {
		entry := &summary.Models[i]
		seen[entry.ModelName] = struct{}{}
		entry.LastProbeAt = service.GetProbeLastRunAt(entry.ModelName)
		if spec, ok := specByModel[entry.ModelName]; ok {
			entry.ProbeTier = spec.Tier
			entry.ProbeTokenParam = spec.TokenParam
		}
	}
	// Models without any samples still matter for the observation checklist
	// (probe coverage), so list them with zero counters.
	specNames := make([]string, 0, len(specByModel))
	for name := range specByModel {
		specNames = append(specNames, name)
	}
	sort.Strings(specNames)
	for _, name := range specNames {
		if _, ok := seen[name]; !ok {
			summary.Models = append(summary.Models, perfmetrics.AdminModelPerf{
				ModelName:       name,
				ErrorCodes:      map[string]int64{},
				ProbeTier:       specByModel[name].Tier,
				ProbeTokenParam: specByModel[name].TokenParam,
				LastProbeAt:     service.GetProbeLastRunAt(name),
			})
		}
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

func GetPerfMetrics(c *gin.Context) {
	modelName := c.Query("model")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "model is required",
		})
		return
	}

	result, err := perfmetrics.Query(perfmetrics.QueryParams{
		Model: modelName,
		Group: c.Query("group"),
		Hours: queryHours(c, 24, 24*30),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	result.Groups = filterActiveGroups(result.Groups)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

func filterActiveGroups(groups []perfmetrics.GroupResult) []perfmetrics.GroupResult {
	activeRatios := ratio_setting.GetGroupRatioCopy()
	return lo.Filter(groups, func(g perfmetrics.GroupResult, _ int) bool {
		_, ok := activeRatios[g.Group]
		return ok || g.Group == "auto"
	})
}
