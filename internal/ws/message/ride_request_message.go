package message

type RideRequestMessage struct {
	RideId   string `json:"ride_id"`
	Accepted bool   `json:"accepted"`
}
