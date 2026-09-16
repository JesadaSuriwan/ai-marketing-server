package repository

import "github.com/jmoiron/sqlx"

type brandRepositoryDB struct {
	db *sqlx.DB
}

func NewBrandRepositoryDB(db *sqlx.DB) BrandRepository {
	return brandRepositoryDB{db}
}

func (r brandRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r brandRepositoryDB) GetAll(companyId int) ([]Brand, error) {
	list := []Brand{}
	query := `SELECT id, company_id, name, domain, industry, description, status, is_own, logo_url, created_at FROM brands WHERE company_id = $1 ORDER BY created_at DESC`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r brandRepositoryDB) Create(tx *sqlx.Tx, b Brand) (int, error) {
	var id int
	query := `INSERT INTO brands (company_id, name, domain, industry, description, status, is_own, logo_url) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`
	err := tx.QueryRowx(query, b.CompanyId, b.Name, b.Domain, b.Industry, b.Description, b.Status, b.IsOwn, b.LogoUrl).Scan(&id)
	return id, err
}

func (r brandRepositoryDB) GetById(id int) (*Brand, error) {
	b := Brand{}
	query := `SELECT id, company_id, name, domain, industry, description, status, is_own, logo_url, created_at FROM brands WHERE id = $1`
	err := r.db.Get(&b, query, id)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r brandRepositoryDB) Update(tx *sqlx.Tx, b Brand) error {
	query := `UPDATE brands SET name=$1, domain=$2, industry=$3, description=$4, status=$5, is_own=$6, logo_url=$7 WHERE id=$8`
	_, err := tx.Exec(query, b.Name, b.Domain, b.Industry, b.Description, b.Status, b.IsOwn, b.LogoUrl, b.Id)
	return err
}

func (r brandRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`DELETE FROM brands WHERE id = $1`, id)
	return err
}

func (r brandRepositoryDB) UpdateStatus(tx *sqlx.Tx, id int, status string) error {
	_, err := tx.Exec(`UPDATE brands SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (r brandRepositoryDB) ClearOwn(tx *sqlx.Tx, companyId, exceptId int) error {
	_, err := tx.Exec(`UPDATE brands SET is_own = FALSE WHERE company_id = $1 AND id != $2`, companyId, exceptId)
	return err
}

func (r brandRepositoryDB) BelongsToUser(id, userId int) (bool, error) {
	var owned bool
	err := r.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM brands b
			JOIN companies c ON b.company_id = c.id
			WHERE b.id = $1 AND (
				c.user_id = $2
				OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id = c.id AND cm.user_id = $2 AND cm.status = 'active')
			)
		)`, id, userId)
	return owned, err
}
