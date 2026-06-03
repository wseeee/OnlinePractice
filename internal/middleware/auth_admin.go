package middleware

import (
	"OnlinePrictice/internal/helper"

	"github.com/gin-gonic/gin"
)

// 验证用户是否为作者
func AuthAdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		userClaim, err := helper.AnalyseToken(auth)
		if err != nil {
			c.Abort()
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "未授权访问",
			})
			return
		}
		if userClaim.IsAdmin != 1 {
			c.Abort()
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "无管理员权限",
			})
			return
		}
		c.Set("user_claims", userClaim)
		c.Next()
	}
}
