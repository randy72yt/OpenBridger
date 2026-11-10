package controller

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
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
	perfSummaryCache.Store(key, perfCacheEntry{payload: payload, expireAt: time.Now().Add(perfSummaryCacheTTL)})
	return payload, nil
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
	hours := queryHours(c, 24, 24*30)
	publicV2 := perf_metrics_setting.GetSetting().PublicV2
	bars := 24
	if publicV2 {
		bars = 24
		if raw := c.Query("bars"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 1 {
				bars = parsed
			}
		}
		if bars > 72 {
			bars = 72
		}
	}

	activeGroups := activePricingGroups()
	cacheKey := fmt.Sprintf("summary:%d:%d:%v:%t", hours, bars, activeGroups, publicV2)
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
// budget, error codes). It is unaffected by public_v2.
func GetAdminPerfSummary(c *gin.Context) {
	hours := queryHours(c, 24, 168)
	activeGroups := activePricingGroups()

	summary, err := perfmetrics.QueryAdminSummary(hours, activeGroups)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	cfg := perf_metrics_setting.GetSetting()
	if left, budgetErr := service.ProbeBudgetLeft(cfg.ProbeTokenName, cfg.ProbeDailyBudgetQuota); budgetErr == nil {
		summary.ProbeBudgetQuota = int64(cfg.ProbeDailyBudgetQuota)
		summary.ProbeBudgetUsedQuota = int64(cfg.ProbeDailyBudgetQuota) - left
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
