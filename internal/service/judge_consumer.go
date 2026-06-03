package service

import (
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/config"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

const maxRetries = 3

// StartJudgeConsumers 启动判题消费者池
func StartJudgeConsumers() {
	if model.MQChannel == nil {
		log.Println("RabbitMQ not available, skipping consumer startup")
		return
	}

	workers := 5
	if w := os.Getenv("JUDGE_WORKERS"); w != "" {
		if n, err := strconv.Atoi(w); err == nil && n > 0 {
			workers = n
		}
	}

	for i := 0; i < workers; i++ {
		go consumeQueue(model.QueueJudgeSubmit, i)
		go consumeQueue(model.QueueJudgeContest, i)
	}
	log.Printf("Judge consumers started: %d workers per queue", workers)
}

func consumeQueue(queue string, workerID int) {
	ch, err := model.MQConn.Channel()
	if err != nil {
		log.Printf("[Worker %s#%d] channel error: %v", queue, workerID, err)
		return
	}
	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		log.Printf("[Worker %s#%d] Qos error: %v", queue, workerID, err)
		return
	}

	msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[Worker %s#%d] consume error: %v", queue, workerID, err)
		return
	}

	for msg := range msgs {
		handleJudgeMessage(queue, msg)
	}
}

func handleJudgeMessage(queue string, msg amqp.Delivery) {
	var task model.JudgeTask
	if err := json.Unmarshal(msg.Body, &task); err != nil {
		log.Printf("[Consumer] invalid message: %v", err)
		msg.Ack(false)
		return
	}

	// 读取重试次数
	retryCount := 0
	if rc, ok := msg.Headers["x-retry-count"]; ok {
		switch v := rc.(type) {
		case int32:
			retryCount = int(v)
		case int64:
			retryCount = int(v)
		case int:
			retryCount = v
		}
	}

	var err error
	switch task.RecordType {
	case "submit":
		err = handleSubmitTask(&task)
	case "contest":
		err = handleContestTask(&task)
	default:
		log.Printf("[Consumer] unknown record_type: %s", task.RecordType)
		msg.Ack(false)
		return
	}

	if err != nil {
		log.Printf("[Consumer] task failed (retry=%d): %v", retryCount, err)
		handleRetry(&task, queue, retryCount)
	}
	msg.Ack(false)
}

func handleSubmitTask(task *model.JudgeTask) error {
	passed, total, status, err := JudgeCodeWithLang(task.CodePath, task.ProblemIdentity, task.Language)
	if err != nil {
		return err
	}

	// 更新 submit_basic.status
	err = model.DB.Model(new(model.SubmitBasic)).
		Where("identity = ?", task.SubmitIdentity).
		Update("status", status).Error
	if err != nil {
		return fmt.Errorf("update submit_basic: %w", err)
	}

	// AC 时更新 pass_num（submit_num 已在 Producer +1）
	if status == 1 {
		model.DB.Model(new(model.UserBasic)).
			Where("identity = ?", task.UserIdentity).
			Update("pass_num", gorm.Expr("pass_num + ?", 1))
		model.DB.Model(new(model.ProblemBasic)).
			Where("identity = ?", task.ProblemIdentity).
			Update("pass_num", gorm.Expr("pass_num + ?", 1))
	}

	// 重算题目难度（无论 AC/WA，submit_num 已变）
	_ = RecalculateProblemDifficulty(task.ProblemIdentity)

	// 写 Redis 结果缓存
	SetJudgeResult(task.SubmitIdentity, &JudgeResultCache{
		RecordType: "submit",
		Status:     status,
		Msg:        StatusMsg(status),
		Passed:     passed,
		Total:      total,
	})

	// 通知 WebSocket
	notifyJudgeResult(task.UserIdentity, task.SubmitIdentity, "submit", status, 0)
	return nil
}

func handleContestTask(task *model.JudgeTask) error {
	passed, total, status, err := JudgeCodeWithLang(task.CodePath, task.ProblemIdentity, task.Language)
	if err != nil {
		return err
	}

	// OI 计分
	score := 0
	if total > 0 {
		score = passed * task.ProblemScore / total
	}

	// 更新 contest_submit
	err = model.DB.Model(new(model.ContestSubmit)).
		Where("identity = ?", task.SubmitIdentity).
		Updates(map[string]interface{}{
			"status": status,
			"score":  score,
		}).Error
	if err != nil {
		return fmt.Errorf("update contest_submit: %w", err)
	}

	// 失效比赛排行榜缓存
	model.RDB.Del(config.CTX, "contest_rank:"+task.ContestIdentity)

	// 写 Redis 结果缓存
	SetJudgeResult(task.SubmitIdentity, &JudgeResultCache{
		RecordType: "contest",
		Status:     status,
		Score:      score,
		Msg:        StatusMsg(status),
		Passed:     passed,
		Total:      total,
	})

	// 通知 WebSocket
	notifyJudgeResult(task.UserIdentity, task.SubmitIdentity, "contest", status, score)
	return nil
}

func handleRetry(task *model.JudgeTask, queue string, retryCount int) {
	if retryCount >= maxRetries {
		finalizeJudgeFailure(task)
		return
	}
	if err := model.PublishToDLX(task, queue, retryCount+1); err != nil {
		log.Printf("[Consumer] DLX publish failed: %v", err)
		finalizeJudgeFailure(task)
	}
}

func finalizeJudgeFailure(task *model.JudgeTask) {
	const failMsg = "判题服务异常，请重新提交"

	switch task.RecordType {
	case "submit":
		model.DB.Model(new(model.SubmitBasic)).
			Where("identity = ?", task.SubmitIdentity).
			Update("status", 6)
	case "contest":
		model.DB.Model(new(model.ContestSubmit)).
			Where("identity = ?", task.SubmitIdentity).
			Updates(map[string]interface{}{"status": 6, "score": 0})
		model.RDB.Del(config.CTX, "contest_rank:"+task.ContestIdentity)
	}

	SetJudgeResult(task.SubmitIdentity, &JudgeResultCache{
		RecordType: task.RecordType,
		Status:     6,
		Msg:        failMsg,
	})

	notifyJudgeResult(task.UserIdentity, task.SubmitIdentity, task.RecordType, 6, 0)
}

func notifyJudgeResult(userIdentity, submitIdentity, recordType string, status, score int) {
	payload := map[string]interface{}{
		"type":            "judge_result",
		"submit_identity": submitIdentity,
		"record_type":     recordType,
		"status":          status,
		"score":           score,
		"msg":             StatusMsg(status),
	}
	data, _ := json.Marshal(payload)
	NotifyJudgeResult(userIdentity, data)
}
