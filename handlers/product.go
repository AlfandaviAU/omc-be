package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/davi/omc-be/database"
	"github.com/davi/omc-be/models"
	"github.com/davi/omc-be/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// isValidCategory allows empty for items that have not been categorised yet.
func isValidCategory(category string) bool {
	switch category {
	case "", "Guns", "Ammo", "Attachments", "Utilities":
		return true
	}
	return false
}

func GetProducts(c *gin.Context) {
	var products []models.Product
	if err := database.DB.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	var urlsToRefresh []string
	var productsToUpdate []*models.Product

	for i := range products {
		if utils.IsDiscordURLExpired(products[i].ImageURL) {
			urlsToRefresh = append(urlsToRefresh, products[i].ImageURL)
			productsToUpdate = append(productsToUpdate, &products[i])
		}
	}

	if len(urlsToRefresh) > 0 {
		newURLs, err := utils.RefreshDiscordURLs(urlsToRefresh)
		if err == nil && len(newURLs) > 0 {
			for _, p := range productsToUpdate {
				if newURL, ok := newURLs[p.ImageURL]; ok {
					p.ImageURL = newURL
					database.DB.Save(p)
				}
			}
		}
	}

	c.JSON(http.StatusOK, products)
}

func CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.Name = SanitizeText(product.Name)
	product.Description = SanitizeText(product.Description)
	product.ImageURL = strings.TrimSpace(product.ImageURL)

	if product.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product name is required"})
		return
	}
	if product.Price < 0 || product.Stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price and stock cannot be negative"})
		return
	}
	if !IsValidImageURL(product.ImageURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image URL format"})
		return
	}
	if !isValidCategory(product.Category) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}
	Broadcast("update_products")
	c.JSON(http.StatusOK, product)
}

func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var req models.Product
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Name = SanitizeText(req.Name)
	req.Description = SanitizeText(req.Description)
	req.ImageURL = strings.TrimSpace(req.ImageURL)

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product name is required"})
		return
	}
	if req.Price < 0 || req.Stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price and stock cannot be negative"})
		return
	}
	if !IsValidImageURL(req.ImageURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image URL format"})
		return
	}
	if !isValidCategory(req.Category) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	oldName := product.Name
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	product.ImageURL = req.ImageURL
	product.Category = req.Category

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&product).Error; err != nil {
			return err
		}
		if oldName == product.Name {
			return nil
		}
		// Attachment links are stored by name, so a rename has to carry them along.
		if err := tx.Model(&models.WeaponAttachment{}).Where("LOWER(weapon_name) = LOWER(?)", oldName).Update("weapon_name", product.Name).Error; err != nil {
			return err
		}
		return tx.Model(&models.WeaponAttachment{}).Where("LOWER(attachment_name) = LOWER(?)", oldName).Update("attachment_name", product.Name).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}
	Broadcast("update_products")
	c.JSON(http.StatusOK, product)
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&product).Error; err != nil {
			return err
		}
		return tx.Where("LOWER(weapon_name) = LOWER(?) OR LOWER(attachment_name) = LOWER(?)", product.Name, product.Name).Delete(&models.WeaponAttachment{}).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}
	Broadcast("update_products")
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func GetWeaponAttachments(c *gin.Context) {
	var attachments []models.WeaponAttachment
	if err := database.DB.Find(&attachments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attachments mapping"})
		return
	}
	
	// Create a map of weaponName -> []attachmentName
	mapping := make(map[string][]string)
	for _, a := range attachments {
		mapping[a.WeaponName] = append(mapping[a.WeaponName], a.AttachmentName)
	}
	
	c.JSON(http.StatusOK, mapping)
}

// UpdateWeaponAttachments replaces the attachment list of each weapon sent.
func UpdateWeaponAttachments(c *gin.Context) {
	var input struct {
		Weapons []struct {
			WeaponID      uint   `json:"weapon_id" binding:"required"`
			AttachmentIDs []uint `json:"attachment_ids"`
		} `json:"weapons" binding:"required,dive"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, w := range input.Weapons {
			var weapon models.Product
			if err := tx.First(&weapon, w.WeaponID).Error; err != nil {
				return err
			}
			if err := tx.Where("LOWER(weapon_name) = LOWER(?)", weapon.Name).Delete(&models.WeaponAttachment{}).Error; err != nil {
				return err
			}
			if len(w.AttachmentIDs) == 0 {
				continue
			}

			var attachments []models.Product
			if err := tx.Where("id IN ?", w.AttachmentIDs).Find(&attachments).Error; err != nil {
				return err
			}
			for _, a := range attachments {
				link := models.WeaponAttachment{WeaponName: weapon.Name, AttachmentName: a.Name}
				if err := tx.Create(&link).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Weapon not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save weapon attachments"})
		return
	}

	Broadcast("update_attachments")
	c.JSON(http.StatusOK, gin.H{"message": "Weapon attachments updated"})
}
