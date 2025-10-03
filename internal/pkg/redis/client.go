package redis

import (
	"context"
	"fmt"
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

// Close closes Redis connection
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
