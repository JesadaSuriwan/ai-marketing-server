package repository

import "github.com/jmoiron/sqlx"

type subdomainRepositoryDB struct {
	db *sqlx.DB
}

func NewSubdomainRepositoryDB(db *sqlx.DB) SubdomainRepository {
	return subdomainRepositoryDB{db}
}

func (r subdomainRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r subdomainRepositoryDB) GetAll(companyId int) ([]Subdomain, error) {
	list := []Subdomain{}
	err := r.db.Select(&list, `SELECT id, company_id, subdomain, status, created_at FROM subdomains WHERE company_id = $1 ORDER BY created_at DESC`, companyId)
	return list, err
}

func (r subdomainRepositoryDB) Create(tx *sqlx.Tx, s Subdomain) (int, error) {
	var id int
	err := tx.QueryRowx(`INSERT INTO subdomains (company_id, subdomain, status) VALUES ($1,$2,$3) RETURNING id`, s.CompanyId, s.Subdomain, s.Status).Scan(&id)
	return id, err
}

func (r subdomainRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`DELETE FROM subdomains WHERE id=$1`, id)
	return err
}

func (r subdomainRepositoryDB) BelongsToUser(id, userId int) (bool, error) {
	var exists bool
	err := r.db.QueryRowx(
		`SELECT EXISTS(SELECT 1 FROM subdomains s JOIN companies c ON s.company_id = c.id WHERE s.id = $1 AND c.user_id = $2)`,
		id, userId,
	).Scan(&exists)
	return exists, err
}
