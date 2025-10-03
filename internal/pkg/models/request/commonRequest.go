package request

type PageInfo struct {
	PageNumber int `form:"pageNumber" json:"pageNumber" binding:"required,number"`
	PageSize   int `form:"pageSize" json:"pageSize" binding:"required,number"`
}
