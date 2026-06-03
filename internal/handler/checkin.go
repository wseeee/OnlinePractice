package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/config"
	"OnlinePrictice/internal/service"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PostCheckIn 执行签到（用户 JWT）
func PostCheckIn(c *gin.Context) {
	u, _ := c.Get("user_claims")
	uid := u.(*helper.UserJwt).Identity
	today := helper.ShanghaiDateString()

	// Redis SETNX 防并发双签（Redis 不可用时降级到 DB 唯一索引）
	lockKey := fmt.Sprintf("checkin:lock:%s:%s", uid, today)
	ok, err := model.RDB.SetNX(config.CTX, lockKey, "1", 48*time.Hour).Result()
	if err == nil && !ok {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "今日已签到"})
		return
	}

	// 查用户
	var user model.UserBasic
	if err := model.DB.Where("identity = ?", uid).First(&user).Error; err != nil {
		model.RDB.Del(config.CTX, lockKey)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户不存在"})
		return
	}

	// 检查今日是否已签（DB 兜底）
	if user.LastCheckInDate.Valid {
		lastDateStr := helper.ShanghaiDateOnly(user.LastCheckInDate.Time)
		if lastDateStr == today {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "今日已签到"})
			return
		}
	}

	// 计算连续天数
	streak := 1
	if user.LastCheckInDate.Valid {
		if helper.IsYesterday(helper.ShanghaiNow(), user.LastCheckInDate.Time) {
			streak = user.CheckInStreak + 1
		}
	}

	points := service.CalcCheckInPoints(streak)

	// 事务：写签到流水 + 更新用户统计
	todayTime, _ := helper.ParseShanghaiDate(today)
	record := &model.UserCheckIn{
		Identity:     helper.GetUUID(),
		UserIdentity: uid,
		CheckDate:    today,
		Streak:       streak,
		Points:       points,
	}

	tx := model.DB.Begin()
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		model.RDB.Del(config.CTX, lockKey)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "签到失败"})
		return
	}

	if err := tx.Model(&model.UserBasic{}).Where("identity = ?", uid).Updates(map[string]interface{}{
		"check_in_total":     gorm.Expr("check_in_total + ?", 1),
		"check_in_streak":    streak,
		"check_in_points":    gorm.Expr("check_in_points + ?", points),
		"last_check_in_date": todayTime,
	}).Error; err != nil {
		tx.Rollback()
		model.RDB.Del(config.CTX, lockKey)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "签到失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		model.RDB.Del(config.CTX, lockKey)
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "签到失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"check_date":     today,
			"streak":         streak,
			"points_earned":  points,
			"check_in_total": user.CheckInTotal + 1,
			"check_in_points": user.CheckInPoints + points,
		},
	})
}

// GetCheckInStatus 查询签到状态（用户 JWT）
func GetCheckInStatus(c *gin.Context) {
	u2, _ := c.Get("user_claims")
	uid := u2.(*helper.UserJwt).Identity

	var user model.UserBasic
	if err := model.DB.Where("identity = ?", uid).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户不存在"})
		return
	}

	today := helper.ShanghaiNow()
	todayStr := helper.ShanghaiDateOnly(today)

	todayChecked := false
	streakActive := false
	if user.LastCheckInDate.Valid {
		lastDateStr := helper.ShanghaiDateOnly(user.LastCheckInDate.Time)
		todayChecked = lastDateStr == todayStr
		streakActive = todayChecked || helper.IsYesterday(today, user.LastCheckInDate.Time)
	}

	nextReward := service.CalcCheckInPoints(func() int {
		if streakActive {
			return user.CheckInStreak + 1
		}
		return 1
	}())

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"today_checked":      todayChecked,
			"today":              todayStr,
			"streak":             user.CheckInStreak,
			"streak_active":      streakActive,
			"check_in_total":     user.CheckInTotal,
			"check_in_points":    user.CheckInPoints,
			"next_reward_preview": nextReward,
		},
	})
}

// GetCheckInCalendar 本月签到日历（用户 JWT）
func GetCheckInCalendar(c *gin.Context) {
	u2, _ := c.Get("user_claims")
	uid := u2.(*helper.UserJwt).Identity

	now := helper.ShanghaiNow()
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))

	dates, err := model.GetCheckInDates(uid, year, month)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"year":  year,
			"month": month,
			"dates": dates,
		},
	})
}

// GetCheckInLeaderboard 签到积分排行榜（公共）
func GetCheckInLeaderboard(c *gin.Context) {
	limit := 50
	users, err := model.GetCheckInLeaderboard(limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}

	type item struct {
		Identity      string `json:"identity"`
		Name          string `json:"name"`
		CheckInPoints int    `json:"check_in_points"`
		CheckInTotal  int    `json:"check_in_total"`
		CheckInStreak int    `json:"check_in_streak"`
	}
	list := make([]item, len(users))
	for i, u := range users {
		list[i] = item{
			Identity:      u.Identity,
			Name:          u.Name,
			CheckInPoints: u.CheckInPoints,
			CheckInTotal:  u.CheckInTotal,
			CheckInStreak: u.CheckInStreak,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": list,
	})
}
