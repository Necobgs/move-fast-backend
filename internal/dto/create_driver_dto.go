package dto

type CreateDriverDto struct {
	CnhNumber           string `json:"cnh_number" binding:"required"`
	VehicleLicensePlate string `json:"vehicle_license_plate" binding:"required"`
	VehicleModel        string `json:"vehicle_model" binding:"required"`
	VehicleBrand        string `json:"vehicle_brand" binding:"required"`
	VehicleColor        string `json:"vehicle_color" binding:"required"`
}
