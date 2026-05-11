package ws

import (
	"encoding/json"
	"net/http"

	"github.com/Necobgs/move-fast-backend/internal/handler"
	"github.com/Necobgs/move-fast-backend/internal/realtime"
	"github.com/Necobgs/move-fast-backend/internal/utils"
	"github.com/Necobgs/move-fast-backend/internal/ws/message"

	"github.com/Necobgs/move-fast-backend/internal/auth"
	"github.com/Necobgs/move-fast-backend/internal/service"
	"github.com/Necobgs/move-fast-backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type HandlerWs struct {
	upgrader      websocket.Upgrader
	eventHandler  *handler.EventHandler
	hub           *Hub
	rdb           *redis.Client
	authService   *service.AuthService
	eventRegistry *EventRegistry
}

func NewWsHandler(rdb *redis.Client, authService *service.AuthService, hub *Hub, eventRegistry *EventRegistry, eventHandler *handler.EventHandler) *HandlerWs {
	return &HandlerWs{
		rdb:           rdb,
		authService:   authService,
		hub:           hub,
		eventRegistry: eventRegistry,
		eventHandler:  eventHandler,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // TODO adicionar validação do dominio caso for pra produção
			},
		},
	}
}

func (h *HandlerWs) HandleConnection(c *gin.Context) {
	tokenString := c.Query("token")
	token, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		logger.Log.Error("token inválido")
		c.AbortWithStatus(401)
		return
	}
	claims := token.Claims.(*auth.CustomClaims)

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("upgrade err", "error", err)
		return
	}

	client := &realtime.Client{
		ID:     utils.BuildClientKey(claims.Id, claims.DriverId),
		Conn:   conn,
		Send:   make(chan []byte),
		Claims: claims,
	}

	h.hub.Register <- client

	h.eventHandler.SyncPassengerConnection(client)

	go h.readPump(client)
	go h.writePump(client)

}

func (h *HandlerWs) readPump(c *realtime.Client) {
	defer func() {
		h.hub.UnRegister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var base message.BaseMessage
		err = json.Unmarshal(msg, &base)
		if err != nil {
			logger.Log.Error("Erro ao deserializar baseMessage", "error", err)
			break
		}

		handler, ok := h.eventRegistry.GetHandler(base.Event)
		if !ok {
			logger.Log.Warn("handler não encontrado", "event", base.Event)
			break
		}
		handler(c, base)
	}
}

func (h *HandlerWs) writePump(
	c *realtime.Client,
) {

	defer func() {
		h.hub.UnRegister <- c
		c.Conn.Close()
	}()

	for msg := range c.Send {

		err := c.Conn.WriteMessage(
			websocket.TextMessage,
			msg,
		)

		if err != nil {
			break
		}
	}
}
