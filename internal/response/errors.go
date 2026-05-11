package response

var (
	ErrInvalidBody          = ErrorResponse{Message: "Corpo da requisição inválido", StatusCode: 400}
	ErrUserAlreadyExists    = ErrorResponse{Message: "Usuário já existe", StatusCode: 409}
	ErrInternalServer       = ErrorResponse{Message: "Erro interno", StatusCode: 500}
	ErrInvalidParam         = ErrorResponse{Message: "Parâmetros da requisição inválidos", StatusCode: 400}
	ErrNotFound             = ErrorResponse{Message: "Não encontrado", StatusCode: 404}
	ErrUnauthorized         = ErrorResponse{Message: "Sem autorização", StatusCode: 401}
	ErrVehicleAlreadyExists = ErrorResponse{Message: "Veículo já existe", StatusCode: 409}
	ErrDriverAlreadyExists  = ErrorResponse{Message: "Motorista já existe", StatusCode: 409}
)
