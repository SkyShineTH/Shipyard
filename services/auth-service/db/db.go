package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/SkyShineTH/shipyard/auth-service/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	var database *gorm.DB
	var err error
	const maxAttempts = 30
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("database connect attempt %d/%d: %v", attempt, maxAttempts, err)
		if attempt < maxAttempts {
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if err := database.AutoMigrate(&model.User{}); err != nil {
		return nil, fmt.Errorf("run automigrate: %w", err)
	}

	return database, nil
}
