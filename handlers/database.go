package handlers

import (
	"net/http"
	"strings"

	"github.com/davi/omc-be/database"

	"github.com/gin-gonic/gin"
)

func RunRawSQL(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or missing query"})
		return
	}

	query := strings.TrimSpace(req.Query)
	isSelect := strings.HasPrefix(strings.ToUpper(query), "SELECT") || strings.HasPrefix(strings.ToUpper(query), "PRAGMA")

	if isSelect {
		var results []map[string]interface{}
		if err := database.DB.Raw(query).Scan(&results).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": results})
	} else {
		res := database.DB.Exec(query)
		if res.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": res.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Query executed successfully", "rows_affected": res.RowsAffected})
	}
}

func DownloadDatabase(c *gin.Context) {
	c.FileAttachment("/data/omc.db", "omc_live_backup.db")
}
