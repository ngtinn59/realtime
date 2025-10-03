package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ENV struct {
	ENV       string
	HOST      string
	USER_NAME string
	PASSWORD  string
	DB_NAME   string
	URL_DIR   string
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
