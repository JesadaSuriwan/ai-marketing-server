package service

type MemberData struct {
	Id        int    `json:"id"`
	CompanyId int    `json:"company_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	JoinedAt  string `json:"joined_at"`
	Status    string `json:"status"`
}

type CreateMemberRequest struct {
	CompanyId int    `json:"company_id,string" binding:"required"`
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

// CreateMemberData carries the one-time temp password back to the
// Admin/Team Lead who added the member. It's empty when the email already
// belonged to an existing account (that account is simply linked instead).
type CreateMemberData struct {
	TempPassword string `json:"temp_password"`
}

type CreateMemberResponse struct {
	Status bool             `json:"status"`
	Desc   string           `json:"desc"`
	Data   CreateMemberData `json:"data"`
}

type MemberService interface {
	GetAll(companyId int) (*MemberListResponse, error)
	Create(req CreateMemberRequest) (*CreateMemberResponse, error)
	Delete(id, userId int) (*SimpleResponse, error)
}
