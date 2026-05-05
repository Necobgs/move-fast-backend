package response

var (
	ErrInvalidBody          = ErrorResponse{Message: "invalid request body", StatusCode: 400}
	ErrUserAlreadyExists    = ErrorResponse{Message: "user already exists", StatusCode: 409}
	ErrInternalServer       = ErrorResponse{Message: "internal server error", StatusCode: 500}
	ErrInvalidParam         = ErrorResponse{Message: "invalid request param", StatusCode: 400}
	ErrNotFound             = ErrorResponse{Message: "not found", StatusCode: 404}
	ErrUnauthorized         = ErrorResponse{Message: "unauthorized", StatusCode: 401}
	ErrVehicleAlreadyExists = ErrorResponse{Message: "vehicle already exists", StatusCode: 409}
	ErrDriverAlreadyExists  = ErrorResponse{Message: "driver already exists", StatusCode: 409}
)
