package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	Client *redis.Client
	ctx    = context.Background()
)

// Config holds Redis configuration
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// Setup initializes Redis client
func Setup(config Config) error {
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
		PoolSize: 10,
	})

	// Test connection
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logrus.Info("✓ Connected to Redis successfully")
	return nil
}

// SetUserOnline sets user as online in Redis
func SetUserOnline(userID uint) error {
	key := fmt.Sprintf("user:online:%d", userID)
	return Client.Set(ctx, key, "1", 0).Err()
}

// SetUserOffline removes user from online list
func SetUserOffline(userID uint) error {
	key := fmt.Sprintf("user:online:%d", userID)
	return Client.Del(ctx, key).Err()
}

// IsUserOnline checks if user is online
func IsUserOnline(userID uint) (bool, error) {
	key := fmt.Sprintf("user:online:%d", userID)
	result, err := Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetOnlineUsers returns list of online user IDs
func GetOnlineUsers() ([]string, error) {
	pattern := "user:online:*"
	var userIDs []string

	iter := Client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// Extract user ID from key (user:online:123 -> 123)
		var userID string
		fmt.Sscanf(key, "user:online:%s", &userID)
		userIDs = append(userIDs, userID)
	}
	
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}

// SetUserTyping sets user as typing in a conversation
func SetUserTyping(userID uint, conversationID string) error {
	key := fmt.Sprintf("typing:%s:%d", conversationID, userID)
	return Client.Set(ctx, key, "1", 10*time.Second).Err()
}

// CleanupExpiredTyping removes expired typing indicators
func CleanupExpiredTyping() error {
	pattern := "typing:*:*"
	var keysToDelete []string

	iter := Client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// Check if key has expired TTL
		ttl, err := Client.TTL(ctx, key).Result()
		if err != nil {
			continue
		}
		// If TTL is -2, key doesn't exist (expired and deleted)
		// If TTL is -1, key exists but has no expiration
		// Only delete keys that have expired (TTL < 0 and exists)
		if ttl < 0 && ttl != -1 {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	if err := iter.Err(); err != nil {
		return err
	}

	if len(keysToDelete) > 0 {
		return Client.Del(ctx, keysToDelete...).Err()
	}
	return nil
}

// GetTypingUsers gets users currently typing in a conversation
func GetTypingUsers(conversationID string) ([]string, error) {
	pattern := fmt.Sprintf("typing:%s:*", conversationID)
	var userIDs []string

	iter := Client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		var userID string
		fmt.Sscanf(key, "typing:"+conversationID+":%s", &userID)
		userIDs = append(userIDs, userID)
	}
	
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}

// PublishMessage publishes a message to a channel for pub/sub
func PublishMessage(channel string, message interface{}) error {
	return Client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to a channel
func Subscribe(channels ...string) *redis.PubSub {
	return Client.Subscribe(ctx, channels...)
}

// SetWebSocketConnection stores WebSocket connection info for a user
func SetWebSocketConnection(userID uint, connectionID string) error {
	key := fmt.Sprintf("ws:connection:%d", userID)
	return Client.Set(ctx, key, connectionID, 24*time.Hour).Err() // Expire after 24 hours
}

// GetWebSocketConnection gets WebSocket connection info for a user
func GetWebSocketConnection(userID uint) (string, error) {
	key := fmt.Sprintf("ws:connection:%d", userID)
	return Client.Get(ctx, key).Result()
}

// RemoveWebSocketConnection removes WebSocket connection info for a user
func RemoveWebSocketConnection(userID uint) error {
	key := fmt.Sprintf("ws:connection:%d", userID)
	return Client.Del(ctx, key).Err()
}

// PublishWebSocketMessage publishes a WebSocket message to all subscribers
func PublishWebSocketMessage(channel string, messageData map[string]interface{}) error {
	jsonData, err := json.Marshal(messageData)
	if err != nil {
		return err
	}
	return Client.Publish(ctx, channel, string(jsonData)).Err()
}

// SubscribeWebSocket subscribes to WebSocket message channel
func SubscribeWebSocket(channel string) *redis.PubSub {
	return Client.Subscribe(ctx, channel)
}

// StoreUserSession stores user session info in Redis
func StoreUserSession(userID uint, sessionData map[string]interface{}) error {
	key := fmt.Sprintf("session:user:%d", userID)
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}
	return Client.Set(ctx, key, string(jsonData), 24*time.Hour).Err()
}

// GetUserSession gets user session info from Redis
func GetUserSession(userID uint) (map[string]interface{}, error) {
	key := fmt.Sprintf("session:user:%d", userID)
	jsonStr, err := Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var sessionData map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &sessionData)
	return sessionData, err
}

// BroadcastToChannel broadcasts a message to a specific channel
func BroadcastToChannel(channel string, event string, data map[string]interface{}) error {
	message := map[string]interface{}{
		"event": event,
		"data":  data,
		"timestamp": time.Now().Unix(),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return Client.Publish(ctx, channel, string(jsonData)).Err()
}

// GetActiveConnections gets all active WebSocket connections
func GetActiveConnections() ([]uint, error) {
	pattern := "ws:connection:*"
	var userIDs []uint

	iter := Client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		var userIDStr string
		fmt.Sscanf(key, "ws:connection:%s", &userIDStr)

		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userIDs = append(userIDs, uint(userID))
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}
