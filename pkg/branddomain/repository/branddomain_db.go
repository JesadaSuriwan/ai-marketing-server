package repository

import "github.com/jmoiron/sqlx"

type brandDomainRepositoryDB struct {
	db *sqlx.DB
}

func NewBrandDomainRepositoryDB(db *sqlx.DB) BrandDomainRepository {
	return brandDomainRepositoryDB{db}
}

func (r brandDomainRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r brandDomainRepositoryDB) GetAll(brandId int) ([]BrandDomain, error) {
	list := []BrandDomain{}
	err := r.db.Select(&list, `SELECT id, brand_id, domain, created_at FROM brand_domains WHERE brand_id = $1 ORDER BY created_at DESC`, brandId)
	return list, err
}

func (r brandDomainRepositoryDB) GetById(id int) (*BrandDomain, error) {
	d := BrandDomain{}
	err := r.db.Get(&d, `SELECT id, brand_id, domain, created_at FROM brand_domains WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r brandDomainRepositoryDB) Create(tx *sqlx.Tx, d BrandDomain) (int, error) {
	var id int
	err := tx.QueryRowx(`INSERT INTO brand_domains (brand_id, domain) VALUES ($1,$2) RETURNING id`, d.BrandId, d.Domain).Scan(&id)
	return id, err
}

func (r brandDomainRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`DELETE FROM brand_domains WHERE id=$1`, id)
	return err
}

func (r brandDomainRepositoryDB) GetCompanyIdByBrandId(brandId int) (int, error) {
	var companyId int
	err := r.db.Get(&companyId, `SELECT company_id FROM brands WHERE id = $1`, brandId)
	return companyId, err
}
