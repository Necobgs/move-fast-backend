package ws

import (
	"sync"

	"github.com/Necobgs/move-fast-backend/internal/realtime"
	"github.com/Necobgs/move-fast-backend/pkg/logger"
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

			var id string
			if client.Claims.DriverId != "" {
				id = client.Claims.DriverId
			} else {
				id = client.Claims.Id
			}
			existing, ok := h.Clients[id]

			// Substitui conexão antiga
			if ok {

				delete(h.Clients, id)

				close(existing.Send)
			}

			h.Clients[id] = client

		case client := <-h.UnRegister:

			var id string
			if client.Claims.DriverId != "" {
				id = client.Claims.DriverId
			} else {
				id = client.Claims.Id
			}

			existing, ok := h.Clients[id]

			if !ok {
				continue
			}

			// Evita remover conexão nova
			if existing != client {
				continue
			}

			delete(h.Clients, id)

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
		logger.Log.Error("Cliente não encontrado", "client_id", clientID)

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
