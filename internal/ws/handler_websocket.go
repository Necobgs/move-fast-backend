package ws

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Necobgs/move-fast-backend/internal/utils"
	"github.com/Necobgs/move-fast-backend/internal/ws/message"

	"github.com/Necobgs/move-fast-backend/internal/auth"
	"github.com/Necobgs/move-fast-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type HandlerWs struct {
	upgrader      websocket.Upgrader
	hub           *Hub
	rdb           *redis.Client
	authService   *service.AuthService
	eventRegistry *EventRegistry
}

func NewWsHandler(rdb *redis.Client, authService *service.AuthService, hub *Hub, eventRegistry *EventRegistry) *HandlerWs {
	return &HandlerWs{
		rdb:           rdb,
		authService:   authService,
		hub:           hub,
		eventRegistry: eventRegistry,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // TODO adicionar validação do dominio caso for pra produção
			},
		},
	}
}

func (h *HandlerWs) HandleConnection(c *gin.Context) {
	fmt.Println("Conectado")
	tokenString := c.Query("token")
	fmt.Println("token: ", tokenString)
	token, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		fmt.Println("token inválido")
		c.AbortWithStatus(401)
		return
	}
	claims := token.Claims.(*auth.CustomClaims)

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("upgrade err:", err.Error())
		return
	}

	client := &ClientWebSocket{
		ID:     utils.BuildClientKey(claims.Id, claims.DriverId),
		Conn:   conn,
		Send:   make(chan []byte),
		Claims: claims,
	}

	h.hub.Register <- client

	go h.readPump(client)
	go h.writePump(client)

}

func (h *HandlerWs) readPump(c *ClientWebSocket) {
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
			fmt.Println("Erro ao deserializar baseMessage: ", err)
			break
		}
		fmt.Println("--base message--")
		fmt.Println("event: ", base.Event)
		fmt.Println("data: ", base.Data)

		handler, ok := h.eventRegistry.GetHandler(base.Event)
		if !ok {
			fmt.Println("handler not não encontrado")
			break
		}
		handler(c, base)
	}
}

func (h *HandlerWs) writePump(c *ClientWebSocket) {
	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}
