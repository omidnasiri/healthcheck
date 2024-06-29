package main

import (
	"healthcheck/cmd/boot"
	"healthcheck/config"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var APP_ENV = os.Getenv("APP_ENV")

func main() {
	// load env
	if APP_ENV == "" {
		if err := godotenv.Load(); err != nil {
			log.Fatal("error loading .env file, err:", err.Error())
		}
	}

	// load config
	conf := &config.Config{}
	if err := setEnvConf(conf); err != nil {
		log.Fatal("could not map configurations", "err", err.Error())
	}

	// boot
	closeFunctions, wg, err := boot.Up(conf)
	if err != nil {
		log.Println("could not boot", "err", err.Error())
	}

	// shutdown
	boot.Down(closeFunctions, wg)
	log.Println("shutdown completed")
}

func setEnvConf(cfg *config.Config) error {
	cfg.DB.Host = os.Getenv("POSTGRES_HOST")
	cfg.DB.Port = os.Getenv("POSTGRES_PORT")
	cfg.DB.User = os.Getenv("POSTGRES_USER")
	cfg.DB.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.DB.DBName = os.Getenv("POSTGRES_DB")

	cfg.WebhookURL = os.Getenv("WEBHOOK_URL")

	return nil
}
