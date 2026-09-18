package handlers

import (
	"net/http"

	"github.com/davi/omc-be/database"
	"github.com/davi/omc-be/models"
	"github.com/gin-gonic/gin"
)

type OrderItemInput struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required"`
}

type OrderInput struct {
	Items         []OrderItemInput `json:"items" binding:"required,min=1"`
	ProofImage    string           `json:"proof_image" binding:"required"`
	DestinationID uint             `json:"destination_id" binding:"required"`
}

func CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var input OrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range input.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
			return
		}

		if product.Stock < item.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock for " + product.Name})
			return
		}

		// Deduct stock
		product.Stock -= item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
			return
		}

		price := product.Price * float64(item.Quantity)
		totalAmount += price

		orderItems = append(orderItems, models.OrderItem{
			ProductID: product.ID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		})
	}

	order := models.Order{
		UserID:        userID.(uint),
		DestinationID: input.DestinationID,
		TotalAmount:   totalAmount,
		ProofImage:    input.ProofImage,
		Status:        "pending",
		Items:         orderItems,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	tx.Commit()
	Broadcast("update_orders")
	Broadcast("update_products")
	c.JSON(http.StatusOK, order)
}

func GetMyOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var orders []models.Order
	if err := database.DB.Where("user_id = ?", userID).Preload("Items").Preload("Items.Product").Preload("User").Preload("Destination").Preload("Verifier").Order("created_at desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func GetAllOrders(c *gin.Context) {
	query := database.DB.Preload("Items").Preload("Items.Product").Preload("User").Preload("Destination").Preload("Verifier").Order("created_at desc")

	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var orders []models.Order
	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	verifierID, exists := c.Get("user_id")
	
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := database.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	order.Status = input.Status
	if exists {
		vid := verifierID.(uint)
		order.VerifierID = &vid
	}

	if err := database.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	Broadcast("update_orders")
	c.JSON(http.StatusOK, order)
}

func DeleteOrder(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.Order{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		return
	}
	Broadcast("update_orders")
	c.JSON(http.StatusOK, gin.H{"message": "Order deleted"})
}
