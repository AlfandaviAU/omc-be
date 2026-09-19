package handlers

import (
	"net/http"
	"strings"

	"github.com/davi/omc-be/database"
	"github.com/davi/omc-be/models"
	"github.com/gin-gonic/gin"
)

func GetProducts(c *gin.Context) {
	var products []models.Product
	if err := database.DB.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
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

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	product.ImageURL = req.ImageURL

	if err := database.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}
	Broadcast("update_products")
	c.JSON(http.StatusOK, product)
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.Product{}, id).Error; err != nil {
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
