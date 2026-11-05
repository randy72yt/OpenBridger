package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 下单时记录的 CreditQuota 必须原样入账，不能被取整后的 Amount 覆盖。
// 回归背景：CNY 展示模式下 ¥10 的订单 Amount 被截成 1 美元，回调只入账
// 500000 额度（=¥7.2），用户少充 28%。
func TestEffectiveCreditQuotaUsesRecordedQuota(t *testing.T) {
	oldQuotaPerUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = oldQuotaPerUnit })

	topUp := &TopUp{
		Amount:          1,
		Money:           10,
		PaymentProvider: PaymentProviderEpay,
		CreditQuota:     694444,
	}
	quota, err := topUp.EffectiveCreditQuota()
	require.NoError(t, err)
	assert.Equal(t, 694444, quota)
}

func TestEffectiveCreditQuotaLegacyFallbacks(t *testing.T) {
	oldQuotaPerUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = oldQuotaPerUnit })

	cases := []struct {
		name  string
		topUp TopUp
		want  int
	}{
		{
			name:  "epay treats Amount as USD",
			topUp: TopUp{Amount: 10, PaymentProvider: PaymentProviderEpay},
			want:  5_000_000,
		},
		{
			name:  "stripe treats Money as USD",
			topUp: TopUp{Amount: 10, Money: 21.6, PaymentProvider: PaymentProviderStripe},
			want:  10_800_000,
		},
		{
			name:  "creem treats Amount as quota",
			topUp: TopUp{Amount: 6_944_444, PaymentProvider: PaymentProviderCreem},
			want:  6_944_444,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			quota, err := tc.topUp.EffectiveCreditQuota()
			require.NoError(t, err)
			assert.Equal(t, tc.want, quota)
		})
	}
}

// 回归：CNY 展示模式下 ¥10 的订单 Amount 只能存成 1 美元（整数字段），
// 若回调按 Amount 换算就只入账 ¥7.2。下单时记下的 CreditQuota 必须优先。
func TestRechargeEpayCreditsRecordedQuota(t *testing.T) {
	truncateTables(t)

	oldQuotaPerUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = oldQuotaPerUnit })

	user := insertUserForPaymentGuardTest(t, 777, 0)
	topUp := TopUp{
		UserId:          user.Id,
		Amount:          1, // 10 / 7.2 取整后的美元数
		Money:           10,
		TradeNo:         "EPAYCNY10",
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		CreditQuota:     694_444, // 10 / 7.2 * 500000
		CreateTime:      common.GetTimestamp(),
		Status:          common.TopUpStatusPending,
	}
	require.NoError(t, DB.Create(&topUp).Error)

	_, err := RechargeEpay(topUp.TradeNo, "alipay", "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, 694_444, getUserQuotaForPaymentGuardTest(t, user.Id))
}
