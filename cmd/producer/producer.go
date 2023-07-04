package producer

import "github.com/segmentio/kafka-go"

// the topic and broker address are initialized as constants
var (
	Topic          = "my-kafka-topic"
	Broker1Address = "localhost:9092"
	// broker2Address = "localhost:9093"
	// broker3Address = "localhost:9094"
)

var KafkaWriter *kafka.Writer

func IntializeKafkaProducerWriter() {
	// intialize the writer with the broker addresses, and the topic
	KafkaWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{Broker1Address},
		Topic:   Topic,
	})

}
