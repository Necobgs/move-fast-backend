package dto

type HandleRideRequestDto struct {
	RideId   string `json:"ride_id"`
	Accepted bool   `json:"accepted"`
}
