package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Endpoint struct {
	gorm.Model
	URL        string
	Interval   int // in seconds
	Retries    int // retries before submitting failure
	LastStatus bool
}

const WebhookURL = "http://localhost:9000/webhook"

var (
	alpha = &Endpoint{
		URL:      "http://localhost:8080/alpha",
		Interval: 5,
		Retries:  3,
	}

	beta = &Endpoint{
		URL:      "http://localhost:8081/beta",
		Interval: 5,
		Retries:  3,
	}
)

func main() {
	_, err := PostgresConn()
	if err != nil {
		log.Fatal("db connection failed, err:", err.Error())
	}

	// _ = RegisterEndpoint(db, alpha)
	// _ = RegisterEndpoint(db, beta)

	ctx, _ := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go Agent(ctx, &wg, alpha)

	// wg.Add(1)
	// go Agent(ctx, &wg, beta)

	time.Sleep(5 * time.Minute)
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

func HealthCheck(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("unhealthy")
	}

	return nil
}

func Agent(ctx context.Context, wg *sync.WaitGroup, endpoint *Endpoint) {
	defer wg.Done()
	tries := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(endpoint.Interval) * time.Second):
			err := HealthCheck(endpoint.URL)
			if err != nil {
				tries++
				log.Println(endpoint.URL, "healthcheck failed, try ", tries, ", err:", err.Error())
				if tries >= endpoint.Retries {
					tries = 0
					log.Println(endpoint.URL, "endpoint is unhealthy")
					if endpoint.LastStatus {
						endpoint.LastStatus = false
						Webhook(endpoint.ID, false)
					}
				}
				continue
			}
			tries = 0
			if !endpoint.LastStatus {
				endpoint.LastStatus = true
				Webhook(endpoint.ID, true)
			}
			log.Println(endpoint.URL, "endpoint is healthy")
		}
	}
}

func Webhook(endpointID uint, status bool) error {
	payload := struct {
		Status bool `json:"status"`
	}{
		Status: status,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s/%v", WebhookURL, endpointID), "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("webhook failed")
	}

	return nil
}
