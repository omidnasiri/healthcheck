package main

import (
	"errors"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Endpoint struct {
	gorm.Model
	URL      string
	Interval int // in seconds
	Retries  int // retries before submitting failure
}

var (
	alpha = &Endpoint{
		URL:      "http://localhost:8080",
		Interval: 30,
		Retries:  3,
	}

	beta = &Endpoint{
		URL:      "http://localhost:8081",
		Interval: 30,
		Retries:  3,
	}
)

func main() {
	db, err := PostgresConn()
	if err != nil {
		log.Fatal("db connection failed, err:", err.Error())
	}

	_ = RegisterEndpoint(db, alpha)
	_ = RegisterEndpoint(db, beta)
}

func PostgresConn() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=mysecretpassword dbname=healthcheck port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Endpoint{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func RegisterEndpoint(db *gorm.DB, endpoint *Endpoint) error {
	if endpoint == nil {
		return errors.New("empty endpoint")
	}

	err := db.Create(endpoint).Error
	if err != nil {
		log.Println("endpoint registration failed, err", err.Error())
		return err
	}

	return nil
}
