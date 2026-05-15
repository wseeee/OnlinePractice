package middleware

import (
	"OnlinePrictice/Helper"

	"github.com/gin-gonic/gin"
)

// 验证用户是否为作者
func AuthUserCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		userClaim, err := Helper.AnalyseToken(auth)
		if err != nil {
			c.Abort()
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "未授权访问",
			})
			return
		}
		if userClaim.IsAdmin != 0 {
			c.Abort()
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "管理员请使用管理端",
			})
			return
		}
		c.Set("user_claims", userClaim)
		c.Next()
	}
}
