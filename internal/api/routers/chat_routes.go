package routers

import (
	"web-api/internal/api/controllers"
	"web-api/internal/api/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupChatRoutes sets up chat-related routes
func SetupChatRoutes(router *gin.Engine) {
	authCtrl := &controllers.AuthController{}
	userCtrl := &controllers.UserController{}
	chatCtrl := &controllers.ChatController{}
	groupCtrl := &controllers.GroupController{}
	fileCtrl := &controllers.FileController{}
	wsCtrl := &controllers.WebSocketController{}

	api := router.Group("/api")
	{
		// Public routes
		api.POST("/register", authCtrl.Register)
		api.POST("/login", authCtrl.Login)

		// Protected routes
		protected := api.Group("")
		protected.Use(middlewares.AuthMiddleware())
		{
			// Auth/Profile
			protected.GET("/profile", authCtrl.GetProfile)

			// Users
			protected.GET("/users/online", userCtrl.GetOnlineUsers)
			protected.GET("/users/search", userCtrl.SearchUsers)
			protected.GET("/users/:id", userCtrl.GetUserByID)

			// Private Messages
			protected.POST("/messages/private", chatCtrl.SendPrivateMessage)
			protected.GET("/messages/private/:userID", chatCtrl.GetPrivateMessages)
			protected.POST("/messages/:messageID/read", chatCtrl.MarkMessageAsRead)
			protected.GET("/messages/unread/count", chatCtrl.GetUnreadCount)

			// Group Messages
			protected.POST("/messages/group", chatCtrl.SendGroupMessage)
			protected.GET("/messages/group/:groupID", chatCtrl.GetGroupMessages)

			// Conversations
			protected.GET("/conversations", chatCtrl.GetConversations)

			// Groups
			protected.POST("/groups/create", groupCtrl.CreateGroup)
			protected.GET("/groups", groupCtrl.GetUserGroups)
			protected.GET("/groups/:id", groupCtrl.GetGroupByID)
			protected.POST("/groups/:id/add-member", groupCtrl.AddMember)
			protected.DELETE("/groups/:id/remove-member/:userID", groupCtrl.RemoveMember)
			protected.GET("/groups/:id/members", groupCtrl.GetGroupMembers)
			protected.DELETE("/groups/:id", groupCtrl.DeleteGroup)

			// Files
			protected.POST("/files/upload", fileCtrl.UploadFile)
			protected.GET("/files", fileCtrl.GetUserFiles)
			protected.GET("/files/:id", fileCtrl.GetFile)
			protected.DELETE("/files/:id", fileCtrl.DeleteFile)
		}
	}

	// WebSocket endpoint (authentication via query parameter)
	router.GET("/ws", wsCtrl.HandleWebSocket)

	// Serve uploaded files
	router.Static("/uploads", "./uploads")
}
