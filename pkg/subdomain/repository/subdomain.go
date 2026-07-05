package repository

import "github.com/jmoiron/sqlx"

type Subdomain struct {
	Id        int    `db:"id"`
	CompanyId int    `db:"company_id"`
	Subdomain string `db:"subdomain"`
	Status    string `db:"status"`
	CreatedAt string `db:"created_at"`
}

type SubdomainRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(companyId int) ([]Subdomain, error)
	Create(tx *sqlx.Tx, s Subdomain) (int, error)
	Delete(tx *sqlx.Tx, id int) error
	BelongsToUser(id, userId int) (bool, error)
}
