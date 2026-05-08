package message

type RequestedRideMessage struct {
	RideId   string `json:"ride_id"`
	Accepted bool   `json:"accepted"`
}
