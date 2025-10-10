package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"web-api/internal/pkg/database"
	"web-api/internal/pkg/models"
	"web-api/internal/pkg/redis"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	// hubInstance is the global hub instance
	hubInstance *Hub
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients (userID -> client)
	Clients map[uint]*Client

	// Mutex for thread-safe access to clients map
	mu sync.RWMutex

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Broadcast messages to clients
	Broadcast chan BroadcastMessage
}

// BroadcastMessage represents a message to be broadcasted
type BroadcastMessage struct {
	Message  Message
	SenderID uint
}

// Message represents a websocket message structure
type Message struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uint]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan BroadcastMessage, 256),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	// Start typing cleanup routine
	go h.typingCleanupRoutine()

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.handleBroadcast(message)
		}
	}
}

// typingCleanupRoutine periodically cleans up expired typing indicators
func (h *Hub) typingCleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := redis.CleanupExpiredTyping(); err != nil {
			logrus.Errorf("Failed to cleanup expired typing indicators: %v", err)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	h.Clients[client.UserID] = client
	h.mu.Unlock()

	// Set user as online in Redis
	if err := redis.SetUserOnline(client.UserID); err != nil {
		logrus.Errorf("failed to set user online: %v", err)
	}

	logrus.Infof("User %d (%s) connected. Total clients: %d", client.UserID, client.Username, len(h.Clients))

	// Broadcast user online status via Redis
	data := map[string]interface{}{
		"user_id":   client.UserID,
		"is_online": true,
	}

	channel := fmt.Sprintf("ws:user:%d", client.UserID)
	if err := redis.BroadcastToChannel(channel, "user_status", data); err != nil {
		logrus.Errorf("Failed to broadcast user online status: %v", err)
	}
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	if _, ok := h.Clients[client.UserID]; ok {
		delete(h.Clients, client.UserID)
		close(client.Send)
	}
	h.mu.Unlock()

	// Stop Redis subscriber
	client.StopRedisSubscriber()

	// Set user as offline in Redis
	if err := redis.SetUserOffline(client.UserID); err != nil {
		logrus.Errorf("failed to set user offline: %v", err)
	}

	logrus.Infof("User %d (%s) disconnected. Total clients: %d", client.UserID, client.Username, len(h.Clients))

	// Broadcast user offline status via Redis
	data := map[string]interface{}{
		"user_id":   client.UserID,
		"is_online": false,
		"last_seen": time.Now().Format(time.RFC3339),
	}

	channel := fmt.Sprintf("ws:user:%d", client.UserID)
	if err := redis.BroadcastToChannel(channel, "user_status", data); err != nil {
		logrus.Errorf("Failed to broadcast user offline status: %v", err)
	}
}

// validateMessage validates incoming WebSocket message structure
func validateMessage(msg Message) error {
	if msg.Event == "" {
		return errors.New("message event cannot be empty")
	}

	if msg.Data == nil {
		return errors.New("message data cannot be nil")
	}

	// Validate specific events
	switch msg.Event {
	case "send_private_message":
		if _, ok := msg.Data["receiver_id"].(float64); !ok {
			return errors.New("private message must have valid receiver_id")
		}
	case "send_group_message":
		if _, ok := msg.Data["group_id"].(float64); !ok {
			return errors.New("group message must have valid group_id")
		}
	case "user_typing":
		if _, ok := msg.Data["conversation_id"].(string); !ok {
			return errors.New("typing message must have conversation_id")
		}
	case "message_read":
		if _, ok := msg.Data["message_id"].(float64); !ok {
			return errors.New("message_read must have valid message_id")
		}
	}

	return nil
}

// handleBroadcast processes broadcast messages
func (h *Hub) handleBroadcast(bm BroadcastMessage) {
	// Validate message structure
	if err := validateMessage(bm.Message); err != nil {
		logrus.Errorf("Invalid message from user %d: %v", bm.SenderID, err)
		return
	}

	switch bm.Message.Event {
	case "send_private_message":
		h.handlePrivateMessage(bm)
	case "send_group_message":
		h.handleGroupMessage(bm)
	case "user_typing":
		h.handleTypingIndicator(bm)
	case "message_read":
		h.handleMessageRead(bm)
	case "ping":
		h.handlePing(bm)
	case "pong":
		logrus.Debugf("Received pong from user %d", bm.SenderID)
	default:
		logrus.Warnf("Unknown event: %s", bm.Message.Event)
	}
}

// handlePing handles ping messages and responds with pong
func (h *Hub) handlePing(bm BroadcastMessage) {
	logrus.Debugf("Received ping from user %d, sending pong", bm.SenderID)
	h.SendToUser(bm.SenderID, "pong", map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// handlePrivateMessage handles private message sending
func (h *Hub) handlePrivateMessage(bm BroadcastMessage) {
	logrus.Infof("Handling private message broadcast: %+v", bm)

	receiverID, ok := bm.Message.Data["receiver_id"].(float64)
	if !ok {
		logrus.Error("Invalid receiver_id in private message")
		return
	}

	content, ok := bm.Message.Data["content"].(string)
	if !ok {
		logrus.Error("Invalid content in private message")
		return
	}

	logrus.Infof("Saving private message to database and sending to receiver %d", uint(receiverID))

	// Save message to database first
	message, err := savePrivateMessageToDB(bm.SenderID, uint(receiverID), content, bm.Message.Data)
	if err != nil {
		logrus.Errorf("Failed to save private message to database: %v", err)
		return
	}

	// Update message data with database ID and timestamps
	updatedData := bm.Message.Data
	updatedData["message_id"] = message.ID
	updatedData["created_at"] = message.CreatedAt
	updatedData["updated_at"] = message.UpdatedAt

	// Send to receiver if online
	h.SendToUser(uint(receiverID), "private_message", updatedData)

	// Also send back to sender for confirmation
	h.SendToUser(bm.SenderID, "message_sent", updatedData)

	logrus.Info("Private message saved and broadcast completed")
}

// handleGroupMessage handles group message sending
func (h *Hub) handleGroupMessage(bm BroadcastMessage) {
	groupID, ok := bm.Message.Data["group_id"].(float64)
	if !ok {
		logrus.Error("Invalid group_id in group message")
		return
	}

	content, ok := bm.Message.Data["content"].(string)
	if !ok {
		logrus.Error("Invalid content in group message")
		return
	}

	logrus.Infof("Saving group message to database and broadcasting to group %d", uint(groupID))

	// Save message to database first
	message, err := saveGroupMessageToDB(bm.SenderID, uint(groupID), content, bm.Message.Data)
	if err != nil {
		logrus.Errorf("Failed to save group message to database: %v", err)
		return
	}

	// Update message data with database ID and timestamps
	updatedData := bm.Message.Data
	updatedData["message_id"] = message.ID
	updatedData["created_at"] = message.CreatedAt
	updatedData["updated_at"] = message.UpdatedAt

	// Broadcast to all group members via Redis
	channel := fmt.Sprintf("ws:group:%d", uint(groupID))
	if err := redis.BroadcastToChannel(channel, "group_message", updatedData); err != nil {
		logrus.Errorf("Failed to broadcast group message: %v", err)
		return
	}

	logrus.Info("Group message saved and broadcast completed")
}

// handleTypingIndicator handles typing indicator
func (h *Hub) handleTypingIndicator(bm BroadcastMessage) {
	conversationID, ok := bm.Message.Data["conversation_id"].(string)
	if !ok {
		return
	}

	// Set typing status in Redis
	redis.SetUserTyping(bm.SenderID, conversationID)

	// Determine chat type and ID from conversation_id (format: "private:123" or "group:456")
	var chatType string
	var chatID uint

	if strings.HasPrefix(conversationID, "private:") {
		chatType = "private"
		chatIDStr := strings.TrimPrefix(conversationID, "private:")
		if chatIDInt, parseErr := strconv.ParseUint(chatIDStr, 10, 32); parseErr == nil {
			chatID = uint(chatIDInt)
		} else {
			logrus.Errorf("Invalid private conversation ID: %s", conversationID)
			return
		}
	} else if strings.HasPrefix(conversationID, "group:") {
		chatType = "group"
		chatIDStr := strings.TrimPrefix(conversationID, "group:")
		if chatIDInt, parseErr := strconv.ParseUint(chatIDStr, 10, 32); parseErr == nil {
			chatID = uint(chatIDInt)
		} else {
			logrus.Errorf("Invalid group conversation ID: %s", conversationID)
			return
		}
	} else {
		logrus.Errorf("Invalid conversation ID format: %s", conversationID)
		return
	}

	// Prepare typing data for broadcast
	var typingChatID uint
	if chatType == "private" {
		typingChatID = bm.SenderID // For recipient, chat_id should be sender's ID
	} else {
		typingChatID = chatID // For groups, chat_id is the group ID
	}
	
	typingData := map[string]interface{}{
		"user_id":   bm.SenderID,
		"username":  bm.Message.Data["username"], // If available
		"is_typing": bm.Message.Data["is_typing"],
		"chat_type": chatType,
		"chat_id":   typingChatID,
	}

	if chatType == "private" {
		// For private chat, broadcast to the other participant
		h.SendToUser(chatID, "typing", typingData)
	} else if chatType == "group" {
		// For group chat, broadcast to all group members except sender
		// Get group members from database
		db := database.GetDB()
		var members []models.GroupMember
		if err := db.Where("group_id = ?", chatID).Find(&members).Error; err != nil {
			logrus.Errorf("Failed to get group members for group %d: %v", chatID, err)
			return
		}

		// Broadcast to all members except sender
		for _, member := range members {
			if member.UserID != bm.SenderID {
				h.SendToUser(member.UserID, "typing", typingData)
			}
		}
	}
}

// handleMessageRead handles message read acknowledgment
func (h *Hub) handleMessageRead(bm BroadcastMessage) {
	messageID, ok := bm.Message.Data["message_id"].(float64)
	if !ok {
		logrus.Error("Invalid message_id in message_read event")
		return
	}

	logrus.Infof("Message %d marked as read by user %d", uint(messageID), bm.SenderID)

	// For now, just log the event
	// TODO: Broadcast read status to relevant users (sender of the message)
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID uint, event string, data map[string]interface{}) {
	logrus.Infof("Attempting to send message to user %d, event: %s", userID, event)

	h.mu.RLock()
	client, ok := h.Clients[userID]
	h.mu.RUnlock()

	if ok {
		logrus.Infof("Found client for user %d, sending message", userID)
		client.SendMessage(event, data)
		logrus.Infof("Message sent to user %d successfully", userID)
	} else {
		logrus.Warnf("No client found for user %d, user may be offline", userID)
	}
}

// BroadcastToGroup sends a message to all members of a group
func (h *Hub) BroadcastToGroup(groupID uint, event string, data map[string]interface{}, excludeUserID uint) {
	// Note: For now, broadcast to all online users
	// TODO: Implement proper group member lookup to avoid import cycle
	h.mu.RLock()
	defer h.mu.RUnlock()

	for userID, client := range h.Clients {
		if userID != excludeUserID {
			client.SendMessage(event, data)
		}
	}
}

// broadcastUserStatus broadcasts user online/offline status
func (h *Hub) broadcastUserStatus(userID uint, isOnline bool) {
	data := map[string]interface{}{
		"user_id":   userID,
		"is_online": isOnline,
	}

	message, _ := json.Marshal(Message{
		Event: "user_online_status",
		Data:  data,
	})

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.Clients {
		if client.UserID != userID {
			select {
			case client.Send <- message:
			default:
				// Channel full, skip
			}
		}
	}
}

// GetOnlineUsers returns list of online user IDs
func (h *Hub) GetOnlineUsers() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uint, 0, len(h.Clients))
	for userID := range h.Clients {
		users = append(users, userID)
	}

	return users
}

// GetConnectionStats returns WebSocket connection statistics
func (h *Hub) GetConnectionStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(h.Clients),
		"clients":          make([]map[string]interface{}, 0, len(h.Clients)),
	}

	for _, client := range h.Clients {
		clientStats := map[string]interface{}{
			"user_id":  client.UserID,
			"username": client.Username,
		}
		stats["clients"] = append(stats["clients"].([]map[string]interface{}), clientStats)
	}

	return stats
}

// GetHub returns the global hub instance
func GetHub() *Hub {
	return hubInstance
}

// BroadcastPrivateMessage broadcasts a private message using Redis pub/sub
func BroadcastPrivateMessage(senderID, receiverID uint, messageData map[string]interface{}) {
	logrus.Infof("Publishing private message from %d to %d via Redis", senderID, receiverID)

	// Publish to Redis channel for the specific receiver
	channel := fmt.Sprintf("ws:user:%d", receiverID)
	if err := redis.BroadcastToChannel(channel, "private_message", messageData); err != nil {
		logrus.Errorf("Failed to publish private message to Redis: %v", err)
		return
	}

	// Also send confirmation back to sender
	senderChannel := fmt.Sprintf("ws:user:%d", senderID)
	confirmationData := map[string]interface{}{
		"type":        "message_sent",
		"message_id":  messageData["message_id"],
		"receiver_id": receiverID,
		"content":     messageData["content"],
		"created_at":  messageData["created_at"],
	}
	if err := redis.BroadcastToChannel(senderChannel, "message_sent", confirmationData); err != nil {
		logrus.Errorf("Failed to send confirmation to sender: %v", err)
		return
	}

	logrus.Info("Private message published to Redis successfully")
}

// savePrivateMessageToDB saves a private message to the database
func savePrivateMessageToDB(senderID, receiverID uint, content string, messageData map[string]interface{}) (*models.PrivateMessage, error) {
	db := database.GetDB()

	// Verify receiver exists
	var receiver models.User
	if err := db.First(&receiver, receiverID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("receiver not found")
		}
		return nil, err
	}

	// Determine message type
	msgType := models.MessageTypeText
	if t, ok := messageData["type"].(string); ok && t != "" {
		msgType = models.MessageType(t)
	}

	// Create message
	message := models.PrivateMessage{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		Type:       msgType,
		IsRead:     false,
	}

	// Handle file_id if present
	if fileID, ok := messageData["file_id"].(float64); ok && fileID > 0 {
		uintFileID := uint(fileID)
		message.FileID = &uintFileID
	}

	// Use transaction to ensure data consistency
	tx := db.Begin()
	if err := tx.Create(&message).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Load relations
	if err := tx.Preload("Sender").Preload("Receiver").Preload("File").First(&message, message.ID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &message, nil
}

// saveGroupMessageToDB saves a group message to the database
func saveGroupMessageToDB(senderID, groupID uint, content string, messageData map[string]interface{}) (*models.GroupMessage, error) {
	db := database.GetDB()

	// Verify user is a member of the group
	var member models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ?", groupID, senderID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("you are not a member of this group")
		}
		return nil, err
	}

	// Determine message type
	msgType := models.MessageTypeText
	if t, ok := messageData["type"].(string); ok && t != "" {
		msgType = models.MessageType(t)
	}

	// Create message
	message := models.GroupMessage{
		GroupID:  groupID,
		SenderID: senderID,
		Content:  content,
		Type:     msgType,
	}

	// Handle file_id if present
	if fileID, ok := messageData["file_id"].(float64); ok && fileID > 0 {
		uintFileID := uint(fileID)
		message.FileID = &uintFileID
	}

	// Use transaction to ensure data consistency
	tx := db.Begin()
	if err := tx.Create(&message).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Load relations
	if err := tx.Preload("Sender").Preload("Group").Preload("File").First(&message, message.ID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &message, nil
}
