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
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
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
	"github.com/alicebob/miniredis/v2"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// These tests drive controller.Relay for a non-streaming chat completion.
// The upstream is an httptest server. Billing uses the production
// PreConsumeBilling / SettleBilling / Refund path against an isolated SQLite
// file. Redis is off, so the assertion store is the database.
//
// TestTextRelayStreamFlushThenDisconnectDoesNotDoubleCharge covers an OpenAI
// SSE frame that is flushed and then cut off without a normal EOF. It checks
// one upstream call and one balance drop. It does not decide whether the
// partial output should be refunded or kept.

const textRelayBillingStartQuota = 1_000_000

var textRelayBillingSeq atomic.Uint64

func TestTextRelayUpstreamFailureRefundsPreConsume(t *testing.T) {
	db, trace := useTextRelayBillingDatabase(t)
	user, token := insertTextRelayAccount(t)
	trace.bind(user.Id, token.Id)
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

	recorder := callTextRelay(t, user, token, server.URL, false)
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
	userSteps, tokenSteps := trace.snapshot()
	assert.Equal(t, []int{heldUser, textRelayBillingStartQuota}, userSteps)
	assert.Equal(t, []int{heldToken, textRelayBillingStartQuota}, tokenSteps)
}

func TestTextRelayPreConsumeDiffersFromSettlement(t *testing.T) {
	db, trace := useTextRelayBillingDatabase(t)
	user, token := insertTextRelayAccount(t)
	trace.bind(user.Id, token.Id)
	var snapMu sync.Mutex
	var sawUpstream bool
	var upstreamReadErr error
	var preConsumedUser, preConsumedToken int
	var userStepsAtUpstream, tokenStepsAtUpstream []int
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
		userNow, tokenNow := trace.snapshot()
		snapMu.Lock()
		sawUpstream = true
		preConsumedUser = storedUser.Quota
		preConsumedToken = storedToken.RemainQuota
		userStepsAtUpstream = userNow
		tokenStepsAtUpstream = tokenNow
		upstreamReadErr = nil
		snapMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl-fixture","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":8,"completion_tokens":1,"total_tokens":9}}`))
	}))
	t.Cleanup(server.Close)

	recorder := callTextRelay(t, user, token, server.URL, false)
	snapMu.Lock()
	readErr, saw, heldUser, heldToken := upstreamReadErr, sawUpstream, preConsumedUser, preConsumedToken
	userAtUpstream := append([]int(nil), userStepsAtUpstream...)
	tokenAtUpstream := append([]int(nil), tokenStepsAtUpstream...)
	snapMu.Unlock()
	require.NoError(t, readErr)
	require.True(t, saw, "upstream was never called; body=%s", recorder.Body.String())
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	preConsumed := textRelayBillingStartQuota - heldUser
	require.Greater(t, preConsumed, 0)
	assert.Equal(t, preConsumed, textRelayBillingStartQuota-heldToken)
	require.Equal(t, []int{heldUser}, userAtUpstream)
	require.Equal(t, []int{heldToken}, tokenAtUpstream)

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
	userSteps, tokenSteps := trace.snapshot()
	require.Equal(t, []int{heldUser, finalUser}, userSteps)
	require.Equal(t, []int{heldToken, finalToken}, tokenSteps)
}

func TestTextRelayStreamFlushThenDisconnectDoesNotDoubleCharge(t *testing.T) {
	_, trace := useTextRelayBillingDatabase(t)
	user, token := insertTextRelayAccount(t)
	trace.bind(user.Id, token.Id)
	previousRetry := common.RetryTimes
	previousStreamTimeout := constant.StreamingTimeout
	common.RetryTimes = 1
	constant.StreamingTimeout = 30
	t.Cleanup(func() {
		common.RetryTimes = previousRetry
		constant.StreamingTimeout = previousStreamTimeout
	})
	poolBaseline := gopool.WorkerCount()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "hijack unsupported", http.StatusInternalServerError)
			return
		}
		conn, writer, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		frame := "data: {\"id\":\"chatcmpl-sse\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"ok\"}}]}\n\n"
		if _, err = fmt.Fprintf(writer, "HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\nTransfer-Encoding: chunked\r\n\r\n%x\r\n%s\r\n", len(frame), frame); err != nil {
			return
		}
		if err = writer.Flush(); err != nil {
			return
		}
		// Close without the terminating chunk. A finished chunked body would be a normal EOF.
	}))
	t.Cleanup(server.Close)

	recorder := callTextRelay(t, user, token, server.URL, true)
	require.Equal(t, int32(1), calls.Load(), "upstream calls after a flushed frame; body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"content":"ok"`)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && gopool.WorkerCount() > poolBaseline {
		time.Sleep(5 * time.Millisecond)
	}

	finalUser := reloadTextRelayQuota(t, user.Id)
	finalToken := reloadTextRelayTokenQuota(t, token.Id)
	netUser := textRelayBillingStartQuota - finalUser
	netToken := textRelayBillingStartQuota - finalToken
	require.Equal(t, netUser, netToken)
	userSteps, tokenSteps := trace.snapshot()
	require.Equal(t, 1, countBalanceDrops(textRelayBillingStartQuota, userSteps))
	require.Equal(t, 1, countBalanceDrops(textRelayBillingStartQuota, tokenSteps))
	require.LessOrEqual(t, countBalanceRises(userSteps), 1)
	require.LessOrEqual(t, countBalanceRises(tokenSteps), 1)

	var logs []model.Log
	require.NoError(t, model.LOG_DB.Where("user_id = ?", user.Id).Find(&logs).Error)
	consume := 0
	for _, item := range logs {
		if item.Type == model.LogTypeConsume {
			consume++
			assert.Equal(t, netUser, item.Quota)
		}
	}
	require.LessOrEqual(t, consume, 1)
	if consume == 0 {
		assert.Equal(t, 0, netUser)
	} else {
		assert.Greater(t, netUser, 0)
	}
	t.Logf("observed stream disconnect: http=%d upstream_calls=%d user_steps=%v token_steps=%v net=%d consume_logs=%d body=%s",
		recorder.Code, calls.Load(), userSteps, tokenSteps, netUser, consume, recorder.Body.String())
}

func countBalanceDrops(start int, steps []int) int {
	drops := 0
	previous := start
	for _, value := range steps {
		if value < previous {
			drops++
		}
		previous = value
	}
	return drops
}

func countBalanceRises(steps []int) int {
	rises := 0
	if len(steps) == 0 {
		return 0
	}
	previous := steps[0]
	for _, value := range steps[1:] {
		if value > previous {
			rises++
		}
		previous = value
	}
	return rises
}

func useTextRelayBillingDatabase(t *testing.T) (*gorm.DB, *textRelayBalanceTrace) {
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
	trace := &textRelayBalanceTrace{}
	require.NoError(t, db.Callback().Update().After("gorm:update").Register("textrelay:balance", func(tx *gorm.DB) {
		if tx.Statement == nil || tx.Statement.Schema == nil {
			return
		}
		trace.note(tx, tx.Statement.Schema.Table)
	}))
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
	return db, trace
}

type textRelayBalanceTrace struct {
	mu         sync.Mutex
	userID     int
	tokenID    int
	userQuota  []int
	tokenQuota []int
}

func (trace *textRelayBalanceTrace) bind(userID int, tokenID int) {
	trace.mu.Lock()
	trace.userID = userID
	trace.tokenID = tokenID
	trace.mu.Unlock()
}

func (trace *textRelayBalanceTrace) note(tx *gorm.DB, table string) {
	trace.mu.Lock()
	userID, tokenID := trace.userID, trace.tokenID
	trace.mu.Unlock()
	switch table {
	case "users":
		if userID == 0 {
			return
		}
		quota, err := textRelayQueryQuota(tx, "SELECT quota FROM users WHERE id = ?", userID)
		if err != nil {
			return
		}
		trace.appendUser(quota)
	case "tokens":
		if tokenID == 0 {
			return
		}
		quota, err := textRelayQueryQuota(tx, "SELECT remain_quota FROM tokens WHERE id = ?", tokenID)
		if err != nil {
			return
		}
		trace.appendToken(quota)
	}
}

func textRelayQueryQuota(tx *gorm.DB, query string, id int) (int, error) {
	if tx.Statement == nil || tx.Statement.ConnPool == nil {
		return 0, fmt.Errorf("missing connection")
	}
	queryer, ok := tx.Statement.ConnPool.(interface {
		QueryRowContext(context.Context, string, ...any) *sql.Row
	})
	if !ok {
		return 0, fmt.Errorf("connection cannot query")
	}
	var quota int
	err := queryer.QueryRowContext(context.Background(), query, id).Scan(&quota)
	return quota, err
}

func (trace *textRelayBalanceTrace) appendUser(quota int) {
	trace.mu.Lock()
	defer trace.mu.Unlock()
	if n := len(trace.userQuota); n > 0 && trace.userQuota[n-1] == quota {
		return
	}
	trace.userQuota = append(trace.userQuota, quota)
}

func (trace *textRelayBalanceTrace) appendToken(quota int) {
	trace.mu.Lock()
	defer trace.mu.Unlock()
	if n := len(trace.tokenQuota); n > 0 && trace.tokenQuota[n-1] == quota {
		return
	}
	trace.tokenQuota = append(trace.tokenQuota, quota)
}

func (trace *textRelayBalanceTrace) snapshot() (userQuota []int, tokenQuota []int) {
	trace.mu.Lock()
	defer trace.mu.Unlock()
	userQuota = append([]int(nil), trace.userQuota...)
	tokenQuota = append([]int(nil), trace.tokenQuota...)
	return userQuota, tokenQuota
}

func insertTextRelayAccount(t *testing.T) (*model.User, *model.Token) {
	t.Helper()
	seq := textRelayBillingSeq.Add(1)
	user := &model.User{
		Username:    fmt.Sprintf("r%x", seq),
		AffCode:     fmt.Sprintf("c%x", seq),
		Status:      common.UserStatusEnabled,
		Group:       "default",
		Quota:       textRelayBillingStartQuota,
		Role:        common.RoleCommonUser,
		AuthVersion: 1,
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

func callTextRelay(t *testing.T, user *model.User, token *model.Token, baseURL string, stream bool) *httptest.ResponseRecorder {
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

	streamJSON := "false"
	if stream {
		streamJSON = "true"
	}
	body := fmt.Sprintf(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}],"max_tokens":200,"stream":%s}`, streamJSON)
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

// textRelayEvalHook records EVAL keys. The billing methods themselves still
// run against miniredis; this only observes which keys the Lua scripts touch.
type textRelayEvalHook struct {
	mu    sync.Mutex
	calls []textRelayEvalCall
}

type textRelayEvalCall struct {
	key    string
	script string
}

func (hook *textRelayEvalHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if cmd.Name() == "eval" {
		args := cmd.Args()
		if len(args) >= 4 {
			key, _ := args[3].(string)
			script, _ := args[1].(string)
			hook.mu.Lock()
			hook.calls = append(hook.calls, textRelayEvalCall{key: key, script: script})
			hook.mu.Unlock()
		}
	}
	return ctx, nil
}

func (hook *textRelayEvalHook) AfterProcess(context.Context, redis.Cmder) error { return nil }

func (hook *textRelayEvalHook) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (hook *textRelayEvalHook) AfterProcessPipeline(context.Context, []redis.Cmder) error { return nil }

func (hook *textRelayEvalHook) quotaMutations(key string) int {
	hook.mu.Lock()
	defer hook.mu.Unlock()
	count := 0
	for _, call := range hook.calls {
		if call.key == key && strings.Contains(call.script, "HINCRBY") {
			count++
		}
	}
	return count
}

func ensureTextRelaySQLiteColumnNames(t *testing.T) {
	t.Helper()
	previousMaster := common.IsMasterNode
	previousPath := common.SQLitePath
	previousDSN, hadDSN := os.LookupEnv("SQL_DSN")
	common.IsMasterNode = false
	common.SQLitePath = filepath.Join(t.TempDir(), "columns.db")
	require.NoError(t, os.Setenv("SQL_DSN", "local"))
	require.NoError(t, model.InitDB())
	if sqlDB, err := model.DB.DB(); err == nil {
		_ = sqlDB.Close()
	}
	common.IsMasterNode = previousMaster
	common.SQLitePath = previousPath
	if hadDSN {
		require.NoError(t, os.Setenv("SQL_DSN", previousDSN))
	} else {
		require.NoError(t, os.Unsetenv("SQL_DSN"))
	}
}

func enableTextRelayMiniRedis(t *testing.T) (*miniredis.Miniredis, *textRelayEvalHook) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	hook := &textRelayEvalHook{}
	client.AddHook(hook)
	previous := common.RDB
	common.RDB = client
	common.RedisEnabled = true
	t.Cleanup(func() {
		_ = client.Close()
		if common.RDB == client {
			common.RDB = previous
		}
	})
	return server, hook
}

func textRelayHashInt(t *testing.T, key string, field string) int {
	t.Helper()
	parsed, err := readTextRelayHashInt(key, field)
	require.NoError(t, err)
	return parsed
}

func readTextRelayHashInt(key string, field string) (int, error) {
	value, err := common.RDB.HGet(context.Background(), key, field).Result()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

func textRelayTokenCacheKey(t *testing.T, server *miniredis.Miniredis) string {
	t.Helper()
	found := ""
	for _, key := range server.Keys() {
		if strings.HasPrefix(key, "token:") && !strings.HasPrefix(key, "token:fence:") {
			require.Empty(t, found)
			found = key
		}
	}
	require.NotEmpty(t, found)
	return found
}

func warmTextRelayQuotaCache(t *testing.T, user *model.User, token *model.Token) {
	t.Helper()
	cachedUser, err := model.GetUserCache(user.Id)
	require.NoError(t, err)
	require.Equal(t, textRelayBillingStartQuota, cachedUser.Quota)
	cachedToken, err := model.GetTokenByKey(token.Key, true)
	require.NoError(t, err)
	require.Equal(t, textRelayBillingStartQuota, cachedToken.RemainQuota)
}

func waitTextRelayQuotaCache(t *testing.T, userID int, tokenID int, userKey string, tokenKey string, expectUser int, expectToken int) {
	t.Helper()
	require.Eventually(t, func() bool {
		dbUser := reloadTextRelayQuota(t, userID)
		dbToken := reloadTextRelayTokenQuota(t, tokenID)
		cacheUser, userErr := readTextRelayHashInt(userKey, "Quota")
		cacheToken, tokenErr := readTextRelayHashInt(tokenKey, "RemainQuota")
		if userErr != nil || tokenErr != nil {
			return false
		}
		return dbUser == expectUser && dbToken == expectToken && cacheUser == dbUser && cacheToken == dbToken
	}, 2*time.Second, 5*time.Millisecond)
}

func TestTextRelayRedisBalanceMatchesDatabase(t *testing.T) {
	runTextRelayRedisBalance(t, func(t *testing.T) (*textRelayEvalHook, func(*testing.T) string) {
		server, hook := enableTextRelayMiniRedis(t)
		return hook, func(t *testing.T) string {
			return textRelayTokenCacheKey(t, server)
		}
	})
}

func TestTextRelayRealRedisBalanceMatchesDatabase(t *testing.T) {
	addr := strings.TrimSpace(os.Getenv("TEST_REDIS_ADDR"))
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR is not set; real Redis is not executed")
	}
	host, _, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		t.Fatalf("refusing non-loopback Redis address %s", host)
	}
	probe := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = probe.Close() })
	info, err := probe.Info(context.Background(), "server").Result()
	require.NoError(t, err)
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "redis_version:") {
			t.Log(line)
		}
	}
	runTextRelayRedisBalance(t, func(t *testing.T) (*textRelayEvalHook, func(*testing.T) string) {
		return openTextRelayRedisAddr(t, addr)
	})
}

func openTextRelayRedisAddr(t *testing.T, addr string) (*textRelayEvalHook, func(*testing.T) string) {
	t.Helper()
	beforeClient := redis.NewClient(&redis.Options{Addr: addr})
	existing, err := beforeClient.Keys(context.Background(), "token:*").Result()
	require.NoError(t, err)
	require.NoError(t, beforeClient.Close())
	before := map[string]struct{}{}
	for _, key := range existing {
		before[key] = struct{}{}
	}
	hook := &textRelayEvalHook{}
	client := redis.NewClient(&redis.Options{Addr: addr})
	client.AddHook(hook)
	previous := common.RDB
	common.RDB = client
	common.RedisEnabled = true
	t.Cleanup(func() {
		_ = client.Close()
		if common.RDB == client {
			common.RDB = previous
		}
	})
	return hook, func(t *testing.T) string {
		t.Helper()
		keys, err := common.RDB.Keys(context.Background(), "token:*").Result()
		require.NoError(t, err)
		found := ""
		for _, key := range keys {
			if strings.HasPrefix(key, "token:fence:") {
				continue
			}
			if _, seen := before[key]; seen {
				continue
			}
			require.Empty(t, found)
			found = key
		}
		require.NotEmpty(t, found)
		return found
	}
}

func runTextRelayRedisBalance(t *testing.T, open func(t *testing.T) (*textRelayEvalHook, func(*testing.T) string)) {
	t.Helper()
	ensureTextRelaySQLiteColumnNames(t)

	t.Run("upstream error refunds once", func(t *testing.T) {
		db, trace := useTextRelayBillingDatabase(t)
		evals, tokenKeyOf := open(t)
		user, token := insertTextRelayAccount(t)
		trace.bind(user.Id, token.Id)
		warmTextRelayQuotaCache(t, user, token)
		userKey := fmt.Sprintf("user:%d", user.Id)
		tokenKey := tokenKeyOf(t)
		t.Cleanup(func() {
			if common.RDB != nil {
				_ = common.RDB.Del(context.Background(), userKey, tokenKey).Err()
			}
		})
		require.Equal(t, textRelayBillingStartQuota, textRelayHashInt(t, userKey, "Quota"))
		require.Equal(t, textRelayBillingStartQuota, textRelayHashInt(t, tokenKey, "RemainQuota"))

		var snapMu sync.Mutex
		var saw bool
		var readErr error
		var heldUser, heldToken, cacheUser, cacheToken int
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var storedUser model.User
			var storedToken model.Token
			if err := db.First(&storedUser, user.Id).Error; err != nil {
				snapMu.Lock()
				readErr = err
				snapMu.Unlock()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := db.First(&storedToken, token.Id).Error; err != nil {
				snapMu.Lock()
				readErr = err
				snapMu.Unlock()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			userNow, userErr := readTextRelayHashInt(userKey, "Quota")
			tokenNow, tokenErr := readTextRelayHashInt(tokenKey, "RemainQuota")
			snapMu.Lock()
			saw = true
			if userErr != nil {
				readErr = userErr
			} else if tokenErr != nil {
				readErr = tokenErr
			}
			heldUser = storedUser.Quota
			heldToken = storedToken.RemainQuota
			cacheUser = userNow
			cacheToken = tokenNow
			snapMu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error":{"message":"upstream failed before any output","type":"upstream_error"}}`))
		}))
		t.Cleanup(upstream.Close)
		recorder := callTextRelay(t, user, token, upstream.URL, false)
		snapMu.Lock()
		gotErr, gotSaw := readErr, saw
		gotHeldUser, gotHeldToken := heldUser, heldToken
		gotCacheUser, gotCacheToken := cacheUser, cacheToken
		snapMu.Unlock()
		require.NoError(t, gotErr)
		require.True(t, gotSaw, recorder.Body.String())
		require.Equal(t, gotHeldUser, gotCacheUser)
		require.Equal(t, gotHeldToken, gotCacheToken)
		require.Less(t, gotCacheUser, textRelayBillingStartQuota)
		heldUser, heldToken = gotHeldUser, gotHeldToken
		waitTextRelayQuotaCache(t, user.Id, token.Id, userKey, tokenKey, textRelayBillingStartQuota, textRelayBillingStartQuota)

		require.GreaterOrEqual(t, evals.quotaMutations(userKey), 2)
		require.GreaterOrEqual(t, evals.quotaMutations(tokenKey), 2)
		assert.Equal(t, textRelayBillingStartQuota, reloadTextRelayQuota(t, user.Id))
		assert.Equal(t, textRelayBillingStartQuota, reloadTextRelayTokenQuota(t, token.Id))
		assert.Equal(t, textRelayBillingStartQuota, textRelayHashInt(t, userKey, "Quota"))
		assert.Equal(t, textRelayBillingStartQuota, textRelayHashInt(t, tokenKey, "RemainQuota"))
		assert.Equal(t, int64(0), countTextRelayConsumeLogs(t, user.Id))
		userSteps, tokenSteps := trace.snapshot()
		assert.Equal(t, []int{heldUser, textRelayBillingStartQuota}, userSteps)
		assert.Equal(t, []int{heldToken, textRelayBillingStartQuota}, tokenSteps)
	})

	t.Run("success settles the difference once", func(t *testing.T) {
		runTextRelayRedisSettlement(t, open, false)
	})

	t.Run("flushed stream scanner error settles once", func(t *testing.T) {
		runTextRelayRedisSettlement(t, open, true)
	})
}

func runTextRelayRedisSettlement(t *testing.T, open func(t *testing.T) (*textRelayEvalHook, func(*testing.T) string), stream bool) {
	t.Helper()
	db, trace := useTextRelayBillingDatabase(t)
	evals, tokenKeyOf := open(t)
	user, token := insertTextRelayAccount(t)
	trace.bind(user.Id, token.Id)
	warmTextRelayQuotaCache(t, user, token)
	userKey := fmt.Sprintf("user:%d", user.Id)
	tokenKey := tokenKeyOf(t)
	t.Cleanup(func() {
		if common.RDB != nil {
			_ = common.RDB.Del(context.Background(), userKey, tokenKey).Err()
		}
	})
	previousRetry := common.RetryTimes
	previousStreamTimeout := constant.StreamingTimeout
	common.RetryTimes = 1
	constant.StreamingTimeout = 30
	t.Cleanup(func() {
		common.RetryTimes = previousRetry
		constant.StreamingTimeout = previousStreamTimeout
	})

	var calls atomic.Int32
	var snapMu sync.Mutex
	var saw bool
	var readErr error
	var heldUser, heldToken, cacheUser, cacheToken int
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var storedUser model.User
		var storedToken model.Token
		if err := db.First(&storedUser, user.Id).Error; err != nil {
			snapMu.Lock()
			readErr = err
			snapMu.Unlock()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := db.First(&storedToken, token.Id).Error; err != nil {
			snapMu.Lock()
			readErr = err
			snapMu.Unlock()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		userNow, userErr := readTextRelayHashInt(userKey, "Quota")
		tokenNow, tokenErr := readTextRelayHashInt(tokenKey, "RemainQuota")
		snapMu.Lock()
		saw = true
		if userErr != nil {
			readErr = userErr
		} else if tokenErr != nil {
			readErr = tokenErr
		}
		heldUser = storedUser.Quota
		heldToken = storedToken.RemainQuota
		cacheUser = userNow
		cacheToken = tokenNow
		snapMu.Unlock()
		if !stream {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"chatcmpl-fixture","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":8,"completion_tokens":1,"total_tokens":9}}`))
			return
		}
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "hijack unsupported", http.StatusInternalServerError)
			return
		}
		conn, writer, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		frame := "data: {\"id\":\"chatcmpl-sse\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"ok\"}}]}\n\n"
		_, _ = fmt.Fprintf(writer, "HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\nTransfer-Encoding: chunked\r\n\r\n%x\r\n%s\r\n", len(frame), frame)
		_ = writer.Flush()
	}))
	t.Cleanup(upstream.Close)

	recorder := callTextRelay(t, user, token, upstream.URL, stream)
	snapMu.Lock()
	gotErr, gotSaw := readErr, saw
	gotHeldUser, gotHeldToken := heldUser, heldToken
	gotCacheUser, gotCacheToken := cacheUser, cacheToken
	snapMu.Unlock()
	require.NoError(t, gotErr)
	require.True(t, gotSaw, recorder.Body.String())
	require.Equal(t, gotHeldUser, gotCacheUser)
	require.Equal(t, gotHeldToken, gotCacheToken)
	require.Less(t, gotHeldUser, textRelayBillingStartQuota)
	heldUser, heldToken = gotHeldUser, gotHeldToken
	if stream {
		require.Equal(t, int32(1), calls.Load(), recorder.Body.String())
		require.Contains(t, recorder.Body.String(), `"content":"ok"`)
	}
	finalUser := reloadTextRelayQuota(t, user.Id)
	finalToken := reloadTextRelayTokenQuota(t, token.Id)
	waitTextRelayQuotaCache(t, user.Id, token.Id, userKey, tokenKey, finalUser, finalToken)
	require.Less(t, finalUser, textRelayBillingStartQuota)
	require.Greater(t, finalUser, heldUser)
	require.Equal(t, textRelayBillingStartQuota-finalUser, textRelayBillingStartQuota-finalToken)
	require.Equal(t, finalUser, textRelayHashInt(t, userKey, "Quota"))
	require.Equal(t, finalToken, textRelayHashInt(t, tokenKey, "RemainQuota"))
	require.GreaterOrEqual(t, evals.quotaMutations(userKey), 2)
	require.GreaterOrEqual(t, evals.quotaMutations(tokenKey), 2)
	assert.Equal(t, int64(1), countTextRelayConsumeLogs(t, user.Id))
	userSteps, tokenSteps := trace.snapshot()
	require.Equal(t, []int{heldUser, finalUser}, userSteps)
	require.Equal(t, []int{heldToken, finalToken}, tokenSteps)
}
