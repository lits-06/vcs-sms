package kafka

import "time"

const (
	minBytes               = 10e3                   // 10KB
	maxBytes               = 20e6                   // 20MB
	maxWait                = 500 * time.Millisecond // Reduced for faster batch collection
	readBatchTimeout       = 500 * time.Millisecond
	queueCapacity          = 10000 // Increased to handle large batches
	heartbeatInterval      = 3 * time.Second
	commitInterval         = 0
	partitionWatchInterval = 5 * time.Second
	maxAttempts            = 3
	dialTimeout            = 3 * time.Minute

	writerReadTimeout  = 5 * time.Second
	writerWriteTimeout = 5 * time.Second
	writerRequiredAcks = -1
	writerMaxAttempts  = 3

	updateTopic       = "server-update"
	updateGroupID     = "update_group_healthcheck_service"
	updateWorkerCount = 5

	createTopic       = "server-create"
	createGroupID     = "create_group_healthcheck_service"
	createWorkerCount = 5

	deleteTopic       = "server-delete"
	deleteGroupID     = "delete_group_healthcheck_service"
	deleteWorkerCount = 1

	deadLetterQueueTopic = "dead-letter-queue"
)
