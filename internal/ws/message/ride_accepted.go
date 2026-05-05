package message

type RideAcceptedMessage struct {
	RideId            string  `json:"ride_id"`
	DriverId          string  `json:"driver_id"`
	DriverLocationLat float64 `json:"driver_location_lat"`
	DriverLocationLng float64 `json:"driver_location_lng"`
}
