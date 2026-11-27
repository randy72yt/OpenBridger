package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/common/testdb"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupOpsSnapshotTestDB(t *testing.T) {
	t.Helper()
	db := testdb.Open()
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{},
		&model.Channel{},
		&model.Token{},
		&model.TopUp{},
		&model.SubscriptionPlan{},
		&model.SubscriptionOrder{},
		&model.Redemption{},
		&model.UpstreamModelOffer{},
		&model.UpstreamCostCalibration{},
		&model.ModelPriceProposal{},
		&model.Log{},
		&model.QuotaData{},
		&model.PerfMetric{},
		&model.PerfMetricAux{},
	))
	model.InitColumnNames()
	for _, tbl := range []any{
		&model.TopUp{},
		&model.SubscriptionOrder{},
		&model.Redemption{},
		&model.Channel{},
		&model.Log{},
		&model.QuotaData{},
		&model.PerfMetric{},
		&model.PerfMetricAux{},
		&model.ModelPriceProposal{},
	} {
		require.NoError(t, db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(tbl).Error)
	}
	resetOpsSnapshotCacheForTest()
}

func TestGetOpsSnapshotAuthAndPayload(t *testing.T) {
	setupOpsSnapshotTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/ops/snapshot", GetOpsSnapshot)

	validSecret := "ops-secret-token-with-at-least-32-chars-998877"

	cases := []struct {
		name       string
		envToken   string
		path       string
		authHeader string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "disabled when env unset",
			envToken:   "",
			path:       "/api/ops/snapshot",
			authHeader: "Bearer " + validSecret,
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   "OPS_SNAPSHOT_DISABLED",
		},
		{
			name:       "disabled when env token shorter than 32 chars",
			envToken:   "too-short-secret",
			path:       "/api/ops/snapshot",
			authHeader: "Bearer too-short-secret",
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   "OPS_SNAPSHOT_DISABLED",
		},
		{
			name:       "missing authorization header rejected",
			envToken:   validSecret,
			path:       "/api/ops/snapshot",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "OPS_UNAUTHORIZED",
		},
		{
			name:       "non-bearer scheme rejected",
			envToken:   validSecret,
			path:       "/api/ops/snapshot",
			authHeader: "Basic " + validSecret,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "OPS_UNAUTHORIZED",
		},
		{
			name:       "query token rejected even with valid bearer header",
			envToken:   validSecret,
			path:       "/api/ops/snapshot?token=" + validSecret,
			authHeader: "Bearer " + validSecret,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "OPS_QUERY_TOKEN_FORBIDDEN",
		},
		{
			name:       "wrong bearer token rejected",
			envToken:   validSecret,
			path:       "/api/ops/snapshot",
			authHeader: "Bearer wrong-token-value-that-is-also-long-enough-0000",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "OPS_UNAUTHORIZED",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(opsSnapshotTokenEnv, tc.envToken)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, tc.wantStatus, w.Code)

			var resp struct {
				Success bool   `json:"success"`
				Code    string `json:"code"`
			}
			require.NoError(t, common.Unmarshal(w.Body.Bytes(), &resp))
			assert.False(t, resp.Success)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.NotContains(t, w.Body.String(), validSecret)
		})
	}

	t.Run("valid bearer returns aggregated read-only snapshot without secrets", func(t *testing.T) {
		t.Setenv(opsSnapshotTokenEnv, validSecret)
		operation_setting.Price = 7.3
		operation_setting.USDExchangeRate = 7.3
		common.QuotaPerUnit = 500000

		now := time.Now().Unix()
		secretUpstreamKey := "sk-upstream-super-secret-do-not-leak"
		require.NoError(t, model.DB.Create(&model.Channel{
			Name:    "primary-openai",
			Key:     secretUpstreamKey,
			Status:  common.ChannelStatusEnabled,
			Balance: 3.5,
		}).Error)
		require.NoError(t, model.DB.Create(&model.Channel{
			Name:    "disabled-claude",
			Key:     "sk-disabled-secret",
			Status:  common.ChannelStatusAutoDisabled,
			Balance: 20.0,
		}).Error)
		require.NoError(t, model.DB.Create(&model.TopUp{
			UserId:          1,
			Money:           100,
			PaymentProvider: model.PaymentProviderEpay,
			PaymentMethod:   "alipay",
			Status:          common.TopUpStatusSuccess,
			CompleteTime:    now - 10,
		}).Error)

		req := httptest.NewRequest(http.MethodGet, "/api/ops/snapshot", nil)
		req.Header.Set("Authorization", "Bearer "+validSecret)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		assert.NotContains(t, w.Body.String(), secretUpstreamKey)
		assert.NotContains(t, w.Body.String(), "primary-openai")

		var resp struct {
			Success bool            `json:"success"`
			Data    OpsSnapshotData `json:"data"`
		}
		require.NoError(t, common.Unmarshal(w.Body.Bytes(), &resp))
		require.True(t, resp.Success)
		assert.True(t, resp.Data.DBHealthy)
		assert.Equal(t, 100.0, resp.Data.Today.CashRevenueCNY)
		assert.Equal(t, int64(1), resp.Data.Today.PaidOrderCount)
		assert.Equal(t, 2, resp.Data.Channels.Total)
		assert.Equal(t, 1, resp.Data.Channels.Enabled)
		assert.Equal(t, 1, resp.Data.Channels.AutoDisabled)
		assert.Equal(t, 1, resp.Data.Channels.LowBalanceCount)
		assert.Equal(t, 23.5, resp.Data.Channels.TotalBalanceUSD)
	})
}
