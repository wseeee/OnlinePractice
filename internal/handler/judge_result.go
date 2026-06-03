package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetSubmitResult 查询判题结果（轮询接口）
// @Tags 用户私有方法
// @Summary 查询判题结果
// @Param Authorization header string true "Authorization"
// @Param identity query string true "submit_identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /user/submit-result [get]
func GetSubmitResult(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不完整"})
		return
	}

	u, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户信息获取失败"})
		return
	}
	userClaim := u.(*helper.UserJwt)

	// 优先读 Redis 缓存
	cached := service.GetJudgeResult(identity)
	if cached != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": cached,
		})
		return
	}

	// 缓存未命中，查 DB
	// 先查 submit_basic
	var sb model.SubmitBasic
	err := model.DB.Where("identity = ?", identity).First(&sb).Error
	if err == nil {
		if sb.UserIdentity != userClaim.Identity {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "无权查看"})
			return
		}
		result := &service.JudgeResultCache{
			RecordType: "submit",
			Status:     sb.Status,
			Msg:        service.StatusMsg(sb.Status),
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
		return
	}

	// 再查 contest_submit
	var cs model.ContestSubmit
	err = model.DB.Where("identity = ?", identity).First(&cs).Error
	if err == nil {
		if cs.UserIdentity != userClaim.Identity {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "无权查看"})
			return
		}
		result := &service.JudgeResultCache{
			RecordType: "contest",
			Status:     cs.Status,
			Score:      cs.Score,
			Msg:        service.StatusMsg(cs.Status),
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "记录不存在"})
}
