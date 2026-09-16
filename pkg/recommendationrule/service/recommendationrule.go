package service

type RecommendationRuleData struct {
	Id        int    `json:"id"`
	CompanyId int    `json:"company_id"`
	Text      string `json:"text"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateRecommendationRuleRequest struct {
	CompanyId int    `json:"company_id,string" binding:"required"`
	Text      string `json:"text" binding:"required"`
}

// Pointers: a field omitted from the request body is left untouched — see
// RecommendationRuleRepository.Update. Lets the frontend toggle active
// without resending text, and vice versa.
type UpdateRecommendationRuleRequest struct {
	Text   *string `json:"text"`
	Active *bool   `json:"active"`
}

type RecommendationRuleListResponse struct {
	Status bool                     `json:"status"`
	Desc   string                   `json:"desc"`
	Data   []RecommendationRuleData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type RecommendationRuleService interface {
	GetAll(companyId int) (*RecommendationRuleListResponse, error)
	Create(req CreateRecommendationRuleRequest) (*SimpleResponse, error)
	Update(id, userId int, req UpdateRecommendationRuleRequest) (*SimpleResponse, error)
	Delete(id, userId int) (*SimpleResponse, error)
}
