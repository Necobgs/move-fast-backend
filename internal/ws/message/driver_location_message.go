package message

type DriverLocationMessage struct {
	DriverId string `json:"driver_id"`
	RideId   string `json:"ride_id"`
}
