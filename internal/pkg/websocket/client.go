package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"web-api/internal/pkg/redis"

	"github.com/gorilla/websocket"
	redispkg "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512 KB
)

// Client represents a websocket client
type Client struct {
	Hub             *Hub
	Conn            *websocket.Conn
	Send            chan []byte
	UserID          uint
	Username        string
	redisSubscriber *redispkg.PubSub
	stopSubscriber  chan struct{}
}

// StartRedisSubscriber starts listening for Redis messages for this user
func (c *Client) StartRedisSubscriber() {
	channel := fmt.Sprintf("ws:user:%d", c.UserID)

	pubsub := redis.SubscribeWebSocket(channel)
	c.redisSubscriber = pubsub
	c.stopSubscriber = make(chan struct{})

	logrus.Infof("Started Redis subscriber for user %d on channel %s", c.UserID, channel)

	go func() {
		defer func() {
			if c.redisSubscriber != nil {
				c.redisSubscriber.Close()
			}
			logrus.Infof("Redis subscriber stopped for user %d", c.UserID)
		}()

		for {
			select {
			case <-c.stopSubscriber:
				logrus.Infof("Stopping Redis subscriber for user %d", c.UserID)
				return
			default:
				if c.redisSubscriber == nil {
					return
				}

				msg, err := c.redisSubscriber.ReceiveMessage(context.Background())
				if err != nil {
					logrus.Errorf("Redis subscriber error for user %d: %v", c.UserID, err)
					return
				}

				// Parse Redis message
				var messageData map[string]interface{}
				if err := json.Unmarshal([]byte(msg.Payload), &messageData); err != nil {
					logrus.Errorf("Failed to unmarshal Redis message: %v", err)
					continue
				}

				// Create WebSocket message
				event, ok := messageData["event"].(string)
				if !ok {
					continue
				}

				data, ok := messageData["data"].(map[string]interface{})
				if !ok {
					data = make(map[string]interface{})
				}

				wsMessage := Message{
					Event: event,
					Data:  data,
				}

				jsonMsg, err := json.Marshal(wsMessage)
				if err != nil {
					logrus.Errorf("Failed to marshal WebSocket message: %v", err)
					continue
				}

				// Send to client's WebSocket connection with timeout
				select {
				case c.Send <- jsonMsg:
					logrus.Debugf("Sent Redis message to user %d: %s", c.UserID, event)
				case <-time.After(1 * time.Second):
					logrus.Warnf("Client %d send channel timeout, dropping message", c.UserID)
				case <-c.stopSubscriber:
					return
				}
			}
		}
	}()
}

// StopRedisSubscriber stops the Redis subscriber
func (c *Client) StopRedisSubscriber() {
	if c.stopSubscriber != nil {
		select {
		case <-c.stopSubscriber:
			// Already closed
		default:
			close(c.stopSubscriber)
		}
	}
	if c.redisSubscriber != nil {
		c.redisSubscriber.Close()
		c.redisSubscriber = nil
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
		c.StopRedisSubscriber()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("websocket error: %v", err)
			}
			break
		}

		// Parse message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			logrus.Errorf("failed to unmarshal message: %v", err)
			continue
		}

		// Add sender info to message data
		if msg.Data == nil {
			msg.Data = make(map[string]interface{})
		}
		msg.Data["sender_id"] = c.UserID
		msg.Data["sender_username"] = c.Username

		// Send to hub for processing
		c.Hub.Broadcast <- BroadcastMessage{
			Message:  msg,
			SenderID: c.UserID,
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		c.StopRedisSubscriber()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(event string, data map[string]interface{}) error {
	msg := Message{
		Event: event,
		Data:  data,
	}

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.Send <- jsonMsg:
		return nil
	default:
		// Channel is full or closed
		return nil
	}
}
