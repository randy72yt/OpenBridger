/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TransferAffQuotaToQuota moves the caller's own aff_quota into quota.
// It is not a withdrawal and it has no idempotency key: two transfers of a
// legal amount both succeed when the balance covers both. A second transfer
// fails only when the remaining aff_quota is too small.

func TestTransferAffQuotaConservesBalance(t *testing.T) {
	for _, kind := range []string{"sqlite", "mysql", "postgres"} {
		t.Run(kind, func(t *testing.T) {
			useAffTransferDatabase(t, kind)
			unit := int(common.QuotaPerUnit)

			t.Run("moves aff quota into the same account", func(t *testing.T) {
				user := insertAffTransferUser(t, 1_000, unit*3)
				require.NoError(t, (&model.User{Id: user.Id}).TransferAffQuotaToQuota(unit))
				require.NoError(t, (&model.User{Id: user.Id}).TransferAffQuotaToQuota(unit*2))
				stored := reloadAffTransferUser(t, user.Id)
				assert.Equal(t, 0, stored.AffQuota)
				assert.Equal(t, 1_000+unit*3, stored.Quota)
				assert.Equal(t, user.AffHistoryQuota, stored.AffHistoryQuota)
			})

			t.Run("rejects amounts below the minimum or above the balance", func(t *testing.T) {
				user := insertAffTransferUser(t, 40, unit*2)
				before := reloadAffTransferUser(t, user.Id)
				for _, amount := range []int{0, -unit, unit - 1} {
					err := (&model.User{Id: user.Id}).TransferAffQuotaToQuota(amount)
					require.Error(t, err)
					assert.Contains(t, err.Error(), "转移额度最小为")
				}
				err := (&model.User{Id: user.Id}).TransferAffQuotaToQuota(unit*2 + 1)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "邀请额度不足")
				// A second independent request for the remaining balance is legal.
				require.NoError(t, (&model.User{Id: user.Id}).TransferAffQuotaToQuota(unit))
				err = (&model.User{Id: user.Id}).TransferAffQuotaToQuota(unit * 2)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "邀请额度不足")
				stored := reloadAffTransferUser(t, user.Id)
				assert.Equal(t, before.AffQuota-unit, stored.AffQuota)
				assert.Equal(t, before.Quota+unit, stored.Quota)
			})

			t.Run("concurrent transfers cannot exceed the aff balance", func(t *testing.T) {
				// Repeated on purpose. SQLite skips FOR UPDATE, so this case
				// shows whether writers still conserve the balance. MySQL and
				// PostgreSQL are the runs that exercise the row lock.
				for round := 1; round <= 8; round++ {
					user := insertAffTransferUser(t, 7, unit*8)
					const attempts = 16
					var wg sync.WaitGroup
					errors := make([]error, attempts)
					wg.Add(attempts)
					for i := 0; i < attempts; i++ {
						go func(i int) {
							defer wg.Done()
							errors[i] = (&model.User{Id: user.Id}).TransferAffQuotaToQuota(unit)
						}(i)
					}
					wg.Wait()
					successes := 0
					insufficient := 0
					busy := 0
					for _, err := range errors {
						switch {
						case err == nil:
							successes++
						case strings.Contains(err.Error(), "邀请额度不足"):
							insufficient++
						case kind == "sqlite" && (strings.Contains(err.Error(), "SQLITE_BUSY") || strings.Contains(err.Error(), "database is locked")):
							busy++
						default:
							require.NoError(t, err)
						}
					}
					require.Equal(t, attempts, successes+insufficient+busy)
					stored := reloadAffTransferUser(t, user.Id)
					moved := successes * unit
					assert.GreaterOrEqual(t, stored.AffQuota, 0)
					assert.LessOrEqual(t, successes, 8)
					assert.Equal(t, user.AffQuota-moved, stored.AffQuota, "round %d dialect %s", round, kind)
					assert.Equal(t, user.Quota+moved, stored.Quota, "round %d dialect %s", round, kind)
					if kind != "sqlite" {
						assert.Zero(t, busy, "round %d dialect %s", round, kind)
						assert.Equal(t, 8, successes, "round %d dialect %s", round, kind)
						assert.Equal(t, attempts-8, insufficient, "round %d dialect %s", round, kind)
					}
				}
			})
		})
	}
}

func TestTransferAffQuotaHTTPBoundary(t *testing.T) {
	useAffTransferDatabase(t, "sqlite")
	require.NoError(t, i18n.Init())
	gin.SetMode(gin.TestMode)
	unit := int(common.QuotaPerUnit)
	user := insertAffTransferUser(t, 25, unit*2)

	post := func(body string, comply bool) *httptest.ResponseRecorder {
		t.Helper()
		previous := *operation_setting.GetPaymentSetting()
		setting := operation_setting.GetPaymentSetting()
		setting.ComplianceConfirmed = comply
		if comply {
			setting.ComplianceTermsVersion = operation_setting.CurrentComplianceTermsVersion
		} else {
			setting.ComplianceTermsVersion = ""
		}
		t.Cleanup(func() { *operation_setting.GetPaymentSetting() = previous })
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/user/aff_transfer", strings.NewReader(body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Set("id", user.Id)
		TransferAffQuota(ctx)
		return recorder
	}

	before := reloadAffTransferUser(t, user.Id)
	denied := post(`{"quota":500000}`, false)
	assert.Contains(t, denied.Body.String(), `"success":false`)
	assert.Equal(t, before, reloadAffTransferUser(t, user.Id))

	rejected := post(`{}`, true)
	assert.Contains(t, rejected.Body.String(), `"success":false`)
	assert.Equal(t, before, reloadAffTransferUser(t, user.Id))

	accepted := post(`{"quota":500000}`, true)
	assert.Contains(t, accepted.Body.String(), `"success":true`)
	stored := reloadAffTransferUser(t, user.Id)
	assert.Equal(t, before.AffQuota-unit, stored.AffQuota)
	assert.Equal(t, before.Quota+unit, stored.Quota)
}

func useAffTransferDatabase(t *testing.T, kind string) {
	t.Helper()
	if kind == "postgres" && os.Getenv("TEST_AFF_TRANSFER_POSTGRES") != "1" {
		t.Skip("PostgreSQL row-lock run is opt-in via TEST_AFF_TRANSFER_POSTGRES=1; not executed")
	}
	if kind != "sqlite" && strings.TrimSpace(os.Getenv("TEST_"+strings.ToUpper(kind)+"_DSN")) == "" {
		t.Skip(kind + " DSN is not configured; row-lock behavior is not executed")
	}
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousMain, previousLog := common.MainDatabaseType(), common.LogDatabaseType()
	previousRedis := common.RedisEnabled
	previousUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	common.RedisEnabled = false

	var db *gorm.DB
	switch kind {
	case "sqlite":
		dsn := filepath.Join(t.TempDir(), "aff.db") + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
		opened, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		require.NoError(t, err)
		db = opened
	default:
		opened, _ := newAuditTestDatabase(t, kind, os.Getenv("TEST_"+strings.ToUpper(kind)+"_DSN"))
		db = opened
	}
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(8)
	sqlDB.SetMaxIdleConns(8)
	databaseType := map[string]common.DatabaseType{
		"sqlite":   common.DatabaseTypeSQLite,
		"mysql":    common.DatabaseTypeMySQL,
		"postgres": common.DatabaseTypePostgreSQL,
	}[kind]
	model.DB, model.LOG_DB = db, db
	common.SetDatabaseTypes(databaseType, databaseType)
	require.NoError(t, db.AutoMigrate(&model.User{}))
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.SetDatabaseTypes(previousMain, previousLog)
		common.RedisEnabled = previousRedis
		common.QuotaPerUnit = previousUnit
		if kind == "sqlite" {
			_ = sqlDB.Close()
		}
	})
}

func insertAffTransferUser(t *testing.T, quota int, affQuota int) *model.User {
	t.Helper()
	id := affTransferUsers.Add(1)
	user := &model.User{
		Username: fmt.Sprintf("a%x", id),
		AffCode:  fmt.Sprintf("c%x", id),
		Status:   common.UserStatusEnabled,
		Quota:    quota,
		AffQuota: affQuota,
	}
	require.NoError(t, model.DB.Create(user).Error)
	return user
}

var affTransferUsers atomic.Uint64

func reloadAffTransferUser(t *testing.T, id int) model.User {
	t.Helper()
	var stored model.User
	require.NoError(t, model.DB.First(&stored, id).Error)
	return stored
}
