package handlers

import (
	"net/http"

	"github.com/davi/omc-be/utils"
	"github.com/gin-gonic/gin"
)

// TestDiscordRefresh is an endpoint for testing Discord link refreshing locally
func TestDiscordRefresh(c *gin.Context) {
	var input struct {
		URLs []string `json:"urls" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refreshed, err := utils.RefreshDiscordURLs(input.URLs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Success",
		"refreshed": refreshed,
	})
}
