package ws

type EventRegistry struct {
	Handlers map[string]EventHandler
}

func NewEventRegistry() *EventRegistry {
	return &EventRegistry{Handlers: make(map[string]EventHandler)}
}

func (er *EventRegistry) Register(event string, handler EventHandler) {
	er.Handlers[event] = handler
}

func (er *EventRegistry) GetHandler(event string) (EventHandler, bool) {
	value, ok := er.Handlers[event]
	return value, ok
}
