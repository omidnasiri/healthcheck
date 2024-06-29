package boot

import (
	"fmt"
	"healthcheck/api"
	"healthcheck/config"
	"healthcheck/internal/model"
	"healthcheck/pkg/postgres"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func Up(cfg *config.Config) (map[string]func(), *sync.WaitGroup, error) {
	// closeFunctions is a map of functions that will be called on shutdown
	closeFunctions := make(map[string]func())

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.User, cfg.DB.Password,
		cfg.DB.DBName, cfg.DB.Port)
	db, err := postgres.Connect(dsn, nil)
	if err != nil {
		log.Println("db connection failed, err:", err.Error())
		return closeFunctions, nil, err
	}
	closeFunctions["db"] = func() { postgres.Disconnect(db) }

	if err := db.AutoMigrate(&model.Endpoint{}, &model.CheckLog{}); err != nil {
		log.Println("db migration failed, err:", err.Error())
		return closeFunctions, nil, err
	}

	wg := &sync.WaitGroup{}
	container, err := Inject(db, wg, cfg, closeFunctions)
	if err != nil {
		log.Println("dependency injection failed, err:", err.Error())
		return closeFunctions, nil, err
	}

	router := api.SetupRoutes(container)

	httpServer := http.Server{
		Addr:    ":8000",
		Handler: router,
	}
	httpServerErrors := make(chan error, 1)
	go func() {
		httpServerErrors <- httpServer.ListenAndServe()
	}()
	closeFunctions["httpServer"] = func() { httpServer.Shutdown(nil) }

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-httpServerErrors:
		log.Println("http server error, err:", err.Error())
	case <-shutdown:
		log.Println("shutdown signal received")
	}

	return closeFunctions, wg, nil
}

func Down(closeFunctions map[string]func(), wg *sync.WaitGroup) {
	for key, fn := range closeFunctions {
		if key != "db" {
			fn()
			log.Println(key, "closed")
		}
	}
	wg.Wait()
	closeFunctions["db"]()
	log.Println("db", "closed")
}
