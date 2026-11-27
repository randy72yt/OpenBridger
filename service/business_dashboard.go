package service

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"gorm.io/gorm"
)

const (
	MaxBusinessDashboardRangeSeconds int64 = 366 * 86400
	maxDirectDailyLogScan            int64 = 100000
)

type BusinessDashboardRates struct {
	QuotaPerUnit    float64
	PriceRatio      float64
	USDExchangeRate float64
	NowTimestamp    int64
}

type BusinessDashboardSummary struct {
	CashRevenueUSD          float64  `json:"cash_revenue_usd"`
	TopUpRevenueUSD         float64  `json:"topup_revenue_usd"`
	SubscriptionRevenueUSD  float64  `json:"subscription_revenue_usd"`
	PaidOrderCount          int64    `json:"paid_order_count"`
	PayingUsers             int64    `json:"paying_users"`
	ARPPUUSD                float64  `json:"arppu_usd"`
	RedemptionUSD           float64  `json:"redemption_usd"`
	RedemptionCount         int64    `json:"redemption_count"`
	ConsumptionUSD          float64  `json:"consumption_usd"`
	CostedConsumptionUSD    float64  `json:"costed_consumption_usd"`
	UncostedConsumptionUSD  float64  `json:"uncosted_consumption_usd"`
	EstimatedCostUSD        float64  `json:"estimated_cost_usd"`
	EstimatedGrossProfitUSD float64  `json:"estimated_gross_profit_usd"`
	EstimatedGrossMarginPct *float64 `json:"estimated_gross_margin_pct"`
	CostCoveragePct         float64  `json:"cost_coverage_pct"`
	TotalRequests           int64    `json:"total_requests"`
	TotalTokens             int64    `json:"total_tokens"`
	TotalUsers              int64    `json:"total_users"`
	NewUsers                int64    `json:"new_users"`
	ActiveUsers             int64    `json:"active_users"`
	WalletLiabilityUSD      float64  `json:"wallet_liability_usd"`
	ChannelBalanceUSD       float64  `json:"channel_balance_usd"`
}

type BusinessDailyPoint struct {
	Date             string  `json:"date"`
	Timestamp        int64   `json:"timestamp"`
	CashRevenueUSD   float64 `json:"cash_revenue_usd"`
	RedemptionUSD    float64 `json:"redemption_usd"`
	ConsumptionUSD   float64 `json:"consumption_usd"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
	NewUsers         int64   `json:"new_users"`
	PaidOrders       int64   `json:"paid_orders"`
}

type PaymentChannelStat struct {
	Channel          string  `json:"channel"`
	Provider         string  `json:"provider"`
	OrderType        string  `json:"order_type"`
	OrderCount       int64   `json:"order_count"`
	OriginalCurrency string  `json:"original_currency"`
	OriginalAmount   float64 `json:"original_amount"`
	RevenueUSD       float64 `json:"revenue_usd"`
}

type ModelMarginStat struct {
	ModelName        string   `json:"model_name"`
	ChannelID        int      `json:"channel_id"`
	ChannelName      string   `json:"channel_name"`
	Requests         int64    `json:"requests"`
	PromptTokens     int64    `json:"prompt_tokens"`
	CompletionTokens int64    `json:"completion_tokens"`
	ConsumptionUSD   float64  `json:"consumption_usd"`
	EstimatedCostUSD *float64 `json:"estimated_cost_usd"`
	GrossProfitUSD   *float64 `json:"gross_profit_usd"`
	GrossMarginPct   *float64 `json:"gross_margin_pct"`
	CostCovered      bool     `json:"cost_covered"`
}

type BusinessDashboardData struct {
	Summary         BusinessDashboardSummary `json:"summary"`
	DailyTrend      []BusinessDailyPoint     `json:"daily_trend"`
	PaymentChannels []PaymentChannelStat     `json:"payment_channels"`
	ModelMargins    []ModelMarginStat        `json:"model_margins"`
}

type resolvedOfferUnitCost struct {
	inputUSDPerMillion  float64
	outputUSDPerMillion float64
}

func roundMoney(val float64) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0
	}
	return math.Round(val*10000) / 10000
}

func roundPct(val float64) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0
	}
	return math.Round(val*100) / 100
}

func utcDayBucket(ts int64) int64 {
	if ts <= 0 {
		return 0
	}
	return ts - (ts % 86400)
}

func formatUTCDay(dayTs int64) string {
	return time.Unix(dayTs, 0).UTC().Format("2006-01-02")
}

func GetAdminBusinessDashboard(startTimestamp int64, endTimestamp int64) (*BusinessDashboardData, error) {
	mainDB := model.DB
	logDB := model.LOG_DB
	if logDB == nil {
		logDB = mainDB
	}
	return ComputeBusinessDashboard(mainDB, logDB, startTimestamp, endTimestamp, BusinessDashboardRates{
		QuotaPerUnit:    common.QuotaPerUnit,
		PriceRatio:      operation_setting.Price,
		USDExchangeRate: operation_setting.USDExchangeRate,
		NowTimestamp:    common.GetTimestamp(),
	})
}

func ComputeBusinessDashboard(
	db *gorm.DB,
	logDB *gorm.DB,
	startTimestamp int64,
	endTimestamp int64,
	rates BusinessDashboardRates,
) (*BusinessDashboardData, error) {
	if logDB == nil {
		logDB = db
	}
	quotaPerUnit := rates.QuotaPerUnit
	if quotaPerUnit <= 0 {
		quotaPerUnit = 500000
	}
	priceRatio := rates.PriceRatio
	if priceRatio <= 0 {
		if rates.USDExchangeRate > 0 {
			priceRatio = rates.USDExchangeRate
		} else {
			priceRatio = 7.2
		}
	}
	usdExchangeRate := rates.USDExchangeRate
	if usdExchangeRate <= 0 {
		usdExchangeRate = priceRatio
	}
	nowTs := rates.NowTimestamp
	if nowTs <= 0 {
		nowTs = time.Now().Unix()
	}

	// Build contiguous daily timeline buckets from startTimestamp to endTimestamp (UTC days).
	startDay := utcDayBucket(startTimestamp)
	endDay := utcDayBucket(endTimestamp)
	dailyMap := make(map[int64]*BusinessDailyPoint)
	dayKeys := make([]int64, 0, int((endDay-startDay)/86400)+1)
	for d := startDay; d <= endDay; d += 86400 {
		pt := &BusinessDailyPoint{
			Date:      formatUTCDay(d),
			Timestamp: d,
		}
		dailyMap[d] = pt
		dayKeys = append(dayKeys, d)
	}
	ensureDay := func(ts int64) *BusinessDailyPoint {
		d := utcDayBucket(ts)
		if pt, ok := dailyMap[d]; ok {
			return pt
		}
		if len(dayKeys) > 0 {
			return dailyMap[dayKeys[len(dayKeys)-1]]
		}
		return nil
	}

	// 1. User totals, wallet liability, and new registrations
	var userSnapshot struct {
		TotalUsers     int64 `gorm:"column:total_users"`
		TotalQuotaLeft int64 `gorm:"column:total_quota_left"`
	}
	if err := db.Model(&model.User{}).
		Select("COUNT(*) AS total_users, COALESCE(SUM(quota), 0) AS total_quota_left").
		Scan(&userSnapshot).Error; err != nil {
		return nil, err
	}

	var newUserCreatedTimes []int64
	if err := db.Model(&model.User{}).
		Where("created_at >= ? AND created_at <= ?", startTimestamp, endTimestamp).
		Pluck("created_at", &newUserCreatedTimes).Error; err != nil {
		return nil, err
	}
	for _, ts := range newUserCreatedTimes {
		if pt := ensureDay(ts); pt != nil {
			pt.NewUsers++
		}
	}

	// 2. Upstream channel balances & names
	var channelRows []struct {
		ID      int     `gorm:"column:id"`
		Name    string  `gorm:"column:name"`
		Balance float64 `gorm:"column:balance"`
	}
	if err := db.Model(&model.Channel{}).
		Select("id, name, balance").
		Find(&channelRows).Error; err != nil {
		return nil, err
	}
	channelNameByID := make(map[int]string, len(channelRows))
	var totalChannelBalanceUSD float64
	for _, ch := range channelRows {
		channelNameByID[ch.ID] = ch.Name
		if ch.Balance > 0 && !math.IsNaN(ch.Balance) && !math.IsInf(ch.Balance, 0) {
			totalChannelBalanceUSD += ch.Balance
		}
	}

	// 3. Cash Top-Ups + Cash Subscriptions + Redemptions
	var topUpRows []struct {
		UserID          int     `gorm:"column:user_id"`
		Amount          int64   `gorm:"column:amount"`
		Money           float64 `gorm:"column:money"`
		CreditQuota     int64   `gorm:"column:credit_quota"`
		PaymentMethod   string  `gorm:"column:payment_method"`
		PaymentProvider string  `gorm:"column:payment_provider"`
		CompleteTime    int64   `gorm:"column:complete_time"`
	}
	if err := db.Model(&model.TopUp{}).
		Select("user_id, amount, money, credit_quota, payment_method, payment_provider, complete_time").
		Where("status = ? AND complete_time >= ? AND complete_time <= ?", common.TopUpStatusSuccess, startTimestamp, endTimestamp).
		Find(&topUpRows).Error; err != nil {
		return nil, err
	}

	payingUserSet := make(map[int]struct{})
	channelStatMap := make(map[string]*PaymentChannelStat)
	addPaymentChannelOrder := func(orderType, provider, method, origCurrency string, origAmount, usd float64) {
		normProvider := strings.ToLower(strings.TrimSpace(provider))
		normMethod := strings.ToLower(strings.TrimSpace(method))
		if normProvider == "" {
			normProvider = normMethod
		}
		if normProvider == "" {
			normProvider = "other"
		}
		channelLabel := normProvider
		if normMethod != "" && normMethod != normProvider {
			channelLabel = normProvider + ":" + normMethod
		}
		key := fmt.Sprintf("%s|%s|%s", orderType, channelLabel, origCurrency)
		stat, ok := channelStatMap[key]
		if !ok {
			stat = &PaymentChannelStat{
				Channel:          channelLabel,
				Provider:         normProvider,
				OrderType:        orderType,
				OriginalCurrency: origCurrency,
			}
			channelStatMap[key] = stat
		}
		stat.OrderCount++
		stat.OriginalAmount += origAmount
		stat.RevenueUSD += usd
	}

	var topUpRevenueUSD float64
	var paidOrderCount int64
	for _, row := range topUpRows {
		provider := strings.ToLower(strings.TrimSpace(row.PaymentProvider))
		method := strings.ToLower(strings.TrimSpace(row.PaymentMethod))
		effectiveProvider := provider
		if effectiveProvider == "" {
			effectiveProvider = method
		}

		var usd float64
		origCurrency := "USD"
		origAmount := row.Money

		switch effectiveProvider {
		case model.PaymentProviderStripe:
			usd = row.Money
			origCurrency = "USD"
			origAmount = row.Money
		case model.PaymentProviderCreem:
			if row.CreditQuota > 0 {
				usd = float64(row.CreditQuota) / quotaPerUnit
			} else {
				usd = float64(row.Amount) / quotaPerUnit
			}
			origCurrency = "USD"
			if origAmount <= 0 {
				origAmount = usd
			}
		case model.PaymentProviderWaffo, model.PaymentProviderWaffoPancake:
			if row.Amount > 0 {
				usd = float64(row.Amount)
			} else if row.CreditQuota > 0 {
				usd = float64(row.CreditQuota) / quotaPerUnit
			} else {
				usd = row.Money
			}
			origCurrency = "USD"
			if origAmount <= 0 {
				origAmount = usd
			}
		case model.PaymentProviderEpay, "alipay", "wxpay":
			origCurrency = "CNY"
			if row.Money > 0 {
				origAmount = row.Money
				usd = row.Money / priceRatio
			} else if row.Amount > 0 {
				usd = float64(row.Amount)
				origAmount = usd * priceRatio
			}
		default:
			if method == "alipay" || method == "wxpay" {
				origCurrency = "CNY"
				if row.Money > 0 {
					origAmount = row.Money
					usd = row.Money / priceRatio
				} else {
					usd = float64(row.Amount)
					origAmount = usd * priceRatio
				}
			} else {
				origCurrency = "USD"
				if row.Money > 0 {
					usd = row.Money
					origAmount = row.Money
				} else if row.CreditQuota > 0 {
					usd = float64(row.CreditQuota) / quotaPerUnit
					origAmount = usd
				} else {
					usd = float64(row.Amount)
					origAmount = usd
				}
			}
		}

		if usd < 0 || math.IsNaN(usd) || math.IsInf(usd, 0) {
			usd = 0
		}
		topUpRevenueUSD += usd
		paidOrderCount++
		if row.UserID > 0 {
			payingUserSet[row.UserID] = struct{}{}
		}
		addPaymentChannelOrder("topup", effectiveProvider, method, origCurrency, origAmount, usd)
		if pt := ensureDay(row.CompleteTime); pt != nil {
			pt.CashRevenueUSD += usd
			pt.PaidOrders++
		}
	}

	// Cash Subscription Orders (exclude balance-paid subscriptions to avoid double-counting wallet deposits)
	var subOrders []struct {
		UserID          int     `gorm:"column:user_id"`
		PlanID          int     `gorm:"column:plan_id"`
		Money           float64 `gorm:"column:money"`
		PaymentMethod   string  `gorm:"column:payment_method"`
		PaymentProvider string  `gorm:"column:payment_provider"`
		CompleteTime    int64   `gorm:"column:complete_time"`
	}
	if err := db.Model(&model.SubscriptionOrder{}).
		Select("user_id, plan_id, money, payment_method, payment_provider, complete_time").
		Where("status = ? AND complete_time >= ? AND complete_time <= ?", common.TopUpStatusSuccess, startTimestamp, endTimestamp).
		Find(&subOrders).Error; err != nil {
		return nil, err
	}

	planCurrencyByID := make(map[int]string)
	if len(subOrders) > 0 {
		planIDs := make([]int, 0, len(subOrders))
		seenPlan := make(map[int]struct{})
		for _, ord := range subOrders {
			if ord.PlanID > 0 {
				if _, ok := seenPlan[ord.PlanID]; !ok {
					seenPlan[ord.PlanID] = struct{}{}
					planIDs = append(planIDs, ord.PlanID)
				}
			}
		}
		if len(planIDs) > 0 {
			var plans []struct {
				ID       int    `gorm:"column:id"`
				Currency string `gorm:"column:currency"`
			}
			if err := db.Model(&model.SubscriptionPlan{}).
				Select("id, currency").
				Where("id IN ?", planIDs).
				Find(&plans).Error; err != nil {
				return nil, err
			}
			for _, pl := range plans {
				planCurrencyByID[pl.ID] = model.NormalizePlanCurrency(pl.Currency)
			}
		}
	}

	var subscriptionRevenueUSD float64
	for _, ord := range subOrders {
		provider := strings.ToLower(strings.TrimSpace(ord.PaymentProvider))
		method := strings.ToLower(strings.TrimSpace(ord.PaymentMethod))
		if provider == model.PaymentProviderBalance || method == model.PaymentMethodBalance {
			continue
		}
		effectiveProvider := provider
		if effectiveProvider == "" {
			effectiveProvider = method
		}

		planCurrency := planCurrencyByID[ord.PlanID]
		isCNY := planCurrency == "CNY" || effectiveProvider == model.PaymentProviderEpay || method == "alipay" || method == "wxpay"
		origCurrency := "USD"
		origAmount := ord.Money
		usd := ord.Money
		if isCNY {
			origCurrency = "CNY"
			usd = ord.Money / priceRatio
		}
		if usd < 0 || math.IsNaN(usd) || math.IsInf(usd, 0) {
			usd = 0
		}

		subscriptionRevenueUSD += usd
		paidOrderCount++
		if ord.UserID > 0 {
			payingUserSet[ord.UserID] = struct{}{}
		}
		addPaymentChannelOrder("subscription", effectiveProvider, method, origCurrency, origAmount, usd)
		if pt := ensureDay(ord.CompleteTime); pt != nil {
			pt.CashRevenueUSD += usd
			pt.PaidOrders++
		}
	}

	// Redemptions (separate promo/voucher bucket, not mixed into cash revenue)
	var redemptionRows []struct {
		Quota        int   `gorm:"column:quota"`
		RedeemedTime int64 `gorm:"column:redeemed_time"`
	}
	if err := db.Model(&model.Redemption{}).
		Select("quota, redeemed_time").
		Where("redeemed_time >= ? AND redeemed_time <= ?", startTimestamp, endTimestamp).
		Find(&redemptionRows).Error; err != nil {
		return nil, err
	}
	var redemptionUSD float64
	for _, r := range redemptionRows {
		if r.Quota <= 0 {
			continue
		}
		valUSD := float64(r.Quota) / quotaPerUnit
		redemptionUSD += valUSD
		if pt := ensureDay(r.RedeemedTime); pt != nil {
			pt.RedemptionUSD += valUSD
		}
	}

	// 4. Load active UpstreamModelOffers + UpstreamCostCalibrations for cost estimation
	var activeOffers []model.UpstreamModelOffer
	if err := db.Model(&model.UpstreamModelOffer{}).
		Where("enabled = ?", true).
		Order("collected_at desc, id desc").
		Find(&activeOffers).Error; err != nil {
		return nil, err
	}

	calibrations := make(map[string]model.UpstreamCostCalibration)
	var calRows []model.UpstreamCostCalibration
	if err := db.Model(&model.UpstreamCostCalibration{}).Find(&calRows).Error; err == nil {
		for _, cal := range calRows {
			calibrations[model.UpstreamCostCalibrationKey(cal.ChannelID, cal.PublicModel)] = cal
		}
	}

	offerUnitCostMap := make(map[string]resolvedOfferUnitCost)
	offerMapKey := func(channelID int, modelName string) string {
		return fmt.Sprintf("%d\x00%s", channelID, strings.ToLower(strings.TrimSpace(modelName)))
	}
	for i := range activeOffers {
		offer := &activeOffers[i]
		if offer.InputCost < 0 || offer.OutputCost < 0 ||
			(offer.InputCost == 0 && offer.OutputCost == 0) {
			continue
		}
		key := offerMapKey(offer.ChannelID, offer.PublicModel)
		if _, exists := offerUnitCostMap[key]; exists {
			continue
		}
		mult, _, usable := costMultiplierForOffer(offer, calibrations, nowTs)
		if !usable || mult <= 0 {
			mult = 1.0
		}
		currencyDivisor := 1.0
		if strings.EqualFold(strings.TrimSpace(offer.Currency), "CNY") ||
			strings.EqualFold(strings.TrimSpace(offer.Currency), "RMB") {
			currencyDivisor = usdExchangeRate
		}
		offerUnitCostMap[key] = resolvedOfferUnitCost{
			inputUSDPerMillion:  (offer.InputCost * mult) / currencyDivisor,
			outputUSDPerMillion: (offer.OutputCost * mult) / currencyDivisor,
		}
	}

	// 5. Consume Logs aggregation by (channel_id, model_name) + Active API users
	var activeLogUsers int64
	if err := logDB.Model(&model.Log{}).
		Where("type = ? AND created_at >= ? AND created_at <= ?", model.LogTypeConsume, startTimestamp, endTimestamp).
		Distinct("user_id").
		Count(&activeLogUsers).Error; err != nil {
		return nil, err
	}

	var logGroupRows []struct {
		ChannelID        int    `gorm:"column:channel_id"`
		ModelName        string `gorm:"column:model_name"`
		Requests         int64  `gorm:"column:requests"`
		Quota            int64  `gorm:"column:quota"`
		PromptTokens     int64  `gorm:"column:prompt_tokens"`
		CompletionTokens int64  `gorm:"column:completion_tokens"`
	}
	if err := logDB.Model(&model.Log{}).
		Select("channel_id, model_name, COUNT(*) AS requests, COALESCE(SUM(quota), 0) AS quota, COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens, COALESCE(SUM(completion_tokens), 0) AS completion_tokens").
		Where("type = ? AND created_at >= ? AND created_at <= ?", model.LogTypeConsume, startTimestamp, endTimestamp).
		Group("channel_id, model_name").
		Scan(&logGroupRows).Error; err != nil {
		return nil, err
	}

	var (
		totalRequests          int64
		totalTokens            int64
		totalConsumptionUSD    float64
		costedConsumptionUSD   float64
		uncostedConsumptionUSD float64
		totalEstimatedCostUSD  float64
	)

	// Store per-(channel, model) cost-to-consumption ratio for hourly QuotaData trend attribution
	costPerDollarBySlice := make(map[string]float64, len(logGroupRows))
	modelMargins := make([]ModelMarginStat, 0, len(logGroupRows))

	for _, row := range logGroupRows {
		totalRequests += row.Requests
		totalTokens += row.PromptTokens + row.CompletionTokens
		consUSD := float64(row.Quota) / quotaPerUnit
		totalConsumptionUSD += consUSD

		chName := channelNameByID[row.ChannelID]
		if chName == "" {
			if row.ChannelID > 0 {
				chName = fmt.Sprintf("#%d", row.ChannelID)
			} else {
				chName = "-"
			}
		}

		sliceKey := offerMapKey(row.ChannelID, row.ModelName)
		unitCost, covered := offerUnitCostMap[sliceKey]
		stat := ModelMarginStat{
			ModelName:        row.ModelName,
			ChannelID:        row.ChannelID,
			ChannelName:      chName,
			Requests:         row.Requests,
			PromptTokens:     row.PromptTokens,
			CompletionTokens: row.CompletionTokens,
			ConsumptionUSD:   roundMoney(consUSD),
			CostCovered:      covered,
		}

		if covered {
			estCost := (float64(row.PromptTokens)*unitCost.inputUSDPerMillion +
				float64(row.CompletionTokens)*unitCost.outputUSDPerMillion) / 1e6
			grossProfit := consUSD - estCost
			costedConsumptionUSD += consUSD
			totalEstimatedCostUSD += estCost

			roundedCost := roundMoney(estCost)
			roundedProfit := roundMoney(grossProfit)
			stat.EstimatedCostUSD = &roundedCost
			stat.GrossProfitUSD = &roundedProfit
			if consUSD > 0 {
				marginPct := roundPct((grossProfit / consUSD) * 100)
				stat.GrossMarginPct = &marginPct
				costPerDollarBySlice[sliceKey] = estCost / consUSD
			}
		} else {
			uncostedConsumptionUSD += consUSD
		}

		modelMargins = append(modelMargins, stat)
	}

	sort.Slice(modelMargins, func(i, j int) bool {
		if modelMargins[i].ConsumptionUSD != modelMargins[j].ConsumptionUSD {
			return modelMargins[i].ConsumptionUSD > modelMargins[j].ConsumptionUSD
		}
		return modelMargins[i].Requests > modelMargins[j].Requests
	})

	// 6. Daily Consumption & Estimated Cost Trend (prefer pre-aggregated QuotaData; fall back to direct Log scan if unflushed)
	var quotaHourlyRows []struct {
		CreatedAt int64  `gorm:"column:created_at"`
		ChannelID int    `gorm:"column:channel_id"`
		ModelName string `gorm:"column:model_name"`
		Quota     int64  `gorm:"column:quota"`
	}
	if err := db.Model(&model.QuotaData{}).
		Select("created_at, channel_id, model_name, COALESCE(SUM(quota), 0) AS quota").
		Where("created_at >= ? AND created_at <= ?", startTimestamp, endTimestamp).
		Group("created_at, channel_id, model_name").
		Scan(&quotaHourlyRows).Error; err != nil {
		return nil, err
	}

	if len(quotaHourlyRows) > 0 {
		for _, qRow := range quotaHourlyRows {
			pt := ensureDay(qRow.CreatedAt)
			if pt == nil {
				continue
			}
			consUSD := float64(qRow.Quota) / quotaPerUnit
			pt.ConsumptionUSD += consUSD
			sliceKey := offerMapKey(qRow.ChannelID, qRow.ModelName)
			if ratio, ok := costPerDollarBySlice[sliceKey]; ok {
				pt.EstimatedCostUSD += consUSD * ratio
			}
		}
	} else if totalRequests > 0 && totalRequests <= maxDirectDailyLogScan {
		var rawLogs []struct {
			CreatedAt        int64  `gorm:"column:created_at"`
			ChannelID        int    `gorm:"column:channel_id"`
			ModelName        string `gorm:"column:model_name"`
			Quota            int64  `gorm:"column:quota"`
			PromptTokens     int64  `gorm:"column:prompt_tokens"`
			CompletionTokens int64  `gorm:"column:completion_tokens"`
		}
		if err := logDB.Model(&model.Log{}).
			Select("created_at, channel_id, model_name, quota, prompt_tokens, completion_tokens").
			Where("type = ? AND created_at >= ? AND created_at <= ?", model.LogTypeConsume, startTimestamp, endTimestamp).
			Find(&rawLogs).Error; err == nil {
			for _, rl := range rawLogs {
				pt := ensureDay(rl.CreatedAt)
				if pt == nil {
					continue
				}
				pt.ConsumptionUSD += float64(rl.Quota) / quotaPerUnit
				if unitCost, ok := offerUnitCostMap[offerMapKey(rl.ChannelID, rl.ModelName)]; ok {
					pt.EstimatedCostUSD += (float64(rl.PromptTokens)*unitCost.inputUSDPerMillion +
						float64(rl.CompletionTokens)*unitCost.outputUSDPerMillion) / 1e6
				}
			}
		}
	}

	// Finalize daily trend array with rounded floats
	dailyTrend := make([]BusinessDailyPoint, 0, len(dayKeys))
	for _, d := range dayKeys {
		pt := dailyMap[d]
		pt.CashRevenueUSD = roundMoney(pt.CashRevenueUSD)
		pt.RedemptionUSD = roundMoney(pt.RedemptionUSD)
		pt.ConsumptionUSD = roundMoney(pt.ConsumptionUSD)
		pt.EstimatedCostUSD = roundMoney(pt.EstimatedCostUSD)
		dailyTrend = append(dailyTrend, *pt)
	}

	// Finalize sorted payment channels
	paymentChannels := make([]PaymentChannelStat, 0, len(channelStatMap))
	for _, stat := range channelStatMap {
		stat.OriginalAmount = roundMoney(stat.OriginalAmount)
		stat.RevenueUSD = roundMoney(stat.RevenueUSD)
		paymentChannels = append(paymentChannels, *stat)
	}
	sort.Slice(paymentChannels, func(i, j int) bool {
		if paymentChannels[i].RevenueUSD != paymentChannels[j].RevenueUSD {
			return paymentChannels[i].RevenueUSD > paymentChannels[j].RevenueUSD
		}
		return paymentChannels[i].OrderCount > paymentChannels[j].OrderCount
	})

	// Finalize Summary KPIs
	cashRevenueUSD := topUpRevenueUSD + subscriptionRevenueUSD
	payingUsers := int64(len(payingUserSet))
	var arppu float64
	if payingUsers > 0 {
		arppu = cashRevenueUSD / float64(payingUsers)
	}
	grossProfitUSD := costedConsumptionUSD - totalEstimatedCostUSD
	var grossMarginPct *float64
	if costedConsumptionUSD > 0 {
		pct := roundPct((grossProfitUSD / costedConsumptionUSD) * 100)
		grossMarginPct = &pct
	}
	var costCoveragePct float64
	if totalConsumptionUSD > 0 {
		costCoveragePct = roundPct((costedConsumptionUSD / totalConsumptionUSD) * 100)
	}

	summary := BusinessDashboardSummary{
		CashRevenueUSD:          roundMoney(cashRevenueUSD),
		TopUpRevenueUSD:         roundMoney(topUpRevenueUSD),
		SubscriptionRevenueUSD:  roundMoney(subscriptionRevenueUSD),
		PaidOrderCount:          paidOrderCount,
		PayingUsers:             payingUsers,
		ARPPUUSD:                roundMoney(arppu),
		RedemptionUSD:           roundMoney(redemptionUSD),
		RedemptionCount:         int64(len(redemptionRows)),
		ConsumptionUSD:          roundMoney(totalConsumptionUSD),
		CostedConsumptionUSD:    roundMoney(costedConsumptionUSD),
		UncostedConsumptionUSD:  roundMoney(uncostedConsumptionUSD),
		EstimatedCostUSD:        roundMoney(totalEstimatedCostUSD),
		EstimatedGrossProfitUSD: roundMoney(grossProfitUSD),
		EstimatedGrossMarginPct: grossMarginPct,
		CostCoveragePct:         costCoveragePct,
		TotalRequests:           totalRequests,
		TotalTokens:             totalTokens,
		TotalUsers:              userSnapshot.TotalUsers,
		NewUsers:                int64(len(newUserCreatedTimes)),
		ActiveUsers:             activeLogUsers,
		WalletLiabilityUSD:      roundMoney(float64(userSnapshot.TotalQuotaLeft) / quotaPerUnit),
		ChannelBalanceUSD:       roundMoney(totalChannelBalanceUSD),
	}

	return &BusinessDashboardData{
		Summary:         summary,
		DailyTrend:      dailyTrend,
		PaymentChannels: paymentChannels,
		ModelMargins:    modelMargins,
	}, nil
}
