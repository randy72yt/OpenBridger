package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUpdateOptionRejectsInvalidTaskBillingExpressions(t *testing.T) {
	const pluginKey = "billing-save-probe"
	const modelName = "billing-save-model"
	source := `
export const meta = {
  apiVersion: 1, key: "billing-save-probe", name: "Billing Save Probe", version: "1.0.0", author: {name: "Test"},
  models: ["billing-save-model"], fetchMode: "per_task",
  usageSchema: {seconds: {type: "number", unit: "second"}}
};
export function buildSubmitRequest() { return {}; }
export function parseSubmitResponse() { return {}; }
export function buildQueryRequest() { return {}; }
export function parseTaskResult() { return {}; }
`
	_, err := jsplugin.DefaultRegistry.Register(source, jsplugin.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { jsplugin.DefaultRegistry.Unregister(pluginKey) })

	tests := []struct {
		name       string
		expression string
		errorText  string
	}{
		{
			name:       "invalid syntax",
			expression: `tier("base",`,
			errorText:  "expr compile error",
		},
		{
			name:       "undeclared usage key",
			expression: `tier("base", u("clips") * 0.1)`,
			errorText:  `usage key \"clips\" is not declared`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			expressions, marshalErr := common.Marshal(map[string]string{modelName: testCase.expression})
			require.NoError(t, marshalErr)
			body, marshalErr := common.Marshal(OptionUpdateRequest{
				Key:   "billing_setting.billing_expr",
				Value: string(expressions),
			})
			require.NoError(t, marshalErr)
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequest(http.MethodPut, "/api/option/", strings.NewReader(string(body)))

			UpdateOption(context)

			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.Contains(t, recorder.Body.String(), `"success":false`)
			assert.Contains(t, recorder.Body.String(), modelName)
			assert.Contains(t, recorder.Body.String(), testCase.errorText)
		})
	}
}

func TestUpdateOptionRejectsUsageExpressionWithoutTaskPlugin(t *testing.T) {
	const modelName = "billing-save-model-without-plugin"
	expressions, err := common.Marshal(map[string]string{
		modelName: `u("mode") == "std" ? 1 : 2`,
	})
	require.NoError(t, err)
	body, err := common.Marshal(OptionUpdateRequest{
		Key:   "billing_setting.billing_expr",
		Value: string(expressions),
	})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/option/",
		strings.NewReader(string(body)),
	)

	UpdateOption(context)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"success":false`)
	assert.Contains(t, recorder.Body.String(), modelName)
	assert.Contains(t, recorder.Body.String(), "mode")
	assert.Contains(t, recorder.Body.String(), "no task plugin usage schema")
}

func setupBillingAliasOptionDB(t *testing.T) {
	t.Helper()
	previousDB := model.DB
	previousLogDB := model.LOG_DB
	previousType := common.MainDatabaseType()
	previousCache := common.MemoryCacheEnabled
	previousMap := common.OptionMap
	previousRedis := common.RedisEnabled
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.AutoMigrate(&model.Channel{}, &model.Option{}, &model.Log{}, &model.AuditLog{}, &model.User{}))
	model.DB = database
	model.LOG_DB = database
	common.SetMainDatabaseType(common.DatabaseTypeSQLite)
	common.MemoryCacheEnabled = false
	common.RedisEnabled = false
	common.OptionMap = map[string]string{}
	t.Cleanup(func() {
		model.DB = previousDB
		model.LOG_DB = previousLogDB
		common.SetMainDatabaseType(previousType)
		common.MemoryCacheEnabled = previousCache
		common.OptionMap = previousMap
		common.RedisEnabled = previousRedis
		model.InitChannelCache()
	})
}

func TestUpdateOptionAliasBillingExprUsesPluginSchema(t *testing.T) {
	setupBillingAliasOptionDB(t)
	const pluginKey = "billing-alias-probe"
	source := `
export const meta = {
  apiVersion: 1, key: "billing-alias-probe", name: "Billing Alias Probe", version: "1.0.0", author: {name: "Test"},
  models: ["declared-model"], fetchMode: "per_task",
  usageSchema: {seconds: {type: "number", unit: "second"}}
};
export function buildSubmitRequest() { return {}; }
export function parseSubmitResponse() { return {}; }
export function buildQueryRequest() { return {}; }
export function parseTaskResult() { return {}; }
`
	_, err := jsplugin.DefaultRegistry.Register(source, jsplugin.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { jsplugin.DefaultRegistry.Unregister(pluginKey) })

	mapping := `{"alias-model":"declared-model"}`
	require.NoError(t, model.DB.Create(&model.Channel{
		Id:           1,
		Type:         54,
		Key:          "key-1",
		Status:       common.ChannelStatusEnabled,
		Name:         "ch-1",
		Group:        "default",
		Models:       "alias-model,declared-model",
		ModelMapping: &mapping,
	}).Error)
	model.InitChannelCache()

	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() {
		require.NoError(t, config.GlobalConfig.LoadFromDB(saved))
	})

	putExpr := func(modelName, expression string) *httptest.ResponseRecorder {
		t.Helper()
		expressions, marshalErr := common.Marshal(map[string]string{modelName: expression})
		require.NoError(t, marshalErr)
		body, marshalErr := common.Marshal(OptionUpdateRequest{
			Key:   "billing_setting.billing_expr",
			Value: string(expressions),
		})
		require.NoError(t, marshalErr)
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Request = httptest.NewRequest(http.MethodPut, "/api/option/", strings.NewReader(string(body)))
		UpdateOption(context)
		return recorder
	}

	accepted := putExpr("alias-model", `u("seconds")`)
	assert.Equal(t, http.StatusOK, accepted.Code)
	assert.Contains(t, accepted.Body.String(), `"success":true`)

	rejectedKey := putExpr("alias-model", `u("clips")`)
	assert.Equal(t, http.StatusOK, rejectedKey.Code)
	assert.Contains(t, rejectedKey.Body.String(), `"success":false`)
	assert.Contains(t, rejectedKey.Body.String(), `usage key \"clips\" is not declared`)

	unresolvable := putExpr("unknown-alias-model", `u("seconds")`)
	assert.Equal(t, http.StatusOK, unresolvable.Code)
	assert.Contains(t, unresolvable.Body.String(), `"success":false`)
	assert.Contains(t, unresolvable.Body.String(), "no task plugin usage schema")
}

func putOptionRequest(t *testing.T, key, value string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := common.Marshal(OptionUpdateRequest{Key: key, Value: value})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPut, "/api/option/", strings.NewReader(string(body)))
	UpdateOption(context)
	return recorder
}

func storedOption(t *testing.T, key string) (string, bool) {
	t.Helper()
	var row model.Option
	err := model.DB.Where(&model.Option{Key: key}).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false
	}
	require.NoError(t, err)
	return row.Value, true
}

func TestLegacySellPriceWritesRejectedWhenPricingControlEnabled(t *testing.T) {
	savedRatios := map[string]struct {
		value  string
		update func(string) error
	}{
		"ModelRatio":           {ratio_setting.ModelRatio2JSONString(), ratio_setting.UpdateModelRatioByJSONString},
		"ModelPrice":           {ratio_setting.ModelPrice2JSONString(), ratio_setting.UpdateModelPriceByJSONString},
		"CompletionRatio":      {ratio_setting.CompletionRatio2JSONString(), ratio_setting.UpdateCompletionRatioByJSONString},
		"CacheRatio":           {ratio_setting.CacheRatio2JSONString(), ratio_setting.UpdateCacheRatioByJSONString},
		"CreateCacheRatio":     {ratio_setting.CreateCacheRatio2JSONString(), ratio_setting.UpdateCreateCacheRatioByJSONString},
		"ImageRatio":           {ratio_setting.ImageRatio2JSONString(), ratio_setting.UpdateImageRatioByJSONString},
		"AudioRatio":           {ratio_setting.AudioRatio2JSONString(), ratio_setting.UpdateAudioRatioByJSONString},
		"AudioCompletionRatio": {ratio_setting.AudioCompletionRatio2JSONString(), ratio_setting.UpdateAudioCompletionRatioByJSONString},
		"GroupRatio":           {ratio_setting.GroupRatio2JSONString(), ratio_setting.UpdateGroupRatioByJSONString},
	}
	savedConfigs := config.GlobalConfig.ExportAllConfigs()
	t.Cleanup(func() {
		for _, item := range savedRatios {
			require.NoError(t, item.update(item.value))
		}
		require.NoError(t, config.GlobalConfig.LoadFromDB(savedConfigs))
	})

	setupBillingAliasOptionDB(t)
	t.Setenv("PRICING_CONTROL_ENABLED", "false")
	seeded := putOptionRequest(t, "ModelRatio", `{"keep-model":1}`)
	require.Equal(t, http.StatusOK, seeded.Code)
	require.Contains(t, seeded.Body.String(), `"success":true`)
	seededRatio, ok := storedOption(t, "ModelRatio")
	require.True(t, ok)
	require.Contains(t, seededRatio, "keep-model")

	t.Setenv("PRICING_CONTROL_ENABLED", "true")
	imageBefore := ratio_setting.ImageRatio2JSONString()
	sellPriceKeys := []string{
		"ModelRatio", "CompletionRatio", "CacheRatio", "ModelPrice",
		"ImageRatio", "AudioRatio", "AudioCompletionRatio", "CreateCacheRatio",
		"billing_setting.billing_expr", "billing_setting.billing_mode",
	}
	for _, key := range sellPriceKeys {
		t.Run(key, func(t *testing.T) {
			value := `{"hijack":1}`
			if key == "billing_setting.billing_expr" {
				value = `{"hijack":"tier(\"base\", p * 1 + c * 1)"}`
			}
			if key == "billing_setting.billing_mode" {
				value = `{"hijack":"tiered_expr"}`
			}
			before, beforeOK := storedOption(t, key)
			recorder := putOptionRequest(t, key, value)
			assert.Equal(t, http.StatusForbidden, recorder.Code)
			assert.Contains(t, recorder.Body.String(), "approving and publishing")
			after, afterOK := storedOption(t, key)
			assert.Equal(t, beforeOK, afterOK)
			assert.Equal(t, before, after)
			assert.NotContains(t, after, "hijack")
		})
	}
	assert.Equal(t, imageBefore, ratio_setting.ImageRatio2JSONString())

	groupRecorder := putOptionRequest(t, "GroupRatio", `{"default":1,"vip":1.2}`)
	assert.Equal(t, http.StatusOK, groupRecorder.Code)
	assert.Contains(t, groupRecorder.Body.String(), `"success":true`)
	assert.Equal(t, 1.2, ratio_setting.GetGroupRatioCopy()["vip"])

	resetRecorder := httptest.NewRecorder()
	resetContext, _ := gin.CreateTestContext(resetRecorder)
	resetContext.Request = httptest.NewRequest(http.MethodPost, "/api/option/rest_model_ratio", nil)
	ResetModelRatio(resetContext)
	assert.Equal(t, http.StatusForbidden, resetRecorder.Code)
	resetRatio, _ := storedOption(t, "ModelRatio")
	assert.Contains(t, resetRatio, "keep-model")
	assert.NotContains(t, resetRatio, "gpt-4")

	bulkErr := model.UpdateOptionsBulk(map[string]string{"CacheRatio": `{"hijack":1}`})
	require.ErrorIs(t, bulkErr, model.ErrManualPricingWriteBlocked)
	require.NoError(t, model.UpdateOptionsBulk(map[string]string{"Notice": "still-writable"}))
	notice, noticeOK := storedOption(t, "Notice")
	require.True(t, noticeOK)
	assert.Equal(t, "still-writable", notice)

	snapshot, err := model.GetModelPricingSnapshot([]string{"keep-model"})
	require.NoError(t, err)
	require.Len(t, snapshot.Entries, 1)
	require.NoError(t, model.UpdateModelPricing([]model.ModelPricingChange{{
		ModelName:       "keep-model",
		ExpectedVersion: snapshot.Entries[0].Version,
		Pricing:         model.PricingValues{"ModelPrice": float64(3)},
	}}))
	published, publishedOK := storedOption(t, "ModelPrice")
	require.True(t, publishedOK)
	assert.Contains(t, published, "keep-model")

	patchBody, err := common.Marshal(map[string]any{
		"changes": []model.ModelPricingChange{{
			ModelName:       "keep-model",
			ExpectedVersion: snapshot.Entries[0].Version,
			Pricing:         model.PricingValues{"ModelPrice": float64(9)},
		}},
		"override": map[string]any{"enabled": true, "reason": "direct API write"},
	})
	require.NoError(t, err)
	patchRecorder := httptest.NewRecorder()
	patchContext, _ := gin.CreateTestContext(patchRecorder)
	patchContext.Request = httptest.NewRequest(http.MethodPatch, "/api/option/model_pricing", strings.NewReader(string(patchBody)))
	UpdateModelPricingConfig(patchContext)
	assert.Equal(t, http.StatusForbidden, patchRecorder.Code)
	unchanged, _ := storedOption(t, "ModelPrice")
	assert.Equal(t, published, unchanged)

	statusRecorder := httptest.NewRecorder()
	statusContext, _ := gin.CreateTestContext(statusRecorder)
	statusContext.Request = httptest.NewRequest(http.MethodGet, "/api/status", nil)
	GetStatus(statusContext)
	var statusPayload struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(statusRecorder.Body.Bytes(), &statusPayload))
	require.True(t, statusPayload.Success)
	assert.Equal(t, true, statusPayload.Data["pricing_control_enabled"])

	t.Setenv("PRICING_CONTROL_ENABLED", "false")
	allowed := putOptionRequest(t, "ModelRatio", `{"keep-model":2}`)
	assert.Equal(t, http.StatusOK, allowed.Code)
	assert.Contains(t, allowed.Body.String(), `"success":true`)
	updatedRatio, _ := storedOption(t, "ModelRatio")
	assert.Contains(t, updatedRatio, "keep-model")
	assert.NotContains(t, updatedRatio, `"keep-model":1`)
}
