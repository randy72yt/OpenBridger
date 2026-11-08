package ratio_setting

import (
	"github.com/QuantumNous/new-api/setting/config"
)

// ModelPoolSetting 定义销售侧的模型池门禁（pricing plan §8.2）。
//
// 语义：基础池模型（basic_pool_models）对所有销售分组开放；基础池之外的模型
// 仅对高级分组（premium_groups，如 pro）开放。上游结算分组是采购侧概念，按
// (channel_id, model) 隔离，绝不进入这一层——销售分组词表自持有（default/pro），
// 不从上游导入。
//
// 门禁仅在两个清单都非空时生效：未配置时保持旧行为（全分组全模型），保证零回归。
// 渠道能力（abilities）在每次重建时从该配置推导，因此模型同步不会冲掉门禁。
type ModelPoolSetting struct {
	BasicPoolModels []string `json:"basic_pool_models"`
	PremiumGroups   []string `json:"premium_groups"`
}

var modelPoolSetting = ModelPoolSetting{
	BasicPoolModels: []string{},
	PremiumGroups:   []string{},
}

func init() {
	config.GlobalConfig.Register("model_pool_setting", &modelPoolSetting)
}

// ModelPoolAllows 报告某销售分组是否可以使用某模型。渠道能力重建据此跳过
// 不允许的 (group, model) 组合。
func ModelPoolAllows(group, model string) bool {
	if len(modelPoolSetting.PremiumGroups) == 0 || len(modelPoolSetting.BasicPoolModels) == 0 {
		return true
	}
	for _, g := range modelPoolSetting.PremiumGroups {
		if g == group {
			return true
		}
	}
	for _, m := range modelPoolSetting.BasicPoolModels {
		if m == model {
			return true
		}
	}
	return false
}

// GetModelPoolSetting 返回配置副本，供控制台展示。
func GetModelPoolSetting() ModelPoolSetting {
	return ModelPoolSetting{
		BasicPoolModels: append([]string(nil), modelPoolSetting.BasicPoolModels...),
		PremiumGroups:   append([]string(nil), modelPoolSetting.PremiumGroups...),
	}
}
