package message

import "encoding/json"

type BaseMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}
