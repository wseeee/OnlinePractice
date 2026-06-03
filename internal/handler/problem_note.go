package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SaveProblemNote 保存做题笔记（用户 JWT）
func SaveProblemNote(c *gin.Context) {
	var req struct {
		ProblemIdentity string `json:"problem_identity"`
		Content         string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProblemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
		return
	}
	req.Content = strings.TrimSpace(req.Content)

	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	if err := model.UpsertNote(helper.GetUUID(), uid, req.ProblemIdentity, req.Content); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "保存成功"})
}

// GetProblemNote 获取做题笔记（用户 JWT）
func GetProblemNote(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
		return
	}
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	note, err := model.GetNote(uid, problemIdentity)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": nil})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": note})
}

// DeleteProblemNote 删除做题笔记（用户 JWT）
func DeleteProblemNote(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
		return
	}
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	if err := model.DeleteNote(uid, problemIdentity); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
