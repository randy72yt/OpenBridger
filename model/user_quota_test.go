package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 配额增加走条件更新（quota <= MaxWalletQuota - quota），结果超过钱包上限时
// 必须失败，而不是静默截断或算术回绕成负数。
func TestIncreaseUserQuotaRejectsWalletOverflow(t *testing.T) {
	truncateTables(t)

	user := createReserveTestUser(t, common.MaxWalletQuota-100)

	require.Error(t, IncreaseUserQuota(user.Id, 101, true))
	assert.Equal(t, common.MaxWalletQuota-100, getUserQuotaFromDB(t, user.Id))

	require.NoError(t, IncreaseUserQuota(user.Id, 100, true))
	assert.Equal(t, common.MaxWalletQuota, getUserQuotaFromDB(t, user.Id))
}

func TestIncreaseUserQuotaRejectsInvalidInput(t *testing.T) {
	truncateTables(t)

	user := createReserveTestUser(t, 100)

	require.Error(t, IncreaseUserQuota(user.Id, -1, true))
	require.Error(t, IncreaseUserQuota(user.Id, common.MaxWalletQuota+1, true))
	require.Error(t, IncreaseUserQuota(99999999, 1, true)) // 不存在的用户
	assert.Equal(t, 100, getUserQuotaFromDB(t, user.Id))
}

func TestDecreaseUserQuota(t *testing.T) {
	truncateTables(t)

	user := createReserveTestUser(t, 1000)

	require.Error(t, DecreaseUserQuota(user.Id, -1, true))
	require.NoError(t, DecreaseUserQuota(user.Id, 300, true))
	assert.Equal(t, 700, getUserQuotaFromDB(t, user.Id))
}

func TestDeltaUpdateUserQuota(t *testing.T) {
	truncateTables(t)

	user := createReserveTestUser(t, 1000)

	require.NoError(t, DeltaUpdateUserQuota(user.Id, 0))
	assert.Equal(t, 1000, getUserQuotaFromDB(t, user.Id))

	require.NoError(t, DeltaUpdateUserQuota(user.Id, 500))
	assert.Equal(t, 1500, getUserQuotaFromDB(t, user.Id))

	require.NoError(t, DeltaUpdateUserQuota(user.Id, -200))
	assert.Equal(t, 1300, getUserQuotaFromDB(t, user.Id))
}

func TestGetUserQuotaAndUsedQuota(t *testing.T) {
	truncateTables(t)

	user := createReserveTestUser(t, 4321)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Update("used_quota", 123).Error)

	quota, err := GetUserQuota(user.Id, true)
	require.NoError(t, err)
	assert.Equal(t, 4321, quota)

	used, err := GetUserUsedQuota(user.Id)
	require.NoError(t, err)
	assert.Equal(t, 123, used)

	// GetUserQuota 用 Find 读取：用户不存在时返回 (0, nil) 而非错误。
	// 锁定该契约，调用方不能依赖错误判断用户是否存在。
	quota, err = GetUserQuota(99999999, true)
	require.NoError(t, err)
	assert.Zero(t, quota)
}

// 邀请额度转账：金额、余额、下限校验与成功路径。
func TestTransferAffQuotaToQuota(t *testing.T) {
	truncateTables(t)

	user := createReserveTestUser(t, 1000)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Update("aff_quota", int(common.QuotaPerUnit)*3).Error)

	// 低于最小转账额度
	require.Error(t, user.TransferAffQuotaToQuota(int(common.QuotaPerUnit)-1))

	// 邀请额度不足
	require.Error(t, user.TransferAffQuotaToQuota(int(common.QuotaPerUnit)*4))

	// 成功：aff 减少、quota 增加
	require.NoError(t, user.TransferAffQuotaToQuota(int(common.QuotaPerUnit)))
	var reloaded User
	require.NoError(t, DB.First(&reloaded, user.Id).Error)
	assert.Equal(t, int(common.QuotaPerUnit)*2, reloaded.AffQuota)
	assert.Equal(t, 1000+int(common.QuotaPerUnit), reloaded.Quota)
}

// 转账必须尊重钱包上限：quota 已接近 MaxWalletQuota 时，
// 再转入邀请额度应被拒绝，而不是把余额推过 JS 安全整数边界。
func TestTransferAffQuotaToQuotaRejectsWalletOverflow(t *testing.T) {
	truncateTables(t)

	user := createReserveTestUser(t, common.MaxWalletQuota-10)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Update("aff_quota", int(common.QuotaPerUnit)).Error)

	require.Error(t, user.TransferAffQuotaToQuota(int(common.QuotaPerUnit)))
	var reloaded User
	require.NoError(t, DB.First(&reloaded, user.Id).Error)
	assert.Equal(t, common.MaxWalletQuota-10, reloaded.Quota)
	assert.Equal(t, int(common.QuotaPerUnit), reloaded.AffQuota)
}

func TestValidateAndRevokeUserAccessToken(t *testing.T) {
	truncateTables(t)

	// 空 token：返回 (nil, nil) 表示"未提供"，不视为错误
	user, err := ValidateAccessToken("")
	require.NoError(t, err)
	assert.Nil(t, user)

	// Bearer 前缀剥离 + 未知 token 不暴露错误细节
	user, err = ValidateAccessToken("Bearer sk-nonexistent-token")
	require.NoError(t, err)
	assert.Nil(t, user)

	owner := createReserveTestUser(t, 100)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", owner.Id).
		Update("access_token", "sk-access-token-test").Error)

	user, err = ValidateAccessToken("sk-access-token-test")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, owner.Id, user.Id)

	user, err = ValidateAccessToken("Bearer sk-access-token-test")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, owner.Id, user.Id)

	// 撤销后立即失效
	_, err = RevokeUserAccessToken(owner.Id)
	require.NoError(t, err)
	user, err = ValidateAccessToken("sk-access-token-test")
	require.NoError(t, err)
	assert.Nil(t, user)

	var reloaded User
	require.NoError(t, DB.First(&reloaded, owner.Id).Error)
	assert.Empty(t, reloaded.AccessToken)
}
