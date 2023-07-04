package main

import (
	"log"

	"github.com/ankit/project/message-quening-system/cmd/producer"
	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/server"
	"github.com/ankit/project/message-quening-system/internal/service"
	"github.com/ankit/project/message-quening-system/internal/utils"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/swaggo/swag/example/basic/docs"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	logger.Info("Main started")

	docs.SwaggerInfo.BasePath = ""
	docs.SwaggerInfo.Host = "0.0.0.0:8080"
	postgres, err := db.New()
	if err != nil {
		log.Fatal("Unable to connect to DB : ", err)
	}

	producer.IntializeKafkaProducerWriter()
	defer producer.KafkaWriter.Close()
	// consumer.IntializeKafkaProducerReader()
	// defer consumer.KafkaReader.Close()
	utils.InitLogClient()
	service.NewProductService(postgres)
	server.Start()
}
