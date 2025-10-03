package controllers

import "github.com/gin-gonic/gin"

type BaseController struct{}

func (c *BaseController) ValidateReqParams(ctx *gin.Context, requestParams interface{}) error {
	var err error

	switch ctx.ContentType() {
	case "application/json":
		err = ctx.ShouldBindJSON(requestParams)
	case "application/xml":
		err = ctx.ShouldBindXML(requestParams)
	case "":
		err = ctx.ShouldBindUri(requestParams)
		err = ctx.ShouldBindQuery(requestParams)
	default:
		err = ctx.ShouldBind(requestParams)
	}

	if err != nil {
		return err
	}

	return nil
}
func getOrderCode(ctx *gin.Context) string {
	// 1. Query string
	if code := ctx.Query("orderCode"); code != "" {
		return code
	}

	// 2. URL params
	if code := ctx.Param("orderCode"); code != "" {
		return code
	}

	// 3. JSON body
	var body struct {
		ProductCode string `json:"orderCode"`
	}
	if err := ctx.ShouldBindJSON(&body); err == nil && body.ProductCode != "" {
		return body.ProductCode
	}

	return ""
}

// getProductCode retrieves the product code from the request context.
func getProductCode(ctx *gin.Context) string {
	// 1. Query string
	if code := ctx.Query("productCode"); code != "" {
		return code
	}

	// 2. URL params
	if code := ctx.Param("productCode"); code != "" {
		return code
	}

	// 3. JSON body
	var body struct {
		ProductCode string `json:"productCode"`
	}
	if err := ctx.ShouldBindJSON(&body); err == nil && body.ProductCode != "" {
		return body.ProductCode
	}

	return ""
}
