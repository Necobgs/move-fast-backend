package ws

import (
	"github.com/Necobgs/move-fast-backend/internal/realtime"
	"github.com/Necobgs/move-fast-backend/internal/ws/message"
)

type EventHandler func(c *realtime.Client, data message.BaseMessage)
