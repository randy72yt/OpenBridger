package service

import (
	"context"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// P0-1 / P0-1b upstream group ratio resolution
// ---------------------------------------------------------------------------

func ratioTestPayload() *upstreamPricingResponse {
	return &upstreamPricingResponse{
		Success:    true,
		AutoGroups: []string{"claude-kiro", "claude-max", "gemini"},
		GroupRatio: map[string]float64{
			"claude-kiro": 0.25,
			"claude-max":  1.8,
			"gemini":      0.3,
			"default":     0.23,
		},
		ModelGroupRatio: map[string]map[string]float64{
			"claude-haiku-4-5": {"claude-max": 0.2647, "auto": 0.1471},
		},
	}
}

func TestPricingGroupRatioPolicySelection(t *testing.T) {
	payload := ratioTestPayload()
	item := upstreamPriceItem{ModelName: "claude-haiku-4-5", EnableGroups: []string{"claude-kiro", "claude-max"}}

	t.Run("key_group follows upstream auto group order", func(t *testing.T) {
		choice, warnings := resolveUpstreamGroupRatio(payload, item, pricingGroupRatioPolicyKeyGroup)
		require.Empty(t, warnings)
		assert.Equal(t, "claude-kiro", choice.Group)
		assert.InDelta(t, 0.25, choice.Ratio, 1e-9)
	})

	t.Run("max picks the most expensive settlement group", func(t *testing.T) {
		choice, warnings := resolveUpstreamGroupRatio(payload, item, pricingGroupRatioPolicyMax)
		require.Empty(t, warnings)
		assert.Equal(t, "claude-max", choice.Group)
		assert.InDelta(t, 1.8, choice.Ratio, 1e-9, "max must never understate procurement cost")
	})

	t.Run("min picks the cheapest settlement group", func(t *testing.T) {
		choice, warnings := resolveUpstreamGroupRatio(payload, item, pricingGroupRatioPolicyMin)
		require.Empty(t, warnings)
		assert.Equal(t, "claude-kiro", choice.Group)
		assert.InDelta(t, 0.25, choice.Ratio, 1e-9)
	})

	t.Run("model level ratio overrides the top level ratio", func(t *testing.T) {
		onlyMax := upstreamPriceItem{ModelName: "claude-haiku-4-5", EnableGroups: []string{"claude-max"}}
		choice, warnings := resolveUpstreamGroupRatio(payload, onlyMax, pricingGroupRatioPolicyKeyGroup)
		require.Empty(t, warnings)
		assert.True(t, choice.Scoped, "model_group_ratio must win over group_ratio")
		assert.InDelta(t, 0.2647, choice.Ratio, 1e-9)
	})

	t.Run("single group model is policy independent", func(t *testing.T) {
		gemini := upstreamPriceItem{ModelName: "gemini-3.1-pro", EnableGroups: []string{"gemini"}}
		for _, policy := range []string{pricingGroupRatioPolicyKeyGroup, pricingGroupRatioPolicyMax, pricingGroupRatioPolicyMin} {
			choice, warnings := resolveUpstreamGroupRatio(payload, gemini, policy)
			require.Empty(t, warnings)
			assert.Equal(t, "gemini", choice.Group)
			assert.InDelta(t, 0.3, choice.Ratio, 1e-9)
		}
	})
}

func TestPricingGroupRatioPolicyRejectsUnknownGroups(t *testing.T) {
	payload := ratioTestPayload()
	item := upstreamPriceItem{ModelName: "unknown-model", EnableGroups: []string{"nope", "also-nope"}}

	for _, policy := range []string{pricingGroupRatioPolicyKeyGroup, pricingGroupRatioPolicyMax, pricingGroupRatioPolicyMin} {
		choice, warnings := resolveUpstreamGroupRatio(payload, item, policy)
		assert.Zero(t, choice.Ratio, "must not fall back to 1 when no group ratio is known")
		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "no usable upstream group ratio")
		assert.Contains(t, warnings[0], "nope")
	}
}

// The measured ratios of one upstream response spanned 0.002 to 1.0 because the
// provider settles each model through a private group it never publishes. max
// therefore over-priced whole families by up to 15x, so median is the default
// and calibration corrects the rest from real invoices.
func TestPricingGroupRatioPolicyDefaultsToMedian(t *testing.T) {
	t.Setenv("PRICING_GROUP_RATIO_POLICY", "")
	assert.Equal(t, pricingGroupRatioPolicyMedian, pricingGroupRatioPolicy())
	t.Setenv("PRICING_GROUP_RATIO_POLICY", "MAX")
	assert.Equal(t, pricingGroupRatioPolicyMax, pricingGroupRatioPolicy())
	t.Setenv("PRICING_GROUP_RATIO_POLICY", "key_group")
	assert.Equal(t, pricingGroupRatioPolicyKeyGroup, pricingGroupRatioPolicy())
	t.Setenv("PRICING_GROUP_RATIO_POLICY", "nonsense")
	assert.Equal(t, pricingGroupRatioPolicyMedian, pricingGroupRatioPolicy())
}

// ---------------------------------------------------------------------------
// P0-6 offer TTL
// ---------------------------------------------------------------------------

func TestPricingOfferTTLConfig(t *testing.T) {
	t.Setenv("PRICING_OFFER_TTL_MINUTES", "")
	assert.Equal(t, int64(120*60), pricingOfferTTLSeconds())

	t.Setenv("PRICING_OFFER_TTL_MINUTES", "840")
	assert.Equal(t, int64(840*60), pricingOfferTTLSeconds(), "two syncs a day need a TTL longer than the interval")

	t.Setenv("PRICING_OFFER_TTL_MINUTES", "1")
	assert.Equal(t, int64(15*60), pricingOfferTTLSeconds(), "below 15 minutes is clamped")
}

// ---------------------------------------------------------------------------
// P1-4 source priority
// ---------------------------------------------------------------------------

func TestPricingOfferSourcePriority(t *testing.T) {
	truncate(t)
	now := common.GetTimestamp()
	automatic := qaOffer(1, "priority-model", 1, 2)
	automatic.SourceType = "new-api-pricing"
	require.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{automatic}))

	manual := qaOffer(1, "priority-model", 9, 9)
	manual.SourceType = "manual"
	stats, err := ValidateAndImportPricingOffersWithStats([]model.UpstreamModelOffer{manual})
	require.NoError(t, err)
	assert.Equal(t, 0, stats.Applied)
	assert.Equal(t, 1, stats.Kept, "a manual import must not overwrite an unexpired automatic offer")

	offers, err := model.ListUpstreamModelOffers("priority-model")
	require.NoError(t, err)
	require.Len(t, offers, 1)
	assert.InDelta(t, 1.0, offers[0].InputCost, 1e-9)
	assert.Equal(t, "new-api-pricing", offers[0].SourceType)

	t.Run("expired automatic offer no longer wins", func(t *testing.T) {
		truncate(t)
		stale := qaOffer(1, "priority-expired", 1, 2)
		stale.SourceType = "new-api-pricing"
		stale.CollectedAt = now - 7200
		stale.ExpiresAt = now - 10
		require.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{stale}))

		manual := qaOffer(1, "priority-expired", 9, 9)
		manual.SourceType = "manual"
		stats, err := ValidateAndImportPricingOffersWithStats([]model.UpstreamModelOffer{manual})
		require.NoError(t, err)
		assert.Equal(t, 1, stats.Applied)
	})
}

// ---------------------------------------------------------------------------
// P2-1 global policy defaults (D4)
// ---------------------------------------------------------------------------

func TestPricingGlobalPolicyDefaults(t *testing.T) {
	truncate(t)
	require.NoError(t, ValidateAndUpsertPricePolicy(&model.ModelPricePolicy{
		PublicModel: model.GlobalPricePolicyModel, ServiceTier: "default",
		TargetMarginBPS: 4000, MinimumMarginBPS: 2500, OverheadBPS: 800, Enabled: true,
	}))
	global, err := model.GlobalPricePolicy()
	require.NoError(t, err)
	require.NotNil(t, global)

	t.Run("model without its own margins inherits the global defaults", func(t *testing.T) {
		policy := &model.ModelPricePolicy{PublicModel: "inherit-model", ServiceTier: "default", PrimaryChannelID: 1}
		policy.ApplyGlobalDefaults(global)
		assert.Equal(t, 4000, policy.TargetMarginBPS)
		assert.Equal(t, 2500, policy.MinimumMarginBPS)
		assert.Equal(t, 800, policy.OverheadBPS)
	})

	t.Run("explicit model values win and do not leak back", func(t *testing.T) {
		policy := &model.ModelPricePolicy{PublicModel: "override-model", ServiceTier: "default", PrimaryChannelID: 1, TargetMarginBPS: 6000}
		policy.ApplyGlobalDefaults(global)
		assert.Equal(t, 6000, policy.TargetMarginBPS)
		assert.Equal(t, 2500, policy.MinimumMarginBPS, "unset fields still inherit")

		other := &model.ModelPricePolicy{PublicModel: "other-model", ServiceTier: "default", PrimaryChannelID: 1}
		other.ApplyGlobalDefaults(global)
		assert.Equal(t, 4000, other.TargetMarginBPS, "one model's override must not change another")
	})

	t.Run("global row never generates proposals", func(t *testing.T) {
		created, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		assert.Zero(t, created)

		proposals, err := model.ListModelPriceProposals("")
		require.NoError(t, err)
		assert.Empty(t, proposals)
	})

	t.Run("global row is excluded from the risk list", func(t *testing.T) {
		risks, err := PricingRisks()
		require.NoError(t, err)
		assert.Empty(t, risks)
	})
}

// ---------------------------------------------------------------------------
// P0-3 manual price write guard (D2)
// ---------------------------------------------------------------------------

func TestPricingManualWriteGuard(t *testing.T) {
	t.Setenv("PRICING_CONTROL_ENABLED", "false")
	require.NoError(t, ValidateManualPricingWrite(10, false, ""))

	t.Setenv("PRICING_CONTROL_ENABLED", "true")
	assert.ErrorIs(t, ValidateManualPricingWrite(10, false, ""), ErrManualPricingWriteBlocked)
	assert.ErrorIs(t, ValidateManualPricingWrite(10, true, "maintenance"), ErrManualPricingWriteBlocked,
		"bulk writes stay blocked even with an override")
	assert.ErrorIs(t, ValidateManualPricingWrite(1, true, ""), ErrManualPricingWriteBlocked,
		"an override needs a reason")
	assert.NoError(t, ValidateManualPricingWrite(1, true, "price correction after upstream outage"))
}

// ---------------------------------------------------------------------------
// P1-3 sync run persistence
// ---------------------------------------------------------------------------

func TestPricingSyncRunPersistence(t *testing.T) {
	truncate(t)
	require.NoError(t, model.RecordPricingSyncRun(&model.PricingSyncRun{
		Trigger: "manual", Status: "succeeded", Imported: 50, Kept: 2, Skipped: 0, CreatedProposals: 12,
		Channels: `[{"channel_id":1,"imported":50}]`,
	}))
	latest, err := model.LatestPricingSyncRun()
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, "manual", latest.Trigger)
	assert.Equal(t, 50, latest.Imported)
	assert.NotZero(t, latest.FinishedAt)
}
