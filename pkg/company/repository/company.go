package repository

import "github.com/jmoiron/sqlx"

type Company struct {
	Id        int     `db:"id"`
	UserId    int     `db:"user_id"`
	Name      string  `db:"name"`
	Initials  string  `db:"initials"`
	LogoUrl   *string `db:"logo_url"`
	Industry  *string `db:"industry"`
	Website   *string `db:"website"`
	Location  *string `db:"location"`
	Size      *string `db:"size"`
	Founded   *string `db:"founded"`
	Phone     *string `db:"phone"`
	Email     *string `db:"email"`
	About     *string `db:"about"`
	Plan      string  `db:"plan"`
	CreatedAt string  `db:"created_at"`
	UpdatedAt string  `db:"updated_at"`
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
}
