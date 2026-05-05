package ws

import (
	"github.com/Necobgs/move-fast-backend/internal/auth"
	"github.com/gorilla/websocket"
)

type ClientWebSocket struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	Claims *auth.CustomClaims
}
