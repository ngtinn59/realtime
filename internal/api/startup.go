package api

import (
	"fmt"

	"web-api/internal/api/controllers"
	"web-api/internal/api/routers"
	"web-api/internal/pkg/config"
	"web-api/internal/pkg/database"
	"web-api/internal/pkg/redis"
	"web-api/internal/pkg/utils"
	"web-api/pkg/logger"
)

func Run(configPath string) {
	if configPath == "" {
		configPath = "data/config.yml"
	}
	
	// Setup configuration
	if err := config.Setup(configPath); err != nil {
		logger.Fatalf("failed to setup config, %s", err)
	}

	cfg := config.GetConfig()
	
	// Initialize JWT secret
	utils.SetJWTSecret(cfg.Server.Secret)

	// Setup database
	if err := database.Setup(); err != nil {
		logger.Fatalf("failed to setup database, %s", err)
	}

	// Setup Redis
	redisConfig := redis.Config{
		Host:     "redis", // Redis container hostname
		Port:     "6379",
		Password: "",
		DB:       0,
	}
	if err := redis.Setup(redisConfig); err != nil {
		logger.Fatalf("failed to setup Redis, %s", err)
	}

	// Initialize WebSocket hub
	controllers.InitWebSocketHub()

	// Setup router
	web := routers.Setup()
	
	// Setup chat routes
	routers.SetupChatRoutes(web)

	fmt.Println("================================>")
	fmt.Println("🚀 Chat API Server Started")
	fmt.Println("   Port: " + cfg.Server.Port)
	fmt.Println("   Mode: " + cfg.Server.Mode)
	fmt.Println("   WebSocket: ws://localhost:" + cfg.Server.Port + "/ws")
	fmt.Println("================================>")
	
	logger.Fatalf("%v", web.Run(":"+cfg.Server.Port))
}
