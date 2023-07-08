package config

import (
	"log"

	"github.com/pelletier/go-toml"
)

var (
	globalConfig GlobalConfig
)

// Global Configuration
type GlobalConfig struct {
	Database Database `toml:"database"`
	Server   Server   `toml:"server"`
	Kafka    Kafka    `toml:"kafka"`
}

// DB configuration
type Database struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	DBname   string `toml:"dbname"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

// server configuration
type Server struct {
	Address      string `toml:"address"`
	ReadTimeOut  int    `toml:"read_time_out"`
	WriteTimeOut int    `toml:"write_time_out"`
}

// kakfa configurations
type Kafka struct {
	Topic          string `toml:"topic"`
	Broker1Address string `toml:"broker_1_address"`
}

// Setter method for GlobalConfig
func SetConfig(cfg GlobalConfig) {
	globalConfig = cfg
}

// Getter method for GlobalConfig
func GetConfig() GlobalConfig {
	return globalConfig
}

// Loading the values from default.toml and assigning them as part of GlobalConfig struct
func InitGlobalConfig() error {
	config, err := toml.LoadFile("./../config/default.toml")
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
