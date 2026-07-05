package service

type PromptData struct {
	Id         int    `json:"id"`
	CompanyId  int    `json:"company_id"`
	CategoryId *int   `json:"category_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

type CreatePromptRequest struct {
	CompanyId  int    `json:"company_id" binding:"required"`
	CategoryId *int   `json:"category_id"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

type UpdatePromptRequest struct {
	CategoryId *int   `json:"category_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
}

type PromptListResponse struct {
	Status bool         `json:"status"`
	Desc   string       `json:"desc"`
	Data   []PromptData `json:"data"`
}

type PromptResponse struct {
	Status bool       `json:"status"`
	Desc   string     `json:"desc"`
	Data   PromptData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type PromptService interface {
	GetAll(companyId int) (*PromptListResponse, error)
	Create(req CreatePromptRequest) (*PromptResponse, error)
	GetById(id, userId int) (*PromptResponse, error)
	Update(id, userId int, req UpdatePromptRequest) (*SimpleResponse, error)
	Delete(id, userId int) (*SimpleResponse, error)
}
