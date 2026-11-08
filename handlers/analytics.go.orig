package handlers

import (
	"net/http"

	"github.com/davi/omc-be/database"
	"github.com/davi/omc-be/models"
	"github.com/gin-gonic/gin"
)

type TopProduct struct {
	ProductName  string  `json:"product_name"`
	TotalQty     int     `json:"total_qty"`
	TotalRevenue float64 `json:"total_revenue"`
}

type RecentOrder struct {
	ID          uint    `json:"id"`
	Username    string  `json:"username"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

type RevenueByVerifier struct {
	VerifierName string  `json:"verifier_name"`
	OrderCount   int64   `json:"order_count"`
	TotalRevenue float64 `json:"total_revenue"`
}

type AnalyticsResponse struct {
	TotalRevenue      float64             `json:"total_revenue"`
	TotalOrders       int64               `json:"total_orders"`
	PendingOrders     int64               `json:"pending_orders"`
	ApprovedOrders    int64               `json:"approved_orders"`
	CompletedOrders   int64               `json:"completed_orders"`
	RejectedOrders    int64               `json:"rejected_orders"`
	TotalMembers      int64               `json:"total_members"`
	TopProducts       []TopProduct        `json:"top_products"`
	RecentOrders      []RecentOrder       `json:"recent_orders"`
	RevenueByVerifier []RevenueByVerifier `json:"revenue_by_verifier"`
}

func GetAnalytics(c *gin.Context) {
	db := database.DB
	var resp AnalyticsResponse

	// Total revenue from approved + completed orders
	db.Model(&models.Order{}).
		Where("status IN ?", []string{"approved", "completed"}).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&resp.TotalRevenue)

	// Order counts
	db.Model(&models.Order{}).Count(&resp.TotalOrders)
	db.Model(&models.Order{}).Where("status = ?", "pending").Count(&resp.PendingOrders)
	db.Model(&models.Order{}).Where("status = ?", "approved").Count(&resp.ApprovedOrders)
	db.Model(&models.Order{}).Where("status = ?", "completed").Count(&resp.CompletedOrders)
	db.Model(&models.Order{}).Where("status = ?", "rejected").Count(&resp.RejectedOrders)

	// Total members
	db.Model(&models.User{}).Where("role = ?", "member").Count(&resp.TotalMembers)

	// Top 5 products by total quantity ordered
	db.Model(&models.OrderItem{}).
		Select("products.name as product_name, SUM(order_items.quantity) as total_qty, SUM(order_items.price * order_items.quantity) as total_revenue").
		Joins("JOIN products ON products.id = order_items.product_id").
		Group("order_items.product_id, products.name").
		Order("total_qty DESC").
		Limit(5).
		Scan(&resp.TopProducts)

	// Recent 10 orders with username
	db.Model(&models.Order{}).
		Select("orders.id, users.username, orders.total_amount, orders.status, orders.created_at").
		Joins("JOIN users ON users.id = orders.user_id").
		Order("orders.created_at DESC").
		Limit(10).
		Scan(&resp.RecentOrders)

	// Revenue by verifier
	db.Model(&models.Order{}).
		Select("users.username as verifier_name, COUNT(orders.id) as order_count, COALESCE(SUM(orders.total_amount), 0) as total_revenue").
		Joins("JOIN users ON users.id = orders.verifier_id").
		Where("orders.status IN ?", []string{"approved", "completed"}).
		Group("orders.verifier_id, users.username").
		Order("total_revenue DESC").
		Scan(&resp.RevenueByVerifier)

	// Ensure slices are never nil in JSON output
	if resp.TopProducts == nil {
		resp.TopProducts = []TopProduct{}
	}
	if resp.RecentOrders == nil {
		resp.RecentOrders = []RecentOrder{}
	}
	if resp.RevenueByVerifier == nil {
		resp.RevenueByVerifier = []RevenueByVerifier{}
	}

	c.JSON(http.StatusOK, resp)
}
