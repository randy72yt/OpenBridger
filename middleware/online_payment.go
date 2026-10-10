package middleware

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
)

// RequireOnlinePayment prevents order creation while online payment is switched off.
//
// 这里是「下单」侧唯一的统一闸门，所有 /api/user/pay* 路由都挂在它下面。
// 它与「回调」侧（例如 EpayNotify 里的 isPaymentComplianceConfirmed）必须是同一道判断：
// 若只在回调侧校验合规声明，用户就能在管理员未确认前真实付款，而回调被拒付、额度不到账，
// 形成「钱扣了没到账」的资损。合规校验因此必须放在这里，而不是逐个 handler 补。
func RequireOnlinePayment() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !common.OnlinePaymentEnabled() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Online payment is unavailable",
			})
			return
		}
		if !operation_setting.IsPaymentComplianceConfirmed() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": common.TranslateMessage(c, i18n.MsgPaymentComplianceRequired),
			})
			return
		}
		c.Next()
	}
}
