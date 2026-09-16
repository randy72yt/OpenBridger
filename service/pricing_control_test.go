package service

import (
	"context"
	"os"
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
			require.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{{
				ChannelID: 101, UpstreamModel: "matrix-upstream", PublicModel: "matrix-public",
				InputCost: 1.25, OutputCost: 6.25, Currency: "USD", SourceType: "test",
				CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
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
	offers := []model.UpstreamModelOffer{
		{
			ChannelID: 1, UpstreamModel: "provider-model", PublicModel: "pricing-control-model",
			InputCost: 1, OutputCost: 5, CacheReadCost: 0.1, Currency: "usd",
			SourceType: "api", SourceVersion: "primary-v1", SuccessRateBPS: 9950,
			CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
		},
		{
			ChannelID: 2, UpstreamModel: "provider-model", PublicModel: "pricing-control-model",
			InputCost: 2, OutputCost: 10, CacheReadCost: 0.2, Currency: "USD",
			SourceType: "api", SourceVersion: "backup-v1", SuccessRateBPS: 9900,
			CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
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
	require.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{
		{ChannelID: 1, UpstreamModel: "m", PublicModel: "stable-model", InputCost: 1, OutputCost: 5, Currency: "USD", SourceType: "manual", CollectedAt: now, ExpiresAt: now + 3600, Enabled: true},
		{ChannelID: 2, UpstreamModel: "m", PublicModel: "stable-model", InputCost: 2, OutputCost: 10, Currency: "USD", SourceType: "manual", CollectedAt: now, ExpiresAt: now + 3600, Enabled: true},
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
	assert.ErrorIs(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{{
		ChannelID: 1, UpstreamModel: "m", PublicModel: "m", Currency: "CNY", SourceType: "manual",
	}}), ErrPricingOfferInvalid)
	assert.ErrorIs(t, ValidateAndUpsertPricePolicy(&model.ModelPricePolicy{
		PublicModel: "m", PrimaryChannelID: 1, TargetMarginBPS: 2000, MinimumMarginBPS: 2000,
	}), ErrPricingPolicyInvalid)
}
