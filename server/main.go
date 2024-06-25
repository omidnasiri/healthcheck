package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"healthcheck/internal/model"
	"healthcheck/pkg/postgres"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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

func main() {
	dsn := "host=localhost user=postgres password=mysecretpassword dbname=healthcheck port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := postgres.Connect(dsn, nil)
	if err != nil {
		log.Fatal("db connection failed, err:", err.Error())
	}

	if err := postgres.Migrate(db, &model.Endpoint{}); err != nil {
		log.Fatal("db migration failed, err:", err.Error())
	}

	// ctx, _ := context.WithCancel(context.Background())
	// var wg sync.WaitGroup

	router := gin.Default()
	endpointRoutes := router.Group("/endpoints")
	{
		endpointRoutes.POST("", func(c *gin.Context) {
			req := struct {
				URL      string `json:"url" binding:"required"`
				Interval int    `json:"interval" binding:"required"`
				Retries  int    `json:"retries" binding:"required"`
			}{}
			err := c.ShouldBindJSON(&req)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err = db.Create(&Endpoint{
				URL:      req.URL,
				Interval: req.Interval,
				Retries:  req.Retries,
			}).Error
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
				IsActive bool `json:"is_active" binding:"required"`
			}{}
			err := c.ShouldBindJSON(&req)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// todo: stop or start agent
			// wg.Add(1)
			// go Agent(ctx, &wg, alpha)

			err = db.Model(&Endpoint{}).Where("id = ?", id).Update("is_active", req.IsActive).Error
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

	if err := router.Run(":8000"); err != nil {
		log.Fatal("router failed, err:", err.Error())
	}
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
