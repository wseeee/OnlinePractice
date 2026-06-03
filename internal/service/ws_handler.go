package service

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// NewWsConn 创建 WsConn 实例
func NewWsConn(conn *websocket.Conn, userIdentity string) *WsConn {
	return &WsConn{
		conn:         conn,
		userIdentity: userIdentity,
		send:         make(chan []byte, 256),
	}
}

// HubRegister 注册连接到 hub
func HubRegister(userIdentity string, wc *WsConn) {
	hub.Register(userIdentity, wc)
}

// ReadPump 读取客户端消息 + 拉取离线通知
func (wc *WsConn) ReadPump() {
	defer func() {
		hub.Unregister(wc.userIdentity, wc)
		wc.conn.Close()
	}()

	// 拉取 Redis 中缓存的离线通知
	go DrainRedisNotifications(wc.userIdentity, wc.send)

	wc.conn.SetReadLimit(maxMessageSize)
	wc.conn.SetReadDeadline(time.Now().Add(pongWait))
	wc.conn.SetPongHandler(func(string) error {
		wc.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, _, err := wc.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// WritePump 向客户端发送消息
func (wc *WsConn) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		wc.conn.Close()
	}()
	for {
		select {
		case message, ok := <-wc.send:
			wc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				wc.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := wc.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			wc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wc.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[WS] ping error: %v", err)
				return
			}
		}
	}
}
