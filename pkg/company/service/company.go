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
	// Role is the calling user's effective role for this company —
	// "admin" | "team_lead" | "specialist" | "customer" — computed
	// server-side, never stored directly against this row.
	Role string `json:"role"`
	// PromptLimit caps how many active prompts this workspace can track.
	PromptLimit int `json:"prompt_limit"`
	// ContractEndDate is nil for companies that predate this feature.
	ContractEndDate *string `json:"contract_end_date"`
}

type UpdatePromptLimitRequest struct {
	Limit int `json:"limit" binding:"required,min=1"`
}

// UpdateContractRequest accepts either a preset duration in months (computed
// from today) or an explicit end date — exactly one must be set. Reused for
// both the initial mandatory contract at creation and later renewals.
type UpdateContractRequest struct {
	DurationMonths  *int    `json:"duration_months"`
	ContractEndDate *string `json:"contract_end_date"`
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
	Website  string  `json:"website" binding:"required"`
	Location *string `json:"location"`
	Size     *string `json:"size"`
	Founded  *string `json:"founded"`
	Phone    *string `json:"phone"`
	Email    *string `json:"email"`
	About    *string `json:"about"`
	Plan     string  `json:"plan"`
	// Duration is mandatory at creation — exactly one of these two must be
	// set (a preset like "1 month"/"1 year", or a custom end date).
	DurationMonths  *int    `json:"duration_months"`
	ContractEndDate *string `json:"contract_end_date"`
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

// AllAiEngines is the full set of platforms a company can toggle on the AI
// Setting page. A company with no rows in company_ai_engines yet has every
// one of these enabled by default (see buildAiEngineData).
var AllAiEngines = []string{"chatgpt", "gemini", "perplexity", "claude", "google-ai"}

var aiEngineLabels = map[string]string{
	"chatgpt":    "ChatGPT",
	"gemini":     "Gemini",
	"perplexity": "Perplexity",
	"claude":     "Claude",
	"google-ai":  "Google AI Overview",
}

type AiEngineData struct {
	Platform string `json:"platform"`
	Label    string `json:"label"`
	Enabled  bool   `json:"enabled"`
}

type UpdateAiEngineRequest struct {
	Enabled bool `json:"enabled"`
}

type NotificationPrefsData struct {
	EmailReports     bool `json:"email_reports"`
	VisibilityAlerts bool `json:"visibility_alerts"`
	MemberActivity   bool `json:"member_activity"`
	WeeklyDigest     bool `json:"weekly_digest"`
}

type UpdateNotificationPrefsRequest struct {
	EmailReports     bool `json:"email_reports"`
	VisibilityAlerts bool `json:"visibility_alerts"`
	MemberActivity   bool `json:"member_activity"`
	WeeklyDigest     bool `json:"weekly_digest"`
}

type NotificationPrefsResponse struct {
	Status bool                  `json:"status"`
	Desc   string                `json:"desc"`
	Data   NotificationPrefsData `json:"data"`
}

type AiEngineListResponse struct {
	Status bool           `json:"status"`
	Desc   string         `json:"desc"`
	Data   []AiEngineData `json:"data"`
}

type CompanyService interface {
	GetAll(userId int) (*CompanyListResponse, error)
	Create(userId int, req CreateCompanyRequest) (*CompanyResponse, error)
	GetById(id, userId int) (*CompanyResponse, error)
	Update(id int, req UpdateCompanyRequest) (*SimpleResponse, error)
	Delete(id int) (*SimpleResponse, error)
	AddTag(companyId int, tag string) (*SimpleResponse, error)
	DeleteTag(companyId int, tag string) (*SimpleResponse, error)
	AddMarket(companyId int, market string) (*SimpleResponse, error)
	DeleteMarket(companyId int, market string) (*SimpleResponse, error)
	UpdateSocialLinks(companyId int, req UpdateSocialLinksRequest) (*SimpleResponse, error)
	AddContact(companyId int, req AddContactRequest) (*SimpleResponse, error)
	DeleteContact(companyId, contactId int) (*SimpleResponse, error)
	GetAiEngines(companyId int) (*AiEngineListResponse, error)
	UpdateAiEngine(companyId int, platform string, enabled bool) (*SimpleResponse, error)
	GetNotificationPrefs(companyId int) (*NotificationPrefsResponse, error)
	UpdateNotificationPrefs(companyId int, req UpdateNotificationPrefsRequest) (*SimpleResponse, error)
	UpdatePromptLimit(companyId int, req UpdatePromptLimitRequest) (*SimpleResponse, error)
	UpdateContract(companyId int, req UpdateContractRequest) (*SimpleResponse, error)
}
