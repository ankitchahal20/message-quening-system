package consumer

import (
	"github.com/ankit/project/message-quening-system/cmd/producer"
	"github.com/segmentio/kafka-go"
)

var KafkaReader *kafka.Reader

func IntializeKafkaConsumerReader() {
	KafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{producer.Broker1Address},
		Topic:    producer.Topic,
		GroupID:  "my-group",
		MaxBytes: 1e6,
		//// if you set it to `kafka.LastOffset` it will only consume new messages
		StartOffset: kafka.LastOffset,
	})
}
