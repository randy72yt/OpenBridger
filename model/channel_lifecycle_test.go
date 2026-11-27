package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedChannel(t *testing.T, id int, tag string, status int, multiKey bool, keys string) *Channel {
	t.Helper()
	baseURL := ""
	tagRef := tag
	weight := uint(0)
	channel := &Channel{
		Id:      id,
		Type:    constant.ChannelTypeOpenAI,
		Key:     keys,
		Name:    "chan-" + tag,
		BaseURL: &baseURL,
		Tag:     &tagRef,
		Status:  status,
		Weight:  &weight,
	}
	if multiKey {
		channel.ChannelInfo = ChannelInfo{IsMultiKey: true, MultiKeySize: 2, MultiKeyMode: constant.MultiKeyModePolling}
	}
	require.NoError(t, channel.Insert())
	return channel
}

// 多 key 轮换：非多 key 模式返回原始 key；多 key 模式连续取用轮换 index；
// 没有 key 时返回可操作的错误（调用方据此禁用渠道）。
func TestGetNextEnabledKeyRotation(t *testing.T) {
	truncateTables(t)

	single := seedChannel(t, 1101, "rot-single", common.ChannelStatusEnabled, false, "sk-single-key")
	key, idx, apiErr := single.GetNextEnabledKey()
	require.Nil(t, apiErr)
	assert.Equal(t, "sk-single-key", key)
	assert.Zero(t, idx)

	multi := seedChannel(t, 1102, "rot-multi", common.ChannelStatusEnabled, true, "key-a\nkey-b")
	first, firstIdx, apiErr := multi.GetNextEnabledKey()
	require.Nil(t, apiErr)
	assert.Contains(t, []string{"key-a", "key-b"}, first)
	_, secondIdx, apiErr := multi.GetNextEnabledKey()
	require.Nil(t, apiErr)
	assert.Equal(t, 1, firstIdx+secondIdx, "两次取用应覆盖两个不同的 key 槽位")

	empty := seedChannel(t, 1103, "rot-empty", common.ChannelStatusEnabled, true, "")
	_, _, apiErr = empty.GetNextEnabledKey()
	require.NotNil(t, apiErr)
}

// 按标签批量启停：状态与能力(abilities)同步切换。
func TestChannelTagEnableDisable(t *testing.T) {
	truncateTables(t)

	on := seedChannel(t, 1104, "tagops", common.ChannelStatusEnabled, false, "k")
	off := seedChannel(t, 1105, "tagops", common.ChannelStatusAutoDisabled, false, "k")

	require.NoError(t, EnableChannelByTag("tagops"))
	var statuses []int
	require.NoError(t, DB.Model(&Channel{}).Where("tag = ?", "tagops").Order("id").Pluck("status", &statuses).Error)
	assert.Equal(t, []int{common.ChannelStatusEnabled, common.ChannelStatusEnabled}, statuses)

	require.NoError(t, DisableChannelByTag("tagops"))
	require.NoError(t, DB.Model(&Channel{}).Where("tag = ?", "tagops").Order("id").Pluck("status", &statuses).Error)
	assert.Equal(t, []int{common.ChannelStatusManuallyDisabled, common.ChannelStatusManuallyDisabled}, statuses)

	channels, err := GetChannelsByTag("tagops", true, true)
	require.NoError(t, err)
	assert.Len(t, channels, 2)
	_ = on
	_ = off
}

// 批量删除只删给定 id，不误伤其他渠道。
func TestBatchDeleteChannels(t *testing.T) {
	truncateTables(t)

	seedChannel(t, 1106, "del-a", common.ChannelStatusEnabled, false, "k")
	keep := seedChannel(t, 1107, "del-b", common.ChannelStatusEnabled, false, "k")

	deleted, err := BatchDeleteChannels([]int{1106})
	require.NoError(t, err)
	assert.EqualValues(t, 1, deleted)

	var remain int64
	require.NoError(t, DB.Model(&Channel{}).Count(&remain).Error)
	assert.EqualValues(t, 1, remain)
	require.NoError(t, DB.Model(&Channel{}).Where("id = ?", keep.Id).Count(&remain).Error)
	assert.EqualValues(t, 1, remain)
}

// 按状态清理：手动禁用与自动禁用的渠道都删除，启用渠道保留。
func TestDeleteDisabledChannel(t *testing.T) {
	truncateTables(t)

	seedChannel(t, 1108, "clean-enabled", common.ChannelStatusEnabled, false, "k")
	seedChannel(t, 1109, "clean-manual", common.ChannelStatusManuallyDisabled, false, "k")
	seedChannel(t, 1110, "clean-auto", common.ChannelStatusAutoDisabled, false, "k")

	deleted, err := DeleteDisabledChannel()
	require.NoError(t, err)
	assert.EqualValues(t, 2, deleted)

	var statuses []int
	require.NoError(t, DB.Model(&Channel{}).Order("id").Pluck("status", &statuses).Error)
	assert.Equal(t, []int{common.ChannelStatusEnabled}, statuses)
}
