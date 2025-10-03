package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ENV struct {
	ENV       string `mapstructure:"APP_ENV"`
	HOST      string `mapstructure:"DB_HOST"`
	USER_NAME string `mapstructure:"DB_USER"`
	PASSWORD  string `mapstructure:"DB_PASSWORD"`
	DB_NAME   string `mapstructure:"DB_NAME"`
	URL_DIR   string `mapstructure:"URL_DIR"`
}

func LoadFileENV() *ENV {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}
	envFile := fmt.Sprintf(".env.%s", env)

	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Error loading %s file", envFile)
	}

	return &ENV{
		ENV:       os.Getenv("APP_ENV"),
		HOST:      os.Getenv("DB_HOST"),
		USER_NAME: os.Getenv("DB_USER"),
		PASSWORD:  os.Getenv("DB_PASSWORD"),
		DB_NAME:   os.Getenv("DB_NAME"),
		URL_DIR:   os.Getenv("URL_DIR"),
	}
}
