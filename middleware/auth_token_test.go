package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/common/testdb"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTokenAuthDB(t *testing.T) {
	t.Helper()
	if model.DB == nil || model.DB != tokenAuthDB {
		tokenAuthDB = testdb.Open()
		model.DB = tokenAuthDB
		require.NoError(t, model.DB.AutoMigrate(&model.User{}, &model.Token{}))
	}
	// model 包外的测试不会走 model 的 TestMain，必须显式初始化
	// 方言相关的保留字列名（key/group 的引用形式）。
	model.InitColumnNames()
	model.DB.Exec("DELETE FROM tokens")
	model.DB.Exec("DELETE FROM users")
	common.RedisEnabled = false
}

var tokenAuthDB *gorm.DB

func seedAuthUser(t *testing.T, id int, status int) {
	t.Helper()
	user := &model.User{
		Id:       id,
		Username: "token-auth-user-" + common.GetRandomString(6),
		Status:   status,
		Group:    "default",
	}
	require.NoError(t, model.DB.Create(user).Error)
}

func seedAuthToken(t *testing.T, id, userId int, key string, status int) {
	t.Helper()
	token := &model.Token{
		Id:          id,
		UserId:      userId,
		Key:         key,
		Name:        "auth-test",
		Status:      status,
		ExpiredTime: -1,
		RemainQuota: 1000,
	}
	require.NoError(t, token.Insert())
}

func runTokenAuthReadOnly(t *testing.T, authHeader string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ro", TokenAuthReadOnly())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ro", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	router.ServeHTTP(w, req)
	return w
}

// 只读令牌认证：缺 key 401；仅 Disabled 状态拒绝，
// exhausted/expired 仍可查只读数据（中间件的显式语义）。
func TestTokenAuthReadOnlyStatusSemantics(t *testing.T) {
	setupTokenAuthDB(t)
	seedAuthUser(t, 500, common.UserStatusEnabled)
	seedAuthToken(t, 501, 500, "rodisabledkey", common.TokenStatusDisabled)
	seedAuthToken(t, 502, 500, "roexhaustedkey", common.TokenStatusExhausted)
	seedAuthToken(t, 503, 500, "roenabledkey", common.TokenStatusEnabled)

	assert.Equal(t, http.StatusUnauthorized, runTokenAuthReadOnly(t, "").Code)

	assert.Equal(t, http.StatusUnauthorized, runTokenAuthReadOnly(t, "sk-rodisabledkey").Code)

	// exhausted：非 disabled 状态放行
	w := runTokenAuthReadOnly(t, "Bearer sk-roexhaustedkey")
	assert.Equal(t, http.StatusOK, w.Code)

	assert.Equal(t, http.StatusOK, runTokenAuthReadOnly(t, "sk-roenabledkey").Code)

	// 不存在的 key
	assert.Equal(t, http.StatusUnauthorized, runTokenAuthReadOnly(t, "sk-nosuchkey").Code)
}

// 被封禁的用户即使令牌有效也必须 403。
func TestTokenAuthReadOnlyRejectsBannedUser(t *testing.T) {
	setupTokenAuthDB(t)
	seedAuthUser(t, 600, common.UserStatusDisabled)
	seedAuthToken(t, 601, 600, "robannedkey", common.TokenStatusEnabled)

	assert.Equal(t, http.StatusForbidden, runTokenAuthReadOnly(t, "sk-robannedkey").Code)
}

// 成功路径必须把用户/令牌身份写进 context，供下游 handler 使用。
func TestTokenAuthReadOnlySetsContextIdentity(t *testing.T) {
	setupTokenAuthDB(t)
	seedAuthUser(t, 700, common.UserStatusEnabled)
	seedAuthToken(t, 701, 700, "roctxkey", common.TokenStatusEnabled)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	var gotID, gotTokenID int
	var gotKey string
	router.GET("/ro", TokenAuthReadOnly(), func(c *gin.Context) {
		gotID = c.GetInt("id")
		gotTokenID = c.GetInt("token_id")
		gotKey = c.GetString("token_key")
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ro", nil)
	req.Header.Set("Authorization", "Bearer sk-roctxkey")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 700, gotID)
	assert.Equal(t, 701, gotTokenID)
	assert.Equal(t, "roctxkey", gotKey)
}
