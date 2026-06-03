package handler

import (
	"OnlinePrictice/internal/config"
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetNotifications 获取通知列表（用户 JWT）
func GetNotifications(c *gin.Context) {
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	size, _ := strconv.Atoi(c.DefaultQuery("size", config.DefaultSize))
	page, _ := strconv.Atoi(c.DefaultQuery("page", config.DefaultPage))

	list, count, err := model.GetNotificationsByUser(uid, page, size)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  list,
			"count": count,
		},
	})
}

// MarkNotificationRead 标记单条通知已读
func MarkNotificationRead(c *gin.Context) {
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	var req struct {
		Identity string `json:"identity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	if err := model.MarkNotificationRead(req.Identity, uid); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已读"})
}

// MarkAllNotificationsRead 全部标记已读
func MarkAllNotificationsRead(c *gin.Context) {
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	if err := model.MarkAllNotificationsRead(uid); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已全部标为已读"})
}

// GetUnreadNotificationCount 获取未读通知数
func GetUnreadNotificationCount(c *gin.Context) {
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity

	count, err := model.GetUnreadNotificationCount(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"count": count},
	})
}
