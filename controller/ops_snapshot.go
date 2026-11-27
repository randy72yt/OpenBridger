package controller

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	perfmetrics "github.com/QuantumNous/new-api/pkg/perf_metrics"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
)

const (
	opsSnapshotTokenEnv        = "OPS_SNAPSHOT_TOKEN"
	minOpsSnapshotTokenLen     = 32
	opsSnapshotCacheTTL        = 30 * time.Second
	defaultOpsTzOffsetMinutes  = 480 // Asia/Shanghai (UTC+8)
	lowChannelBalanceThreshold = 5.0 // USD
)

type OpsCurrencyRates struct {
	DisplayCurrency string  `json:"display_currency"`
	USDExchangeRate float64 `json:"usd_exchange_rate"`
	PriceRatio      float64 `json:"price_ratio"`
}

type OpsPeriodKPI struct {
	CashRevenueCNY          float64  `json:"cash_revenue_cny"`
	TopUpRevenueCNY         float64  `json:"topup_revenue_cny"`
	SubscriptionRevenueCNY  float64  `json:"subscription_revenue_cny"`
	ConsumptionCNY          float64  `json:"consumption_cny"`
	EstimatedCostCNY        float64  `json:"estimated_cost_cny"`
	EstimatedGrossProfitCNY float64  `json:"estimated_gross_profit_cny"`
	WalletLiabilityCNY      float64  `json:"wallet_liability_cny"`
	CashRevenueUSD          float64  `json:"cash_revenue_usd"`
	ConsumptionUSD          float64  `json:"consumption_usd"`
	EstimatedCostUSD        float64  `json:"estimated_cost_usd"`
	EstimatedGrossProfitUSD float64  `json:"estimated_gross_profit_usd"`
	EstimatedGrossMarginPct *float64 `json:"estimated_gross_margin_pct"`
	PaidOrderCount          int64    `json:"paid_order_count"`
	PayingUsers             int64    `json:"paying_users"`
	RedemptionCount         int64    `json:"redemption_count"`
	TotalRequests           int64    `json:"total_requests"`
	TotalTokens             int64    `json:"total_tokens"`
	TotalUsers              int64    `json:"total_users"`
	NewUsers                int64    `json:"new_users"`
	ActiveUsers             int64    `json:"active_users"`
}

type OpsDailyTrendPoint struct {
	Date           string  `json:"date"`
	CashRevenueCNY float64 `json:"cash_revenue_cny"`
	ConsumptionCNY float64 `json:"consumption_cny"`
	GrossProfitCNY float64 `json:"gross_profit_cny"`
	PaidOrders     int64   `json:"paid_orders"`
	NewUsers       int64   `json:"new_users"`
}

type OpsChannelsSummary struct {
	Total            int     `json:"total"`
	Enabled          int     `json:"enabled"`
	AutoDisabled     int     `json:"auto_disabled"`
	ManuallyDisabled int     `json:"manually_disabled"`
	LowBalanceCount  int     `json:"low_balance_count"`
	TotalBalanceUSD  float64 `json:"total_balance_usd"`
}

type OpsSnapshotData struct {
	GeneratedAt             int64                        `json:"generated_at"`
	Version                 string                       `json:"version"`
	StartTime               int64                        `json:"start_time"`
	UptimeSeconds           int64                        `json:"uptime_seconds"`
	DBHealthy               bool                         `json:"db_healthy"`
	PricingControlEnabled   bool                         `json:"pricing_control_enabled"`
	PendingPricingProposals int64                        `json:"pending_pricing_proposals"`
	Currency                OpsCurrencyRates             `json:"currency"`
	Today                   OpsPeriodKPI                 `json:"today"`
	Last7d                  OpsPeriodKPI                 `json:"last_7d"`
	DailyTrend7d            []OpsDailyTrendPoint         `json:"daily_trend_7d"`
	Channels                OpsChannelsSummary           `json:"channels"`
	Models                  perfmetrics.AdminPerfSummary `json:"models"`
}

var (
	opsSnapshotCacheMu sync.Mutex
	opsSnapshotCached  map[int]opsCachedSnapshot
)

type opsCachedSnapshot struct {
	data     OpsSnapshotData
	expireAt time.Time
}

func resetOpsSnapshotCacheForTest() {
	opsSnapshotCacheMu.Lock()
	defer opsSnapshotCacheMu.Unlock()
	opsSnapshotCached = nil
}

func roundTwoDecimals(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return math.Round(v*100) / 100
}

func verifyOpsSnapshotAuth(c *gin.Context) (int, string, string) {
	expected := strings.TrimSpace(os.Getenv(opsSnapshotTokenEnv))
	if len(expected) < minOpsSnapshotTokenLen {
		return http.StatusServiceUnavailable, "OPS_SNAPSHOT_DISABLED", "ops snapshot endpoint is disabled"
	}

	// Reject secrets passed via URL query string so tokens never leak into access logs or Referer headers.
	for _, qk := range []string{"token", "access_token", "key", "api_key"} {
		if _, present := c.GetQuery(qk); present {
			logger.LogWarn(c.Request.Context(), fmt.Sprintf("ops_snapshot auth rejected: reason=query_token_forbidden ip=%s", c.ClientIP()))
			return http.StatusUnauthorized, "OPS_QUERY_TOKEN_FORBIDDEN", "credentials must be sent via Authorization Bearer header"
		}
	}

	rawHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if rawHeader == "" {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("ops_snapshot auth rejected: reason=missing_header ip=%s", c.ClientIP()))
		return http.StatusUnauthorized, "OPS_UNAUTHORIZED", "missing Authorization Bearer token"
	}
	parts := strings.SplitN(rawHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("ops_snapshot auth rejected: reason=invalid_scheme ip=%s", c.ClientIP()))
		return http.StatusUnauthorized, "OPS_UNAUTHORIZED", "invalid Authorization scheme"
	}
	candidate := strings.TrimSpace(parts[1])
	if candidate == "" {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("ops_snapshot auth rejected: reason=empty_bearer ip=%s", c.ClientIP()))
		return http.StatusUnauthorized, "OPS_UNAUTHORIZED", "empty bearer token"
	}

	expectedDigest := sha256.Sum256([]byte(expected))
	candidateDigest := sha256.Sum256([]byte(candidate))
	if subtle.ConstantTimeCompare(expectedDigest[:], candidateDigest[:]) != 1 {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("ops_snapshot auth rejected: reason=token_mismatch ip=%s", c.ClientIP()))
		return http.StatusUnauthorized, "OPS_UNAUTHORIZED", "invalid ops snapshot token"
	}

	return 0, "", ""
}

func convertSummaryToOpsKPI(s service.BusinessDashboardSummary, priceRatio, usdExchangeRate float64) OpsPeriodKPI {
	return OpsPeriodKPI{
		CashRevenueCNY:          roundTwoDecimals(s.CashRevenueUSD * priceRatio),
		TopUpRevenueCNY:         roundTwoDecimals(s.TopUpRevenueUSD * priceRatio),
		SubscriptionRevenueCNY:  roundTwoDecimals(s.SubscriptionRevenueUSD * priceRatio),
		ConsumptionCNY:          roundTwoDecimals(s.ConsumptionUSD * usdExchangeRate),
		EstimatedCostCNY:        roundTwoDecimals(s.EstimatedCostUSD * usdExchangeRate),
		EstimatedGrossProfitCNY: roundTwoDecimals(s.EstimatedGrossProfitUSD * usdExchangeRate),
		WalletLiabilityCNY:      roundTwoDecimals(s.WalletLiabilityUSD * usdExchangeRate),
		CashRevenueUSD:          s.CashRevenueUSD,
		ConsumptionUSD:          s.ConsumptionUSD,
		EstimatedCostUSD:        s.EstimatedCostUSD,
		EstimatedGrossProfitUSD: s.EstimatedGrossProfitUSD,
		EstimatedGrossMarginPct: s.EstimatedGrossMarginPct,
		PaidOrderCount:          s.PaidOrderCount,
		PayingUsers:             s.PayingUsers,
		RedemptionCount:         s.RedemptionCount,
		TotalRequests:           s.TotalRequests,
		TotalTokens:             s.TotalTokens,
		TotalUsers:              s.TotalUsers,
		NewUsers:                s.NewUsers,
		ActiveUsers:             s.ActiveUsers,
	}
}

func parseOpsTzOffsetMinutes(raw string) int {
	if raw == "" {
		return defaultOpsTzOffsetMinutes
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val < -720 || val > 840 {
		return defaultOpsTzOffsetMinutes
	}
	return val
}

func GetOpsSnapshot(c *gin.Context) {
	if status, code, msg := verifyOpsSnapshotAuth(c); status != 0 {
		c.JSON(status, gin.H{
			"success": false,
			"code":    code,
			"message": msg,
		})
		return
	}

	tzOffsetMin := parseOpsTzOffsetMinutes(c.Query("tz_offset_minutes"))
	now := time.Now()

	opsSnapshotCacheMu.Lock()
	if opsSnapshotCached != nil {
		if entry, ok := opsSnapshotCached[tzOffsetMin]; ok && now.Before(entry.expireAt) {
			data := entry.data
			opsSnapshotCacheMu.Unlock()
			c.Header("Cache-Control", "no-store")
			c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
			return
		}
	}
	opsSnapshotCacheMu.Unlock()

	snapshot, err := computeOpsSnapshot(now, tzOffsetMin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    "OPS_SNAPSHOT_ERROR",
			"message": "failed to compute ops snapshot",
		})
		return
	}

	opsSnapshotCacheMu.Lock()
	if opsSnapshotCached == nil {
		opsSnapshotCached = make(map[int]opsCachedSnapshot, 4)
	}
	opsSnapshotCached[tzOffsetMin] = opsCachedSnapshot{
		data:     snapshot,
		expireAt: now.Add(opsSnapshotCacheTTL),
	}
	opsSnapshotCacheMu.Unlock()

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    snapshot,
	})
}

func computeOpsSnapshot(now time.Time, tzOffsetMin int) (OpsSnapshotData, error) {
	nowUnix := now.Unix()
	loc := time.FixedZone("ops", tzOffsetMin*60)
	localNow := now.In(loc)
	startOfToday := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc).Unix()
	startOf7d := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -6).Unix()

	priceRatio := operation_setting.Price
	if priceRatio <= 0 {
		if operation_setting.USDExchangeRate > 0 {
			priceRatio = operation_setting.USDExchangeRate
		} else {
			priceRatio = 7.2
		}
	}
	usdExchangeRate := operation_setting.USDExchangeRate
	if usdExchangeRate <= 0 {
		usdExchangeRate = priceRatio
	}

	todayDash, err := service.GetAdminBusinessDashboard(startOfToday, nowUnix)
	if err != nil {
		return OpsSnapshotData{}, err
	}
	last7dDash, err := service.GetAdminBusinessDashboard(startOf7d, nowUnix)
	if err != nil {
		return OpsSnapshotData{}, err
	}

	trend7d := make([]OpsDailyTrendPoint, 0, len(last7dDash.DailyTrend))
	for _, pt := range last7dDash.DailyTrend {
		profitUSD := pt.ConsumptionUSD - pt.EstimatedCostUSD
		trend7d = append(trend7d, OpsDailyTrendPoint{
			Date:           pt.Date,
			CashRevenueCNY: roundTwoDecimals(pt.CashRevenueUSD * priceRatio),
			ConsumptionCNY: roundTwoDecimals(pt.ConsumptionUSD * usdExchangeRate),
			GrossProfitCNY: roundTwoDecimals(profitUSD * usdExchangeRate),
			PaidOrders:     pt.PaidOrders,
			NewUsers:       pt.NewUsers,
		})
	}

	var chRows []struct {
		Status  int     `gorm:"column:status"`
		Balance float64 `gorm:"column:balance"`
	}
	if model.DB != nil {
		_ = model.DB.Model(&model.Channel{}).Select("status, balance").Find(&chRows).Error
	}
	var chSummary OpsChannelsSummary
	chSummary.Total = len(chRows)
	for _, ch := range chRows {
		switch ch.Status {
		case common.ChannelStatusEnabled:
			chSummary.Enabled++
			if ch.Balance > 0 && ch.Balance < lowChannelBalanceThreshold {
				chSummary.LowBalanceCount++
			}
		case common.ChannelStatusAutoDisabled:
			chSummary.AutoDisabled++
		case common.ChannelStatusManuallyDisabled:
			chSummary.ManuallyDisabled++
		}
		if ch.Balance > 0 && !math.IsNaN(ch.Balance) && !math.IsInf(ch.Balance, 0) {
			chSummary.TotalBalanceUSD += ch.Balance
		}
	}
	chSummary.TotalBalanceUSD = roundTwoDecimals(chSummary.TotalBalanceUSD)

	var pendingProposals int64
	if model.DB != nil && model.DB.Migrator().HasTable(&model.ModelPriceProposal{}) {
		_ = model.DB.Model(&model.ModelPriceProposal{}).
			Where("status = ?", model.PricingProposalPending).
			Count(&pendingProposals).Error
	}

	modelsPerf, perfErr := buildAdminPerfSummary(24)
	if perfErr != nil {
		modelsPerf = perfmetrics.AdminPerfSummary{
			Models:      []perfmetrics.AdminModelPerf{},
			GeneratedAt: nowUnix,
		}
	}

	uptime := int64(0)
	if common.StartTime > 0 && nowUnix >= common.StartTime {
		uptime = nowUnix - common.StartTime
	}

	return OpsSnapshotData{
		GeneratedAt:             nowUnix,
		Version:                 common.Version,
		StartTime:               common.StartTime,
		UptimeSeconds:           uptime,
		DBHealthy:               model.PingDB() == nil,
		PricingControlEnabled:   service.PricingControlEnabled(),
		PendingPricingProposals: pendingProposals,
		Currency: OpsCurrencyRates{
			DisplayCurrency: "CNY",
			USDExchangeRate: usdExchangeRate,
			PriceRatio:      priceRatio,
		},
		Today:        convertSummaryToOpsKPI(todayDash.Summary, priceRatio, usdExchangeRate),
		Last7d:       convertSummaryToOpsKPI(last7dDash.Summary, priceRatio, usdExchangeRate),
		DailyTrend7d: trend7d,
		Channels:     chSummary,
		Models:       modelsPerf,
	}, nil
}
