package main

import (
	"log"

	"github.com/ankit/project/message-quening-system/internal/config"
	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/kafka"
	"github.com/ankit/project/message-quening-system/internal/server"
	"github.com/ankit/project/message-quening-system/internal/service"
	"github.com/ankit/project/message-quening-system/internal/utils"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {

	// Initializing the Log client
	utils.InitLogClient()

	// Initializing the GlobalConfig
	err := config.InitGlobalConfig()
	if err != nil {
		log.Fatalf("Unable to initialize global config")
	}

	// Establishing the connection to DB.
	postgres, err := db.New()
	if err != nil {
		log.Fatal("Unable to connect to DB : ", err)
	}

	// Initializing the kakfa producer and consumer
	kafkaWriter := kafka.IntializeKafkaProducerWriter()
	defer kafkaWriter.Close()
	kafkaReader := kafka.IntializeKafkaConsumerReader()
	defer kafkaReader.Close()

	// Initializing the client for product service
	_ = service.NewProductService(postgres, kafkaWriter, kafkaReader)

	// Starting the server
	server.Start()
}
