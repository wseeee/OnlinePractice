package middleware

import (
	"OnlinePrictice/Helper"

	"github.com/gin-gonic/gin"
)

// 验证用户是否为作者
func AuthAdiminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		userClaim, err := Helper.AnalyseToken(auth)
		if err != nil {
			c.Abort()
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "UnAuthorized Authorization",
			})
			return
		}
		if userClaim.IsAdmin != 1 {
			c.Abort()
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "UnAuthorized Admin",
			})
			return
		}
		c.Next()
	}
}
