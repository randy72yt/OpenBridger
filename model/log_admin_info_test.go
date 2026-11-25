package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// admin_info 的可见性契约：普通用户日志视图必须剥离 admin_info 只留 op，
// 管理员视图必须保留。这是非管理员看不到操作者身份的强制保证。
func TestRecordLogWithAdminInfoVisibilitySplit(t *testing.T) {
	truncateTables(t)

	user := createReserveTestUser(t, 100)
	RecordLogWithAdminInfo(user.Id, LogTypeSystem, "op done",
		&AuditAdminInfo{AdminID: 42, AdminUsername: "root-admin", AdminRole: 10},
		&AuditOperation{Action: "update_ratio", Params: map[string]interface{}{"k": "v"}})

	// 用户视图：op 可见、admin_info 剥离
	logs, _, err := GetUserLogsForTest(user.Id)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Contains(t, logs[0].Other, `"op"`)
	assert.Contains(t, logs[0].Other, "update_ratio")
	assert.NotContains(t, logs[0].Other, "admin_info", "user view must strip admin_info")
	assert.NotContains(t, logs[0].Other, "root-admin")

	// 管理员视图：admin_info 保留
	adminLogs, _, err := GetAllLogsForTest()
	require.NoError(t, err)
	require.Len(t, adminLogs, 1)
	assert.Contains(t, adminLogs[0].Other, "admin_info")
	assert.Contains(t, adminLogs[0].Other, "root-admin")
}

// consume 日志受全局开关控制：关闭时不落库。
func TestRecordLogSkipsConsumeWhenDisabled(t *testing.T) {
	truncateTables(t)

	user := createReserveTestUser(t, 100)

	previous := common.LogConsumeEnabled
	common.LogConsumeEnabled = false
	t.Cleanup(func() { common.LogConsumeEnabled = previous })

	RecordLog(user.Id, LogTypeConsume, "consume entry")
	RecordLog(user.Id, LogTypeSystem, "system entry")

	var count int64
	require.NoError(t, DB.Model(&Log{}).Where("type = ?", LogTypeConsume).Count(&count).Error)
	assert.Zero(t, count)
	require.NoError(t, DB.Model(&Log{}).Where("type = ?", LogTypeSystem).Count(&count).Error)
	assert.EqualValues(t, 1, count)
}

func GetUserLogsForTest(userId int) ([]*Log, int64, error) {
	return GetUserLogs(userId, LogTypeUnknown, 0, 0, "", "", 0, 10, "", "", "")
}

func GetAllLogsForTest() ([]*Log, int64, error) {
	return GetAllLogs(LogTypeUnknown, 0, 0, "", "", "", 0, 10, 0, "", "", "")
}
