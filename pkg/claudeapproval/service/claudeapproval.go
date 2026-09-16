package service

type CreateClaudeApprovalRequest struct {
	CompanyId int `json:"company_id,string" binding:"required"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

// EmailProvider is implemented by whatever transactional-email client sends
// the actual request/approval emails (providers/resend.Client in production)
// — mirrors pkg/member/service's identical interface for the same reason:
// each package defines exactly the small interface it needs.
type EmailProvider interface {
	Send(to, subject, html string) error
}

type ClaudeApprovalService interface {
	// Create is called by a Specialist requesting Claude access — emails the
	// workspace creator and any active Team Leads a one-click approval link.
	Create(companyId, requestedByUserId int) (*SimpleResponse, error)
	// Approve is the public, no-auth landing the emailed link points at.
	Approve(token string) (*SimpleResponse, error)
}
