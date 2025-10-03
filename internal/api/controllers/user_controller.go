package controllers

import (
	"net/http"
	"strconv"

	"web-api/internal/api/services"

	"github.com/gin-gonic/gin"
)

type UserController struct{}

// GetOnlineUsers returns list of online users
// @Summary Get online users
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.UserResponse
// @Router /api/users/online [get]
func (ctrl *UserController) GetOnlineUsers(c *gin.Context) {
	users, err := services.User.GetOnlineUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// SearchUsers searches for users
// @Summary Search users
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} models.UserResponse
// @Router /api/users/search [get]
func (ctrl *UserController) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	users, err := services.User.SearchUsers(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUserByID returns user by ID
// @Summary Get user by ID
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.UserResponse
// @Router /api/users/:id [get]
func (ctrl *UserController) GetUserByID(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := services.User.GetUserByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}
