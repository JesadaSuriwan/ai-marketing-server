package service

type SubdomainData struct {
	Id        int    `json:"id"`
	CompanyId int    `json:"company_id"`
	Subdomain string `json:"subdomain"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type CreateSubdomainRequest struct {
	CompanyId int    `json:"company_id" binding:"required"`
	Subdomain string `json:"subdomain" binding:"required"`
}

type SubdomainListResponse struct {
	Status bool            `json:"status"`
	Desc   string          `json:"desc"`
	Data   []SubdomainData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type SubdomainService interface {
	GetAll(companyId int) (*SubdomainListResponse, error)
	Create(req CreateSubdomainRequest) (*SimpleResponse, error)
	Delete(id, userId int) (*SimpleResponse, error)
}
