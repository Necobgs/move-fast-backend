package ws

type Hub struct {
	Clients    map[string]*ClientWebSocket
	Register   chan *ClientWebSocket
	UnRegister chan *ClientWebSocket
	Broadcast  chan []byte
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*ClientWebSocket),
		Register:   make(chan *ClientWebSocket),
		UnRegister: make(chan *ClientWebSocket),
		Broadcast:  make(chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {

		case client := <-h.Register:
			h.Clients[client.ID] = client

		case client := <-h.UnRegister:
			delete(h.Clients, client.ID)
			close(client.Send)

		case msg := <-h.Broadcast:
			for _, c := range h.Clients {
				select {
				case c.Send <- msg:
				default:
					close(c.Send)
					delete(h.Clients, c.ID)
				}
			}
		}
	}
}
