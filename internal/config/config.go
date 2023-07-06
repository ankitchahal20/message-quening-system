package config

import (
	"log"

	"github.com/pelletier/go-toml"
)

var (
	globalConfig GlobalConfig
)

type GlobalConfig struct {
	Database Database `toml:"database"`
	Server   Server   `toml:"server"`
	Kafka    Kafka    `toml:"kafka"`
}

type Database struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	DBname   string `toml:"dbname"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type Server struct {
	Address      string `toml:"address"`
	ReadTimeOut  int    `toml:"read_time_out"`
	WriteTimeOut int    `toml:"write_time_out"`
}

type Kafka struct {
	Topic          string `toml:"topic"`
	Broker1Address string `toml:"broker_1_address"`
}

func SetConfig(cfg GlobalConfig) {
	globalConfig = cfg
}
func GetConfig() GlobalConfig {
	return globalConfig
}

func InitGlobalConfig() error {
	config, err := toml.LoadFile("./config/default.toml")
	if err != nil {
		log.Printf("Error while loading deafault.toml file : %v ", err)
		return err
	}

	var appConfig GlobalConfig
	err = config.Unmarshal(&appConfig)
	if err != nil {
		log.Printf("Error while unmarshalling config : %v", err)
		return err
	}

	SetConfig(appConfig)
	return nil
}
