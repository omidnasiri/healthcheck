package postgres

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string, conf *gorm.Config) (*gorm.DB, error) {
	if conf == nil {
		conf = &gorm.Config{}
	}
	return gorm.Open(postgres.Open(dsn), conf)
}

func Disconnect(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
