package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境放宽
	},
}

// WsHandler WebSocket 升级 + JWT 鉴权
// @Tags 公共方法
// @Summary WebSocket 连接
// @Param token query string true "JWT token"
// @Success 101 {string} Switching Protocols
// @Router /ws [get]
func WsHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 token"})
		return
	}

	userClaim, err := helper.AnalyseToken(token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "token 无效"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] upgrade error: %v", err)
		return
	}

	wc := service.NewWsConn(conn, userClaim.Identity)
	service.HubRegister(userClaim.Identity, wc)

	go wc.WritePump()
	go wc.ReadPump()
}
