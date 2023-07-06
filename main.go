package main

import (
	"log"

	"github.com/ankit/project/message-quening-system/cmd/consumer"
	"github.com/ankit/project/message-quening-system/cmd/producer"
	"github.com/ankit/project/message-quening-system/internal/config"
	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/server"
	"github.com/ankit/project/message-quening-system/internal/service"
	"github.com/ankit/project/message-quening-system/internal/utils"
	_ "github.com/jackc/pgx/v5/stdlib"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	logger.Info("Main started")

	err := config.InitGlobalConfig()
	if err != nil {
		log.Fatalf("Unable to initialize global config")
	}

	postgres, err := db.New()
	if err != nil {
		log.Fatal("Unable to connect to DB : ", err)
	}

	producer.IntializeKafkaProducerWriter()
	defer producer.KafkaWriter.Close()
	consumer.IntializeKafkaConsumerReader()
	defer consumer.KafkaReader.Close()

	utils.InitLogClient()
	_ = service.NewProductService(postgres)

	server.Start()
}
