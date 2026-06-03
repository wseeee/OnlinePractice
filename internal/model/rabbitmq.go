package model

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueJudgeSubmit  = "judge.submit"
	QueueJudgeContest = "judge.contest"
	ExchangeJudge     = "judge"
	ExchangeJudgeDLX  = "judge.dlx"
	QueueJudgeDLQ     = "judge.dlq"
	RoutingKeySubmit  = "submit"
	RoutingKeyContest = "contest"
)

type JudgeTask struct {
	SubmitIdentity  string `json:"submit_identity"`
	RecordType      string `json:"record_type"` // "submit" | "contest"
	CodePath        string `json:"code_path"`
	ProblemIdentity string `json:"problem_identity"`
	UserIdentity    string `json:"user_identity"`
	Language        string `json:"language"`
	IsContest       bool   `json:"is_contest"`
	ContestIdentity string `json:"contest_identity,omitempty"`
	ProblemScore    int    `json:"problem_score,omitempty"`
}

var (
	MQConn    *amqp.Connection
	MQChannel *amqp.Channel
	mqMu      sync.Mutex
)

func InitRabbitMQ() {
	host := getEnv("RABBITMQ_HOST", "127.0.0.1")
	port := getEnv("RABBITMQ_PORT", "5672")
	user := getEnv("RABBITMQ_USER", "guest")
	pass := getEnv("RABBITMQ_PASSWORD", "guest")
	vhost := getEnv("RABBITMQ_VHOST", "/")

	addr := fmt.Sprintf("amqp://%s:%s@%s:%s%s", user, pass, host, port, vhost)

	var err error
	for i := range 30 {
		MQConn, err = amqp.Dial(addr)
		if err == nil {
			log.Println("RabbitMQ connected successfully")
			break
		}
		log.Printf("RabbitMQ not ready (attempt %d/30): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Printf("Warning: RabbitMQ not available: %v", err)
		return
	}

	MQChannel, err = MQConn.Channel()
	if err != nil {
		log.Printf("Warning: RabbitMQ channel error: %v", err)
		return
	}

	if err := declareTopology(MQChannel); err != nil {
		log.Printf("Warning: RabbitMQ topology error: %v", err)
		return
	}
	log.Println("RabbitMQ topology declared")

	go watchReconnect(addr)
}

func declareTopology(ch *amqp.Channel) error {
	// 主 Exchange
	if err := ch.ExchangeDeclare(ExchangeJudge, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	// DLX Exchange
	if err := ch.ExchangeDeclare(ExchangeJudgeDLX, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	// judge.submit 队列
	submitArgs := amqp.Table{
		"x-dead-letter-exchange": ExchangeJudgeDLX,
	}
	if _, err := ch.QueueDeclare(QueueJudgeSubmit, true, false, false, false, submitArgs); err != nil {
		return err
	}
	if err := ch.QueueBind(QueueJudgeSubmit, RoutingKeySubmit, ExchangeJudge, false, nil); err != nil {
		return err
	}

	// judge.contest 队列
	contestArgs := amqp.Table{
		"x-dead-letter-exchange": ExchangeJudgeDLX,
	}
	if _, err := ch.QueueDeclare(QueueJudgeContest, true, false, false, false, contestArgs); err != nil {
		return err
	}
	if err := ch.QueueBind(QueueJudgeContest, RoutingKeyContest, ExchangeJudge, false, nil); err != nil {
		return err
	}

	// judge.dlq 队列 — 消息过期后死信回 judge exchange，routing key 保持不变
	dlqArgs := amqp.Table{
		"x-dead-letter-exchange": ExchangeJudge,
	}
	if _, err := ch.QueueDeclare(QueueJudgeDLQ, true, false, false, false, dlqArgs); err != nil {
		return err
	}
	// DLQ 接收来自 DLX 的消息，routing key 为原队列名
	if err := ch.QueueBind(QueueJudgeDLQ, QueueJudgeSubmit, ExchangeJudgeDLX, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(QueueJudgeDLQ, QueueJudgeContest, ExchangeJudgeDLX, false, nil); err != nil {
		return err
	}

	return nil
}

func watchReconnect(addr string) {
	for {
		errCh := make(chan *amqp.Error)
		MQConn.NotifyClose(errCh)

		err := <-errCh
		if err == nil {
			return
		}
		log.Printf("RabbitMQ connection lost: %v, reconnecting in 5s...", err)
		time.Sleep(5 * time.Second)

		for i := 0; i < 30; i++ {
			conn, dErr := amqp.Dial(addr)
			if dErr == nil {
				mqMu.Lock()
				MQConn = conn
				MQChannel, dErr = conn.Channel()
				if dErr == nil {
					dErr = declareTopology(MQChannel)
				}
				mqMu.Unlock()
				if dErr == nil {
					log.Println("RabbitMQ reconnected successfully")
					break // 跳出内层循环，回到外层重新监听关闭
				}
				log.Printf("RabbitMQ reconnect failed (attempt %d/30): %v", i+1, dErr)
			} else {
				log.Printf("RabbitMQ reconnect failed (attempt %d/30): %v", i+1, dErr)
			}
			time.Sleep(2 * time.Second)
		}
	}
}

// PublishJudgeTask 发布判题任务到指定队列
func PublishJudgeTask(routingKey string, task *JudgeTask) error {
	mqMu.Lock()
	defer mqMu.Unlock()

	if MQChannel == nil {
		return fmt.Errorf("RabbitMQ not connected")
	}

	body, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return MQChannel.Publish(ExchangeJudge, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

// PublishToDLX 发布到死信交换机（重试用）
// routingKey = 原队列名（judge.submit / judge.contest），DLQ 绑定了这两个 key
func PublishToDLX(task *JudgeTask, originalQueue string, retryCount int) error {
	mqMu.Lock()
	defer mqMu.Unlock()

	if MQChannel == nil {
		return fmt.Errorf("RabbitMQ not connected")
	}

	body, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return MQChannel.Publish(ExchangeJudgeDLX, originalQueue, false, false, amqp.Publishing{
		ContentType:   "application/json",
		DeliveryMode:  amqp.Persistent,
		Expiration:    "30000", // 30s TTL，过期后死信回 judge exchange
		Headers: amqp.Table{
			"x-retry-count":    int32(retryCount),
			"x-original-queue": originalQueue,
		},
		Body: body,
	})
}
