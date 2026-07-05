package repository

import "github.com/jmoiron/sqlx"

type companyRepositoryDB struct {
	db *sqlx.DB
}

func NewCompanyRepositoryDB(db *sqlx.DB) CompanyRepository {
	return companyRepositoryDB{db}
}

func (r companyRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r companyRepositoryDB) Create(tx *sqlx.Tx, c Company) (int, error) {
	var id int
	query := `INSERT INTO companies (user_id, name, initials, logo_url, industry, website, location, size, founded, phone, email, about, plan)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id`
	err := tx.QueryRowx(query, c.UserId, c.Name, c.Initials, c.LogoUrl, c.Industry, c.Website, c.Location, c.Size, c.Founded, c.Phone, c.Email, c.About, c.Plan).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r companyRepositoryDB) GetAll(userId int) ([]Company, error) {
	list := []Company{}
	query := `SELECT id, user_id, name, initials, logo_url, industry, website, location, size, founded, phone, email, about, plan, created_at, updated_at FROM companies WHERE user_id = $1 ORDER BY created_at DESC`
	err := r.db.Select(&list, query, userId)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r companyRepositoryDB) GetById(id int) (*Company, error) {
	c := Company{}
	query := `SELECT id, user_id, name, initials, logo_url, industry, website, location, size, founded, phone, email, about, plan, created_at, updated_at FROM companies WHERE id = $1`
	err := r.db.Get(&c, query, id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r companyRepositoryDB) Update(tx *sqlx.Tx, c Company) error {
	query := `UPDATE companies SET name=$1, initials=$2, logo_url=$3, industry=$4, website=$5, location=$6, size=$7, founded=$8, phone=$9, email=$10, about=$11, plan=$12, updated_at=NOW() WHERE id=$13`
	_, err := tx.Exec(query, c.Name, c.Initials, c.LogoUrl, c.Industry, c.Website, c.Location, c.Size, c.Founded, c.Phone, c.Email, c.About, c.Plan, c.Id)
	return err
}

func (r companyRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`DELETE FROM companies WHERE id = $1`, id)
	return err
}

func (r companyRepositoryDB) GetTags(companyId int) ([]CompanyTag, error) {
	list := []CompanyTag{}
	err := r.db.Select(&list, `SELECT id, company_id, tag FROM company_tags WHERE company_id = $1`, companyId)
	return list, err
}

func (r companyRepositoryDB) AddTag(tx *sqlx.Tx, companyId int, tag string) error {
	_, err := tx.Exec(`INSERT INTO company_tags (company_id, tag) VALUES ($1, $2)`, companyId, tag)
	return err
}

func (r companyRepositoryDB) DeleteTag(tx *sqlx.Tx, companyId int, tag string) error {
	_, err := tx.Exec(`DELETE FROM company_tags WHERE company_id = $1 AND tag = $2`, companyId, tag)
	return err
}

func (r companyRepositoryDB) GetMarkets(companyId int) ([]CompanyMarket, error) {
	list := []CompanyMarket{}
	err := r.db.Select(&list, `SELECT id, company_id, market FROM company_markets WHERE company_id = $1`, companyId)
	return list, err
}

func (r companyRepositoryDB) AddMarket(tx *sqlx.Tx, companyId int, market string) error {
	_, err := tx.Exec(`INSERT INTO company_markets (company_id, market) VALUES ($1, $2)`, companyId, market)
	return err
}

func (r companyRepositoryDB) DeleteMarket(tx *sqlx.Tx, companyId int, market string) error {
	_, err := tx.Exec(`DELETE FROM company_markets WHERE company_id = $1 AND market = $2`, companyId, market)
	return err
}

func (r companyRepositoryDB) GetSocialLinks(companyId int) (*CompanySocialLinks, error) {
	sl := CompanySocialLinks{}
	err := r.db.Get(&sl, `SELECT id, company_id, linkedin, twitter, facebook FROM company_social_links WHERE company_id = $1`, companyId)
	if err != nil {
		return nil, err
	}
	return &sl, nil
}

func (r companyRepositoryDB) UpsertSocialLinks(tx *sqlx.Tx, companyId int, linkedin, twitter, facebook *string) error {
	query := `INSERT INTO company_social_links (company_id, linkedin, twitter, facebook) VALUES ($1,$2,$3,$4)
		ON CONFLICT (company_id) DO UPDATE SET linkedin=$2, twitter=$3, facebook=$4`
	_, err := tx.Exec(query, companyId, linkedin, twitter, facebook)
	return err
}

func (r companyRepositoryDB) GetKeyContacts(companyId int) ([]CompanyKeyContact, error) {
	list := []CompanyKeyContact{}
	err := r.db.Select(&list, `SELECT id, company_id, name, role, email, phone FROM company_key_contacts WHERE company_id = $1`, companyId)
	return list, err
}

func (r companyRepositoryDB) AddKeyContact(tx *sqlx.Tx, contact CompanyKeyContact) (int, error) {
	var id int
	query := `INSERT INTO company_key_contacts (company_id, name, role, email, phone) VALUES ($1,$2,$3,$4,$5) RETURNING id`
	err := tx.QueryRowx(query, contact.CompanyId, contact.Name, contact.Role, contact.Email, contact.Phone).Scan(&id)
	return id, err
}

func (r companyRepositoryDB) DeleteKeyContact(tx *sqlx.Tx, companyId, contactId int) error {
	_, err := tx.Exec(`DELETE FROM company_key_contacts WHERE id = $1 AND company_id = $2`, contactId, companyId)
	return err
}
