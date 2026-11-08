package database

import (
	"log"
	"os"

	"github.com/davi/omc-be/models"
	"golang.org/x/crypto/bcrypt"
)

func Seed() {
	seedUsers()
	seedProducts()

	// Dev-only seeding
	if os.Getenv("GIN_MODE") != "release" {
		seedDevData()
	}
}

func hashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}
	return string(hashed)
}

func seedUsers() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	users := []models.User{
		{Username: "admin", Password: hashPassword("admin"), Role: "admin"},
	}

	if err := DB.Create(&users).Error; err != nil {
		log.Fatal("Failed to seed users:", err)
	}
	log.Println("Seeded default admin user (username: admin, password: admin)")
}

func seedProducts() {
	var count int64
	DB.Model(&models.Product{}).Count(&count)
	if count > 0 {
		return
	}

	products := []models.Product{
		// Firearms
		{Name: "Pistol Kacang", Price: 5000, Stock: 0},
		{Name: "P50", Price: 12000, Stock: 0},
		{Name: "Ceramic", Price: 20000, Stock: 0},
		{Name: "Micro SMG", Price: 23000, Stock: 0},
		{Name: "Navi Revolver", Price: 76000, Stock: 0},
		{Name: "SMG", Price: 45000, Stock: 0},
		{Name: "Vektor KVR", Price: 78000, Stock: 0},
		{Name: "Mini SMG", Price: 23000, Stock: 0},
		{Name: "Machine Pistol", Price: 20000, Stock: 0},
		{Name: "X17", Price: 45000, Stock: 0},
		{Name: "Black Revolver", Price: 96000, Stock: 0},
		{Name: "Shotgun", Price: 70000, Stock: 0},
		{Name: "AK", Price: 260000, Stock: 0},
		{Name: "Virtus", Price: 285000, Stock: 0},
		{Name: "Carbine", Price: 320000, Stock: 0},
		{Name: "Sawn Off Shotgun", Price: 50000, Stock: 0},

		// Ammo
		{Name: "Ammo P.50", Price: 1200, Stock: 0},
		{Name: "9 MM", Price: 3000, Stock: 0},
		{Name: "44 Magnum (Navy-BR)", Price: 5700, Stock: 0},
		{Name: "45 ACP (KVR)", Price: 5700, Stock: 0},
		{Name: "Ammo 5.56 (AK)", Price: 6000, Stock: 0},
		{Name: "Ammo Virtus", Price: 6000, Stock: 0},
		{Name: "Peluru 12 (SG)", Price: 6700, Stock: 0},

		// Attachments
		{Name: "Suppressor", Price: 16000, Stock: 0},
		{Name: "Tactical Suppressor", Price: 16000, Stock: 0},
		{Name: "SMG Drum Mag", Price: 16000, Stock: 0},
		{Name: "Rifle Drum Mag", Price: 29000, Stock: 0},
		{Name: "Extended SMG Mag", Price: 9000, Stock: 0},
		{Name: "Extended Pistol Mag", Price: 7000, Stock: 0},
		{Name: "Macro Scope", Price: 7000, Stock: 0},
		{Name: "Medium Scope", Price: 7000, Stock: 0},
		{Name: "Tactical Flashlight", Price: 7000, Stock: 0},
		{Name: "Grip", Price: 7000, Stock: 0},
		{Name: "Extended Rifle Mag", Price: 23000, Stock: 0},

		// Other items
		{Name: "Vest Merah", Price: 1200, Stock: 0},
		{Name: "Vest Biru", Price: 3000, Stock: 0},
		{Name: "Weed", Price: 300, Stock: 0},
		{Name: "Meth", Price: 360, Stock: 0},
		{Name: "Opium", Price: 360, Stock: 0},
		{Name: "Cocain", Price: 1000, Stock: 0},
		{Name: "Baggy", Price: 70, Stock: 0},
		{Name: "Lockpick", Price: 3000, Stock: 0},
	}

	if err := DB.Create(&products).Error; err != nil {
		log.Fatal("Failed to seed products:", err)
	}
	log.Println("Seeded", len(products), "products")
}

// seedDevData populates stock for development
func seedDevData() {
	// Add stock to all products
	result := DB.Model(&models.Product{}).Where("stock = 0").Update("stock", 50)
	if result.RowsAffected > 0 {
		log.Println("Dev: Populated stock (50 each) for", result.RowsAffected, "products")
	}
}
