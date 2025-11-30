package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func New(databaseURL string) (*DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Println("Database connected successfully")

	return &DB{db}, nil
}

func (db *DB) RunMigrations(models ...interface{}) error {
	if err := db.AutoMigrate(models...); err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}
