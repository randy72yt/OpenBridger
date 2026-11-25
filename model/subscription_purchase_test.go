package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedSubscriptionPlan(t *testing.T, id int, price float64, enabled bool, allowBalance bool) *SubscriptionPlan {
	t.Helper()
	allow := allowBalance
	plan := &SubscriptionPlan{
		Id:           id,
		Title:        "plan-" + common.GetRandomString(4),
		PriceAmount:  price,
		Currency:     "USD",
		DurationUnit: "month",
		DurationValue: 1,
		Enabled:      enabled,
		AllowBalancePay: &allow,
	}
	require.NoError(t, DB.Create(plan).Error)
	return plan
}

// 余额购买套餐：成功扣减余额；余额不足、套餐禁用、不允许余额支付全部拒绝。
func TestPurchaseSubscriptionWithBalance(t *testing.T) {
	truncateTables(t)

	plan := seedSubscriptionPlan(t, 1201, 10.0, true, true)
	expectedCost := int(10.0 * common.QuotaPerUnit)

	// 成功购买
	rich := createReserveTestUser(t, expectedCost*10)
	require.NoError(t, PurchaseSubscriptionWithBalance(rich.Id, plan.Id))
	assert.Equal(t, expectedCost*10-expectedCost, getUserQuotaFromDB(t, rich.Id))
	subs, err := GetAllUserSubscriptions(rich.Id)
	require.NoError(t, err)
	require.Len(t, subs, 1)
	require.NotNil(t, subs[0].Subscription)
	assert.Equal(t, plan.Id, subs[0].Subscription.PlanId)

	// 余额不足
	poor := createReserveTestUser(t, expectedCost-1)
	require.Error(t, PurchaseSubscriptionWithBalance(poor.Id, plan.Id))
	assert.Equal(t, expectedCost-1, getUserQuotaFromDB(t, poor.Id))
	poorSubs, err := GetAllUserSubscriptions(poor.Id)
	require.NoError(t, err)
	assert.Empty(t, poorSubs)

	// 套餐禁用
	disabled := seedSubscriptionPlan(t, 1202, 10.0, false, true)
	buyer := createReserveTestUser(t, expectedCost*10)
	require.Error(t, PurchaseSubscriptionWithBalance(buyer.Id, disabled.Id))
	assert.Equal(t, expectedCost*10, getUserQuotaFromDB(t, buyer.Id))

	// 不允许余额支付
	noBalance := seedSubscriptionPlan(t, 1203, 10.0, true, false)
	require.Error(t, PurchaseSubscriptionWithBalance(buyer.Id, noBalance.Id))
	assert.Equal(t, expectedCost*10, getUserQuotaFromDB(t, buyer.Id))

	// 无效参数
	require.Error(t, PurchaseSubscriptionWithBalance(0, plan.Id))
	require.Error(t, PurchaseSubscriptionWithBalance(buyer.Id, 0))
}
