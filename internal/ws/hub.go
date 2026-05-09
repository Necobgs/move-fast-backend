package ws

import (
	"sync"

	"github.com/Necobgs/move-fast-backend/internal/realtime"
)

type Hub struct {
	mu sync.RWMutex

	Clients map[string]*realtime.Client

	Register   chan *realtime.Client
	UnRegister chan *realtime.Client
	Broadcast  chan []byte
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*realtime.Client),
		Register:   make(chan *realtime.Client),
		UnRegister: make(chan *realtime.Client),
		Broadcast:  make(chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {

		case client := <-h.Register:

			existing, ok := h.Clients[client.ID]

			// Substitui conexão antiga
			if ok {

				delete(h.Clients, client.ID)

				close(existing.Send)
			}

			h.Clients[client.ID] = client

		case client := <-h.UnRegister:

			existing, ok := h.Clients[client.ID]

			if !ok {
				continue
			}

			// Evita remover conexão nova
			if existing != client {
				continue
			}

			delete(h.Clients, client.ID)

			close(client.Send)

		case msg := <-h.Broadcast:

			for id, client := range h.Clients {

				select {

				case client.Send <- msg:

				default:

					delete(h.Clients, id)

					close(client.Send)
				}
			}
		}
	}
}

func (h *Hub) SendToClient(
	clientID string,
	msg []byte,
) bool {

	client, ok := h.Clients[clientID]

	if !ok {
		return false
	}

	select {

	case client.Send <- msg:
		return true

	default:

		delete(h.Clients, clientID)

		close(client.Send)

		return false
	}
}
