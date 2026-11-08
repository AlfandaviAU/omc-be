package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/davi/omc-be/database"
	"github.com/davi/omc-be/models"
	"github.com/davi/omc-be/utils"
	"github.com/gin-gonic/gin"
)

type OrderItemInput struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type OrderInput struct {
	Items         []OrderItemInput `json:"items" binding:"required,min=1,dive"`
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

	input.ProofImage = strings.TrimSpace(input.ProofImage)
	if !IsValidDiscordImageURL(input.ProofImage) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proof image must be a valid Discord CDN URL"})
		return
	}

	tx := database.DB.Begin()

	usage, err := loadLimitUsage(tx, userID.(uint))
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check purchase limits"})
		return
	}

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

		if limit, limited := usage.Limits[product.ID]; limited {
			left := limit - usage.Purchased[product.ID]
			if item.Quantity > left {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Weekly limit reached for %s: %d left this week", product.Name, max(left, 0))})
				return
			}
			usage.Purchased[product.ID] += item.Quantity
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

	var urlsToRefresh []string
	var ordersToUpdate []*models.Order

	for i := range orders {
		if utils.IsDiscordURLExpired(orders[i].ProofImage) {
			urlsToRefresh = append(urlsToRefresh, orders[i].ProofImage)
			ordersToUpdate = append(ordersToUpdate, &orders[i])
		}
	}

	if len(urlsToRefresh) > 0 {
		newURLs, err := utils.RefreshDiscordURLs(urlsToRefresh)
		if err == nil && len(newURLs) > 0 {
			for _, o := range ordersToUpdate {
				if newURL, ok := newURLs[o.ProofImage]; ok {
					o.ProofImage = newURL
					database.DB.Save(o)
				}
			}
		}
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

	var urlsToRefresh []string
	var ordersToUpdate []*models.Order

	for i := range orders {
		if utils.IsDiscordURLExpired(orders[i].ProofImage) {
			urlsToRefresh = append(urlsToRefresh, orders[i].ProofImage)
			ordersToUpdate = append(ordersToUpdate, &orders[i])
		}
	}

	if len(urlsToRefresh) > 0 {
		newURLs, err := utils.RefreshDiscordURLs(urlsToRefresh)
		if err == nil && len(newURLs) > 0 {
			for _, o := range ordersToUpdate {
				if newURL, ok := newURLs[o.ProofImage]; ok {
					o.ProofImage = newURL
					database.DB.Save(o)
				}
			}
		}
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
	// Preload items so we can revert stock if needed
	if err := database.DB.Preload("Items").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if order.Status == "completed" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Order is already completed and locked"})
		return
	}

	if input.Status == "rejected" {
		tx := database.DB.Begin()
		// Revert the stock
		for _, item := range order.Items {
			var product models.Product
			if err := tx.First(&product, item.ProductID).Error; err == nil {
				product.Stock += item.Quantity
				tx.Save(&product)
			}
		}
		// Hard delete the order and its items
		tx.Unscoped().Where("order_id = ?", order.ID).Delete(&models.OrderItem{})
		if err := tx.Unscoped().Delete(&order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rejected order"})
			return
		}
		tx.Commit()

		Broadcast("update_orders")
		Broadcast("update_products")
		c.JSON(http.StatusOK, gin.H{"message": "Order rejected, deleted, and stock reverted"})
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

	var order models.Order
	if err := database.DB.Preload("Items").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if order.Status == "completed" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Order is already completed and locked from deletion"})
		return
	}

	tx := database.DB.Begin()
	// Revert the stock upon manual deletion as well
	for _, item := range order.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err == nil {
			product.Stock += item.Quantity
			tx.Save(&product)
		}
	}
	
	tx.Unscoped().Where("order_id = ?", order.ID).Delete(&models.OrderItem{})
	if err := tx.Unscoped().Delete(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		return
	}
	tx.Commit()

	Broadcast("update_orders")
	Broadcast("update_products")
	c.JSON(http.StatusOK, gin.H{"message": "Order deleted and stock reverted"})
}
