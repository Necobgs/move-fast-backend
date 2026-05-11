package dto

type CreateRideDto struct {
	OriginLocationLat          float64 `json:"origin_location_lat"`
	OriginLocationLng          float64 `json:"origin_location_lng"`
	OriginLocationAddress      string  `json:"origin_location_address"`
	DestinationLocationLat     float64 `json:"destination_location_lat"`
	DestinationLocationLng     float64 `json:"destination_location_lng"`
	DestinationLocationAddress string  `json:"destination_location_address"`
}
