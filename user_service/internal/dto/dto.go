package dto

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

type AddUserScopeRequest struct {
	Email  string   `json:"email" binding:"required" example:"user@example.com"`
	Scopes []string `json:"scopes" binding:"required" example:"server:view,server:create"`
}

type RemoveUserScopeRequest struct {
	Email  string   `json:"email" binding:"required" example:"user@example.com"`
	Scopes []string `json:"scopes" binding:"required" example:"server:view"`
}

// Response structures for Swagger documentation
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Error message"`
}
