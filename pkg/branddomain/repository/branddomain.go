package repository

import "github.com/jmoiron/sqlx"

type BrandDomain struct {
	Id        int    `db:"id"`
	BrandId   int    `db:"brand_id"`
	Domain    string `db:"domain"`
	CreatedAt string `db:"created_at"`
}

type BrandDomainRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(brandId int) ([]BrandDomain, error)
	GetById(id int) (*BrandDomain, error)
	Create(tx *sqlx.Tx, d BrandDomain) (int, error)
	Delete(tx *sqlx.Tx, id int) error
	// GetCompanyIdByBrandId resolves the owning company for a brand — used
	// to compute the caller's effective role for a permission check, since
	// brand_domains only carries brand_id, not company_id.
	GetCompanyIdByBrandId(brandId int) (int, error)
}
