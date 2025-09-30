package kafka

import "time"

const (
	minBytes               = 10e3 // 10KB
	maxBytes               = 10e6 // 10MB
	queueCapacity          = 1000
	heartbeatInterval      = 3 * time.Second
	commitInterval         = 0
	partitionWatchInterval = 5 * time.Second
	maxAttempts            = 3
	dialTimeout            = 3 * time.Minute

	writerReadTimeout  = 10 * time.Second
	writerWriteTimeout = 10 * time.Second
	writerRequiredAcks = -1
	writerMaxAttempts  = 3

	updateTopic       = "server-update"
	updateGroupID     = "update_group_healthcheck_service"
	updateWorkerCount = 3

	createTopic       = "server-create"
	createGroupID     = "create_group_healthcheck_service"
	createWorkerCount = 3

	deleteTopic       = "server-delete"
	deleteGroupID     = "delete_group_healthcheck_service"
	deleteWorkerCount = 1

	deadLetterQueueTopic = "dead-letter-queue"
)
