package service

type CategoryData struct {
	Id        int    `json:"id"`
	CompanyId int    `json:"company_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type CreateCategoryRequest struct {
	CompanyId int    `json:"company_id,string" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type CategoryListResponse struct {
	Status bool           `json:"status"`
	Desc   string         `json:"desc"`
	Data   []CategoryData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type CategoryService interface {
	GetAll(companyId int) (*CategoryListResponse, error)
	Create(req CreateCategoryRequest) (*SimpleResponse, error)
	Update(id, userId int, req UpdateCategoryRequest) (*SimpleResponse, error)
	Delete(id, userId int) (*SimpleResponse, error)
}
