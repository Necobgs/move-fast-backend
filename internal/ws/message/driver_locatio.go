package message

type DriverLocation struct {
	Lat      float64 `json:"driver_location_lat"`
	Lng      float64 `json:"driver_location_lng"`
	RideId   string  `json:"ride_id"`
	DriverId string  `json:"driver_id"`
}
