package service

import (
	"OnlinePrictice/internal/config"
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"context"
	"encoding/json"
	"log"
)

// SendNotification 创建通知并推送（WebSocket + Redis 兜底）
func SendNotification(userIdentity, notiType, title, content, relatedIdentity string) {
	n := &model.Notification{
		Identity:        helper.GetUUID(),
		UserIdentity:    userIdentity,
		Type:            notiType,
		Title:           title,
		Content:         content,
		RelatedIdentity: relatedIdentity,
	}
	if err := model.CreateNotification(n); err != nil {
		log.Printf("[Notification] create error: %v", err)
		return
	}

	// 构建推送消息
	payload, _ := json.Marshal(map[string]interface{}{
		"type":              "notification",
		"identity":          n.Identity,
		"notification_type": notiType,
		"title":             title,
		"content":           content,
		"related_identity":  relatedIdentity,
	})

	// 尝试 WebSocket 推送
	sent := NotifyJudgeResult(userIdentity, payload)

	// WebSocket 未送达，发布 Redis Pub/Sub 供下次连接消费
	if !sent {
		if err := model.RDB.Publish(config.CTX, "notification:"+userIdentity, string(payload)).Err(); err != nil {
			log.Printf("[Notification] redis publish error: %v", err)
		}
	}
}

// DrainRedisNotifications 拉取 Redis 中缓存的离线通知并推送到 WebSocket
func DrainRedisNotifications(userIdentity string, sendCh chan<- []byte) {
	pubsub := model.RDB.Subscribe(context.Background(), "notification:"+userIdentity)
	defer pubsub.Close()

	// 非阻塞读取积压消息
	ctx := context.Background()
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			break
		}
		select {
		case sendCh <- []byte(msg.Payload):
		default:
			break
		}
	}
}
