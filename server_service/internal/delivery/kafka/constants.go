package kafka

import "time"

const (
	minBytes               = 10e3 // 10KB
	maxBytes               = 20e6 // 20MB
	maxWait                = 500 * time.Microsecond
	readBatchTimeout       = 500 * time.Millisecond
	queueCapacity          = 10000
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
	updateGroupID     = "update_group_server_service"
	updateWorkerCount = 5

	deadLetterQueueTopic = "dead-letter-queue"

	createTopic = "server-create"
	deleteTopic = "server-delete"
)
