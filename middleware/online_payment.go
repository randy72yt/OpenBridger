package middleware

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

// RequireOnlinePayment prevents order creation in the invite-only release.
func RequireOnlinePayment() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !common.OnlinePaymentEnabled() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Online payment is unavailable",
			})
			return
		}
		c.Next()
	}
}
