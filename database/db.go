package database

import (
	"log"
	"os"

	"github.com/davi/omc-be/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

var InitErr error

func Connect() {
	var db *gorm.DB
	var err error

	dsn := os.Getenv("POSTGRES_URL_NON_POOLING")
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_URL") // Fallback
	}

	if dsn != "" {
		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: true, // disables implicit prepared statement usage
		}), &gorm.Config{})
	} else {
		db, err = gorm.Open(sqlite.Open("omc.db"), &gorm.Config{})
	}

	if err != nil {
		log.Println("Failed to connect to database:", err)
		InitErr = err
		return
	}

	err = db.AutoMigrate(&models.User{}, &models.Product{}, &models.Order{}, &models.OrderItem{}, &models.WeaponAttachment{})
	if err != nil {
		log.Println("Failed to auto migrate database:", err)
		InitErr = err
		return
	}

	DB = db

	Seed()
}
