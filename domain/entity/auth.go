package entity

type TokenProvider interface {
	GenerateToken(data *TokenData) (string, error)
	ValidateToken(token string) (*TokenData, error)
	RevokeToken(token string) error
}

type TokenRepository interface {
}

type AuthorizationProvider interface {
	HasPermission(context *AuthContext, resource, action string) bool
	GetUserPermissions(userID string) ([]Permission, error)
	ValidateScope(context *AuthContext, requiredScope string) bool
}

type TokenData struct {
	UserID string   `json:"user_id"`
	Role   string   `json:"role"`
	Scopes []string `json:"scopes"`
}

type AuthContext struct {
	UserID      string                 `json:"user_id"`
	Email       string                 `json:"email"`
	Scopes      []string               `json:"scopes"`
	Permissions []Permission           `json:"permissions"`
	TokenType   string                 `json:"token_type"`
	Claims      map[string]interface{} `json:"claims"`
}

type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Scope    string `json:"scope"`
}
