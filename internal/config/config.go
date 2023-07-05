package config

type GlobalConfig struct {
	Database Database `toml:"database"`
	Server   Server   `toml:"server"`
}

var (
	globalConfig GlobalConfig
)

func SetConfig(cfg GlobalConfig) {
	globalConfig = cfg
}
func GetConfig() GlobalConfig {
	return globalConfig
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
