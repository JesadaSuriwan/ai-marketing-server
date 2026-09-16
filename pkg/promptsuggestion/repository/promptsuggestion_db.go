package repository

import "github.com/jmoiron/sqlx"

type promptSuggestionRepositoryDB struct {
	db *sqlx.DB
}

func NewPromptSuggestionRepositoryDB(db *sqlx.DB) PromptSuggestionRepository {
	return promptSuggestionRepositoryDB{db}
}

func (r promptSuggestionRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r promptSuggestionRepositoryDB) Create(tx *sqlx.Tx, ps PromptSuggestion) (int, error) {
	var id int
	err := tx.QueryRowx(
		`INSERT INTO prompt_suggestions (company_id, title, content, rationale, category) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		ps.CompanyId, ps.Title, ps.Content, ps.Rationale, ps.Category,
	).Scan(&id)
	return id, err
}

func (r promptSuggestionRepositoryDB) GetById(id int) (*PromptSuggestion, error) {
	ps := PromptSuggestion{}
	err := r.db.Get(&ps, `SELECT id, company_id, title, content, rationale, category, status, created_prompt_id, created_at FROM prompt_suggestions WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &ps, nil
}

func (r promptSuggestionRepositoryDB) GetByCompanyId(companyId int) ([]PromptSuggestion, error) {
	list := []PromptSuggestion{}
	err := r.db.Select(&list, `SELECT id, company_id, title, content, rationale, category, status, created_prompt_id, created_at FROM prompt_suggestions WHERE company_id = $1 ORDER BY created_at DESC`, companyId)
	return list, err
}

func (r promptSuggestionRepositoryDB) UpdateStatus(tx *sqlx.Tx, id int, status string, createdPromptId *int) error {
	_, err := tx.Exec(`UPDATE prompt_suggestions SET status=$1, created_prompt_id=$2 WHERE id=$3`, status, createdPromptId, id)
	return err
}

func (r promptSuggestionRepositoryDB) BelongsToUser(id, userId int) (bool, error) {
	var owned bool
	err := r.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM prompt_suggestions ps
			JOIN companies c ON ps.company_id = c.id
			WHERE ps.id = $1 AND (
				c.user_id = $2
				OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id = c.id AND cm.user_id = $2 AND cm.status = 'active')
			)
		)`, id, userId)
	return owned, err
}
