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

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Endpoint struct {
	gorm.Model
	URL        string
	Interval   int // in seconds
	Retries    int // retries before submitting failure
	LastStatus bool
	IsActive   bool
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
	db, err := PostgresConn()
	if err != nil {
		log.Fatal("db connection failed, err:", err.Error())
	}

	// ctx, _ := context.WithCancel(context.Background())
	// var wg sync.WaitGroup

	router := gin.Default()
	endpointRoutes := router.Group("/endpoints")
	{
		endpointRoutes.POST("", func(c *gin.Context) {
			var endpoint Endpoint
			err := c.BindJSON(&endpoint)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err = db.Create(endpoint).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "endpoint registered"})
		})

		endpointRoutes.GET("", func(c *gin.Context) {
			var endpoints []*Endpoint
			err := db.Find(&endpoints).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, endpoints)
		})

		endpointRoutes.PATCH("/:id", func(c *gin.Context) {
			id := c.Param("id")
			req := struct {
				IsActive bool `json:"is_active"`
			}{}
			err := c.BindJSON(&req)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// todo: stop or start agent
			// wg.Add(1)
			// go Agent(ctx, &wg, alpha)

			err = db.Model(&Endpoint{}).Where("id = ?", id).Update("active", req.IsActive).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "endpoint updated"})
		})

		endpointRoutes.DELETE("/:id", func(c *gin.Context) {
			id := c.Param("id")
			err := db.Delete(&Endpoint{}, id).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "endpoint deleted"})
		})
	}

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
