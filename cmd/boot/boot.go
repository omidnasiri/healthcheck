package boot

import (
	"fmt"
	"healthcheck/api"
	"healthcheck/config"
	"healthcheck/internal/model"
	"healthcheck/pkg/postgres"
	"log"
	"sync"
)

func Up(cfg *config.Config) (map[string]func(), error) {
	// closeFunctions is a map of functions that will be called on shutdown
	closeFunctions := make(map[string]func())

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.User, cfg.DB.Password,
		cfg.DB.DBName, cfg.DB.Port)
	db, err := postgres.Connect(dsn, nil)
	if err != nil {
		log.Println("db connection failed, err:", err.Error())
		return closeFunctions, err
	}
	closeFunctions["db"] = func() { postgres.Disconnect(db) }

	if err := db.AutoMigrate(&model.Endpoint{}); err != nil {
		log.Println("db migration failed, err:", err.Error())
		return closeFunctions, err
	}

	wg := &sync.WaitGroup{}
	container, err := Inject(db, wg, cfg)
	if err != nil {
		log.Println("dependency injection failed, err:", err.Error())
		return closeFunctions, err
	}

	router := api.SetupRoutes(container)

	if err := router.Run(":8000"); err != nil {
		log.Println("router failed, err:", err.Error())
		return closeFunctions, err
	}

	wg.Wait()
	return closeFunctions, nil
}

func Down(closeFunctions map[string]func()) {
	for key, fn := range closeFunctions {
		fn()
		log.Println(key, "closed")
	}
}
