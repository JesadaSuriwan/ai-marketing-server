package service

type VisibilityData struct {
	Id        int     `json:"id"`
	BrandId   int     `json:"brand_id"`
	PromptId  *int    `json:"prompt_id"`
	Platform  string  `json:"platform"`
	Score     float64 `json:"score"`
	Mentions  int     `json:"mentions"`
	Date      string  `json:"date"`
	CreatedAt string  `json:"created_at"`
}

type PlatformBreakdownData struct {
	Platform string  `json:"platform"`
	Score    float64 `json:"score"`
	Mentions int     `json:"mentions"`
}

type CreateVisibilityRequest struct {
	BrandId  int     `json:"brand_id" binding:"required"`
	PromptId *int    `json:"prompt_id"`
	Platform string  `json:"platform" binding:"required"`
	Score    float64 `json:"score"`
	Mentions int     `json:"mentions"`
	Date     string  `json:"date" binding:"required"`
}

type VisibilityListResponse struct {
	Status bool             `json:"status"`
	Desc   string           `json:"desc"`
	Data   []VisibilityData `json:"data"`
}

type BreakdownResponse struct {
	Status bool                    `json:"status"`
	Desc   string                  `json:"desc"`
	Data   []PlatformBreakdownData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type VisibilityService interface {
	GetAll(brandId int, from, to string) (*VisibilityListResponse, error)
	Create(req CreateVisibilityRequest) (*SimpleResponse, error)
	GetBreakdown(brandId int) (*BreakdownResponse, error)
}
