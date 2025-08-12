package entity

type AuthService interface {
	CreateCredential(userID string) (string, error)
	ValidateAccess(credentialStr string) (string, error)
	RefreshCredential(refreshStr string) (string, error)
	Revoke(credentialStr string) error
}
