package service

import (
	"os"
	"strconv"
)

// CalcCheckInPoints 计算签到积分
func CalcCheckInPoints(streak int) int {
	base := EnvInt("CHECKIN_BASE_POINTS", 10)
	bonus := EnvInt("CHECKIN_STREAK_BONUS", 2)
	cap := EnvInt("CHECKIN_STREAK_CAP", 7)

	s := streak
	if s > cap {
		s = cap
	}
	return base + s*bonus
}

// EnvInt 读取环境变量并转为 int，失败返回 fallback
func EnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}
