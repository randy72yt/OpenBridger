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
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// These tests drive controller.Relay for a non-streaming chat completion.
// The upstream is an httptest server. Billing uses the production
// PreConsumeBilling / SettleBilling / Refund path against an isolated SQLite
// file. Redis is off, so the assertion store is the database.
//
// A failure that happens after the response has already been flushed to the
// client (SSE) is not covered here.

const textRelayBillingStartQuota = 1_000_000

var textRelayBillingSeq atomic.Uint64

func TestTextRelayUpstreamFailureRefundsPreConsume(t *testing.T) {
	db := useTextRelayBillingDatabase(t)
	user, token := insertTextRelayAccount(t)
	var snapMu sync.Mutex
	var sawUpstream bool
	var upstreamReadErr error
	var preConsumedUser, preConsumedToken int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var storedUser model.User
		var storedToken model.Token
		if err := db.First(&storedUser, user.Id).Error; err != nil {
			snapMu.Lock()
			upstreamReadErr = err
			snapMu.Unlock()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := db.First(&storedToken, token.Id).Error; err != nil {
			snapMu.Lock()
			upstreamReadErr = err
			snapMu.Unlock()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		snapMu.Lock()
		sawUpstream = true
		preConsumedUser = storedUser.Quota
		preConsumedToken = storedToken.RemainQuota
		upstreamReadErr = nil
		snapMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"message":"upstream failed before any output","type":"upstream_error"}}`))
	}))
	t.Cleanup(server.Close)

	recorder := callTextRelay(t, user, token, server.URL)
	snapMu.Lock()
	readErr, saw, heldUser, heldToken := upstreamReadErr, sawUpstream, preConsumedUser, preConsumedToken
	snapMu.Unlock()
	require.NoError(t, readErr)
	require.True(t, saw, "upstream was never called; body=%s", recorder.Body.String())
	assert.NotEqual(t, http.StatusOK, recorder.Code, recorder.Body.String())
	preConsumed := textRelayBillingStartQuota - heldUser
	require.Greater(t, preConsumed, 0)
	assert.Equal(t, preConsumed, textRelayBillingStartQuota-heldToken)

	require.Eventually(t, func() bool {
		return reloadTextRelayQuota(t, user.Id) == textRelayBillingStartQuota &&
			reloadTextRelayTokenQuota(t, token.Id) == textRelayBillingStartQuota
	}, 3*time.Second, 20*time.Millisecond)

	assert.Equal(t, textRelayBillingStartQuota, reloadTextRelayQuota(t, user.Id))
	assert.Equal(t, textRelayBillingStartQuota, reloadTextRelayTokenQuota(t, token.Id))
	assert.Equal(t, int64(0), countTextRelayConsumeLogs(t, user.Id))
}

func TestTextRelayPreConsumeDiffersFromSettlement(t *testing.T) {
	db := useTextRelayBillingDatabase(t)
	user, token := insertTextRelayAccount(t)
	var snapMu sync.Mutex
	var sawUpstream bool
	var upstreamReadErr error
	var preConsumedUser, preConsumedToken int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var storedUser model.User
		var storedToken model.Token
		if err := db.First(&storedUser, user.Id).Error; err != nil {
			snapMu.Lock()
			upstreamReadErr = err
			snapMu.Unlock()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := db.First(&storedToken, token.Id).Error; err != nil {
			snapMu.Lock()
			upstreamReadErr = err
			snapMu.Unlock()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		snapMu.Lock()
		sawUpstream = true
		preConsumedUser = storedUser.Quota
		preConsumedToken = storedToken.RemainQuota
		upstreamReadErr = nil
		snapMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl-fixture","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":8,"completion_tokens":1,"total_tokens":9}}`))
	}))
	t.Cleanup(server.Close)

	recorder := callTextRelay(t, user, token, server.URL)
	snapMu.Lock()
	readErr, saw, heldUser, heldToken := upstreamReadErr, sawUpstream, preConsumedUser, preConsumedToken
	snapMu.Unlock()
	require.NoError(t, readErr)
	require.True(t, saw, "upstream was never called; body=%s", recorder.Body.String())
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	preConsumed := textRelayBillingStartQuota - heldUser
	require.Greater(t, preConsumed, 0)
	assert.Equal(t, preConsumed, textRelayBillingStartQuota-heldToken)

	finalUser := reloadTextRelayQuota(t, user.Id)
	finalToken := reloadTextRelayTokenQuota(t, token.Id)
	settled := textRelayBillingStartQuota - finalUser
	require.Greater(t, settled, 0)
	assert.Equal(t, settled, textRelayBillingStartQuota-finalToken)
	assert.NotEqual(t, preConsumed, settled)

	var logs []model.Log
	require.NoError(t, model.LOG_DB.Where("user_id = ? AND type = ?", user.Id, model.LogTypeConsume).Find(&logs).Error)
	require.Len(t, logs, 1)
	assert.Equal(t, settled, logs[0].Quota)
	assert.Equal(t, 8, logs[0].PromptTokens)
	assert.Equal(t, 1, logs[0].CompletionTokens)
}

func useTextRelayBillingDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousMain, previousLog := common.MainDatabaseType(), common.LogDatabaseType()
	previousRedis := common.RedisEnabled
	previousBatch := common.BatchUpdateEnabled
	previousUnit := common.QuotaPerUnit
	common.RedisEnabled = false
	common.BatchUpdateEnabled = false
	common.QuotaPerUnit = 500000
	ratio_setting.InitRatioSettings()

	dsn := filepath.Join(t.TempDir(), "relay.db") + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)
	model.DB, model.LOG_DB = db, db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Token{}, &model.Channel{}, &model.Log{}))
	poolBaseline := gopool.WorkerCount()
	t.Cleanup(func() {
		// Refund and perf sampling run on the shared pool and still read
		// process-wide flags after Relay returns. Wait until this test's
		// extra workers exit before restoring those flags. A pool that other
		// tests already occupied is left alone.
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) && gopool.WorkerCount() > poolBaseline {
			time.Sleep(5 * time.Millisecond)
		}
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.SetDatabaseTypes(previousMain, previousLog)
		common.RedisEnabled = previousRedis
		common.BatchUpdateEnabled = previousBatch
		common.QuotaPerUnit = previousUnit
		_ = sqlDB.Close()
	})
	return db
}

func insertTextRelayAccount(t *testing.T) (*model.User, *model.Token) {
	t.Helper()
	seq := textRelayBillingSeq.Add(1)
	user := &model.User{
		Username: fmt.Sprintf("r%x", seq),
		AffCode:  fmt.Sprintf("c%x", seq),
		Status:   common.UserStatusEnabled,
		Group:    "default",
		Quota:    textRelayBillingStartQuota,
		Role:     common.RoleCommonUser,
	}
	require.NoError(t, model.DB.Create(user).Error)
	token := &model.Token{
		UserId:      user.Id,
		Key:         fmt.Sprintf("sk-relay-%x", seq),
		Status:      common.TokenStatusEnabled,
		Name:        "relay",
		ExpiredTime: -1,
		RemainQuota: textRelayBillingStartQuota,
		Group:       "default",
	}
	require.NoError(t, model.DB.Create(token).Error)
	return user, token
}

func callTextRelay(t *testing.T, user *model.User, token *model.Token, baseURL string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	autoBan := 0
	channel := &model.Channel{
		Type:          constant.ChannelTypeOpenAI,
		Key:           "sk-upstream-fixture",
		Name:          "text-relay-fixture",
		Status:        common.ChannelStatusEnabled,
		BaseURL:       &baseURL,
		Models:        "gpt-4o-mini",
		Group:         "default",
		OtherSettings: "{}",
		AutoBan:       &autoBan,
	}
	require.NoError(t, model.DB.Create(channel).Error)

	body := `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}],"max_tokens":200}`
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	require.Nil(t, middleware.SetupContextForSelectedChannel(ctx, channel, "gpt-4o-mini"))
	common.SetContextKey(ctx, constant.ContextKeyOriginalModel, "gpt-4o-mini")
	common.SetContextKey(ctx, constant.ContextKeyUserId, user.Id)
	common.SetContextKey(ctx, constant.ContextKeyUserGroup, "default")
	common.SetContextKey(ctx, constant.ContextKeyUsingGroup, "default")
	common.SetContextKey(ctx, constant.ContextKeyUserQuota, textRelayBillingStartQuota)
	common.SetContextKey(ctx, constant.ContextKeyTokenId, token.Id)
	common.SetContextKey(ctx, constant.ContextKeyTokenKey, token.Key)
	common.SetContextKey(ctx, constant.ContextKeyTokenGroup, "default")
	common.SetContextKey(ctx, constant.ContextKeyTokenUnlimited, false)
	common.SetContextKey(ctx, constant.ContextKeyUserSetting, dto.UserSetting{BillingPreference: "wallet_only"})
	ctx.Set("token_quota", textRelayBillingStartQuota)
	ctx.Set("username", user.Username)

	Relay(ctx, types.RelayFormatOpenAI)
	return recorder
}

func reloadTextRelayQuota(t *testing.T, userID int) int {
	t.Helper()
	quota, err := model.GetUserQuota(userID, true)
	require.NoError(t, err)
	return quota
}

func reloadTextRelayTokenQuota(t *testing.T, tokenID int) int {
	t.Helper()
	var token model.Token
	require.NoError(t, model.DB.First(&token, tokenID).Error)
	return token.RemainQuota
}

func countTextRelayConsumeLogs(t *testing.T, userID int) int64 {
	t.Helper()
	var count int64
	require.NoError(t, model.LOG_DB.Model(&model.Log{}).Where("user_id = ? AND type = ?", userID, model.LogTypeConsume).Count(&count).Error)
	return count
}
