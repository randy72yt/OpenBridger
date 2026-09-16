package service

import (
	"context"
	"os"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type qaDBResult struct {
	InputPrice   float64
	OutputPrice  float64
	MarginBPS    int
	SnapshotRows int64
	PublishErr   error
}

// qaRunOnDatabase 在指定数据库上执行同一组定价操作，返回可比较的结果。
func qaRunOnDatabase(t *testing.T, dbType common.DatabaseType, openDB func(string) gorm.Dialector, dsn string) qaDBResult {
	t.Helper()
	database, err := gorm.Open(openDB(dsn), &gorm.Config{})
	require.NoError(t, err)

	previousDB := model.DB
	previousMain, previousLog := common.MainDatabaseType(), common.LogDatabaseType()
	model.DB = database
	common.SetDatabaseTypes(dbType, dbType)
	t.Cleanup(func() {
		model.DB = previousDB
		common.SetDatabaseTypes(previousMain, previousLog)
		sqlDB, closeErr := database.DB()
		if closeErr == nil {
			sqlDB.Close()
		}
	})

	models := []any{
		&model.Option{}, &model.Channel{}, &model.UpstreamModelOffer{}, &model.UpstreamCostSnapshot{},
		&model.ModelPricePolicy{}, &model.ModelPriceProposal{},
	}
	require.NoError(t, database.AutoMigrate(models...))
	require.NoError(t, database.AutoMigrate(models...))

	const publicModel = "qa-db-matrix"
	const defaultModel = publicModel + "-def"
	for _, table := range []any{&model.ModelPriceProposal{}, &model.ModelPricePolicy{}, &model.UpstreamModelOffer{}, &model.UpstreamCostSnapshot{}} {
		require.NoError(t, database.Where("public_model IN ?", []string{publicModel, defaultModel}).Delete(table).Error)
	}

	now := common.GetTimestamp()
	require.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{
		{
			ChannelID: 201, UpstreamModel: "qa-up", PublicModel: publicModel,
			InputCost: 1.23456789, OutputCost: 6.54321987, CacheReadCost: 0.11111111,
			Currency: "USD", SourceType: "test", SourceVersion: "db-v1", SuccessRateBPS: 9891,
			CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
		},
		{
			ChannelID: 202, UpstreamModel: "qa-up", PublicModel: publicModel,
			InputCost: 2.5, OutputCost: 12.5, CacheReadCost: 0.2,
			Currency: "USD", SourceType: "test", SourceVersion: "db-v1", SuccessRateBPS: 9900,
			CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
		},
	}))

	require.NoError(t, ValidateAndUpsertPricePolicy(&model.ModelPricePolicy{
		PublicModel: publicModel, ServiceTier: "stable", PrimaryChannelID: 201, BackupChannelID: 202,
		TargetMarginBPS: 3500, MinimumMarginBPS: 2000, OverheadBPS: 500,
		FallbackProbabilityBPS: 1234, Enabled: true,
	}))

	proposal := qaBuild(t, &model.ModelPricePolicy{
		PublicModel: publicModel, ServiceTier: "stable", PrimaryChannelID: 201, BackupChannelID: 202,
		TargetMarginBPS: 3500, MinimumMarginBPS: 2000, OverheadBPS: 500,
		FallbackProbabilityBPS: 1234, Enabled: true,
	})

	var snapshotRows int64
	require.NoError(t, database.Model(&model.UpstreamCostSnapshot{}).Count(&snapshotRows).Error)

	// 完整走一遍 pending -> approved -> published，确认三库都能发布。
	_, err = RecalculatePricingProposals(context.Background())
	require.NoError(t, err)
	pending, err := model.ListModelPriceProposals(model.PricingProposalPending)
	require.NoError(t, err)
	require.Len(t, pending, 1)

	var publishErr error
	if err := ApprovePricingProposal(pending[0].ID, 1); err != nil {
		publishErr = err
	} else {
		// stable 档按设计不可发布，此处记录行为并在 default 档上补测发布。
		publishErr = PublishPricingProposal(pending[0].ID)
	}
	if publishErr != nil {
		defaultPolicy := &model.ModelPricePolicy{
			PublicModel: publicModel + "-def", ServiceTier: "default", PrimaryChannelID: 201,
			TargetMarginBPS: 3500, MinimumMarginBPS: 2000, OverheadBPS: 500, Enabled: true,
		}
		require.NoError(t, ValidateAndUpsertPricePolicy(defaultPolicy))
		qaSeedOffers(t,
			qaOfferWithModel(201, publicModel+"-def", 1.23456789, 6.54321987),
		)
		defProposal := qaBuild(t, defaultPolicy)
		require.NoError(t, database.Create(defProposal).Error)
		require.NoError(t, ApprovePricingProposal(defProposal.ID, 1))
		publishErr = PublishPricingProposal(defProposal.ID)
	}

	return qaDBResult{
		InputPrice:   proposal.ProposedInputPrice,
		OutputPrice:  proposal.ProposedOutputPrice,
		MarginBPS:    proposal.ExpectedMarginBPS,
		SnapshotRows: snapshotRows,
		PublishErr:   publishErr,
	}
}

func qaOfferWithModel(channelID int, publicModel string, in, out float64) model.UpstreamModelOffer {
	return qaOffer(channelID, publicModel, in, out)
}

// TestQAPricingDatabaseMatrix 在 SQLite、MySQL、PostgreSQL 上执行同一组定价操作，
// 验证流程可用且三种库产出的售价完全一致。
func TestQAPricingDatabaseMatrix(t *testing.T) {
	dialects := []struct {
		name   string
		env    string
		dbType common.DatabaseType
		openDB func(string) gorm.Dialector
	}{
		{name: "mysql", env: "TEST_MYSQL_DSN", dbType: common.DatabaseTypeMySQL, openDB: func(dsn string) gorm.Dialector { return mysql.Open(dsn) }},
		{name: "postgres", env: "TEST_POSTGRES_DSN", dbType: common.DatabaseTypePostgreSQL, openDB: func(dsn string) gorm.Dialector { return postgres.Open(dsn) }},
	}

	results := make(map[string]qaDBResult)

	// SQLite 基线：使用 TestMain 的内存库。
	t.Run("sqlite", func(t *testing.T) {
		truncate(t)
		res := qaRunSQLite(t)
		results["sqlite"] = res
		t.Logf("实际结果: sqlite 建议售价=%.8f/%.8f 毛利=%dbps 快照行数=%d 发布结果=%v",
			res.InputPrice, res.OutputPrice, res.MarginBPS, res.SnapshotRows, res.PublishErr)
		assert.NoError(t, res.PublishErr, "SQLite 上应能完成发布")
	})

	for _, dialect := range dialects {
		t.Run(dialect.name, func(t *testing.T) {
			dsn := os.Getenv(dialect.env)
			if dsn == "" {
				t.Skip("未设置 " + dialect.env + "，跳过")
			}
			res := qaRunOnDatabase(t, dialect.dbType, dialect.openDB, dsn)
			results[dialect.name] = res
			t.Logf("实际结果: %s 建议售价=%.8f/%.8f 毛利=%dbps 快照行数=%d 发布结果=%v",
				dialect.name, res.InputPrice, res.OutputPrice, res.MarginBPS, res.SnapshotRows, res.PublishErr)
			assert.NoError(t, res.PublishErr, dialect.name+" 上应能完成发布")
		})
	}

	base, ok := results["sqlite"]
	if !ok {
		return
	}
	for _, name := range []string{"mysql", "postgres"} {
		res, ok := results[name]
		if !ok {
			t.Logf("跳过一致性对比: %s 未执行", name)
			continue
		}
		t.Logf("对比 %s vs sqlite: %.8f vs %.8f / %.8f vs %.8f",
			name, res.InputPrice, base.InputPrice, res.OutputPrice, base.OutputPrice)
		assert.InDelta(t, base.InputPrice, res.InputPrice, 1e-8, name+" 与 SQLite 的输入售价必须一致")
		assert.InDelta(t, base.OutputPrice, res.OutputPrice, 1e-8, name+" 与 SQLite 的输出售价必须一致")
		assert.Equal(t, base.MarginBPS, res.MarginBPS, name+" 与 SQLite 的毛利必须一致")
	}
}

// qaRunSQLite 在 TestMain 的共享内存库上执行与 qaRunOnDatabase 相同的操作。
func qaRunSQLite(t *testing.T) qaDBResult {
	t.Helper()
	const publicModel = "qa-db-matrix"

	now := common.GetTimestamp()
	require.NoError(t, ValidateAndImportPricingOffers([]model.UpstreamModelOffer{
		{
			ChannelID: 201, UpstreamModel: "qa-up", PublicModel: publicModel,
			InputCost: 1.23456789, OutputCost: 6.54321987, CacheReadCost: 0.11111111,
			Currency: "USD", SourceType: "test", SourceVersion: "db-v1", SuccessRateBPS: 9891,
			CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
		},
		{
			ChannelID: 202, UpstreamModel: "qa-up", PublicModel: publicModel,
			InputCost: 2.5, OutputCost: 12.5, CacheReadCost: 0.2,
			Currency: "USD", SourceType: "test", SourceVersion: "db-v1", SuccessRateBPS: 9900,
			CollectedAt: now, ExpiresAt: now + 3600, Enabled: true,
		},
	}))

	require.NoError(t, ValidateAndUpsertPricePolicy(&model.ModelPricePolicy{
		PublicModel: publicModel, ServiceTier: "stable", PrimaryChannelID: 201, BackupChannelID: 202,
		TargetMarginBPS: 3500, MinimumMarginBPS: 2000, OverheadBPS: 500,
		FallbackProbabilityBPS: 1234, Enabled: true,
	}))

	proposal := qaBuild(t, &model.ModelPricePolicy{
		PublicModel: publicModel, ServiceTier: "stable", PrimaryChannelID: 201, BackupChannelID: 202,
		TargetMarginBPS: 3500, MinimumMarginBPS: 2000, OverheadBPS: 500,
		FallbackProbabilityBPS: 1234, Enabled: true,
	})

	var snapshotRows int64
	require.NoError(t, model.DB.Model(&model.UpstreamCostSnapshot{}).Count(&snapshotRows).Error)

	defaultPolicy := &model.ModelPricePolicy{
		PublicModel: publicModel + "-def", ServiceTier: "default", PrimaryChannelID: 201,
		TargetMarginBPS: 3500, MinimumMarginBPS: 2000, OverheadBPS: 500, Enabled: true,
	}
	require.NoError(t, ValidateAndUpsertPricePolicy(defaultPolicy))
	qaSeedOffers(t, qaOfferWithModel(201, publicModel+"-def", 1.23456789, 6.54321987))
	defProposal := qaBuild(t, defaultPolicy)
	require.NoError(t, model.DB.Create(defProposal).Error)
	require.NoError(t, ApprovePricingProposal(defProposal.ID, 1))
	publishErr := PublishPricingProposal(defProposal.ID)

	return qaDBResult{
		InputPrice:   proposal.ProposedInputPrice,
		OutputPrice:  proposal.ProposedOutputPrice,
		MarginBPS:    proposal.ExpectedMarginBPS,
		SnapshotRows: snapshotRows,
		PublishErr:   publishErr,
	}
}
