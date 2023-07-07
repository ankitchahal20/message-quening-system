package kafka

import (
	config "github.com/ankit/project/message-quening-system/internal/config"
	"github.com/ankit/project/message-quening-system/internal/constants"
	"github.com/segmentio/kafka-go"
)

func IntializeKafkaConsumerReader() *kafka.Reader {
	cfg := config.GetConfig()
	KafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.Kafka.Broker1Address},
		Topic:    cfg.Kafka.Topic,
		GroupID:  constants.Group,
		MaxBytes: 1e6,
		// MaxWait:  1000 * time.Millisecond,
		//// if you set it to `kafka.LastOffset` it will only consume new messages
		StartOffset: kafka.FirstOffset,
	})
	return KafkaReader
}
