package routers

import (
	"fmt"
	"io"
	"os"

	"web-api/internal/api/middlewares"
	router_v1 "web-api/internal/api/routers/v1"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	app := gin.New()

	// Logging to a file.
	f, _ := os.Create("log/application.log")
	gin.DisableConsoleColor()
	gin.DefaultWriter = io.MultiWriter(f)

	// Middlewares
	app.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - - [%s] \"%s %s %s %d %s \" \" %s\" \" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	app.Use(middlewares.RequestLogger())
	app.Use(middlewares.RecoveryHandler)
	app.Use(middlewares.CORS())
	app.NoMethod(middlewares.NoMethodHandler())
	app.NoRoute(middlewares.NoRouteHandler())

	router_v1.Register(app)

	return app
}
