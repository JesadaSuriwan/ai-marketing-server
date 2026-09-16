package repository

import "github.com/jmoiron/sqlx"

type Company struct {
	Id       int     `db:"id"`
	UserId   int     `db:"user_id"`
	Name     string  `db:"name"`
	Initials string  `db:"initials"`
	LogoUrl  *string `db:"logo_url"`
	Industry *string `db:"industry"`
	Website  *string `db:"website"`
	Location *string `db:"location"`
	Size     *string `db:"size"`
	Founded  *string `db:"founded"`
	Phone    *string `db:"phone"`
	Email    *string `db:"email"`
	About    *string `db:"about"`
	Plan     string  `db:"plan"`
	// PromptLimit caps how many active prompts this workspace can track.
	PromptLimit int `db:"prompt_limit"`
	// ContractEndDate is nil for companies created before this feature
	// existed; new workspaces are required to set one at creation.
	ContractEndDate *string `db:"contract_end_date"`
	CreatedAt       string  `db:"created_at"`
	UpdatedAt       string  `db:"updated_at"`
	// MemberRole is only populated by GetAll's LEFT JOIN — nil for the
	// caller's own companies (they're "admin" by virtue of ownership, not a
	// company_members row) or when the field wasn't selected at all.
	MemberRole *string `db:"member_role"`
}

type CompanyTag struct {
	Id        int    `db:"id"`
	CompanyId int    `db:"company_id"`
	Tag       string `db:"tag"`
}

type CompanyMarket struct {
	Id        int    `db:"id"`
	CompanyId int    `db:"company_id"`
	Market    string `db:"market"`
}

type CompanySocialLinks struct {
	Id        int     `db:"id"`
	CompanyId int     `db:"company_id"`
	Linkedin  *string `db:"linkedin"`
	Twitter   *string `db:"twitter"`
	Facebook  *string `db:"facebook"`
}

type CompanyKeyContact struct {
	Id        int     `db:"id"`
	CompanyId int     `db:"company_id"`
	Name      string  `db:"name"`
	Role      string  `db:"role"`
	Email     *string `db:"email"`
	Phone     *string `db:"phone"`
}

type CompanyAiEngine struct {
	Id        int    `db:"id"`
	CompanyId int    `db:"company_id"`
	Platform  string `db:"platform"`
	Enabled   bool   `db:"enabled"`
}

type CompanyNotificationPrefs struct {
	Id               int  `db:"id"`
	CompanyId        int  `db:"company_id"`
	EmailReports     bool `db:"email_reports"`
	VisibilityAlerts bool `db:"visibility_alerts"`
	MemberActivity   bool `db:"member_activity"`
	WeeklyDigest     bool `db:"weekly_digest"`
}

type CompanyRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	Create(tx *sqlx.Tx, c Company) (int, error)
	GetAll(userId int) ([]Company, error)
	GetById(id int) (*Company, error)
	Update(tx *sqlx.Tx, c Company) error
	Delete(tx *sqlx.Tx, id int) error
	GetTags(companyId int) ([]CompanyTag, error)
	AddTag(tx *sqlx.Tx, companyId int, tag string) error
	DeleteTag(tx *sqlx.Tx, companyId int, tag string) error
	GetMarkets(companyId int) ([]CompanyMarket, error)
	AddMarket(tx *sqlx.Tx, companyId int, market string) error
	DeleteMarket(tx *sqlx.Tx, companyId int, market string) error
	GetSocialLinks(companyId int) (*CompanySocialLinks, error)
	UpsertSocialLinks(tx *sqlx.Tx, companyId int, linkedin, twitter, facebook *string) error
	GetKeyContacts(companyId int) ([]CompanyKeyContact, error)
	AddKeyContact(tx *sqlx.Tx, contact CompanyKeyContact) (int, error)
	DeleteKeyContact(tx *sqlx.Tx, companyId, contactId int) error
	GetAiEngines(companyId int) ([]CompanyAiEngine, error)
	SetAiEngine(tx *sqlx.Tx, companyId int, platform string, enabled bool) error
	GetNotificationPrefs(companyId int) (*CompanyNotificationPrefs, error)
	UpsertNotificationPrefs(tx *sqlx.Tx, companyId int, prefs CompanyNotificationPrefs) error
	SetPromptLimit(tx *sqlx.Tx, companyId, limit int) error
	SetContractEndDate(tx *sqlx.Tx, companyId int, contractEndDate string) error
}
