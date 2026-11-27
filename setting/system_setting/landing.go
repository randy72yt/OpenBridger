package system_setting

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/setting/config"
)

const maxLandingFeaturedModelsLen = 200

type LandingSettings struct {
	FeaturedModels string `json:"featured_models"`
}

var defaultLandingSettings = LandingSettings{
	FeaturedModels: "",
}

func init() {
	config.GlobalConfig.Register("landing", &defaultLandingSettings)
}

func GetLandingSettings() *LandingSettings {
	return &defaultLandingSettings
}

func ValidateLandingOption(key string, value string) error {
	trimmed := strings.TrimSpace(value)
	if key == "landing.featured_models" && utf8.RuneCountInString(trimmed) > maxLandingFeaturedModelsLen {
		return errors.New("首页主推模型标语长度不能超过 200 个字符")
	}
	return nil
}
