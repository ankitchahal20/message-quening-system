package main

import (
	"log"

	"github.com/ankit/project/message-quening-system/cmd/consumer"
	"github.com/ankit/project/message-quening-system/cmd/producer"
	globalconfig "github.com/ankit/project/message-quening-system/internal/config"
	"github.com/ankit/project/message-quening-system/internal/db"
	"github.com/ankit/project/message-quening-system/internal/server"
	"github.com/ankit/project/message-quening-system/internal/service"
	"github.com/ankit/project/message-quening-system/internal/utils"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pelletier/go-toml"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	logger.Info("Main started")

	config, err := toml.LoadFile("./config/default.toml")
	if err != nil {
		log.Fatalf("Error while loading deafault.toml file")
	}

	var appConfig globalconfig.GlobalConfig
	err = config.Unmarshal(&appConfig)
	if err != nil {
		log.Fatalf("Error while unmarshalling config")
	}

	globalconfig.SetConfig(appConfig)

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
