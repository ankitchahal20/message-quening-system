package producer

import (
	"github.com/ankit/project/message-quening-system/internal/config"
	"github.com/segmentio/kafka-go"
)

var KafkaWriter *kafka.Writer

func IntializeKafkaProducerWriter() {
	cfg := config.GetConfig()

	// intialize the writer with the broker addresses, and the topic
	KafkaWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.Kafka.Broker1Address},
		Topic:   cfg.Kafka.Topic,
	})
}
