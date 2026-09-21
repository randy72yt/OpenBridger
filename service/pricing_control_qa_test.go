package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func qaOffer(channelID int, publicModel string, in, out float64) model.UpstreamModelOffer {
	now := common.GetTimestamp()
	groupRatio := 1.0
	return model.UpstreamModelOffer{
		ChannelID: channelID, UpstreamModel: "upstream-" + publicModel, PublicModel: publicModel,
		InputCost: in, OutputCost: out, CacheReadCost: 0,
		UpstreamGroupRatio: &groupRatio,
		Currency:           "USD", SourceType: "manual", SourceVersion: "qa-v1",
		SuccessRateBPS: 9900, CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
	}
}

func qaPolicy(publicModel, tier string, primary, backup, target, minimum, overhead, fallback int) *model.ModelPricePolicy {
	return &model.ModelPricePolicy{
		PublicModel: publicModel, ServiceTier: tier,
		PrimaryChannelID: primary, BackupChannelID: backup,
		TargetMarginBPS: target, MinimumMarginBPS: minimum,
		OverheadBPS: overhead, FallbackProbabilityBPS: fallback,
		Enabled: true,
	}
}

// qaSeedOffers imports offers and fails the test immediately on error.
func qaSeedOffers(t *testing.T, offers ...model.UpstreamModelOffer) {
	t.Helper()
	require.NoError(t, ValidateAndImportPricingOffers(offers))
}

// qaBuild builds a proposal for the given policy, requiring success.
func qaBuild(t *testing.T, policy *model.ModelPricePolicy) *model.ModelPriceProposal {
	t.Helper()
	proposal, err := buildPriceProposal(policy)
	require.NoError(t, err)
	return proposal
}

func qaReport(t *testing.T, p *model.ModelPriceProposal) {
	t.Logf("实际结果: 成本=%.8f/%.8f 压力成本=%.8f/%.8f 建议售价=%.8f/%.8f 预期毛利=%dbps",
		p.ExpectedInputCost, p.ExpectedOutputCost, p.StressInputCost, p.StressOutputCost,
		p.ProposedInputPrice, p.ProposedOutputPrice, p.ExpectedMarginBPS)
}

// ---------------------------------------------------------------------------
// PC-CALC 计算正确性
// ---------------------------------------------------------------------------

func TestQAPricingCalc(t *testing.T) {
	t.Run("CALC-001 default档无备用渠道", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "calc-001", 1, 5))
		p := qaBuild(t, qaPolicy("calc-001", "default", 1, 0, 3500, 2000, 500, 0))
		qaReport(t, p)
		assert.InDelta(t, 1.05, p.ExpectedInputCost, 1e-9)
		assert.InDelta(t, 5.25, p.ExpectedOutputCost, 1e-9)
		assert.InDelta(t, 1.61538462, p.ProposedInputPrice, 1e-9)
		assert.InDelta(t, 8.07692308, p.ProposedOutputPrice, 1e-9)
		assert.Equal(t, 3500, p.ExpectedMarginBPS)
	})

	t.Run("CALC-002 default档有备用渠道", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "calc-002", 1, 5), qaOffer(2, "calc-002", 2, 10))
		p := qaBuild(t, qaPolicy("calc-002", "default", 1, 2, 3500, 2000, 500, 1000))
		qaReport(t, p)
		assert.InDelta(t, 1.155, p.ExpectedInputCost, 1e-9)
		assert.InDelta(t, 5.775, p.ExpectedOutputCost, 1e-9)
		assert.InDelta(t, 1.77692308, p.ProposedInputPrice, 1e-9)
		assert.InDelta(t, 8.88461538, p.ProposedOutputPrice, 1e-9)
		assert.Equal(t, 3500, p.ExpectedMarginBPS)
	})

	t.Run("CALC-003 stable档压力保护价生效", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "calc-003", 1, 5), qaOffer(2, "calc-003", 2, 10))
		p := qaBuild(t, qaPolicy("calc-003", "stable", 1, 2, 3500, 2000, 500, 1000))
		qaReport(t, p)
		assert.InDelta(t, 2.625, p.ProposedInputPrice, 1e-9)
		assert.InDelta(t, 13.125, p.ProposedOutputPrice, 1e-9)
		assert.Equal(t, 5600, p.ExpectedMarginBPS)
	})

	t.Run("CALC-004 stable档无备用渠道退化为default", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "calc-004", 1, 5))
		p := qaBuild(t, qaPolicy("calc-004", "stable", 1, 0, 3500, 2000, 500, 0))
		qaReport(t, p)
		assert.InDelta(t, 1.61538462, p.ProposedInputPrice, 1e-9)
		assert.InDelta(t, 8.07692308, p.ProposedOutputPrice, 1e-9)
	})

	t.Run("CALC-005 压力价抬高售价后毛利回算跟随", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "calc-005", 1, 5), qaOffer(2, "calc-005", 2, 10))
		p := qaBuild(t, qaPolicy("calc-005", "stable", 1, 2, 3500, 2000, 500, 1000))
		qaReport(t, p)
		assert.NotEqual(t, 3500, p.ExpectedMarginBPS, "售价被压力价抬高后，预期毛利不应仍等于目标毛利")
		assert.Equal(t, 5600, p.ExpectedMarginBPS)
	})

	t.Run("CALC-006 overhead为0", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "calc-006", 1, 5), qaOffer(2, "calc-006", 2, 10))
		p := qaBuild(t, qaPolicy("calc-006", "default", 1, 2, 3500, 2000, 0, 1000))
		qaReport(t, p)
		assert.InDelta(t, 1.1, p.ExpectedInputCost, 1e-9)
		assert.InDelta(t, 5.5, p.ExpectedOutputCost, 1e-9)
		assert.InDelta(t, 1.69230769, p.ProposedInputPrice, 1e-9)
		assert.InDelta(t, 8.46153846, p.ProposedOutputPrice, 1e-9)
	})

	t.Run("CALC-007 发布表达式与建议价一致", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "calc-007", 1, 5), qaOffer(2, "calc-007", 2, 10))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("calc-007", "default", 1, 2, 3500, 2000, 500, 1000)))
		created, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		require.Equal(t, 1, created)

		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, proposals, 1)
		proposal := proposals[0]

		require.NoError(t, ApprovePricingProposal(proposal.ID, 1))
		require.NoError(t, PublishPricingProposal(proposal.ID))

		pricing, err := model.GetModelPricingSnapshot([]string{"calc-007"})
		require.NoError(t, err)
		require.Len(t, pricing.Entries, 1)
		expr, ok := pricing.Entries[0].Configured["billing_setting.billing_expr"].(string)
		require.True(t, ok, "发布后应有计费表达式")
		t.Logf("实际结果: 建议价=%.8f/%.8f 发布表达式=%s",
			proposal.ProposedInputPrice, proposal.ProposedOutputPrice, expr)

		assert.Contains(t, expr, fmt.Sprintf("%v", proposal.ProposedInputPrice))
		assert.Contains(t, expr, fmt.Sprintf("%v", proposal.ProposedOutputPrice))
	})
}

// ---------------------------------------------------------------------------
// PC-EDGE 边界与异常输入
// ---------------------------------------------------------------------------

func TestQAPricingEdge(t *testing.T) {
	t.Run("EDGE-001 零成本渠道不得产出0售价却报目标毛利", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "edge-001", 0, 0))
		_, err := buildPriceProposal(qaPolicy("edge-001", "default", 1, 0, 3500, 2000, 500, 0))
		assert.ErrorIs(t, err, ErrPricingCostInvalid, "零成本报价必须阻止生成可审批方案")
	})

	t.Run("EDGE-017 上游分组倍率计入真实采购成本", func(t *testing.T) {
		truncate(t)
		offer := qaOffer(1, "edge-017", 0.22, 0.66)
		ratio := 6.8
		offer.UpstreamGroupRatio = &ratio
		qaSeedOffers(t, offer)
		p := qaBuild(t, qaPolicy("edge-017", "default", 1, 0, 3500, 2000, 0, 0))
		assert.InDelta(t, 1.496, p.ExpectedInputCost, 1e-9)
		assert.InDelta(t, 4.488, p.ExpectedOutputCost, 1e-9)
	})

	t.Run("EDGE-002 stable档备用渠道零成本", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "edge-002", 1, 5), qaOffer(2, "edge-002", 0, 0))
		p := qaBuild(t, qaPolicy("edge-002", "stable", 1, 2, 3500, 2000, 500, 1000))
		qaReport(t, p)
		assert.Greater(t, p.ProposedInputPrice, 0.0,
			"stable 档保护价应保证售价为正，当前备用渠道零成本时保护价退化为 0")
	})

	t.Run("EDGE-003 极低非零成本不下溢为0", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "edge-003", 0.000001, 0.000005))
		p := qaBuild(t, qaPolicy("edge-003", "default", 1, 0, 3500, 2000, 500, 0))
		qaReport(t, p)
		assert.Greater(t, p.ProposedInputPrice, 0.0)
		assert.Greater(t, p.ProposedOutputPrice, 0.0)
	})

	t.Run("EDGE-004 fallback为0", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "edge-004", 1, 5), qaOffer(2, "edge-004", 2, 10))

		def := qaBuild(t, qaPolicy("edge-004", "default", 1, 2, 3500, 2000, 0, 0))
		qaReport(t, def)
		assert.InDelta(t, 1.53846154, def.ProposedInputPrice, 1e-9, "fallback=0 时加权成本应等于主渠道成本")
		assert.InDelta(t, 7.69230769, def.ProposedOutputPrice, 1e-9)

		stable := qaBuild(t, qaPolicy("edge-004", "stable", 1, 2, 3500, 2000, 0, 0))
		qaReport(t, stable)
		assert.InDelta(t, 2.5, stable.ProposedInputPrice, 1e-9, "stable 档压力成本仍应取备用渠道成本")
		assert.InDelta(t, 12.5, stable.ProposedOutputPrice, 1e-9)
	})

	t.Run("EDGE-005 fallback为10000全部切备用", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "edge-005", 1, 5), qaOffer(2, "edge-005", 2, 10))
		p := qaBuild(t, qaPolicy("edge-005", "default", 1, 2, 3500, 2000, 0, 10000))
		qaReport(t, p)
		assert.InDelta(t, 3.07692308, p.ProposedInputPrice, 1e-9)
		assert.InDelta(t, 15.38461538, p.ProposedOutputPrice, 1e-9)
	})

	t.Run("EDGE-006 极端目标毛利9999不溢出", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "edge-006", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("edge-006", "default", 1, 0, 9999, 0, 0, 0)))
		p := qaBuild(t, qaPolicy("edge-006", "default", 1, 0, 9999, 0, 0, 0))
		qaReport(t, p)
		assert.False(t, math.IsInf(p.ProposedInputPrice, 0) || math.IsNaN(p.ProposedInputPrice))
		// 目标毛利 9999bps 时分母接近 1e-4，浮点误差会放大到 1e-8 量级，此处按 1e-7 容差判断。
		assert.InDelta(t, 10000, p.ProposedInputPrice, 1e-7)
		assert.InDelta(t, 50000, p.ProposedOutputPrice, 1e-7)
	})

	t.Run("EDGE-007 备用远贵于主时取保护价", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "edge-007", 1, 5), qaOffer(2, "edge-007", 50, 100))
		p := qaBuild(t, qaPolicy("edge-007", "stable", 1, 2, 3500, 2000, 0, 5000))
		qaReport(t, p)
		assert.InDelta(t, 62.5, p.ProposedInputPrice, 1e-9)
		assert.InDelta(t, 125.0, p.ProposedOutputPrice, 1e-9)
		assert.Equal(t, 2000, minimumPriceMarginBPS(p.ProposedInputPrice, p.ProposedOutputPrice, p.StressInputCost, p.StressOutputCost))
	})

	t.Run("EDGE-008 舍入后表达式无二次精度损失", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "edge-008", 1, 3))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("edge-008", "default", 1, 0, 3000, 2000, 0, 0)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, proposals, 1)
		require.NoError(t, ApprovePricingProposal(proposals[0].ID, 1))
		require.NoError(t, PublishPricingProposal(proposals[0].ID))

		pricing, err := model.GetModelPricingSnapshot([]string{"edge-008"})
		require.NoError(t, err)
		require.Len(t, pricing.Entries, 1)
		expr := pricing.Entries[0].Configured["billing_setting.billing_expr"].(string)
		t.Logf("实际结果: 建议价=%.8f 表达式=%s", proposals[0].ProposedInputPrice, expr)
		assert.Contains(t, expr, fmt.Sprintf("%v", proposals[0].ProposedInputPrice),
			"发布表达式中的价格必须与建议价逐位一致")
	})

	t.Run("EDGE-016 发布价格的计费单位为每百万token", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "calc-008", 1, 5), qaOffer(2, "calc-008", 2, 10))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("calc-008", "default", 1, 2, 3500, 2000, 500, 1000)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, proposals, 1)
		require.NoError(t, ApprovePricingProposal(proposals[0].ID, 1))
		require.NoError(t, PublishPricingProposal(proposals[0].ID))

		pricing, err := model.GetModelPricingSnapshot([]string{"calc-008"})
		require.NoError(t, err)
		require.Len(t, pricing.Entries, 1)
		expr := pricing.Entries[0].Configured["billing_setting.billing_expr"].(string)

		// RunExpr 返回 p*系数 的原始值，/1e6 的换算在 relay/helper/price.go 的调用方完成。
		rawCost, _, err := billingexpr.RunExpr(expr, billingexpr.TokenParams{P: 1_000_000, C: 1_000_000, Len: 1_000_000})
		require.NoError(t, err)
		expectedRaw := (proposals[0].ProposedInputPrice + proposals[0].ProposedOutputPrice) * 1_000_000
		t.Logf("实际结果: 表达式=%s 100万输入+100万输出 rawCost=%.8f 期望=%.8f 换算后费用=%.8f",
			expr, rawCost, expectedRaw, rawCost/1_000_000)
		assert.InDelta(t, expectedRaw, rawCost, 1e-6,
			"表达式系数应为每百万 token 价格，调用方再除以 1e6 得到实际费用")
		assert.InDelta(t, proposals[0].ProposedInputPrice+proposals[0].ProposedOutputPrice, rawCost/1_000_000, 1e-6,
			"换算后的费用应等于 100 万输入 + 100 万输出的建议售价之和")
	})

	t.Run("EDGE-009 缓存读取成本是否计入售价", func(t *testing.T) {
		truncate(t)
		now := common.GetTimestamp()
		withCache := qaOffer(1, "edge-009", 1, 5)
		withCache.CacheReadCost = 0.5
		require.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{withCache}))
		p := qaBuild(t, qaPolicy("edge-009", "default", 1, 0, 3500, 2000, 0, 0))

		without := qaOffer(1, "edge-009", 1, 5)
		without.CacheReadCost = 0
		_ = now
		qaReport(t, p)
		t.Logf("实际结果: CacheReadCost=0.5，建议售价=%.8f/%.8f（与缓存成本 0 时相同则未计入）",
			p.ProposedInputPrice, p.ProposedOutputPrice)
		assert.InDelta(t, 1.53846154, p.ProposedInputPrice, 1e-9,
			"缓存读取成本未参与计算时售价与缓存成本为 0 时一致，此处如实记录该行为")
	})

	t.Run("EDGE-010 非法成本值被拒", func(t *testing.T) {
		truncate(t)
		for _, cost := range []float64{-1, math.NaN(), math.Inf(1), math.Inf(-1)} {
			bad := qaOffer(9, "edge-010", cost, 5)
			err := ValidateAndImportPricingOffers([]model.UpstreamModelOffer{bad})
			assert.ErrorIs(t, err, ErrPricingOfferInvalid, "成本 %v 应被拒绝", cost)
		}
		var count int64
		require.NoError(t, model.DB.Model(&model.UpstreamModelOffer{}).Where("public_model = ?", "edge-010").Count(&count).Error)
		assert.Equal(t, int64(0), count, "非法成本不得入库")
	})

	t.Run("EDGE-011 币种校验", func(t *testing.T) {
		truncate(t)
		assert.ErrorIs(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{
			qaOfferWithCurrency(1, "edge-011", "CNY"),
		}), ErrPricingOfferInvalid, "CNY 应被拒绝")
		assert.ErrorIs(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{
			qaOfferWithCurrency(1, "edge-011", ""),
		}), ErrPricingOfferInvalid, "空币种应被拒绝")
		assert.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{
			qaOfferWithCurrency(1, "edge-011", "usd"),
		}), "小写 usd 应被规范化为 USD 并接受")
	})

	t.Run("EDGE-012 成功率越界被拒", func(t *testing.T) {
		truncate(t)
		for _, bps := range []int{-1, 10001} {
			o := qaOffer(1, "edge-012", 1, 5)
			o.SuccessRateBPS = bps
			assert.ErrorIs(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{o}),
				ErrPricingOfferInvalid, "success_rate_bps=%d 应被拒绝", bps)
		}
	})

	t.Run("EDGE-013 批量导入中部分非法应整批拒绝", func(t *testing.T) {
		truncate(t)
		first := qaOffer(1, "edge-013", 1, 5)
		bad := qaOffer(2, "edge-013", 1, 5)
		bad.Currency = "CNY"
		third := qaOffer(3, "edge-013", 1, 5)

		err := ValidateAndImportPricingOffers([]model.UpstreamModelOffer{first, bad, third})
		require.Error(t, err)

		var count int64
		require.NoError(t, model.DB.Model(&model.UpstreamModelOffer{}).Where("public_model = ?", "edge-013").Count(&count).Error)
		t.Logf("实际结果: 批量导入返回错误，但已入库条数=%d（期望 0）", count)
		assert.Equal(t, int64(0), count,
			"批量导入失败时不得留下部分入库的数据，当前实现为逐条 upsert 无回滚")
	})

	t.Run("EDGE-014 成本无变化时涨跌为0", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "edge-014", 1, 5), qaOffer(2, "edge-014", 2, 10))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("edge-014", "default", 1, 2, 3500, 2000, 500, 1000)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)

		_, err = RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		reviews, err := ListPricingProposalReviews(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, reviews, 1)
		if reviews[0].InputPriceChangeBPS != nil {
			t.Logf("实际结果: InputPriceChangeBPS=%d OutputPriceChangeBPS=%d PricingChanged=%v",
				*reviews[0].InputPriceChangeBPS, *reviews[0].OutputPriceChangeBPS, reviews[0].PricingChanged)
			assert.Equal(t, 0, *reviews[0].InputPriceChangeBPS)
			assert.Equal(t, 0, *reviews[0].OutputPriceChangeBPS)
		}
		assert.False(t, reviews[0].PricingChanged)
	})

	t.Run("EDGE-015 缺失必填字段被拒", func(t *testing.T) {
		truncate(t)
		noModel := qaOffer(1, "edge-015", 1, 5)
		noModel.PublicModel = ""
		assert.ErrorIs(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{noModel}), ErrPricingOfferInvalid)

		noChannel := qaOffer(0, "edge-015", 1, 5)
		assert.ErrorIs(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{noChannel}), ErrPricingOfferInvalid)

		noGroupRatio := qaOffer(1, "edge-015", 1, 5)
		noGroupRatio.UpstreamGroupRatio = nil
		assert.ErrorIs(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{noGroupRatio}), ErrPricingOfferInvalid,
			"缺少上游结算倍率时必须拒绝导入，不能暗中按 1 计算")

		assert.ErrorIs(t, ValidateAndImportPricingOffers(nil), ErrPricingOfferInvalid, "空批次应被拒绝")
	})
}

func TestQAPricingOfferSync(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{
          "success": true,
          "pricing_version": "upstream-v1",
          "auto_groups": ["default", "deepseek"],
          "group_ratio": {"default": 0.23, "deepseek": 6.8},
          "data": [
            {"model_name":"deepseek-chat","quota_type":0,"model_ratio":0.11,"completion_ratio":3,"cache_ratio":0.1,"enable_groups":["deepseek"]},
            {"model_name":"fixed-image","quota_type":1,"model_price":0.04,"enable_groups":["default"]}
          ]
        }`))
		require.NoError(t, err)
	}))
	defer server.Close()

	baseURL := server.URL + "/v1"
	mapping := `{"deepseek-v3":"deepseek-chat"}`
	channel := &model.Channel{
		Id: 42, Key: "test-key", BaseURL: &baseURL,
		Models: "deepseek-v3,fixed-image", ModelMapping: &mapping,
	}
	offers, warnings, err := fetchPricingOffersForChannel(context.Background(), channel)
	require.NoError(t, err)
	require.Len(t, offers, 1)
	assert.Equal(t, "Bearer test-key", authorization)
	assert.Equal(t, "deepseek-v3", offers[0].PublicModel)
	assert.Equal(t, "deepseek-chat", offers[0].UpstreamModel)
	assert.InDelta(t, 0.22, offers[0].InputCost, 1e-9)
	assert.InDelta(t, 0.66, offers[0].OutputCost, 1e-9)
	assert.InDelta(t, 0.022, offers[0].CacheReadCost, 1e-9)
	require.NotNil(t, offers[0].UpstreamGroupRatio)
	assert.InDelta(t, 6.8, *offers[0].UpstreamGroupRatio, 1e-9)
	assert.Equal(t, "upstream-v1", offers[0].SourceVersion)
	assert.Contains(t, warnings, "fixed-image: fixed-price model is not token-priced")
}

func qaOfferWithCurrency(channelID int, publicModel, currency string) model.UpstreamModelOffer {
	o := qaOffer(channelID, publicModel, 1, 5)
	o.Currency = currency
	return o
}

// ---------------------------------------------------------------------------
// PC-STATE 状态机
// ---------------------------------------------------------------------------

func TestQAPricingState(t *testing.T) {
	t.Run("STATE-001 正常流转", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "state-001", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("state-001", "default", 1, 0, 3500, 2000, 500, 0)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, proposals, 1)
		id := proposals[0].ID

		require.NoError(t, ApprovePricingProposal(id, 7))
		require.NoError(t, PublishPricingProposal(id))

		var final model.ModelPriceProposal
		require.NoError(t, model.DB.Where("id = ?", id).First(&final).Error)
		t.Logf("实际结果: status=%s approved_by=%d approved_at=%d published_at=%d",
			final.Status, final.ApprovedBy, final.ApprovedAt, final.PublishedAt)
		assert.Equal(t, model.PricingProposalPublished, final.Status)
		assert.Equal(t, 7, final.ApprovedBy)
		assert.Greater(t, final.ApprovedAt, int64(0))
		assert.Greater(t, final.PublishedAt, int64(0))
	})

	t.Run("STATE-002 重复发布被拒", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "state-002", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("state-002", "default", 1, 0, 3500, 2000, 500, 0)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, proposals, 1)

		require.NoError(t, ApprovePricingProposal(proposals[0].ID, 1))
		require.NoError(t, PublishPricingProposal(proposals[0].ID))
		assert.ErrorIs(t, PublishPricingProposal(proposals[0].ID), ErrPricingProposalState)
	})

	t.Run("STATE-003 未批准直接发布被拒", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "state-003", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("state-003", "default", 1, 0, 3500, 2000, 500, 0)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, proposals, 1)
		assert.ErrorIs(t, PublishPricingProposal(proposals[0].ID), ErrPricingProposalState)
	})

	t.Run("STATE-004 rejected后不可再批准或发布", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "state-004", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("state-004", "default", 1, 0, 3500, 2000, 500, 0)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, proposals, 1)

		require.NoError(t, RejectPricingProposal(proposals[0].ID))
		assert.ErrorIs(t, ApprovePricingProposal(proposals[0].ID, 1), ErrPricingProposalState)
		assert.ErrorIs(t, PublishPricingProposal(proposals[0].ID), ErrPricingProposalState)
	})

	t.Run("STATE-005 stable档方案发布被拒", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "state-005", 1, 5), qaOffer(2, "state-005", 2, 10))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("state-005", "stable", 1, 2, 3500, 2000, 500, 1000)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, proposals, 1)

		require.NoError(t, ApprovePricingProposal(proposals[0].ID, 1))
		err = PublishPricingProposal(proposals[0].ID)
		require.Error(t, err)
		t.Logf("实际结果: stable 档发布被拒，错误信息=%q", err.Error())
		assert.Contains(t, strings.ToLower(err.Error()), "default",
			"错误信息应说明仅 default 档可发布，便于前端展示")
	})

	t.Run("STATE-006 价格锁定不生成方案", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "state-006", 1, 5))
		policy := qaPolicy("state-006", "default", 1, 0, 3500, 2000, 500, 0)
		policy.PriceLocked = true
		require.NoError(t, ValidateAndUpsertPricePolicy(policy))
		created, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, created)
	})

	t.Run("STATE-007 策略禁用不生成方案", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "state-007", 1, 5))
		policy := qaPolicy("state-007", "default", 1, 0, 3500, 2000, 500, 0)
		policy.Enabled = false
		require.NoError(t, ValidateAndUpsertPricePolicy(policy))
		created, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, created)
	})

	t.Run("STATE-008 连续重算不累积pending方案", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "state-008", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("state-008", "default", 1, 0, 3500, 2000, 500, 0)))
		for i := 0; i < 3; i++ {
			_, err := RecalculatePricingProposals(context.Background())
			require.NoError(t, err)
		}
		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		t.Logf("实际结果: 连续重算 3 次后 pending 方案数=%d", len(proposals))
		assert.Len(t, proposals, 1)
	})

	t.Run("STATE-009 发布后定价模式为tiered_expr", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "state-009", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("state-009", "default", 1, 0, 3500, 2000, 500, 0)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, proposals, 1)
		require.NoError(t, ApprovePricingProposal(proposals[0].ID, 1))
		require.NoError(t, PublishPricingProposal(proposals[0].ID))

		pricing, err := model.GetModelPricingSnapshot([]string{"state-009"})
		require.NoError(t, err)
		require.Len(t, pricing.Entries, 1)
		mode := pricing.Entries[0].Configured["billing_setting.billing_mode"]
		t.Logf("实际结果: billing_mode=%v version=%s", mode, pricing.Entries[0].Version)
		assert.Equal(t, "tiered_expr", mode)
	})
}

// ---------------------------------------------------------------------------
// PC-CONC 过期、并发与风险
// ---------------------------------------------------------------------------

func TestQAPricingConcurrency(t *testing.T) {
	t.Run("CONC-001 报价过期后不得批准旧方案", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "conc-001", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("conc-001", "default", 1, 0, 3500, 2000, 500, 0)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, proposals, 1)

		now := common.GetTimestamp()
		require.NoError(t, model.DB.Model(&model.UpstreamModelOffer{}).
			Where("public_model = ?", "conc-001").Update("expires_at", now-1).Error)

		err = ApprovePricingProposal(proposals[0].ID, 1)
		if err == nil {
			t.Logf("实际结果: 报价已过期，批准仍成功（proposal id=%d）", proposals[0].ID)
		}
		assert.Error(t, err, "DEC-004 要求过期方案禁止审批")
	})

	t.Run("CONC-002 价格版本变化后不得批准", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "conc-002", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("conc-002", "default", 1, 0, 3500, 2000, 500, 0)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		proposals, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, proposals, 1)
		staleVersion := proposals[0].PricingVersion

		require.NoError(t, model.UpdateModelPricing([]model.ModelPricingChange{{
			ModelName: "conc-002", ExpectedVersion: staleVersion,
			Pricing: model.PricingValues{
				"billing_setting.billing_mode": "tiered_expr",
				"billing_setting.billing_expr": `tier("base", p * 9 + c * 9)`,
			},
		}}))

		err = ApprovePricingProposal(proposals[0].ID, 1)
		if err == nil {
			t.Logf("实际结果: PricingVersion 已由 %q 变化，旧方案批准仍成功", staleVersion)
		}
		assert.Error(t, err, "DEC-004 要求价格版本变化的方案禁止审批")
	})

	t.Run("CONC-003 并发发布只有一个生效", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "conc-003", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("conc-003", "default", 1, 0, 3500, 2000, 500, 0)))
		_, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)

		base, err := model.ListModelPriceProposals(model.PricingProposalPending)
		require.NoError(t, err)
		require.Len(t, base, 1)

		// 两个方案基于同一价格版本，模拟两个管理员同时发布。
		second := base[0]
		second.ID = 0
		second.ProposedInputPrice = 99
		second.ProposedOutputPrice = 99
		require.NoError(t, model.DB.Create(&second).Error)
		require.NoError(t, ApprovePricingProposal(base[0].ID, 1))
		require.NoError(t, ApprovePricingProposal(second.ID, 1))

		var wg sync.WaitGroup
		errs := make([]error, 2)
		ids := []int64{base[0].ID, second.ID}
		for i := range ids {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				errs[idx] = PublishPricingProposal(ids[idx])
			}(i)
		}
		wg.Wait()

		success := 0
		for i, e := range errs {
			if e == nil {
				success++
			} else {
				t.Logf("实际结果: 发布 id=%d 失败: %v", ids[i], e)
			}
		}
		t.Logf("实际结果: 并发发布成功数=%d", success)

		pricing, err := model.GetModelPricingSnapshot([]string{"conc-003"})
		require.NoError(t, err)
		require.Len(t, pricing.Entries, 1)
		expr := pricing.Entries[0].Configured["billing_setting.billing_expr"].(string)
		t.Logf("实际结果: 最终表达式=%s（两个候选价分别为 %.8f 和 99）", expr, base[0].ProposedInputPrice)
		assert.Equal(t, 1, success, "基于同一价格版本的并发发布只应有一个生效")
		assert.True(t,
			strings.Contains(expr, fmt.Sprintf("%v", base[0].ProposedInputPrice)) || strings.Contains(expr, "99"),
			"最终价格必须是两个候选价之一，不能是混合或未写入")
	})

	t.Run("CONC-004 导入空批次不覆盖有效成本", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "conc-004", 1, 5))
		assert.Error(t, ValidateAndImportPricingOffers(nil))

		var offer model.UpstreamModelOffer
		require.NoError(t, model.DB.Where("public_model = ?", "conc-004").First(&offer).Error)
		t.Logf("实际结果: 空批次导入失败后，有效成本仍为 %.8f/%.8f", offer.InputCost, offer.OutputCost)
		assert.InDelta(t, 1, offer.InputCost, 1e-9)
		assert.InDelta(t, 5, offer.OutputCost, 1e-9)
	})

	t.Run("CONC-005 主渠道报价过期进风险列表", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "conc-005", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("conc-005", "default", 1, 0, 3500, 2000, 500, 0)))

		now := common.GetTimestamp()
		require.NoError(t, model.DB.Model(&model.UpstreamModelOffer{}).
			Where("public_model = ?", "conc-005").Update("expires_at", now-1).Error)

		created, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, created, "主渠道报价过期时不应生成方案")

		risks, err := PricingRisks()
		require.NoError(t, err)
		require.Len(t, risks, 1)
		t.Logf("实际结果: 风险码=%v", risks[0]["code"])
		assert.Equal(t, "primary_offer_missing_or_expired", risks[0]["code"])
	})

	t.Run("CONC-006 备用渠道报价过期进风险列表", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "conc-006", 1, 5), qaOffer(2, "conc-006", 2, 10))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("conc-006", "stable", 1, 2, 3500, 2000, 500, 1000)))

		now := common.GetTimestamp()
		require.NoError(t, model.DB.Model(&model.UpstreamModelOffer{}).
			Where("public_model = ? AND channel_id = ?", "conc-006", 2).Update("expires_at", now-1).Error)

		created, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, created, "备用渠道报价过期时 stable 档不应生成方案")

		risks, err := PricingRisks()
		require.NoError(t, err)
		require.Len(t, risks, 1)
		t.Logf("实际结果: 风险码=%v", risks[0]["code"])
		assert.Equal(t, "backup_offer_missing_or_expired", risks[0]["code"])
	})

	t.Run("CONC-007 disabled报价等同缺失", func(t *testing.T) {
		truncate(t)
		qaSeedOffers(t, qaOffer(1, "conc-007", 1, 5))
		require.NoError(t, ValidateAndUpsertPricePolicy(qaPolicy("conc-007", "default", 1, 0, 3500, 2000, 500, 0)))
		require.NoError(t, model.DB.Model(&model.UpstreamModelOffer{}).
			Where("public_model = ?", "conc-007").Update("enabled", false).Error)

		created, err := RecalculatePricingProposals(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, created)
	})

	t.Run("CONC-008 同模型同渠道取最新报价", func(t *testing.T) {
		truncate(t)
		now := common.GetTimestamp()
		older := qaOffer(1, "conc-008", 1, 5)
		older.CollectedAt = now - 100
		older.SourceVersion = "older"
		require.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{older}))

		newer := qaOffer(1, "conc-008", 3, 15)
		newer.CollectedAt = now
		newer.SourceVersion = "newer"
		require.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{newer}))

		p := qaBuild(t, qaPolicy("conc-008", "default", 1, 0, 3500, 2000, 0, 0))
		qaReport(t, p)
		assert.InDelta(t, 3, p.ExpectedInputCost, 1e-9, "应取 collected_at 最新的报价")
		assert.InDelta(t, 15, p.ExpectedOutputCost, 1e-9)
	})
}
