package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 邮箱掩码（日志 PII 防泄漏）：本地部分必须完全遮蔽，只允许域名可见。
func TestMaskEmail(t *testing.T) {
	assert.Equal(t, "***masked***", MaskEmail(""))
	assert.Equal(t, "***masked***", MaskEmail("not-an-email"))
	assert.Equal(t, "***@example.com", MaskEmail("alice@example.com"))
	assert.Equal(t, "***@sub.example.com", MaskEmail("a.b+c@sub.example.com"))
}

// 计费偏好归一化：四个合法值原样保留，其余一律回退 subscription_first。
// 该值参与扣费优先级决策，未知值不得放行。
func TestNormalizeBillingPreference(t *testing.T) {
	for _, valid := range []string{"subscription_first", "wallet_first", "subscription_only", "wallet_only"} {
		assert.Equal(t, valid, NormalizeBillingPreference(valid))
		assert.Equal(t, valid, NormalizeBillingPreference(" "+valid+" "))
	}
	assert.Equal(t, "subscription_first", NormalizeBillingPreference(""))
	assert.Equal(t, "subscription_first", NormalizeBillingPreference("wallet-first"))
	assert.Equal(t, "subscription_first", NormalizeBillingPreference("nonsense"))
}

func TestJSONShapeDetectors(t *testing.T) {
	assert.True(t, IsJsonArray(`[1,2,3]`))
	assert.True(t, IsJsonArray(`  [ ] `))
	assert.False(t, IsJsonArray(`{"a":1}`))
	assert.False(t, IsJsonArray(""))
	assert.False(t, IsJsonArray("not json"))

	assert.True(t, IsJsonObject(`{"a":1}`))
	assert.False(t, IsJsonObject(`[1,2]`))
	assert.False(t, IsJsonObject(""))

	arr, err := StrToJsonArray(`["a","b"]`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"a", "b"}, arr)
	_, err = StrToJsonArray("nope")
	assert.Error(t, err)

	m, err := StrToMap(`{"k":"v"}`)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"k": "v"}, m)
}

func TestString2IntAndContains(t *testing.T) {
	assert.Equal(t, 42, String2Int("42"))
	assert.Equal(t, 0, String2Int("abc"))
	assert.Equal(t, 0, String2Int(""))

	assert.True(t, StringsContains([]string{"a", "b"}, "b"))
	assert.False(t, StringsContains(nil, "b"))
}
