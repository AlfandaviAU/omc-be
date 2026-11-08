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

func Connect() {
	var db *gorm.DB
	var err error

	dsn := os.Getenv("POSTGRES_URL")
	if dsn != "" {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "omc.db"
		}
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	}

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Product{}, &models.Order{}, &models.OrderItem{}, &models.WeaponAttachment{}, &models.PurchaseLimit{})
	if err != nil {
		log.Fatal("Failed to auto migrate database:", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	DB = db

	Seed()
}
