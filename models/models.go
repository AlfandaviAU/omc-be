package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
	Role     string `gorm:"not null;default:'member'" json:"role"` // admin, sgt, member
	Rank     string `gorm:"not null;default:'prospect'" json:"rank"` // prospect, member, petinggi; sets weekly purchase limits
	PhotoURL string `json:"photo_url"`
}

type Product struct {
	gorm.Model
	Name        string  `gorm:"not null" json:"name"`
	Description string  `json:"description"`
	Price       float64 `gorm:"not null" json:"price"`
	Stock       int     `gorm:"not null;default:0" json:"stock"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"` // Empty means the store infers it from the name
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

// PurchaseLimit caps how many of a product a rank can buy per week.
// A product with no row for a rank has no limit.
type PurchaseLimit struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	ProductID   uint   `gorm:"not null;uniqueIndex:idx_purchase_limit_product_rank" json:"product_id"`
	Rank        string `gorm:"not null;uniqueIndex:idx_purchase_limit_product_rank" json:"rank"`
	MaxQuantity int    `gorm:"not null" json:"max_quantity"`
}

type WeaponAttachment struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	WeaponName     string `gorm:"not null" json:"weapon_name"`
	AttachmentName string `gorm:"not null" json:"attachment_name"`
}
