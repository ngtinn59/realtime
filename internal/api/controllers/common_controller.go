package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommonController struct{}

var Common = &CommonController{}

// Ping handles health check endpoint
func (c *CommonController) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"status":  "ok",
	})
}
