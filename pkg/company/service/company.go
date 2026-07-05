package service

type CompanyData struct {
	Id          int             `json:"id"`
	UserId      int             `json:"user_id"`
	Name        string          `json:"name"`
	Initials    string          `json:"initials"`
	LogoUrl     *string         `json:"logo_url"`
	Industry    *string         `json:"industry"`
	Website     *string         `json:"website"`
	Location    *string         `json:"location"`
	Size        *string         `json:"size"`
	Founded     *string         `json:"founded"`
	Phone       *string         `json:"phone"`
	Email       *string         `json:"email"`
	About       *string         `json:"about"`
	Plan        string          `json:"plan"`
	Tags        []string        `json:"tags"`
	Markets     []string        `json:"markets"`
	SocialLinks *SocialLinksDTO `json:"social_links"`
	Contacts    []ContactDTO    `json:"key_contacts"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type SocialLinksDTO struct {
	Linkedin *string `json:"linkedin"`
	Twitter  *string `json:"twitter"`
	Facebook *string `json:"facebook"`
}

type ContactDTO struct {
	Id    int     `json:"id"`
	Name  string  `json:"name"`
	Role  string  `json:"role"`
	Email *string `json:"email"`
	Phone *string `json:"phone"`
}

type CreateCompanyRequest struct {
	Name     string  `json:"name" binding:"required"`
	Initials string  `json:"initials"`
	LogoUrl  *string `json:"logo_url"`
	Industry *string `json:"industry"`
	Website  *string `json:"website"`
	Location *string `json:"location"`
	Size     *string `json:"size"`
	Founded  *string `json:"founded"`
	Phone    *string `json:"phone"`
	Email    *string `json:"email"`
	About    *string `json:"about"`
	Plan     string  `json:"plan"`
}

type UpdateCompanyRequest struct {
	Name     string  `json:"name"`
	Initials string  `json:"initials"`
	LogoUrl  *string `json:"logo_url"`
	Industry *string `json:"industry"`
	Website  *string `json:"website"`
	Location *string `json:"location"`
	Size     *string `json:"size"`
	Founded  *string `json:"founded"`
	Phone    *string `json:"phone"`
	Email    *string `json:"email"`
	About    *string `json:"about"`
	Plan     string  `json:"plan"`
}

type AddTagRequest struct {
	Tag string `json:"tag" binding:"required"`
}

type AddMarketRequest struct {
	Market string `json:"market" binding:"required"`
}

type UpdateSocialLinksRequest struct {
	Linkedin *string `json:"linkedin"`
	Twitter  *string `json:"twitter"`
	Facebook *string `json:"facebook"`
}

type AddContactRequest struct {
	Name  string  `json:"name" binding:"required"`
	Role  string  `json:"role" binding:"required"`
	Email *string `json:"email"`
	Phone *string `json:"phone"`
}

type CompanyListResponse struct {
	Status bool          `json:"status"`
	Desc   string        `json:"desc"`
	Data   []CompanyData `json:"data"`
}

type CompanyResponse struct {
	Status bool        `json:"status"`
	Desc   string      `json:"desc"`
	Data   CompanyData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type CompanyService interface {
	GetAll(userId int) (*CompanyListResponse, error)
	Create(userId int, req CreateCompanyRequest) (*CompanyResponse, error)
	GetById(id int) (*CompanyResponse, error)
	Update(id int, req UpdateCompanyRequest) (*SimpleResponse, error)
	Delete(id int) (*SimpleResponse, error)
	AddTag(companyId int, tag string) (*SimpleResponse, error)
	DeleteTag(companyId int, tag string) (*SimpleResponse, error)
	AddMarket(companyId int, market string) (*SimpleResponse, error)
	DeleteMarket(companyId int, market string) (*SimpleResponse, error)
	UpdateSocialLinks(companyId int, req UpdateSocialLinksRequest) (*SimpleResponse, error)
	AddContact(companyId int, req AddContactRequest) (*SimpleResponse, error)
	DeleteContact(companyId, contactId int) (*SimpleResponse, error)
}
