package domain

type Repository interface {
	GetServerSnapshot(serverID string) (*ServerSnapshot, error)
}
