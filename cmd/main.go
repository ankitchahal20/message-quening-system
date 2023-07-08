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
	utils.InitLogClient()
	err := config.InitGlobalConfig()
	if err != nil {
		log.Fatalf("Unable to initialize global config")
	}

	postgres, err := db.New()
	if err != nil {
		log.Fatal("Unable to connect to DB : ", err)
	}

	kafkaWriter := kafka.IntializeKafkaProducerWriter()
	defer kafkaWriter.Close()
	kafkaReader := kafka.IntializeKafkaConsumerReader()
	defer kafkaReader.Close()

	_ = service.NewProductService(postgres, kafkaWriter, kafkaReader)

	server.Start()
}
