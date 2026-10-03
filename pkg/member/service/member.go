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
	// Password only applies when Role is "customer" — lets an Admin/Team
	// Lead hand the client a specific, memorable password up front instead
	// of a random generated one. Ignored for every other role, which keep
	// the existing random-temp-password-then-forced-change flow unchanged.
	Password string `json:"password"`
}

// SetMemberPasswordRequest is for an existing Customer role member — an
// Admin/Team Lead changing the password they already set.
type SetMemberPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}

// MemberPasswordData.Password is nil when this member predates the
// encrypted-copy feature (created before it shipped, or role was changed to
// customer after creation) — the frontend shows "not available, reset it"
// rather than a blank field in that case.
type MemberPasswordData struct {
	Password *string `json:"password"`
}

type MemberPasswordResponse struct {
	Status bool               `json:"status"`
	Desc   string             `json:"desc"`
	Data   MemberPasswordData `json:"data"`
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
	// SetPassword and GetPassword are restricted to Customer role targets —
	// see their implementations for exactly why.
	SetPassword(id, requesterUserId int, req SetMemberPasswordRequest) (*SimpleResponse, error)
	GetPassword(id, requesterUserId int) (*MemberPasswordResponse, error)
}
