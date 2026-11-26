package system_setting

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/setting/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContactSettingsRegistrationAndValidation(t *testing.T) {
	cfg := config.GlobalConfig.Get("contact")
	require.NotNil(t, cfg)
	assert.Equal(t, DefaultContactEmail, GetContactSettings().Email)

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "valid support email",
			key:     "contact.email",
			value:   "support@openbridger.com",
			wantErr: false,
		},
		{
			name:    "invalid email without at sign",
			key:     "contact.email",
			value:   "not-an-email",
			wantErr: true,
		},
		{
			name:    "valid wechat id",
			key:     "contact.wechat_id",
			value:   "OpenBridger_Official",
			wantErr: false,
		},
		{
			name:    "valid png data url qrcode",
			key:     "contact.wechat_qrcode",
			value:   "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
			wantErr: false,
		},
		{
			name:    "valid https qrcode url",
			key:     "contact.wechat_qrcode",
			value:   "https://u.wechat.com/example",
			wantErr: false,
		},
		{
			name:    "reject svg data url to prevent stored script execution",
			key:     "contact.wechat_qrcode",
			value:   "data:image/svg+xml;base64,PHN2Zz48L3N2Zz4=",
			wantErr: true,
		},
		{
			name:    "reject javascript scheme",
			key:     "contact.wechat_qrcode",
			value:   "javascript:alert(1)",
			wantErr: true,
		},
		{
			name:    "reject oversized qrcode payload",
			key:     "contact.wechat_qrcode",
			value:   "data:image/png;base64," + strings.Repeat("A", maxContactQRCodeLen),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateContactOption(tc.key, tc.value)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
