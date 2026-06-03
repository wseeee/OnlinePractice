package service

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*WsConn]bool // user_identity → connections
}

type WsConn struct {
	conn         *websocket.Conn
	userIdentity string
	send         chan []byte
}

var hub = &Hub{
	clients: make(map[string]map[*WsConn]bool),
}

func (h *Hub) Register(userIdentity string, wc *WsConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[userIdentity] == nil {
		h.clients[userIdentity] = make(map[*WsConn]bool)
	}
	h.clients[userIdentity][wc] = true
}

func (h *Hub) Unregister(userIdentity string, wc *WsConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[userIdentity]; ok {
		delete(conns, wc)
		if len(conns) == 0 {
			delete(h.clients, userIdentity)
		}
	}
}

func (h *Hub) Broadcast(userIdentity string, payload []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, ok := h.clients[userIdentity]
	if !ok || len(conns) == 0 {
		return false
	}
	sent := false
	for wc := range conns {
		select {
		case wc.send <- payload:
			sent = true
		default:
			log.Printf("[WS] send buffer full for user %s", userIdentity)
		}
	}
	return sent
}

// NotifyJudgeResult 推送消息给用户（Consumer / Service 调用），返回是否送达
func NotifyJudgeResult(userIdentity string, payload []byte) bool {
	return hub.Broadcast(userIdentity, payload)
}
