package repository

import "github.com/jmoiron/sqlx"

type promptRepositoryDB struct {
	db *sqlx.DB
}

func NewPromptRepositoryDB(db *sqlx.DB) PromptRepository {
	return promptRepositoryDB{db}
}

func (r promptRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r promptRepositoryDB) GetAll(companyId int) ([]Prompt, error) {
	list := []Prompt{}
	err := r.db.Select(&list, `SELECT id, company_id, category_id, title, content, created_at FROM prompts WHERE company_id = $1 ORDER BY created_at DESC`, companyId)
	return list, err
}

func (r promptRepositoryDB) Create(tx *sqlx.Tx, p Prompt) (int, error) {
	var id int
	err := tx.QueryRowx(`INSERT INTO prompts (company_id, category_id, title, content) VALUES ($1,$2,$3,$4) RETURNING id`, p.CompanyId, p.CategoryId, p.Title, p.Content).Scan(&id)
	return id, err
}

func (r promptRepositoryDB) GetById(id int) (*Prompt, error) {
	p := Prompt{}
	err := r.db.Get(&p, `SELECT id, company_id, category_id, title, content, created_at FROM prompts WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r promptRepositoryDB) Update(tx *sqlx.Tx, p Prompt) error {
	_, err := tx.Exec(`UPDATE prompts SET category_id=$1, title=$2, content=$3 WHERE id=$4`, p.CategoryId, p.Title, p.Content, p.Id)
	return err
}

func (r promptRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`DELETE FROM prompts WHERE id = $1`, id)
	return err
}

func (r promptRepositoryDB) BelongsToUser(id, userId int) (bool, error) {
	var owned bool
	err := r.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM prompts p
			JOIN companies c ON p.company_id = c.id
			WHERE p.id = $1 AND c.user_id = $2
		)`, id, userId)
	return owned, err
}
