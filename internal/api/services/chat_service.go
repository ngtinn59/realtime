package services

import (
	"errors"

	"web-api/internal/pkg/database"
	"web-api/internal/pkg/models"
	"web-api/internal/pkg/websocket"

	"gorm.io/gorm"
)

type ChatService struct{}

var Chat = &ChatService{}

// SendPrivateMessageRequest represents a private message request
type SendPrivateMessageRequest struct {
	ReceiverID uint               `json:"receiver_id" binding:"required"`
	Content    string             `json:"content" binding:"required"`
	Type       models.MessageType `json:"type"`
	FileID     *uint              `json:"file_id"`
}

// SendGroupMessageRequest represents a group message request
type SendGroupMessageRequest struct {
	GroupID uint               `json:"group_id" binding:"required"`
	Content string             `json:"content" binding:"required"`
	Type    models.MessageType `json:"type"`
	FileID  *uint              `json:"file_id"`
}

// SendPrivateMessage sends a private message
func (s *ChatService) SendPrivateMessage(senderID uint, req SendPrivateMessageRequest) (*models.PrivateMessage, error) {
	db := database.GetDB()

	// Verify receiver exists
	var receiver models.User
	if err := db.First(&receiver, req.ReceiverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("receiver not found")
		}
		return nil, err
	}

	// Create message
	message := models.PrivateMessage{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		Type:       req.Type,
		FileID:     req.FileID,
		IsRead:     false,
	}

	if message.Type == "" {
		message.Type = models.MessageTypeText
	}

	if err := db.Create(&message).Error; err != nil {
		return nil, err
	}

	// Broadcast message to WebSocket clients
	messageData := map[string]interface{}{
		"message_id":  message.ID,
		"sender_id":   message.SenderID,
		"receiver_id": message.ReceiverID,
		"content":     message.Content,
		"type":        string(message.Type),
		"file_id":     message.FileID,
		"created_at":  message.CreatedAt,
	}
	websocket.BroadcastPrivateMessage(senderID, req.ReceiverID, messageData)

	// Load sender and receiver info
	db.Preload("Sender").Preload("Receiver").Preload("File").First(&message, message.ID)

	return &message, nil
}

// GetPrivateMessages retrieves private messages between two users
func (s *ChatService) GetPrivateMessages(userID, otherUserID uint, limit, offset int) ([]models.PrivateMessage, error) {
	db := database.GetDB()

	var messages []models.PrivateMessage
	if err := db.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, otherUserID, otherUserID, userID,
	).
		Preload("Sender").
		Preload("Receiver").
		Preload("File").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// MarkMessageAsRead marks a message as read
func (s *ChatService) MarkMessageAsRead(messageID, userID uint) error {
	db := database.GetDB()

	// Verify user is the receiver
	var message models.PrivateMessage
	if err := db.First(&message, messageID).Error; err != nil {
		return err
	}

	if message.ReceiverID != userID {
		return errors.New("unauthorized to mark this message as read")
	}

	return db.Model(&message).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": gorm.Expr("NOW()"),
	}).Error
}

// GetUnreadMessageCount returns count of unread messages for a user
func (s *ChatService) GetUnreadMessageCount(userID uint) (int64, error) {
	db := database.GetDB()

	var count int64
	if err := db.Model(&models.PrivateMessage{}).
		Where("receiver_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// SendGroupMessage sends a message to a group
func (s *ChatService) SendGroupMessage(senderID uint, req SendGroupMessageRequest) (*models.GroupMessage, error) {
	db := database.GetDB()

	// Verify user is a member of the group
	var member models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ?", req.GroupID, senderID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("you are not a member of this group")
		}
		return nil, err
	}

	// Create message
	message := models.GroupMessage{
		GroupID:  req.GroupID,
		SenderID: senderID,
		Content:  req.Content,
		Type:     req.Type,
		FileID:   req.FileID,
	}

	if message.Type == "" {
		message.Type = models.MessageTypeText
	}

	if err := db.Create(&message).Error; err != nil {
		return nil, err
	}

	// Broadcast message to WebSocket clients
	messageData := map[string]interface{}{
		"message_id": message.ID,
		"group_id":   message.GroupID,
		"sender_id":  message.SenderID,
		"content":    message.Content,
		"type":       string(message.Type),
		"file_id":    message.FileID,
		"created_at": message.CreatedAt,
	}
	websocket.BroadcastGroupMessage(senderID, req.GroupID, messageData)

	// Load relations
	db.Preload("Sender").Preload("Group").Preload("File").First(&message, message.ID)

	return &message, nil
}

// GetGroupMessages retrieves messages from a group
func (s *ChatService) GetGroupMessages(userID, groupID uint, limit, offset int) ([]models.GroupMessage, error) {
	db := database.GetDB()

	// Verify user is a member
	var member models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("you are not a member of this group")
		}
		return nil, err
	}

	var messages []models.GroupMessage
	if err := db.Where("group_id = ?", groupID).
		Preload("Sender").
		Preload("File").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// GetConversations returns list of conversations for a user
func (s *ChatService) GetConversations(userID uint) ([]map[string]interface{}, error) {
	db := database.GetDB()

	// Get latest message with each user
	var conversations []map[string]interface{}

	// Query for private conversations
	query := `
    WITH conversations AS (
        SELECT 
            CASE 
                WHEN sender_id = ? THEN receiver_id 
                ELSE sender_id 
            END as other_user_id,
            MAX(created_at) as last_message_at
        FROM private_messages
        WHERE sender_id = ? OR receiver_id = ?
        GROUP BY other_user_id
    )
    SELECT 
        c.other_user_id,
        c.last_message_at,
        (
            SELECT pm2.content
            FROM private_messages pm2
            WHERE (pm2.sender_id = ? AND pm2.receiver_id = c.other_user_id)
               OR (pm2.sender_id = c.other_user_id AND pm2.receiver_id = ?)
            ORDER BY pm2.created_at DESC
            LIMIT 1
        ) as last_message
    FROM conversations c
    ORDER BY c.last_message_at DESC
`

	rows, err := db.Raw(query, userID, userID, userID, userID, userID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var conv map[string]interface{}
		var otherUserID uint
		var lastMessageAt, lastMessage string

		if err := rows.Scan(&otherUserID, &lastMessageAt, &lastMessage); err != nil {
			continue
		}

		// Get other user info
		var otherUser models.User
		db.First(&otherUser, otherUserID)

		conv = map[string]interface{}{
			"type":            "private",
			"user":            otherUser.ToResponse(),
			"last_message":    lastMessage,
			"last_message_at": lastMessageAt,
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}
