package websocket

import (
	"encoding/json"
	"sync"

	"web-api/internal/pkg/redis"

	"github.com/sirupsen/logrus"
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

	// Broadcast user online status
	h.broadcastUserStatus(client.UserID, true)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	if _, ok := h.Clients[client.UserID]; ok {
		delete(h.Clients, client.UserID)
		close(client.Send)
	}
	h.mu.Unlock()

	// Set user as offline in Redis
	if err := redis.SetUserOffline(client.UserID); err != nil {
		logrus.Errorf("failed to set user offline: %v", err)
	}

	logrus.Infof("User %d (%s) disconnected. Total clients: %d", client.UserID, client.Username, len(h.Clients))

	// Broadcast user offline status
	h.broadcastUserStatus(client.UserID, false)
}

// handleBroadcast processes broadcast messages
func (h *Hub) handleBroadcast(bm BroadcastMessage) {
	switch bm.Message.Event {
	case "send_private_message":
		h.handlePrivateMessage(bm)
	case "send_group_message":
		h.handleGroupMessage(bm)
	case "user_typing":
		h.handleTypingIndicator(bm)
	default:
		logrus.Warnf("Unknown event: %s", bm.Message.Event)
	}
}

// handlePrivateMessage handles private message sending
func (h *Hub) handlePrivateMessage(bm BroadcastMessage) {
	logrus.Infof("Handling private message broadcast: %+v", bm)

	receiverID, ok := bm.Message.Data["receiver_id"].(float64)
	if !ok {
		logrus.Error("Invalid receiver_id in private message")
		return
	}

	logrus.Infof("Sending private message to receiver %d", uint(receiverID))

	// Send to receiver if online
	h.SendToUser(uint(receiverID), "private_message", bm.Message.Data)

	// Also send back to sender for confirmation
	h.SendToUser(bm.SenderID, "message_sent", bm.Message.Data)

	logrus.Info("Private message broadcast completed")
}

// handleGroupMessage handles group message sending
func (h *Hub) handleGroupMessage(bm BroadcastMessage) {
	groupID, ok := bm.Message.Data["group_id"].(float64)
	if !ok {
		logrus.Error("Invalid group_id in group message")
		return
	}

	// Broadcast to all group members (implement group member lookup)
	h.BroadcastToGroup(uint(groupID), "group_message", bm.Message.Data, bm.SenderID)
}

// handleTypingIndicator handles typing indicator
func (h *Hub) handleTypingIndicator(bm BroadcastMessage) {
	conversationID, ok := bm.Message.Data["conversation_id"].(string)
	if !ok {
		return
	}

	// Set typing status in Redis
	redis.SetUserTyping(bm.SenderID, conversationID)

	// Broadcast to conversation participants
	if receiverID, ok := bm.Message.Data["receiver_id"].(float64); ok {
		h.SendToUser(uint(receiverID), "user_typing", bm.Message.Data)
	}
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
	// Note: You'll need to implement group member lookup from database
	// For now, this is a placeholder
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

// GetHub returns the global hub instance
func GetHub() *Hub {
	return hubInstance
}

// BroadcastPrivateMessage broadcasts a private message to WebSocket clients
func BroadcastPrivateMessage(senderID, receiverID uint, messageData map[string]interface{}) {
	logrus.Infof("Broadcasting private message from %d to %d: %v", senderID, receiverID, messageData)

	if hubInstance == nil {
		logrus.Error("Hub instance is nil, cannot broadcast message")
		return
	}

	logrus.Infof("Sending message to broadcast channel, current clients: %d", len(hubInstance.Clients))

	// Use select with timeout to avoid blocking
	select {
	case hubInstance.Broadcast <- BroadcastMessage{
		Message: Message{
			Event: "send_private_message",
			Data:  messageData,
		},
		SenderID: senderID,
	}:
		logrus.Info("Message sent to broadcast channel successfully")
	default:
		logrus.Error("Broadcast channel is full, message dropped")
	}
}

// BroadcastGroupMessage broadcasts a group message to WebSocket clients
func BroadcastGroupMessage(senderID, groupID uint, messageData map[string]interface{}) {
	if hubInstance == nil {
		return
	}

	hubInstance.Broadcast <- BroadcastMessage{
		Message: Message{
			Event: "send_group_message",
			Data:  messageData,
		},
		SenderID: senderID,
	}
}
