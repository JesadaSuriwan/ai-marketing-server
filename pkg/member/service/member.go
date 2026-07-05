package service

type MemberData struct {
	Id        int    `json:"id"`
	CompanyId int    `json:"company_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	JoinedAt  string `json:"joined_at"`
}

type CreateMemberRequest struct {
	CompanyId int    `json:"company_id" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
}

type MemberListResponse struct {
	Status bool         `json:"status"`
	Desc   string       `json:"desc"`
	Data   []MemberData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type MemberService interface {
	GetAll(companyId int) (*MemberListResponse, error)
	Create(req CreateMemberRequest) (*SimpleResponse, error)
	Delete(id, userId int) (*SimpleResponse, error)
}
