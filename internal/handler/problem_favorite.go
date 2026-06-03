package handler

import (
	"OnlinePrictice/internal/config"
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AddFavorite 收藏题目（用户 JWT）
func AddFavorite(c *gin.Context) {
	var req struct {
		ProblemIdentity string `json:"problem_identity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProblemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
		return
	}
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	if model.IsFavorited(uid, req.ProblemIdentity) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "已收藏"})
		return
	}

	err := model.AddFavorite(helper.GetUUID(), uid, req.ProblemIdentity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "收藏失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "收藏成功"})
}

// RemoveFavorite 取消收藏（用户 JWT）
func RemoveFavorite(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
		return
	}
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	if err := model.RemoveFavorite(uid, problemIdentity); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "取消失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已取消收藏"})
}

// GetFavorites 收藏列表（用户 JWT）
func GetFavorites(c *gin.Context) {
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	size, _ := strconv.Atoi(c.DefaultQuery("size", config.DefaultSize))
	page, _ := strconv.Atoi(c.DefaultQuery("page", config.DefaultPage))

	list, count, err := model.GetFavoriteProblems(uid, page, size)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{"list": list, "count": count},
	})
}

// CheckFavorite 检查是否已收藏（用户 JWT）
func CheckFavorite(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
		return
	}
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	favorited := model.IsFavorited(uid, problemIdentity)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"favorited": favorited}})
}
