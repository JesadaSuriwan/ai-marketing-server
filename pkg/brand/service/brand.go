package service

type BrandData struct {
	Id          int     `json:"id"`
	CompanyId   int     `json:"company_id"`
	Name        string  `json:"name"`
	Domain      string  `json:"domain"`
	Industry    *string `json:"industry"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	IsOwn       bool    `json:"is_own"`
	LogoUrl     *string `json:"logo_url"`
	CreatedAt   string  `json:"created_at"`
}

type CreateBrandRequest struct {
	CompanyId   int     `json:"company_id,string" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Domain      string  `json:"domain" binding:"required"`
	Industry    *string `json:"industry"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	IsOwn       bool    `json:"is_own"`
	LogoUrl     *string `json:"logo_url"`
}

type UpdateBrandRequest struct {
	Name        string  `json:"name"`
	Domain      string  `json:"domain"`
	Industry    *string `json:"industry"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	IsOwn       bool    `json:"is_own"`
	LogoUrl     *string `json:"logo_url"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type BrandListResponse struct {
	Status bool        `json:"status"`
	Desc   string      `json:"desc"`
	Data   []BrandData `json:"data"`
}

type BrandResponse struct {
	Status bool      `json:"status"`
	Desc   string    `json:"desc"`
	Data   BrandData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type BrandService interface {
	GetAll(companyId int) (*BrandListResponse, error)
	Create(req CreateBrandRequest) (*BrandResponse, error)
	GetById(id, userId int) (*BrandResponse, error)
	Update(id, userId int, req UpdateBrandRequest) (*SimpleResponse, error)
	Delete(id, userId int) (*SimpleResponse, error)
	UpdateStatus(id, userId int, status string) (*SimpleResponse, error)
}
