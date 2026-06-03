package service

import (
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/config"
	"encoding/json"
	"time"
)

const judgeResultTTL = 5 * time.Minute

type JudgeResultCache struct {
	RecordType string `json:"record_type"`
	Status     int    `json:"status"`
	Score      int    `json:"score,omitempty"`
	Msg        string `json:"msg"`
	Passed     int    `json:"passed,omitempty"`
	Total      int    `json:"total,omitempty"`
}

func SetJudgeResult(identity string, result *JudgeResultCache) {
	data, _ := json.Marshal(result)
	model.RDB.Set(config.CTX, "judge:result:"+identity, string(data), judgeResultTTL)
}

func GetJudgeResult(identity string) *JudgeResultCache {
	val, err := model.RDB.Get(config.CTX, "judge:result:"+identity).Result()
	if err != nil {
		return nil
	}
	var result JudgeResultCache
	if json.Unmarshal([]byte(val), &result) != nil {
		return nil
	}
	return &result
}

func StatusMsg(status int) string {
	msgMap := map[int]string{
		0: "排队中",
		1: "答案正确",
		2: "答案错误",
		3: "运行超时",
		4: "运行超内存",
		5: "编译错误",
		6: "无效代码",
	}
	if m, ok := msgMap[status]; ok {
		return m
	}
	return "未知状态"
}
