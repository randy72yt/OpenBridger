package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 掩码不得泄露 key 中段：短 key 全遮，长 key 只露首尾。
func TestMaskTokenKey(t *testing.T) {
	assert.Empty(t, MaskTokenKey(""))
	assert.Equal(t, "****", MaskTokenKey("abcd"))
	assert.Equal(t, "ab****gh", MaskTokenKey("abcdefgh"))
	assert.Equal(t, "sk-1**********xyz9", MaskTokenKey("sk-1234567890abcdefghxyz9"))
}

func TestTokenQuotaIncreaseDecrease(t *testing.T) {
	truncateTables(t)

	token := createReserveTestToken(t, 10_000)
	require.NoError(t, DB.Model(&Token{}).Where("id = ?", token.Id).
		Update("used_quota", 2000).Error)

	require.Error(t, IncreaseTokenQuota(token.Id, token.Key, -1))
	require.Error(t, DecreaseTokenQuota(token.Id, token.Key, -1))

	require.NoError(t, DecreaseTokenQuota(token.Id, token.Key, 1500))
	var reloaded Token
	require.NoError(t, DB.First(&reloaded, token.Id).Error)
	assert.Equal(t, 8500, reloaded.RemainQuota)
	assert.Equal(t, 3500, reloaded.UsedQuota)

	require.NoError(t, IncreaseTokenQuota(token.Id, token.Key, 500))
	require.NoError(t, DB.First(&reloaded, token.Id).Error)
	assert.Equal(t, 9000, reloaded.RemainQuota)
	assert.Equal(t, 3000, reloaded.UsedQuota)
	assert.Greater(t, reloaded.AccessedTime, int64(0))
}

// 令牌校验：状态与过期时间的每个拒绝分支。
func TestValidateUserToken(t *testing.T) {
	truncateTables(t)

	valid := createReserveTestToken(t, 1000)
	token, err := ValidateUserToken(valid.Key)
	require.NoError(t, err)
	assert.Equal(t, valid.Id, token.Id)

	_, err = ValidateUserToken("")
	require.ErrorIs(t, err, ErrTokenNotProvided)

	_, err = ValidateUserToken("sk-nonexistent-key")
	require.ErrorIs(t, err, ErrTokenInvalid)

	// 禁用状态
	disabled := createReserveTestToken(t, 1000)
	require.NoError(t, DB.Model(&Token{}).Where("id = ?", disabled.Id).
		Update("status", common.TokenStatusDisabled).Error)
	_, err = ValidateUserToken(disabled.Key)
	require.ErrorIs(t, err, ErrTokenInvalid)

	// 已过期（ExpiredTime < 当前时间）
	expired := createReserveTestToken(t, 1000)
	require.NoError(t, DB.Model(&Token{}).Where("id = ?", expired.Id).
		Update("expired_time", common.GetTimestamp()-1).Error)
	_, err = ValidateUserToken(expired.Key)
	require.ErrorIs(t, err, ErrTokenInvalid)
}

// IP 白名单解析：nil、空串、多行混合空格与逗号。
func TestGetIpLimits(t *testing.T) {
	truncateTables(t)

	empty := Token{}
	assert.Empty(t, empty.GetIpLimits())

	space := "  "
	token := Token{AllowIps: &space}
	assert.Empty(t, token.GetIpLimits())

	multi := "1.2.3.4\n 5.6.7.8 ,\n\n9.9.9.9"
	token = Token{AllowIps: &multi}
	assert.Equal(t, []string{"1.2.3.4", "5.6.7.8", "9.9.9.9"}, token.GetIpLimits())
}

func TestCountUserTokensAndBatchDelete(t *testing.T) {
	truncateTables(t)

	alice := createReserveTestUser(t, 100)
	bob := createReserveTestUser(t, 100)
	keep := createReserveTestToken(t, 10_000)
	require.NoError(t, DB.Model(&Token{}).Where("id = ?", keep.Id).Update("user_id", alice.Id).Error)
	t1 := createReserveTestToken(t, 10_000)
	t2 := createReserveTestToken(t, 10_000)
	require.NoError(t, DB.Model(&Token{}).Where("id IN ?", []int{t1.Id, t2.Id}).Update("user_id", bob.Id).Error)

	total, err := CountUserTokens(bob.Id)
	require.NoError(t, err)
	assert.EqualValues(t, 2, total)

	// 只删 bob 的令牌；试图带上 alice 的 id 也不得删除
	deleted, err := BatchDeleteTokens([]int{t1.Id, t2.Id, keep.Id}, bob.Id)
	require.NoError(t, err)
	assert.Equal(t, 2, deleted)

	var remain int64
	require.NoError(t, DB.Model(&Token{}).Where("user_id = ?", bob.Id).Count(&remain).Error)
	assert.Zero(t, remain)
	require.NoError(t, DB.Model(&Token{}).Where("id = ?", keep.Id).Count(&remain).Error)
	assert.EqualValues(t, 1, remain)
}
