package system_setting

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

const (
	DefaultContactEmail   = "support@openbridger.com"
	maxContactEmailLen    = 254
	maxContactWeChatIDLen = 128
	maxContactQRCodeLen   = 512 * 1024
)

type ContactSettings struct {
	Email        string `json:"email"`
	WeChatID     string `json:"wechat_id"`
	WeChatQRCode string `json:"wechat_qrcode"`
}

var defaultContactSettings = ContactSettings{
	Email:        DefaultContactEmail,
	WeChatID:     "",
	WeChatQRCode: "",
}

func init() {
	config.GlobalConfig.Register("contact", &defaultContactSettings)
}

func GetContactSettings() *ContactSettings {
	return &defaultContactSettings
}

func ValidateContactOption(key string, value string) error {
	trimmed := strings.TrimSpace(value)

	switch key {
	case "contact.email":
		if len(trimmed) > maxContactEmailLen {
			return errors.New("联系邮箱长度超出限制")
		}
		if trimmed != "" && !strings.Contains(trimmed, "@") {
			return errors.New("请输入有效的联系邮箱地址")
		}
	case "contact.wechat_id":
		if len(trimmed) > maxContactWeChatIDLen {
			return errors.New("微信号长度超出限制")
		}
	case "contact.wechat_qrcode":
		if len(trimmed) > maxContactQRCodeLen {
			return errors.New("微信二维码内容过大，请压缩图片或使用图片 URL")
		}
		if trimmed != "" && !isSafeQRCodeSource(trimmed) {
			return errors.New("微信二维码仅支持 HTTPS/HTTP 链接、站内路径或 PNG/JPEG/WebP/GIF 图片")
		}
	}

	return nil
}

func isSafeQRCodeSource(raw string) bool {
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "/") {
		return true
	}

	for _, prefix := range []string{
		"data:image/png;base64,",
		"data:image/jpeg;base64,",
		"data:image/jpg;base64,",
		"data:image/webp;base64,",
		"data:image/gif;base64,",
	} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}

	return false
}
