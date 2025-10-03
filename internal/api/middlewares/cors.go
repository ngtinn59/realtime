package middlewares

import (
	"web-api/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

// CORS middleware
func CORS() gin.HandlerFunc {
	config := config.GetConfig()

	return func(ctx *gin.Context) {
		if config.Cors.Global {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", config.Cors.Ips)
		}
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next()
	}
}
