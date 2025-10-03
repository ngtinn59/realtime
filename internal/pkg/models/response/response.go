package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommonResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func Result(ctx *gin.Context, code int, data interface{}, message string) {
	ctx.JSON(code, CommonResponse{
		code,
		data,
		message,
	})
}

func Ok(ctx *gin.Context) {
	Result(ctx, http.StatusOK, nil, "success")
}

func OkWithMessage(ctx *gin.Context, message string) {
	Result(ctx, http.StatusOK, nil, message)
}

func OkWithData(ctx *gin.Context, data interface{}) {
	Result(ctx, http.StatusOK, data, "success")
}

func OkWithDetailed(ctx *gin.Context, code int, data interface{}, message string) {
	Result(ctx, code, data, message)
}

func Fail(ctx *gin.Context) {
	Result(ctx, http.StatusInternalServerError, nil, "failure")
}

func FailWithMessage(ctx *gin.Context, message string) {
	Result(ctx, http.StatusInternalServerError, nil, message)
}

func FailWithDetailed(ctx *gin.Context, code int, data interface{}, message string) {
	Result(ctx, code, data, message)
}
