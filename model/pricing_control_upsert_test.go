package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func truncatePricingTables(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		DB.Exec("DELETE FROM upstream_model_offers")
		DB.Exec("DELETE FROM upstream_cost_snapshots")
		DB.Exec("DELETE FROM model_price_policies")
		DB.Exec("DELETE FROM model_price_proposals")
		DB.Exec("DELETE FROM pricing_sync_runs")
		DB.Exec("DELETE FROM upstream_cost_calibrations")
	})
	DB.Exec("DELETE FROM upstream_model_offers")
	DB.Exec("DELETE FROM upstream_cost_snapshots")
	DB.Exec("DELETE FROM model_price_policies")
	DB.Exec("DELETE FROM model_price_proposals")
	DB.Exec("DELETE FROM pricing_sync_runs")
	DB.Exec("DELETE FROM upstream_cost_calibrations")
}

func seedOffer(t *testing.T, channelID int, publicModel, sourceType string, inputCost float64, expiresAt int64) *UpstreamModelOffer {
	t.Helper()
	ratio := 1.0
	offer := &UpstreamModelOffer{
		ChannelID:          channelID,
		UpstreamModel:      publicModel + "-up",
		PublicModel:        publicModel,
		InputCost:          inputCost,
		OutputCost:         inputCost,
		CacheReadCost:      inputCost,
		UpstreamGroupRatio: &ratio,
		UpstreamGroup:      "default",
		Currency:           "USD",
		SourceType:         sourceType,
		SourceURL:          "https://example.com/pricing",
		SourceVersion:      "v1",
		SuccessRateBPS:     10000,
		CollectedAt:        common.GetTimestamp(),
		ExpiresAt:          expiresAt,
		Enabled:            true,
	}
	require.NoError(t, UpsertUpstreamModelOffer(offer))
	return offer
}

// 报价导入的来源优先级（D3）：new-api-pricing(100) > manual(50) > 其他(10)，
// 高优先级来源拥有的行不被低优先级写入覆盖。
func TestImportUpstreamModelOffersSourcePriority(t *testing.T) {
	truncatePricingTables(t)

	// 高优先级先写入
	seedOffer(t, 701, "prio-model", "new-api-pricing", 1.0, common.GetTimestamp()+3600)
	// 低优先级尝试覆盖同 (channel, upstream, public) 行
	manual := UpstreamModelOffer{
		ChannelID: 701, UpstreamModel: "prio-model-up", PublicModel: "prio-model",
		InputCost: 99.0, OutputCost: 99.0, CacheReadCost: 99.0,
		UpstreamGroupRatio: ptrOf(1.0), UpstreamGroup: "default",
		Currency: "USD", SourceType: "manual",
		SourceURL: "https://example.com/manual", SourceVersion: "v2",
		SuccessRateBPS: 9000, CollectedAt: common.GetTimestamp(),
		ExpiresAt: common.GetTimestamp() + 3600, Enabled: true,
	}
	stats, err := ImportUpstreamModelOffersWithStats([]UpstreamModelOffer{manual})
	require.NoError(t, err)
	assert.Equal(t, 0, stats.Applied, "manual import must not override automated collection")
	assert.Equal(t, 1, stats.Kept, "the rejected row counts as kept")

	var stored UpstreamModelOffer
	require.NoError(t, DB.Where("channel_id = ? AND public_model = ?", 701, "prio-model").First(&stored).Error)
	assert.Equal(t, 1.0, stored.InputCost, "original cost must survive")
	assert.Equal(t, "new-api-pricing", stored.SourceType)

	// 无效数据整批拒绝：channel=0、非 USD、NaN 成本、SuccessRateBPS 越界
	bad := manual
	bad.ChannelID = 0
	_, err = ImportUpstreamModelOffersWithStats([]UpstreamModelOffer{bad})
	require.Error(t, err)
}

// 策略 upsert：同 (public_model, service_tier) 冲突时更新而非插入第二行。
func TestUpsertModelPricePolicyReplacesOnConflict(t *testing.T) {
	truncatePricingTables(t)

	first := &ModelPricePolicy{
		PublicModel: "policy-model", ServiceTier: "stable",
		PrimaryChannelID: 801, TargetMarginBPS: 2000, MinimumMarginBPS: 500,
		OverheadBPS: 300, FallbackProbabilityBPS: 100, Enabled: true,
	}
	require.NoError(t, UpsertModelPricePolicy(first))

	first.TargetMarginBPS = 3000
	first.OverheadBPS = 400
	require.NoError(t, UpsertModelPricePolicy(first))

	var count int64
	require.NoError(t, DB.Model(&ModelPricePolicy{}).Where("public_model = ?", "policy-model").Count(&count).Error)
	assert.EqualValues(t, 1, count, "upsert must not create a second row")

	var stored ModelPricePolicy
	require.NoError(t, DB.Where("public_model = ?", "policy-model").First(&stored).Error)
	assert.Equal(t, 3000, stored.TargetMarginBPS)
	assert.Equal(t, 400, stored.OverheadBPS)
}

// 全局默认策略：Enabled 的全局行胜出；ApplyGlobalDefaults 把空字段回填为全局值。
func TestGlobalPricePolicyAndDefaults(t *testing.T) {
	truncatePricingTables(t)

	global := &ModelPricePolicy{
		PublicModel: GlobalPricePolicyModel, ServiceTier: "default",
		Scope:           PricePolicyScopeGlobal,
		TargetMarginBPS: 2500, MinimumMarginBPS: 800,
		OverheadBPS: 500, FallbackProbabilityBPS: 200,
		AutoPublishChangeBPS: 700, Enabled: true,
	}
	require.NoError(t, UpsertModelPricePolicy(global))

	got, err := GlobalPricePolicy()
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 2500, got.TargetMarginBPS)

	policy := &ModelPricePolicy{PublicModel: "model-needs-defaults", ServiceTier: "stable"}
	policy.ApplyGlobalDefaults(got)
	assert.Equal(t, 2500, policy.TargetMarginBPS)
	assert.Equal(t, 800, policy.MinimumMarginBPS)
	assert.Equal(t, 500, policy.OverheadBPS)
	assert.Equal(t, 200, policy.FallbackProbabilityBPS)
}

// 同模型同档位只允许一个 pending 提案：新的把旧的顶成 rejected。
func TestReplacePendingModelPriceProposal(t *testing.T) {
	truncatePricingTables(t)

	now := common.GetTimestamp()
	first := &ModelPriceProposal{
		PublicModel: "proposal-model", ServiceTier: "stable", PolicyID: 901,
		ProposedInputPrice: 1.0, ProposedOutputPrice: 2.0,
		Status: PricingProposalPending, PricingVersion: "v1", CreatedAt: now,
	}
	require.NoError(t, ReplacePendingModelPriceProposal(first))

	second := &ModelPriceProposal{
		PublicModel: "proposal-model", ServiceTier: "stable", PolicyID: 901,
		ProposedInputPrice: 1.5, ProposedOutputPrice: 2.5,
		Status: PricingProposalPending, PricingVersion: "v1", CreatedAt: now + 1,
	}
	require.NoError(t, ReplacePendingModelPriceProposal(second))

	var statuses []string
	require.NoError(t, DB.Model(&ModelPriceProposal{}).Where("public_model = ?", "proposal-model").
		Order("created_at").Pluck("status", &statuses).Error)
	require.Len(t, statuses, 2)
	assert.Equal(t, PricingProposalRejected, statuses[0])
	assert.Equal(t, PricingProposalPending, statuses[1])
}

// 过期报价清理：只删过期的，未过期保留。
func TestPruneExpiredUpstreamModelOffers(t *testing.T) {
	truncatePricingTables(t)

	now := common.GetTimestamp()
	seedOffer(t, 901, "expired-model", "manual", 1.0, now-1)
	seedOffer(t, 902, "live-model", "manual", 1.0, now+3600)

	deleted, err := PruneExpiredUpstreamModelOffers(now)
	require.NoError(t, err)
	assert.EqualValues(t, 1, deleted)

	var remain int64
	require.NoError(t, DB.Model(&UpstreamModelOffer{}).Where("public_model = ?", "live-model").Count(&remain).Error)
	assert.EqualValues(t, 1, remain)
	require.NoError(t, DB.Model(&UpstreamModelOffer{}).Where("public_model = ?", "expired-model").Count(&remain).Error)
	assert.Zero(t, remain)
}

// 渠道成本校准：upsert 复合同一键、Map 汇总、Find 命中。
func TestUpstreamCostCalibrationRoundTrip(t *testing.T) {
	truncatePricingTables(t)

	cal := &UpstreamCostCalibration{
		ChannelID: 1001, PublicModel: "cal-model", UpstreamModel: "cal-up",
		EffectiveRatio: 1.25, PolicyRatio: 1.1, ObservedGroup: "default",
	}
	require.NoError(t, UpsertUpstreamCostCalibration(cal))
	cal.EffectiveRatio = 1.35
	require.NoError(t, UpsertUpstreamCostCalibration(cal))

	var count int64
	require.NoError(t, DB.Model(&UpstreamCostCalibration{}).Where("channel_id = ?", 1001).Count(&count).Error)
	assert.EqualValues(t, 1, count, "upsert must not duplicate")

	m, err := UpstreamCostCalibrationMap()
	require.NoError(t, err)
	key := UpstreamCostCalibrationKey(1001, "cal-model")
	require.Contains(t, m, key)
	assert.Equal(t, 1.35, m[key].EffectiveRatio)

	found, err := FindUpstreamCostCalibration(1001, "cal-model")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, 1.35, found.EffectiveRatio)

	missing, err := FindUpstreamCostCalibration(9999, "no-such-model")
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func ptrOf(v float64) *float64 { return &v }
