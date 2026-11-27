package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func seedTopUpUser(t *testing.T, quota int) User {
	t.Helper()
	return createReserveTestUser(t, quota)
}

func seedTopUpOrder(t *testing.T, userId int, tradeNo string, status string, provider string, creditQuota int64) *TopUp {
	t.Helper()
	topUp := &TopUp{
		UserId:          userId,
		Amount:          10,
		Money:           10,
		TradeNo:         tradeNo,
		PaymentMethod:   provider,
		PaymentProvider: provider,
		CreditQuota:     creditQuota,
		CreateTime:      common.GetTimestamp(),
		Status:          status,
	}
	require.NoError(t, topUp.Insert())
	return topUp
}

// 支付回调结算：pending 的 stripe 订单成功入账，重复回调幂等拒绝。
func TestRechargeCompletesPendingStripeOrder(t *testing.T) {
	truncateTables(t)

	user := seedTopUpUser(t, 1000)
	seedTopUpOrder(t, user.Id, "recharge-ok-1", common.TopUpStatusPending, PaymentProviderStripe, 5000)

	require.NoError(t, Recharge("recharge-ok-1", "cus_test", "127.0.0.1"))
	assert.Equal(t, 1000+5000, getUserQuotaFromDB(t, user.Id))

	var reloaded TopUp
	require.NoError(t, DB.Where("trade_no = ?", "recharge-ok-1").First(&reloaded).Error)
	assert.Equal(t, common.TopUpStatusSuccess, reloaded.Status)
	assert.Greater(t, reloaded.CompleteTime, int64(0))

	// 幂等：已成功的订单再次回调必须失败且不重复入账
	require.Error(t, Recharge("recharge-ok-1", "cus_test", "127.0.0.1"))
	assert.Equal(t, 1000+5000, getUserQuotaFromDB(t, user.Id))
}

// 无效订单全部拒绝：不存在、provider 不匹配、状态不是 pending。
func TestRechargeRejectsInvalidOrders(t *testing.T) {
	truncateTables(t)

	user := seedTopUpUser(t, 1000)

	require.Error(t, Recharge("recharge-missing", "cus_test", "127.0.0.1"))

	// 非 stripe 渠道的订单不能走 stripe 回调（防止跨渠道串单）
	creem := seedTopUpOrder(t, user.Id, "recharge-creem-1", common.TopUpStatusPending, PaymentProviderCreem, 5000)
	require.Error(t, Recharge("recharge-creem-1", "cus_test", "127.0.0.1"))
	var reloaded TopUp
	require.NoError(t, DB.Where("trade_no = ?", "recharge-creem-1").First(&reloaded).Error)
	assert.Equal(t, common.TopUpStatusPending, reloaded.Status) // 订单保持待支付
	_ = creem

	// 已过期订单不可再充值
	seedTopUpOrder(t, user.Id, "recharge-expired-1", "expired", PaymentProviderStripe, 5000)
	require.Error(t, Recharge("recharge-expired-1", "cus_test", "127.0.0.1"))
	assert.Equal(t, 1000, getUserQuotaFromDB(t, user.Id))
}

// 入账必须尊重钱包上限：余额已接近上限时回调失败，事务回滚订单保持 pending。
func TestRechargeRejectsWalletOverflow(t *testing.T) {
	truncateTables(t)

	user := seedTopUpUser(t, common.MaxWalletQuota-100)
	seedTopUpOrder(t, user.Id, "recharge-overflow-1", common.TopUpStatusPending, PaymentProviderStripe, 5000)

	require.Error(t, Recharge("recharge-overflow-1", "cus_test", "127.0.0.1"))
	assert.Equal(t, common.MaxWalletQuota-100, getUserQuotaFromDB(t, user.Id))

	// 结算失败必须回滚订单状态，否则丢单
	var reloaded TopUp
	require.NoError(t, DB.Where("trade_no = ?", "recharge-overflow-1").First(&reloaded).Error)
	assert.Equal(t, common.TopUpStatusPending, reloaded.Status)
}

// 下单前的用户可见预检：余额加本次充值超过钱包上限时拒绝。
func TestValidateTopUpQuotaCapacity(t *testing.T) {
	truncateTables(t)

	user := seedTopUpUser(t, common.MaxWalletQuota-5000+1)
	require.ErrorIs(t, ValidateTopUpQuotaCapacity(user.Id, 5000), ErrTopUpQuotaLimitExceeded)

	okUser := seedTopUpUser(t, common.MaxWalletQuota-5000)
	require.NoError(t, ValidateTopUpQuotaCapacity(okUser.Id, 5000))

	require.ErrorIs(t, ValidateTopUpQuotaCapacity(99999999, 5000), gorm.ErrRecordNotFound)
}

// 管理员补单：成功入账、已成功幂等、状态不符拒绝。
func TestManualCompleteTopUp(t *testing.T) {
	truncateTables(t)

	user := seedTopUpUser(t, 1000)
	seedTopUpOrder(t, user.Id, "manual-complete-1", common.TopUpStatusPending, PaymentProviderCreem, 3000)

	require.NoError(t, ManualCompleteTopUp("manual-complete-1", "127.0.0.1"))
	assert.Equal(t, 1000+3000, getUserQuotaFromDB(t, user.Id))

	// 幂等：已成功订单补单返回 nil 且不重复入账
	require.NoError(t, ManualCompleteTopUp("manual-complete-1", "127.0.0.1"))
	assert.Equal(t, 1000+3000, getUserQuotaFromDB(t, user.Id))

	// 状态不符
	seedTopUpOrder(t, user.Id, "manual-complete-2", "expired", PaymentProviderCreem, 3000)
	require.Error(t, ManualCompleteTopUp("manual-complete-2", "127.0.0.1"))

	require.Error(t, ManualCompleteTopUp("manual-complete-missing", "127.0.0.1"))
}

// 用户侧查询按属主过滤并受 30 天窗口限制；管理员侧无窗口。
func TestGetUserTopUpsScopesOwnerAndWindow(t *testing.T) {
	truncateTables(t)

	alice := seedTopUpUser(t, 100)
	bob := seedTopUpUser(t, 100)
	now := common.GetTimestamp()
	window := int64(30 * 24 * 60 * 60)

	recent := seedTopUpOrder(t, alice.Id, "query-recent-1", common.TopUpStatusSuccess, PaymentProviderStripe, 100)
	old := seedTopUpOrder(t, alice.Id, "query-old-1", common.TopUpStatusSuccess, PaymentProviderStripe, 100)
	require.NoError(t, DB.Model(&TopUp{}).Where("id = ?", old.Id).Update("create_time", now-window-1).Error)
	seedTopUpOrder(t, bob.Id, "query-bob-1", common.TopUpStatusSuccess, PaymentProviderStripe, 100)

	pageInfo := &common.PageInfo{Page: 1, PageSize: 10}
	topups, total, err := GetUserTopUps(alice.Id, pageInfo)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, topups, 1)
	assert.Equal(t, recent.TradeNo, topups[0].TradeNo)

	all, allTotal, err := GetAllTopUps(pageInfo)
	require.NoError(t, err)
	assert.EqualValues(t, 3, allTotal)
	assert.Len(t, all, 3)
}
