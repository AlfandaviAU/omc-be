package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
	Role     string `gorm:"not null;default:'member'" json:"role"` // admin, sgt, member
}

type Product struct {
	gorm.Model
	Name        string  `gorm:"not null" json:"name"`
	Description string  `json:"description"`
	Price       float64 `gorm:"not null" json:"price"`
	Stock       int     `gorm:"not null;default:0" json:"stock"`
	ImageURL    string  `json:"image_url"`
}

type Order struct {
	gorm.Model
	UserID        uint        `gorm:"not null" json:"user_id"`
	User          User        `gorm:"foreignKey:UserID" json:"user"`
	DestinationID uint        `gorm:"not null" json:"destination_id"`
	Destination   User        `gorm:"foreignKey:DestinationID" json:"destination"`
	VerifierID    *uint       `json:"verifier_id"`
	Verifier      *User       `gorm:"foreignKey:VerifierID" json:"verifier"`
	TotalAmount   float64     `gorm:"not null" json:"total_amount"`
	ProofImage  string      `json:"proof_image"` // Base64 or URL of the transfer receipt
	Status      string      `gorm:"not null;default:'pending'" json:"status"` // pending, approved, rejected, completed
	Items       []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

type OrderItem struct {
	gorm.Model
	OrderID   uint    `gorm:"not null" json:"order_id"`
	ProductID uint    `gorm:"not null" json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	Price     float64 `gorm:"not null" json:"price"`
}

type WeaponAttachment struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	WeaponName     string `gorm:"not null" json:"weapon_name"`
	AttachmentName string `gorm:"not null" json:"attachment_name"`
}
