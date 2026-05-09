package realtime

type Hub interface {
	SendToClient(id string, msg []byte) bool
}
