package domain

type UseCase interface {
	CreateServerStateRecord(serverID string, status string) error
	PublishCreateServerStateRecordEvent(serverID string, status string) error
}
