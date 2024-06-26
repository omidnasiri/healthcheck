package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"healthcheck/cmd/boot"
	"healthcheck/internal/model"
	"log"
	"net/http"
	"sync"
	"time"
)

const WebhookURL = "http://localhost:9000/webhook"

func main() {
	// boot
	closeFunctions, err := boot.Up()
	if err != nil {
		log.Println("could not boot", "err", err.Error())
	}

	// shutdown
	boot.Down(closeFunctions)
	log.Println("shutdown completed")
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

func Agent(ctx context.Context, wg *sync.WaitGroup, endpoint *model.Endpoint) {
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
