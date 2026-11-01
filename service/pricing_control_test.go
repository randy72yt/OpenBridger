package service

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPricingControlDatabaseMatrix(t *testing.T) {
	for _, dialect := range []struct {
		name   string
		env    string
		dbType common.DatabaseType
		openDB func(string) gorm.Dialector
	}{
		{name: "mysql", env: "TEST_MYSQL_DSN", dbType: common.DatabaseTypeMySQL, openDB: func(dsn string) gorm.Dialector { return mysql.Open(dsn) }},
		{name: "postgres", env: "TEST_POSTGRES_DSN", dbType: common.DatabaseTypePostgreSQL, openDB: func(dsn string) gorm.Dialector { return postgres.Open(dsn) }},
	} {
		t.Run(dialect.name, func(t *testing.T) {
			dsn := os.Getenv(dialect.env)
			if dsn == "" {
				t.Skip("set " + dialect.env + " to run this database")
			}
			database, err := gorm.Open(dialect.openDB(dsn), &gorm.Config{})
			require.NoError(t, err)
			previousDB := model.DB
			previousMain, previousLog := common.MainDatabaseType(), common.LogDatabaseType()
			model.DB = database
			common.SetDatabaseTypes(dialect.dbType, dialect.dbType)
			t.Cleanup(func() {
				model.DB = previousDB
				common.SetDatabaseTypes(previousMain, previousLog)
				sqlDB, closeErr := database.DB()
				if closeErr == nil {
					require.NoError(t, sqlDB.Close())
				}
			})

			models := []any{
				&model.Option{}, &model.Channel{}, &model.UpstreamModelOffer{}, &model.UpstreamCostSnapshot{},
				&model.ModelPricePolicy{}, &model.ModelPriceProposal{},
			}
			require.NoError(t, database.AutoMigrate(models...))
			require.NoError(t, database.AutoMigrate(models...))
			now := common.GetTimestamp()
			groupRatio := 1.0
			require.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{{
				ChannelID: 101, UpstreamModel: "matrix-upstream", PublicModel: "matrix-public",
				InputCost: 1.25, OutputCost: 6.25, Currency: "USD", SourceType: "test",
				UpstreamGroupRatio: &groupRatio, CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
			}}))
			require.NoError(t, ValidateAndUpsertPricePolicy(&model.ModelPricePolicy{
				PublicModel: "matrix-public", ServiceTier: "default", PrimaryChannelID: 101,
				TargetMarginBPS: 3500, MinimumMarginBPS: 2000, OverheadBPS: 500, Enabled: true,
			}))
			created, err := RecalculatePricingProposals(context.Background())
			require.NoError(t, err)
			assert.Equal(t, 1, created)
		})
	}
}

func TestPricingControlImportRecalculateAndPublish(t *testing.T) {
	truncate(t)
	now := common.GetTimestamp()
	groupRatio := 1.0
	offers := []model.UpstreamModelOffer{
		{
			ChannelID: 1, UpstreamModel: "provider-model", PublicModel: "pricing-control-model",
			InputCost: 1, OutputCost: 5, CacheReadCost: 0.1, Currency: "usd",
			SourceType: "api", SourceVersion: "primary-v1", SuccessRateBPS: 9950,
			UpstreamGroupRatio: &groupRatio, CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
		},
		{
			ChannelID: 2, UpstreamModel: "provider-model", PublicModel: "pricing-control-model",
			InputCost: 2, OutputCost: 10, CacheReadCost: 0.2, Currency: "USD",
			SourceType: "api", SourceVersion: "backup-v1", SuccessRateBPS: 9900,
			UpstreamGroupRatio: &groupRatio, CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
		},
	}
	require.NoError(t, ValidateAndImportPricingOffers(offers))

	var snapshotCount int64
	require.NoError(t, model.DB.Model(&model.UpstreamCostSnapshot{}).Count(&snapshotCount).Error)
	assert.Equal(t, int64(2), snapshotCount)

	policy := &model.ModelPricePolicy{
		PublicModel: "pricing-control-model", ServiceTier: "default",
		PrimaryChannelID: 1, BackupChannelID: 2,
		TargetMarginBPS: 3500, MinimumMarginBPS: 2000, OverheadBPS: 500,
		FallbackProbabilityBPS: 1000, Enabled: true,
	}
	require.NoError(t, ValidateAndUpsertPricePolicy(policy))

	created, err := RecalculatePricingProposals(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, created)

	proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
	require.NoError(t, err)
	require.Len(t, proposals, 1)
	proposal := proposals[0]
	assert.InDelta(t, 1.155, proposal.ExpectedInputCost, 0.00000001)
	assert.InDelta(t, 5.775, proposal.ExpectedOutputCost, 0.00000001)
	assert.InDelta(t, 1.77692308, proposal.ProposedInputPrice, 0.00000001)
	assert.InDelta(t, 8.88461538, proposal.ProposedOutputPrice, 0.00000001)
	assert.Equal(t, 3500, proposal.ExpectedMarginBPS)

	require.NoError(t, ApprovePricingProposal(proposal.ID, 1))
	require.NoError(t, PublishPricingProposal(proposal.ID))
	assert.ErrorIs(t, PublishPricingProposal(proposal.ID), ErrPricingProposalState)

	pricing, err := model.GetModelPricingSnapshot([]string{"pricing-control-model"})
	require.NoError(t, err)
	require.Len(t, pricing.Entries, 1)
	assert.Equal(t, "tiered_expr", pricing.Entries[0].Configured["billing_setting.billing_mode"])
	assert.Equal(t, `tier("base", p * 1.77692308 + c * 8.88461538)`, pricing.Entries[0].Configured["billing_setting.billing_expr"])

	created, err = RecalculatePricingProposals(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, created)
	reviews, err := ListPricingProposalReviews(model.PricingProposalPending)
	require.NoError(t, err)
	require.Len(t, reviews, 1)
	require.NotNil(t, reviews[0].CurrentInputPrice)
	require.NotNil(t, reviews[0].CurrentOutputPrice)
	assert.InDelta(t, proposal.ProposedInputPrice, *reviews[0].CurrentInputPrice, 0.00000001)
	assert.InDelta(t, proposal.ProposedOutputPrice, *reviews[0].CurrentOutputPrice, 0.00000001)
	assert.Equal(t, 0, *reviews[0].InputPriceChangeBPS)
	assert.Equal(t, 0, *reviews[0].OutputPriceChangeBPS)
	assert.Equal(t, -1818, reviews[0].WorstMarginBPS)
	assert.False(t, reviews[0].PricingChanged)
}

func TestPricingControlStableTierAndMissingOfferRisk(t *testing.T) {
	truncate(t)
	now := common.GetTimestamp()
	groupRatio := 1.0
	require.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{
		{ChannelID: 1, UpstreamModel: "m", PublicModel: "stable-model", InputCost: 1, OutputCost: 5, UpstreamGroupRatio: &groupRatio, Currency: "USD", SourceType: "manual", CollectedAt: now, ExpiresAt: now + 3600, Enabled: true},
		{ChannelID: 2, UpstreamModel: "m", PublicModel: "stable-model", InputCost: 2, OutputCost: 10, UpstreamGroupRatio: &groupRatio, Currency: "USD", SourceType: "manual", CollectedAt: now, ExpiresAt: now + 3600, Enabled: true},
	}))
	require.NoError(t, ValidateAndUpsertPricePolicy(&model.ModelPricePolicy{
		PublicModel: "stable-model", ServiceTier: "stable", PrimaryChannelID: 1, BackupChannelID: 2,
		TargetMarginBPS: 3500, MinimumMarginBPS: 2000, OverheadBPS: 500,
		FallbackProbabilityBPS: 1000, Enabled: true,
	}))

	created, err := RecalculatePricingProposals(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, created)
	proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
	require.NoError(t, err)
	require.Len(t, proposals, 1)
	assert.InDelta(t, 2.625, proposals[0].ProposedInputPrice, 0.00000001)
	assert.InDelta(t, 13.125, proposals[0].ProposedOutputPrice, 0.00000001)
	assert.Equal(t, 5600, proposals[0].ExpectedMarginBPS)

	require.NoError(t, model.DB.Model(&model.UpstreamModelOffer{}).Where("channel_id = ?", 2).Update("expires_at", now-1).Error)
	risks, err := PricingRisks()
	require.NoError(t, err)
	require.Len(t, risks, 1)
	assert.Equal(t, "backup_offer_missing_or_expired", risks[0]["code"])

	created, err = RecalculatePricingProposals(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, created)
}

func TestPricingControlRejectsUnsupportedCurrencyAndInvalidMargin(t *testing.T) {
	groupRatio := 1.0
	assert.ErrorIs(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{{
		ChannelID: 1, UpstreamModel: "m", PublicModel: "m", UpstreamGroupRatio: &groupRatio, Currency: "CNY", SourceType: "manual",
	}}), ErrPricingOfferInvalid)
	assert.ErrorIs(t, ValidateAndUpsertPricePolicy(&model.ModelPricePolicy{
		PublicModel: "m", PrimaryChannelID: 1, TargetMarginBPS: 2000, MinimumMarginBPS: 2000,
	}), ErrPricingPolicyInvalid)
}

func TestPricingAutoPublishStaysOffUnlessEnabled(t *testing.T) {
	t.Setenv("PRICING_AUTO_PUBLISH", "")
	assert.False(t, pricingAutoPublishEnabled())
	t.Setenv("PRICING_AUTO_PUBLISH", "false")
	assert.False(t, pricingAutoPublishEnabled())
	t.Setenv("PRICING_AUTO_PUBLISH", "true")
	assert.True(t, pricingAutoPublishEnabled())
}

func TestPricingAutoPublishGates(t *testing.T) {
	require.NoError(t, model.LOG_DB.AutoMigrate(&model.AuditLog{}))

	t.Run("disabled sync does not publish", func(t *testing.T) {
		truncate(t)
		t.Setenv("PRICING_AUTO_PUBLISH", "false")
		const name = "auto-off"
		setSimpleSellPrice(t, name, 1.5, 7.5)
		prepareAutoPolicy(t, name, 500)
		result := PricingOfferSyncResult{}
		require.NoError(t, completePricingOfferSync(context.Background(), &result))
		assert.Nil(t, result.AutoPublish)
		assert.Equal(t, 1, result.CreatedProposals)
		requirePendingModel(t, name)
	})

	t.Run("recalculate task does not publish", func(t *testing.T) {
		truncate(t)
		t.Setenv("PRICING_AUTO_PUBLISH", "true")
		seedAutoPublishRoot(t, 41, common.UserStatusEnabled)
		const name = "auto-recalc-only"
		setSimpleSellPrice(t, name, 1.5, 7.5)
		prepareAutoPolicy(t, name, 500)
		created, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, created)
		requirePendingModel(t, name)
	})

	t.Run("publishes an in-band increase and retires older approvals", func(t *testing.T) {
		truncate(t)
		t.Setenv("PRICING_AUTO_PUBLISH", "true")
		seedAutoPublishRoot(t, 42, common.UserStatusEnabled)
		require.NoError(t, model.LOG_DB.Exec("DELETE FROM audit_logs").Error)
		const name = "auto-publish-up"
		setSimpleSellPrice(t, name, 1.5, 7.5)
		prepareAutoPolicy(t, name, 500)
		result := PricingOfferSyncResult{}
		require.NoError(t, completePricingOfferSync(context.Background(), &result))
		require.NotNil(t, result.AutoPublish)
		assert.Empty(t, result.AutoPublish.Error)
		assert.Equal(t, 1, result.AutoPublish.Published)
		assert.Equal(t, 0, result.AutoPublish.Skipped)

		var published model.ModelPriceProposal
		require.NoError(t, model.DB.Where("public_model = ? AND status = ?", name, model.PricingProposalPublished).First(&published).Error)
		assert.Equal(t, 42, published.ApprovedBy)
		assert.NotZero(t, published.ApprovedAt)
		assert.NotZero(t, published.PublishedAt)

		pricing, err := model.GetModelPricingSnapshot([]string{name})
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("tier(\"base\", p * %s + c * %s)", formatPricingValue(published.ProposedInputPrice), formatPricingValue(published.ProposedOutputPrice)), pricing.Entries[0].Configured["billing_setting.billing_expr"])

		var audits []model.AuditLog
		require.NoError(t, model.LOG_DB.Where("user_id = ?", 42).Order("id").Find(&audits).Error)
		require.Len(t, audits, 2)
		assert.Equal(t, "pricing.proposal.approve", audits[0].Action)
		assert.Equal(t, "pricing.proposal.publish", audits[1].Action)
		assert.Contains(t, audits[0].Content, fmt.Sprintf("%d", published.ID))
	})

	t.Run("rejects a previously approved proposal for the same model", func(t *testing.T) {
		truncate(t)
		seedAutoPublishRoot(t, 43, common.UserStatusEnabled)
		const name = "auto-sibling"
		setSimpleSellPrice(t, name, 1.5, 7.5)
		ids := prepareAndRecalculate(t, name, 500)
		stale := model.ModelPriceProposal{}
		require.NoError(t, model.DB.First(&stale, ids[0]).Error)
		stale.ID = 0
		stale.Status = model.PricingProposalApproved
		stale.ProposedInputPrice = 99
		stale.ProposedOutputPrice = 99
		require.NoError(t, model.DB.Create(&stale).Error)

		stats := AutoPublishEligibleProposals(context.Background(), ids)
		assert.Equal(t, 1, stats.Published)
		var retired model.ModelPriceProposal
		require.NoError(t, model.DB.First(&retired, stale.ID).Error)
		assert.Equal(t, model.PricingProposalRejected, retired.Status)
	})

	t.Run("zero band stays manual", func(t *testing.T) {
		truncate(t)
		seedAutoPublishRoot(t, 44, common.UserStatusEnabled)
		const name = "auto-zero-band"
		setSimpleSellPrice(t, name, 1.5, 7.5)
		assert.Equal(t, "auto_publish_disabled", autoPublishReason(t, prepareAndRecalculate(t, name, 0)))
		requirePendingModel(t, name)
	})

	t.Run("increase above the band stays manual", func(t *testing.T) {
		truncate(t)
		seedAutoPublishRoot(t, 45, common.UserStatusEnabled)
		const name = "auto-wide"
		setSimpleSellPrice(t, name, 1.5, 7.5)
		assert.Equal(t, "change_band", autoPublishReason(t, prepareAndRecalculate(t, name, 100)))
		requirePendingModel(t, name)
	})

	t.Run("a price cut stays manual", func(t *testing.T) {
		truncate(t)
		seedAutoPublishRoot(t, 46, common.UserStatusEnabled)
		const name = "auto-cut"
		setSimpleSellPrice(t, name, 2, 10)
		assert.Equal(t, "price_cut", autoPublishReason(t, prepareAndRecalculate(t, name, 5000)))
		requirePendingModel(t, name)
	})

	t.Run("matching price is not rewritten", func(t *testing.T) {
		truncate(t)
		seedAutoPublishRoot(t, 47, common.UserStatusEnabled)
		const name = "auto-same"
		qaSeedOffers(t, qaOffer(1, name, 1, 5))
		proposal := qaBuild(t, qaPolicy(name, "default", 1, 0, 3500, 2000, 0, 0))
		setSimpleSellPrice(t, name, proposal.ProposedInputPrice, proposal.ProposedOutputPrice)
		assert.Equal(t, "unchanged", autoPublishReason(t, prepareAndRecalculate(t, name, 500)))
		pricing, err := model.GetModelPricingSnapshot([]string{name})
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("tier(\"base\", p * %s + c * %s)", formatPricingValue(proposal.ProposedInputPrice), formatPricingValue(proposal.ProposedOutputPrice)), pricing.Entries[0].Configured["billing_setting.billing_expr"])
	})

	t.Run("margin below the policy floor stays manual", func(t *testing.T) {
		truncate(t)
		seedAutoPublishRoot(t, 48, common.UserStatusEnabled)
		const name = "auto-margin"
		setSimpleSellPrice(t, name, 1.5, 7.5)
		ids := prepareAndRecalculate(t, name, 500)
		require.NoError(t, model.DB.Model(&model.ModelPriceProposal{}).Where("id = ?", ids[0]).Update("expected_margin_bps", 100).Error)
		assert.Equal(t, "margin", autoPublishReason(t, ids))
		requirePendingModel(t, name)
	})

	t.Run("complex expression stays manual", func(t *testing.T) {
		truncate(t)
		seedAutoPublishRoot(t, 49, common.UserStatusEnabled)
		const name = "auto-complex"
		setSellExpression(t, name, `tier("base", p * 1.5 + c * 7.5 + cr * 0.1)`)
		assert.Equal(t, "pricing_shape", autoPublishReason(t, prepareAndRecalculate(t, name, 5000)))
		pricing, err := model.GetModelPricingSnapshot([]string{name})
		require.NoError(t, err)
		assert.Equal(t, `tier("base", p * 1.5 + c * 7.5 + cr * 0.1)`, pricing.Entries[0].Configured["billing_setting.billing_expr"])
	})

	t.Run("cache ratio stays manual", func(t *testing.T) {
		truncate(t)
		seedAutoPublishRoot(t, 50, common.UserStatusEnabled)
		const name = "auto-cache"
		snapshot, err := model.GetModelPricingSnapshot([]string{name})
		require.NoError(t, err)
		require.NoError(t, model.UpdateModelPricing([]model.ModelPricingChange{{
			ModelName: name, ExpectedVersion: snapshot.Entries[0].Version,
			Pricing: model.PricingValues{
				"billing_setting.billing_mode": "tiered_expr",
				"billing_setting.billing_expr": `tier("base", p * 1.5 + c * 7.5)`,
				"CacheRatio":                   0.2,
			},
		}}))
		assert.Equal(t, "pricing_shape", autoPublishReason(t, prepareAndRecalculate(t, name, 5000)))
		pricing, err := model.GetModelPricingSnapshot([]string{name})
		require.NoError(t, err)
		assert.Equal(t, 0.2, pricing.Entries[0].Configured["CacheRatio"])
	})

	t.Run("no live price stays manual", func(t *testing.T) {
		truncate(t)
		seedAutoPublishRoot(t, 51, common.UserStatusEnabled)
		const name = "auto-first"
		assert.Equal(t, "change_band", autoPublishReason(t, prepareAndRecalculate(t, name, 5000)))
		pricing, err := model.GetModelPricingSnapshot([]string{name})
		require.NoError(t, err)
		assert.Nil(t, pricing.Entries[0].Configured["billing_setting.billing_expr"])
	})

	t.Run("locked disabled stable and stale offers stay manual", func(t *testing.T) {
		truncate(t)
		seedAutoPublishRoot(t, 52, common.UserStatusEnabled)

		locked := prepareAndRecalculate(t, "auto-locked", 500)
		require.NoError(t, model.DB.Model(&model.ModelPricePolicy{}).Where("public_model = ?", "auto-locked").Update("price_locked", true).Error)
		assert.Equal(t, "price_locked", autoPublishReason(t, locked))

		disabled := prepareAndRecalculate(t, "auto-disabled", 500)
		require.NoError(t, model.DB.Model(&model.ModelPricePolicy{}).Where("public_model = ?", "auto-disabled").Update("enabled", false).Error)
		assert.Equal(t, "policy_disabled", autoPublishReason(t, disabled))

		qaSeedOffers(t, qaOffer(1, "auto-stable", 1, 5), qaOffer(2, "auto-stable", 2, 10))
		stablePolicy := qaPolicy("auto-stable", "stable", 1, 2, 3500, 2000, 0, 1000)
		stablePolicy.AutoPublishChangeBPS = 5000
		require.NoError(t, ValidateAndUpsertPricePolicy(stablePolicy))
		_, _, err := recalculatePricingProposals(context.Background())
		require.NoError(t, err)
		var stable model.ModelPriceProposal
		require.NoError(t, model.DB.Where("public_model = ? AND status = ?", "auto-stable", model.PricingProposalPending).Order("id desc").First(&stable).Error)
		assert.Equal(t, "service_tier", autoPublishReason(t, []int64{stable.ID}))

		expired := prepareAndRecalculate(t, "auto-expired", 5000)
		require.NoError(t, model.DB.Model(&model.UpstreamModelOffer{}).Where("public_model = ?", "auto-expired").Update("expires_at", common.GetTimestamp()-1).Error)
		assert.Equal(t, "offer_stale", autoPublishReason(t, expired))
	})

	t.Run("operator must be an enabled root user", func(t *testing.T) {
		truncate(t)
		const name = "auto-operator"
		setSimpleSellPrice(t, name, 1.5, 7.5)
		ids := prepareAndRecalculate(t, name, 500)

		t.Setenv("PRICING_AUTO_PUBLISH_OPERATOR_ID", "")
		stats := AutoPublishEligibleProposals(context.Background(), ids)
		assert.Contains(t, stats.Error, "PRICING_AUTO_PUBLISH_OPERATOR_ID")
		assert.Equal(t, 0, stats.Published)

		seedAutoPublishRoot(t, 53, common.UserStatusDisabled)
		stats = AutoPublishEligibleProposals(context.Background(), ids)
		assert.Contains(t, stats.Error, "enabled root user")
		requirePendingModel(t, name)

		require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", 53).Update("role", common.RoleAdminUser).Error)
		require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", 53).Update("status", common.UserStatusEnabled).Error)
		stats = AutoPublishEligibleProposals(context.Background(), ids)
		assert.Contains(t, stats.Error, "enabled root user")
		requirePendingModel(t, name)
	})
}

func prepareAutoPolicy(t *testing.T, name string, band int) {
	t.Helper()
	qaSeedOffers(t, qaOffer(1, name, 1, 5))
	policy := qaPolicy(name, "default", 1, 0, 3500, 2000, 0, 0)
	policy.AutoPublishChangeBPS = band
	require.NoError(t, ValidateAndUpsertPricePolicy(policy))
}

func prepareAndRecalculate(t *testing.T, name string, band int) []int64 {
	t.Helper()
	prepareAutoPolicy(t, name, band)
	_, ids, err := recalculatePricingProposals(context.Background())
	require.NoError(t, err)
	var mine []int64
	for _, id := range ids {
		var proposal model.ModelPriceProposal
		require.NoError(t, model.DB.First(&proposal, id).Error)
		if proposal.PublicModel == name {
			mine = append(mine, id)
		}
	}
	require.NotEmpty(t, mine)
	return mine
}

func autoPublishReason(t *testing.T, ids []int64) string {
	t.Helper()
	stats := AutoPublishEligibleProposals(context.Background(), ids)
	require.Empty(t, stats.Error)
	require.Equal(t, 0, stats.Published)
	require.Equal(t, len(ids), stats.Skipped)
	require.Len(t, stats.Skips, len(ids))
	return stats.Skips[0].Reason
}

func requirePendingModel(t *testing.T, name string) {
	t.Helper()
	var proposal model.ModelPriceProposal
	require.NoError(t, model.DB.Where("public_model = ?", name).Order("id desc").First(&proposal).Error)
	assert.Equal(t, model.PricingProposalPending, proposal.Status)
}

func setSimpleSellPrice(t *testing.T, name string, input, output float64) {
	t.Helper()
	setSellExpression(t, name, fmt.Sprintf("tier(\"base\", p * %s + c * %s)", formatPricingValue(input), formatPricingValue(output)))
}

func setSellExpression(t *testing.T, name, expression string) {
	t.Helper()
	snapshot, err := model.GetModelPricingSnapshot([]string{name})
	require.NoError(t, err)
	require.Len(t, snapshot.Entries, 1)
	require.NoError(t, model.UpdateModelPricing([]model.ModelPricingChange{{
		ModelName: name, ExpectedVersion: snapshot.Entries[0].Version,
		Pricing: model.PricingValues{
			"billing_setting.billing_mode": "tiered_expr",
			"billing_setting.billing_expr": expression,
		},
	}}))
}

func seedAutoPublishRoot(t *testing.T, id int, status int) {
	t.Helper()
	user := &model.User{
		Id: id, Username: "root" + strconv.Itoa(id), Password: "not-used",
		Role: common.RoleRootUser, Status: status, AffCode: "aff" + strconv.Itoa(id),
	}
	require.NoError(t, model.DB.Create(user).Error)
	t.Setenv("PRICING_AUTO_PUBLISH_OPERATOR_ID", strconv.Itoa(id))
}
