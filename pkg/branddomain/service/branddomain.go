package service

type BrandDomainData struct {
	Id        int    `json:"id"`
	BrandId   int    `json:"brand_id"`
	Domain    string `json:"domain"`
	CreatedAt string `json:"created_at"`
}

type CreateBrandDomainRequest struct {
	BrandId int    `json:"brand_id,string" binding:"required"`
	Domain  string `json:"domain" binding:"required"`
}

type BrandDomainListResponse struct {
	Status bool              `json:"status"`
	Desc   string            `json:"desc"`
	Data   []BrandDomainData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type BrandDomainService interface {
	GetAll(brandId int) (*BrandDomainListResponse, error)
	Create(userId int, req CreateBrandDomainRequest) (*SimpleResponse, error)
	Delete(id, userId int) (*SimpleResponse, error)
}
