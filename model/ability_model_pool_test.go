package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func updatePool(t *testing.T, basicJSON, premiumJSON string) {
	t.Helper()
	cfg := config.GlobalConfig.Get("model_pool_setting")
	require.NotNil(t, cfg)
	require.NoError(t, config.UpdateConfigFromMap(cfg, map[string]string{
		"basic_pool_models": basicJSON,
		"premium_groups":    premiumJSON,
	}))
}

// TestModelPoolAbilityGating 覆盖模型池门禁的三个契约：
//  1. 未配置时零回归（全分组全模型）；
//  2. 配置后非高级分组只能拿到基础池模型；
//  3. 渠道能力重建（UpdateAbilities/FixAbility）从配置推导，结果一致。
func TestModelPoolAbilityGating(t *testing.T) {
	require.NoError(t, DB.AutoMigrate(&Ability{}, &Channel{}))
	t.Cleanup(func() {
		DB.Exec("DELETE FROM abilities")
		DB.Exec("DELETE FROM channels")
		updatePool(t, "[]", "[]")
	})

	channel := &Channel{
		Id: 9001, Name: "pool-test", Status: common.ChannelStatusEnabled,
		Models: "m-basic,m-premium", Group: "default,pro",
	}

	t.Run("unconfigured_allows_everything", func(t *testing.T) {
		require.NoError(t, DB.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error)
		require.NoError(t, channel.AddAbilities(nil))
		var count int64
		require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ?", channel.Id).Count(&count).Error)
		assert.Equal(t, int64(4), count, "未配置门禁时 2 模型 × 2 分组 = 4 条能力")
	})

	t.Run("configured_gates_non_premium_group", func(t *testing.T) {
		updatePool(t, `["m-basic"]`, `["pro"]`)
		t.Cleanup(func() { updatePool(t, "[]", "[]") })

		require.NoError(t, DB.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error)
		require.NoError(t, channel.AddAbilities(nil))
		var defaultCount int64
		require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ? AND `group` = ?", channel.Id, "default").Count(&defaultCount).Error)
		assert.Equal(t, int64(1), defaultCount, "default 组只应有基础池的 1 个模型")
		var proCount int64
		require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ? AND `group` = ?", channel.Id, "pro").Count(&proCount).Error)
		assert.Equal(t, int64(2), proCount, "pro 组应拿到全部模型")
		var premiumModelDefault int64
		require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ? AND `group` = ? AND model = ?", channel.Id, "default", "m-premium").Count(&premiumModelDefault).Error)
		assert.Equal(t, int64(0), premiumModelDefault, "非基础池模型不得对 default 组开放")

		// 重建路径（UpdateAbilities → 渠道保存/同步）必须推导出同样的结果。
		require.NoError(t, channel.UpdateAbilities(nil))
		var rebuilt int64
		require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ?", channel.Id).Count(&rebuilt).Error)
		assert.Equal(t, int64(3), rebuilt, "重建后 1×default + 2×pro = 3 条能力")
	})
}
