package realtime

import (
	"github.com/Necobgs/move-fast-backend/internal/auth"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	Claims *auth.CustomClaims
}
