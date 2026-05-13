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
				"msg":  "Authorization Userization",
			})
			return
		}
		if userClaim.IsAdmin != 0 {
			c.Abort()
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "Userization User",
			})
			return
		}
		c.Set("user_claims", userClaim)
		c.Next()
	}
}
