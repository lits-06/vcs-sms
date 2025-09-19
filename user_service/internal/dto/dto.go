package dto

type AddUserScopeRequest struct {
	UserID string   `json:"user_id" binding:"required"`
	Scopes []string `json:"scopes" binding:"required"`
}

type RemoveUserScopeRequest struct {
	UserID string   `json:"user_id" binding:"required"`
	Scopes []string `json:"scopes" binding:"required"`
}
