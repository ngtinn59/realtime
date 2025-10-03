package controllers

import (
	"net/http"

	"web-api/internal/api/services"
	"web-api/internal/pkg/utils"
	"web-api/internal/pkg/websocket"

	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var (
	upgrader = gorillaws.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// In production, implement proper origin checking
			return true
		},
	}

	// Hub is the global WebSocket hub
	Hub *websocket.Hub
)

// InitWebSocketHub initializes the WebSocket hub
func InitWebSocketHub() {
	Hub = websocket.NewHub()
	go Hub.Run()
	logrus.Info("✓ WebSocket hub initialized")
}

type WebSocketController struct{}

// HandleWebSocket handles WebSocket connections
// @Summary WebSocket endpoint
// @Description Establishes WebSocket connection for realtime chat
// @Tags WebSocket
// @Security BearerAuth
// @Param token query string true "JWT token"
// @Router /ws [get]
func (ctrl *WebSocketController) HandleWebSocket(c *gin.Context) {
	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	// Validate token
	claims, err := utils.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("Failed to upgrade connection: %v", err)
		return
	}

	// Create client
	client := &websocket.Client{
		Hub:      Hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   claims.UserID,
		Username: claims.Username,
	}

	// Register client
	client.Hub.Register <- client

	// Update user status in database
	services.User.UpdateUserStatus(claims.UserID, true)

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}
