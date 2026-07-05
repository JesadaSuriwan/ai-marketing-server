package repository

import "github.com/jmoiron/sqlx"

type categoryRepositoryDB struct {
	db *sqlx.DB
}

func NewCategoryRepositoryDB(db *sqlx.DB) CategoryRepository {
	return categoryRepositoryDB{db}
}

func (r categoryRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r categoryRepositoryDB) GetAll(companyId int) ([]Category, error) {
	list := []Category{}
	err := r.db.Select(&list, `SELECT id, company_id, name, created_at FROM prompt_categories WHERE company_id = $1 ORDER BY created_at DESC`, companyId)
	return list, err
}

func (r categoryRepositoryDB) Create(tx *sqlx.Tx, c Category) (int, error) {
	var id int
	err := tx.QueryRowx(`INSERT INTO prompt_categories (company_id, name) VALUES ($1,$2) RETURNING id`, c.CompanyId, c.Name).Scan(&id)
	return id, err
}

func (r categoryRepositoryDB) Update(tx *sqlx.Tx, id int, name string) error {
	_, err := tx.Exec(`UPDATE prompt_categories SET name=$1 WHERE id=$2`, name, id)
	return err
}

func (r categoryRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`DELETE FROM prompt_categories WHERE id = $1`, id)
	return err
}

func (r categoryRepositoryDB) BelongsToUser(id, userId int) (bool, error) {
	var owned bool
	err := r.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM prompt_categories pc
			JOIN companies c ON pc.company_id = c.id
			WHERE pc.id = $1 AND c.user_id = $2
		)`, id, userId)
	return owned, err
}
