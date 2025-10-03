package controllers

import (
	"net/http"
	"strconv"

	"web-api/internal/api/middlewares"
	"web-api/internal/api/services"

	"github.com/gin-gonic/gin"
)

type ChatController struct{}

// SendPrivateMessage sends a private message
// @Summary Send private message
// @Tags Chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body services.SendPrivateMessageRequest true "Message request"
// @Success 201 {object} models.PrivateMessage
// @Router /api/messages/private [post]
func (ctrl *ChatController) SendPrivateMessage(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	var req services.SendPrivateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := services.Chat.SendPrivateMessage(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetPrivateMessages retrieves private messages with a user
// @Summary Get private messages
// @Tags Chat
// @Security BearerAuth
// @Produce json
// @Param userID path int true "Other User ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} models.PrivateMessage
// @Router /api/messages/private/:userID [get]
func (ctrl *ChatController) GetPrivateMessages(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	otherUserID, err := strconv.ParseUint(c.Param("userID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	messages, err := services.Chat.GetPrivateMessages(userID, uint(otherUserID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

// SendGroupMessage sends a group message
// @Summary Send group message
// @Tags Chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body services.SendGroupMessageRequest true "Message request"
// @Success 201 {object} models.GroupMessage
// @Router /api/messages/group [post]
func (ctrl *ChatController) SendGroupMessage(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	var req services.SendGroupMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := services.Chat.SendGroupMessage(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetGroupMessages retrieves messages from a group
// @Summary Get group messages
// @Tags Chat
// @Security BearerAuth
// @Produce json
// @Param groupID path int true "Group ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} models.GroupMessage
// @Router /api/messages/group/:groupID [get]
func (ctrl *ChatController) GetGroupMessages(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	groupID, err := strconv.ParseUint(c.Param("groupID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	messages, err := services.Chat.GetGroupMessages(userID, uint(groupID), limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetConversations returns user's conversations
// @Summary Get conversations
// @Tags Chat
// @Security BearerAuth
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /api/conversations [get]
func (ctrl *ChatController) GetConversations(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	conversations, err := services.Chat.GetConversations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

// MarkMessageAsRead marks a message as read
// @Summary Mark message as read
// @Tags Chat
// @Security BearerAuth
// @Param messageID path int true "Message ID"
// @Success 200
// @Router /api/messages/:messageID/read [post]
func (ctrl *ChatController) MarkMessageAsRead(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	messageID, err := strconv.ParseUint(c.Param("messageID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := services.Chat.MarkMessageAsRead(uint(messageID), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message marked as read"})
}

// GetUnreadCount returns unread message count
// @Summary Get unread message count
// @Tags Chat
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/messages/unread/count [get]
func (ctrl *ChatController) GetUnreadCount(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	count, err := services.Chat.GetUnreadMessageCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
