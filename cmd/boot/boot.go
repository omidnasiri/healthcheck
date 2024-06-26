package boot

import (
	"healthcheck/api"
	"healthcheck/internal/model"
	"healthcheck/pkg/postgres"
	"log"
)

func Up() (map[string]func(), error) {
	// closeFunctions is a map of functions that will be called on shutdown
	closeFunctions := make(map[string]func())

	dsn := "host=localhost user=postgres password=mysecretpassword dbname=healthcheck port=5432 sslmode=disable TimeZone=Asia/Shanghai"
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

	container := Inject(db)
	router := api.SetupRoutes(container)

	if err := router.Run(":8000"); err != nil {
		log.Fatal("router failed, err:", err.Error())
	}

	return closeFunctions, nil
}

func Down(closeFunctions map[string]func()) {
	for key, fn := range closeFunctions {
		fn()
		log.Println(key, "closed")
	}
}
