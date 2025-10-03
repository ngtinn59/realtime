package response

type PaginationResponse struct {
	List       interface{} `json:"list"`
	TotalRow   int64       `json:"totalRow"`
	TotalPage  int         `json:"totalPage"`
	PageNumber int         `json:"pageNumber"`
}
