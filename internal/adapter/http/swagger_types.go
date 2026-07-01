package http

// ErrorResponse representa o formato padrão de erro retornado pela API.
type ErrorResponse struct {
	Error string `json:"error" example:"birthDate must be in format YYYY-MM-DD"`
}
