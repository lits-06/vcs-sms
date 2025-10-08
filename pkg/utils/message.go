package utils

import "github.com/segmentio/kafka-go"

const timestampSize = 8

func TotalSize(msg *kafka.Message) int32 {
	return 4 + 1 + 1 + sizeofBytes(msg.Key) + sizeofBytes(msg.Value) + timestampSize
}

func sizeofBytes(b []byte) int32 {
	return 4 + int32(len(b))
}
