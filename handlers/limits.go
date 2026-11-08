package handlers

import (
	"net/http"
	"time"

	"github.com/davi/omc-be/database"
	"github.com/davi/omc-be/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var jakartaTime = time.FixedZone("WIB", 7*60*60)

func isValidRank(rank string) bool {
	return rank == "prospect" || rank == "member" || rank == "petinggi"
}

// weekStart returns the most recent Monday 00:00 Jakarta time.
func weekStart(now time.Time) time.Time {
	t := now.In(jakartaTime)
	daysSinceMonday := (int(t.Weekday()) + 6) % 7
	return time.Date(t.Year(), t.Month(), t.Day()-daysSinceMonday, 0, 0, 0, 0, jakartaTime)
}

type limitUsage struct {
	Rank      string
	Exempt    bool
	Limits    map[uint]int // product ID -> weekly max; empty when exempt
	Purchased map[uint]int // product ID -> quantity already ordered this week
}

func loadLimitUsage(db *gorm.DB, userID uint) (*limitUsage, error) {
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	usage := &limitUsage{Rank: user.Rank, Limits: map[uint]int{}, Purchased: map[uint]int{}}
	if user.Role == "admin" || user.Role == "sgt" {
		usage.Exempt = true
		return usage, nil
	}

	var limits []models.PurchaseLimit
	if err := db.Where("rank = ?", user.Rank).Find(&limits).Error; err != nil {
		return nil, err
	}
	for _, l := range limits {
		usage.Limits[l.ProductID] = l.MaxQuantity
	}

	// SQLite compares these timestamps as text, so query a day wider and apply the exact cutoff in Go.
	start := weekStart(time.Now())
	var orders []models.Order
	err := db.Preload("Items").
		Where("user_id = ? AND status <> ? AND created_at >= ?", userID, "rejected", start.AddDate(0, 0, -1)).
		Find(&orders).Error
	if err != nil {
		return nil, err
	}
	for _, o := range orders {
		if o.CreatedAt.Before(start) {
			continue
		}
		for _, item := range o.Items {
			usage.Purchased[item.ProductID] += item.Quantity
		}
	}
	return usage, nil
}

func GetPurchaseLimits(c *gin.Context) {
	limits := []models.PurchaseLimit{}
	if err := database.DB.Find(&limits).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase limits"})
		return
	}
	c.JSON(http.StatusOK, limits)
}

// UpdatePurchaseLimits applies changed cells; a null max_quantity removes the limit.
func UpdatePurchaseLimits(c *gin.Context) {
	var input struct {
		Limits []struct {
			ProductID   uint   `json:"product_id" binding:"required"`
			Rank        string `json:"rank" binding:"required"`
			MaxQuantity *int   `json:"max_quantity"`
		} `json:"limits" binding:"required,dive"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, l := range input.Limits {
		if !isValidRank(l.Rank) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
			return
		}
		if l.MaxQuantity != nil && *l.MaxQuantity < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Limit cannot be negative"})
			return
		}
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, l := range input.Limits {
			if err := tx.Where("product_id = ? AND rank = ?", l.ProductID, l.Rank).Delete(&models.PurchaseLimit{}).Error; err != nil {
				return err
			}
			if l.MaxQuantity == nil {
				continue
			}
			row := models.PurchaseLimit{ProductID: l.ProductID, Rank: l.Rank, MaxQuantity: *l.MaxQuantity}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save purchase limits"})
		return
	}

	Broadcast("update_limits")
	c.JSON(http.StatusOK, gin.H{"message": "Purchase limits updated"})
}

// GetMyPurchaseLimits tells the store what the logged-in user can still add this week.
func GetMyPurchaseLimits(c *gin.Context) {
	userID, _ := c.Get("user_id")
	usage, err := loadLimitUsage(database.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase limits"})
		return
	}

	type productLimit struct {
		ProductID   uint `json:"product_id"`
		MaxQuantity int  `json:"max_quantity"`
		Purchased   int  `json:"purchased"`
	}
	limits := []productLimit{}
	for productID, maxQuantity := range usage.Limits {
		limits = append(limits, productLimit{productID, maxQuantity, usage.Purchased[productID]})
	}
	c.JSON(http.StatusOK, gin.H{"rank": usage.Rank, "exempt": usage.Exempt, "limits": limits})
}

func UpdateUserRank(c *gin.Context) {
	var input struct {
		Rank string `json:"rank" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !isValidRank(input.Rank) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if user.Role == "admin" || user.Role == "sgt" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Officers are exempt from purchase limits"})
		return
	}

	if err := database.DB.Model(&user).Update("rank", input.Rank).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	Broadcast("update_limits")
	c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
}
