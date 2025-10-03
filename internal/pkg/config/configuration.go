package config

import (
	"log"

	"github.com/spf13/viper"
)

type Configuration struct {
	Server   ServerConfiguration
	Cors     CorsConfiguration
	Database DatabaseConfiguration
}

type ServerConfiguration struct {
	Port   string
	Secret string
	Mode   string
}

type CorsConfiguration struct {
	Global bool
	Ips    string
}

type DatabaseConfiguration struct {
	Driver   string
	Dbname   string
	Username string
	Password string
	Host     string
	Port     string
	Sslmode  bool
	Logmode  bool
}

var Config *Configuration

func Setup(configPath string) error {
	var Configuration *Configuration

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
		return err
	}

	if err := viper.Unmarshal(&Configuration); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
		return err
	}

	Config = Configuration

	return nil
}

func GetConfig() *Configuration {
	return Config
}
