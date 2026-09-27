package repository

import "github.com/jmoiron/sqlx"

type Brand struct {
	Id          int     `db:"id"`
	CompanyId   int     `db:"company_id"`
	Name        string  `db:"name"`
	Domain      string  `db:"domain"`
	Industry    *string `db:"industry"`
	Description *string `db:"description"`
	Status      string  `db:"status"`
	IsOwn       bool    `db:"is_own"`
	LogoUrl     *string `db:"logo_url"`
	CreatedAt   string  `db:"created_at"`
}

// BrandAlias is an alternate name a brand is also known by — see
// brand_aliases in the migration for why this exists.
type BrandAlias struct {
	Id      int    `db:"id"`
	BrandId int    `db:"brand_id"`
	Name    string `db:"name"`
}

type BrandRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(companyId int) ([]Brand, error)
	Create(tx *sqlx.Tx, b Brand) (int, error)
	GetById(id int) (*Brand, error)
	Update(tx *sqlx.Tx, b Brand) error
	Delete(tx *sqlx.Tx, id int) error
	UpdateStatus(tx *sqlx.Tx, id int, status string) error
	// ClearOwn unsets is_own on every other brand in the company — enforces
	// "at most one own brand per company", which the dashboard queries assume.
	ClearOwn(tx *sqlx.Tx, companyId, exceptId int) error
	BelongsToUser(id, userId int) (bool, error)
	// AddAlias records that brandId is also known as name. A duplicate
	// (brand_id, name) is a silent no-op — the alias already exists.
	AddAlias(tx *sqlx.Tx, brandId int, name string) error
	// GetAliasesForCompany returns every alias across every brand the
	// company tracks — used to build the "tracked brands" list handed to
	// citation extraction, so a mention under an alias resolves to its
	// canonical brand instead of surfacing as a new candidate.
	GetAliasesForCompany(companyId int) ([]BrandAlias, error)
}
