package service

import (
	"OnlinePrictice/internal/model"
	"math"
	"os"
	"strconv"
)

// CalcPassRate 计算通过率×100 的整数，submitNum==0 时返回 -1
func CalcPassRate(passNum, submitNum int) int {
	if submitNum <= 0 {
		return -1
	}
	return int(math.Round(float64(passNum) / float64(submitNum) * 10000))
}

// CalcAutoDifficulty 根据通过率自动计算难度
func CalcAutoDifficulty(passNum, submitNum int) int {
	minSubs := EnvInt("DIFFICULTY_MIN_SUBMISSIONS", 10)
	easyRate := envFloat("DIFFICULTY_EASY_RATE", 0.60)
	mediumRate := envFloat("DIFFICULTY_MEDIUM_RATE", 0.30)

	if submitNum < minSubs {
		return 0 // 未知
	}
	rate := float64(passNum) / float64(submitNum)
	if rate >= easyRate {
		return 1 // Easy
	}
	if rate >= mediumRate {
		return 2 // Medium
	}
	return 3 // Hard
}

// RecalculateProblemDifficulty 重算题目的通过率和自动难度
func RecalculateProblemDifficulty(problemIdentity string) error {
	var p model.ProblemBasic
	if err := model.DB.Where("identity = ?", problemIdentity).First(&p).Error; err != nil {
		return err
	}

	passRate := CalcPassRate(p.PassNum, p.SubmitNum)
	updates := map[string]interface{}{
		"pass_rate": passRate,
	}

	if p.DifficultyMode == 1 {
		updates["difficulty"] = CalcAutoDifficulty(p.PassNum, p.SubmitNum)
	}

	return model.DB.Model(new(model.ProblemBasic)).
		Where("identity = ?", problemIdentity).
		Updates(updates).Error
}

func envInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func envFloat(key string, defaultVal float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return defaultVal
}
