package ws

import (
	"github.com/Necobgs/move-fast-backend/internal/ws/message"
)

type EventHandler func(c *ClientWebSocket, data message.BaseMessage)
