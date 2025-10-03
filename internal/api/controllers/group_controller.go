package controllers

import (
	"net/http"
	"strconv"

	"web-api/internal/api/middlewares"
	"web-api/internal/api/services"

	"github.com/gin-gonic/gin"
)

type GroupController struct{}

// CreateGroup creates a new group
// @Summary Create a new group
// @Tags Groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body services.CreateGroupRequest true "Group request"
// @Success 201 {object} models.Group
// @Router /api/groups/create [post]
func (ctrl *GroupController) CreateGroup(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	var req services.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := services.Group.CreateGroup(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// AddMember adds a member to a group
// @Summary Add member to group
// @Tags Groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param request body services.AddMemberRequest true "Add member request"
// @Success 200
// @Router /api/groups/:id/add-member [post]
func (ctrl *GroupController) AddMember(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var req services.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.Group.AddMember(uint(groupID), userID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member added successfully"})
}

// RemoveMember removes a member from a group
// @Summary Remove member from group
// @Tags Groups
// @Security BearerAuth
// @Param id path int true "Group ID"
// @Param userID path int true "User ID to remove"
// @Success 200
// @Router /api/groups/:id/remove-member/:userID [delete]
func (ctrl *GroupController) RemoveMember(c *gin.Context) {
	requestorID, _ := middlewares.GetUserID(c)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	userID, err := strconv.ParseUint(c.Param("userID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := services.Group.RemoveMember(uint(groupID), requestorID, uint(userID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// GetGroupMembers retrieves all members of a group
// @Summary Get group members
// @Tags Groups
// @Security BearerAuth
// @Produce json
// @Param id path int true "Group ID"
// @Success 200 {array} models.GroupMember
// @Router /api/groups/:id/members [get]
func (ctrl *GroupController) GetGroupMembers(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	members, err := services.Group.GetGroupMembers(uint(groupID), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

// GetUserGroups retrieves all groups a user is member of
// @Summary Get user groups
// @Tags Groups
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Group
// @Router /api/groups [get]
func (ctrl *GroupController) GetUserGroups(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	groups, err := services.Group.GetUserGroups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

// GetGroupByID retrieves a group by ID
// @Summary Get group by ID
// @Tags Groups
// @Security BearerAuth
// @Produce json
// @Param id path int true "Group ID"
// @Success 200 {object} models.Group
// @Router /api/groups/:id [get]
func (ctrl *GroupController) GetGroupByID(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	group, err := services.Group.GetGroupByID(uint(groupID), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup deletes a group
// @Summary Delete group
// @Tags Groups
// @Security BearerAuth
// @Param id path int true "Group ID"
// @Success 200
// @Router /api/groups/:id [delete]
func (ctrl *GroupController) DeleteGroup(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	if err := services.Group.DeleteGroup(uint(groupID), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}
