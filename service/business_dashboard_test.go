package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func runBusinessDashboardVerification(t *testing.T, db *gorm.DB) {
	t.Helper()

	tablesToClean := []any{
		&model.User{},
		&model.Channel{},
		&model.TopUp{},
		&model.SubscriptionPlan{},
		&model.SubscriptionOrder{},
		&model.Redemption{},
		&model.UpstreamModelOffer{},
		&model.UpstreamCostCalibration{},
		&model.Log{},
		&model.QuotaData{},
	}
	cleanTables := func() {
		for _, tbl := range tablesToClean {
			_ = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(tbl).Error
		}
	}
	require.NoError(t, db.AutoMigrate(tablesToClean...))
	cleanTables()
	t.Cleanup(cleanTables)

	day1 := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC).Unix()
	day2 := time.Date(2026, 11, 21, 14, 0, 0, 0, time.UTC).Unix()
	beforeWindow := time.Date(2026, 11, 10, 8, 0, 0, 0, time.UTC).Unix()
	startTs := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC).Unix()
	endTs := time.Date(2026, 11, 21, 23, 59, 59, 0, time.UTC).Unix()

	// Seed 3 users (1 registered before window, 2 inside window); total remaining wallet = 7,500,000 ($15.00)
	tok1, tok2, tok3 := "tok00000000000000000000000000001", "tok00000000000000000000000000002", "tok00000000000000000000000000003"
	users := []model.User{
		{Id: 101, Username: "u_old", Password: "password123", Quota: 5000000, CreatedAt: beforeWindow, AccessToken: &tok1, AffCode: "aff101"},
		{Id: 102, Username: "u_new1", Password: "password123", Quota: 2500000, CreatedAt: day1, AccessToken: &tok2, AffCode: "aff102"},
		{Id: 103, Username: "u_new2", Password: "password123", Quota: 0, CreatedAt: day2, AccessToken: &tok3, AffCode: "aff103"},
	}
	for i := range users {
		require.NoError(t, db.Create(&users[i]).Error)
	}
	// Ensure explicit CreatedAt timestamps are preserved even with autoCreateTime tag
	require.NoError(t, db.Model(&model.User{}).Where("id = ?", 101).Update("created_at", beforeWindow).Error)
	require.NoError(t, db.Model(&model.User{}).Where("id = ?", 102).Update("created_at", day1).Error)
	require.NoError(t, db.Model(&model.User{}).Where("id = ?", 103).Update("created_at", day2).Error)

	// Seed 2 channels ($80.50 + $19.50 = $100.00 upstream balance)
	channels := []model.Channel{
		{Id: 1, Name: "OpenAI Direct", Key: "sk-1", Balance: 80.5},
		{Id: 2, Name: "Uncosted Relay", Key: "sk-2", Balance: 19.5},
	}
	for i := range channels {
		require.NoError(t, db.Create(&channels[i]).Error)
	}

	// Seed TopUps:
	// - Stripe $20 USD (user 101, day1, success)
	// - Epay ¥72 CNY @ 7.2 = $10 USD (user 102, day2, success)
	// - Pending Stripe $99 USD (must be ignored)
	topUps := []model.TopUp{
		{UserId: 101, Amount: 20, Money: 20, CreditQuota: 10000000, TradeNo: "t_stripe_1", PaymentMethod: "stripe", PaymentProvider: "stripe", CompleteTime: day1, Status: common.TopUpStatusSuccess},
		{UserId: 102, Amount: 10, Money: 72, CreditQuota: 5000000, TradeNo: "t_epay_1", PaymentMethod: "alipay", PaymentProvider: "epay", CompleteTime: day2, Status: common.TopUpStatusSuccess},
		{UserId: 103, Amount: 99, Money: 99, CreditQuota: 49500000, TradeNo: "t_pending_1", PaymentMethod: "stripe", PaymentProvider: "stripe", CompleteTime: day2, Status: "pending"},
	}
	for i := range topUps {
		require.NoError(t, db.Create(&topUps[i]).Error)
	}

	// Seed Subscriptions:
	// - Plan 1 (USD $15), paid via Waffo by user 102 on day2 -> counts as $15 cash revenue
	// - Plan 1 paid via wallet balance by user 101 -> excluded from cash revenue
	plan := model.SubscriptionPlan{Id: 1, Title: "Pro Monthly", PriceAmount: 15, Currency: "USD", DurationUnit: "month", DurationValue: 1, Enabled: true}
	require.NoError(t, db.Create(&plan).Error)
	subOrders := []model.SubscriptionOrder{
		{UserId: 102, PlanId: 1, Money: 15, TradeNo: "sub_waffo_1", PaymentMethod: "waffo", PaymentProvider: "waffo", Status: common.TopUpStatusSuccess, CompleteTime: day2},
		{UserId: 101, PlanId: 1, Money: 15, TradeNo: "sub_bal_1", PaymentMethod: "balance", PaymentProvider: "balance", Status: common.TopUpStatusSuccess, CompleteTime: day2},
	}
	for i := range subOrders {
		require.NoError(t, db.Create(&subOrders[i]).Error)
	}

	// Seed Redemptions: 2,500,000 quota ($5.00 promo credit) on day1 — must stay separate from cash revenue
	redemption := model.Redemption{UserId: 1, Key: "redem000000000000000000000000001", Name: "Promo", Quota: 2500000, RedeemedTime: day1, UsedUserId: 103}
	require.NoError(t, db.Create(&redemption).Error)

	// Seed UpstreamModelOffer for Channel 1 + gpt-4o:
	// InputCost = $2.00 / 1M, OutputCost = $8.00 / 1M, UpstreamGroupRatio = 0.8 -> effective $1.60 / $6.40 per 1M
	groupRatio := 0.8
	offer := model.UpstreamModelOffer{
		ChannelID:          1,
		UpstreamModel:      "gpt-4o",
		PublicModel:        "gpt-4o",
		InputCost:          2.0,
		OutputCost:         8.0,
		CacheReadCost:      0.2,
		UpstreamGroupRatio: &groupRatio,
		Currency:           "USD",
		SourceType:         "manual",
		SourceURL:          "https://upstream.example",
		SourceVersion:      "v1",
		CollectedAt:        day1,
		ExpiresAt:          endTs + 86400,
		Enabled:            true,
	}
	require.NoError(t, db.Create(&offer).Error)

	// Seed Consume Logs:
	// 1) Channel 1, gpt-4o (costed): 1,000,000 prompt + 500,000 completion tokens
	//    -> Upstream cost = (1.0 * 1.60) + (0.5 * 6.40) = $1.60 + $3.20 = $4.80
	//    -> Billed user quota = 5,000,000 ($10.00) -> Gross Profit = $5.20 (52.00% margin)
	// 2) Channel 2, custom-llm (uncosted): billed user quota = 1,000,000 ($2.00)
	logs := []model.Log{
		{UserId: 101, CreatedAt: day1, Type: model.LogTypeConsume, ModelName: "gpt-4o", ChannelId: 1, Quota: 5000000, PromptTokens: 1000000, CompletionTokens: 500000, RequestId: "req-1"},
		{UserId: 102, CreatedAt: day2, Type: model.LogTypeConsume, ModelName: "custom-llm", ChannelId: 2, Quota: 1000000, PromptTokens: 200000, CompletionTokens: 100000, RequestId: "req-2"},
	}
	for i := range logs {
		require.NoError(t, db.Create(&logs[i]).Error)
	}

	result, err := ComputeBusinessDashboard(db, db, startTs, endTs, BusinessDashboardRates{
		QuotaPerUnit:    500000,
		PriceRatio:      7.2,
		USDExchangeRate: 7.2,
		NowTimestamp:    day2,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify Cash Revenue ($20 Stripe + $10 Epay + $15 Waffo Sub = $45.00; balance sub & pending ignored; redemption separate)
	assert.InDelta(t, 30.0, result.Summary.TopUpRevenueUSD, 1e-4)
	assert.InDelta(t, 15.0, result.Summary.SubscriptionRevenueUSD, 1e-4)
	assert.InDelta(t, 45.0, result.Summary.CashRevenueUSD, 1e-4)
	assert.EqualValues(t, 3, result.Summary.PaidOrderCount)
	assert.EqualValues(t, 2, result.Summary.PayingUsers) // users 101 and 102
	assert.InDelta(t, 22.5, result.Summary.ARPPUUSD, 1e-4)
	assert.InDelta(t, 5.0, result.Summary.RedemptionUSD, 1e-4)
	assert.EqualValues(t, 1, result.Summary.RedemptionCount)

	// Verify User, Liability & Channel Balance KPIs
	assert.EqualValues(t, 3, result.Summary.TotalUsers)
	assert.EqualValues(t, 2, result.Summary.NewUsers)
	assert.EqualValues(t, 2, result.Summary.ActiveUsers)
	assert.InDelta(t, 15.0, result.Summary.WalletLiabilityUSD, 1e-4)
	assert.InDelta(t, 100.0, result.Summary.ChannelBalanceUSD, 1e-4)

	// Verify Consumption, Cost, Gross Profit & Coverage
	assert.InDelta(t, 12.0, result.Summary.ConsumptionUSD, 1e-4)
	assert.InDelta(t, 10.0, result.Summary.CostedConsumptionUSD, 1e-4)
	assert.InDelta(t, 2.0, result.Summary.UncostedConsumptionUSD, 1e-4)
	assert.InDelta(t, 4.8, result.Summary.EstimatedCostUSD, 1e-4)
	assert.InDelta(t, 5.2, result.Summary.EstimatedGrossProfitUSD, 1e-4)
	require.NotNil(t, result.Summary.EstimatedGrossMarginPct)
	assert.InDelta(t, 52.0, *result.Summary.EstimatedGrossMarginPct, 1e-2)
	assert.InDelta(t, 83.33, result.Summary.CostCoveragePct, 1e-2)

	// Verify Daily Trend (2 UTC days)
	require.Len(t, result.DailyTrend, 2)
	assert.Equal(t, "2026-11-20", result.DailyTrend[0].Date)
	assert.InDelta(t, 20.0, result.DailyTrend[0].CashRevenueUSD, 1e-4)
	assert.InDelta(t, 5.0, result.DailyTrend[0].RedemptionUSD, 1e-4)
	assert.InDelta(t, 10.0, result.DailyTrend[0].ConsumptionUSD, 1e-4)
	assert.InDelta(t, 4.8, result.DailyTrend[0].EstimatedCostUSD, 1e-4)
	assert.EqualValues(t, 1, result.DailyTrend[0].NewUsers)

	assert.Equal(t, "2026-11-21", result.DailyTrend[1].Date)
	assert.InDelta(t, 25.0, result.DailyTrend[1].CashRevenueUSD, 1e-4)
	assert.InDelta(t, 2.0, result.DailyTrend[1].ConsumptionUSD, 1e-4)
	assert.EqualValues(t, 1, result.DailyTrend[1].NewUsers)

	// Verify Model Margins breakdown
	require.Len(t, result.ModelMargins, 2)
	assert.Equal(t, "gpt-4o", result.ModelMargins[0].ModelName)
	assert.True(t, result.ModelMargins[0].CostCovered)
	require.NotNil(t, result.ModelMargins[0].EstimatedCostUSD)
	assert.InDelta(t, 4.8, *result.ModelMargins[0].EstimatedCostUSD, 1e-4)

	assert.Equal(t, "custom-llm", result.ModelMargins[1].ModelName)
	assert.False(t, result.ModelMargins[1].CostCovered)
	assert.Nil(t, result.ModelMargins[1].EstimatedCostUSD)
}

func TestComputeBusinessDashboardSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	runBusinessDashboardVerification(t, db)
}

func TestComputeBusinessDashboardPostgreSQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not configured")
	}
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	runBusinessDashboardVerification(t, db)
}

func TestComputeBusinessDashboardMySQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_MYSQL_DSN"))
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not configured")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	runBusinessDashboardVerification(t, db)
}
