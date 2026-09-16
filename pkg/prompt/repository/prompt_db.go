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
	query := `SELECT p.id, p.company_id, p.tag_id, pc.name AS tag_name, p.title, p.content, p.active, p.created_at,
			COALESCE((SELECT STRING_AGG(pco.country_code, ',' ORDER BY pco.country_code) FROM prompt_countries pco WHERE pco.prompt_id = p.id), '') AS countries
		FROM prompts p
		LEFT JOIN prompt_categories pc ON pc.id = p.tag_id
		WHERE p.company_id = $1 ORDER BY p.created_at DESC`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

// GetAllAcrossCompanies returns every ACTIVE prompt for every company — used
// by the scheduler, which has no single company_id to scope to. Inactive
// prompts are paused and never picked up here.
func (r promptRepositoryDB) GetAllAcrossCompanies() ([]Prompt, error) {
	list := []Prompt{}
	err := r.db.Select(&list, `SELECT id, company_id, tag_id, title, content, active, created_at FROM prompts WHERE active = TRUE ORDER BY id ASC`)
	return list, err
}

func (r promptRepositoryDB) Create(tx *sqlx.Tx, p Prompt) (int, error) {
	var id int
	err := tx.QueryRowx(`INSERT INTO prompts (company_id, tag_id, title, content) VALUES ($1,$2,$3,$4) RETURNING id`, p.CompanyId, p.TagId, p.Title, p.Content).Scan(&id)
	return id, err
}

func (r promptRepositoryDB) GetById(id int) (*Prompt, error) {
	p := Prompt{}
	query := `SELECT p.id, p.company_id, p.tag_id, pc.name AS tag_name, p.title, p.content, p.active, p.created_at,
			COALESCE((SELECT STRING_AGG(pco.country_code, ',' ORDER BY pco.country_code) FROM prompt_countries pco WHERE pco.prompt_id = p.id), '') AS countries
		FROM prompts p
		LEFT JOIN prompt_categories pc ON pc.id = p.tag_id
		WHERE p.id = $1`
	err := r.db.Get(&p, query, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r promptRepositoryDB) Update(tx *sqlx.Tx, id int, tagId *int, title, content *string) error {
	_, err := tx.Exec(
		`UPDATE prompts SET tag_id=COALESCE($1, tag_id), title=COALESCE($2, title), content=COALESCE($3, content) WHERE id=$4`,
		tagId, title, content, id,
	)
	return err
}

func (r promptRepositoryDB) SetCountries(tx *sqlx.Tx, promptId int, countryCodes []string) error {
	if _, err := tx.Exec(`DELETE FROM prompt_countries WHERE prompt_id = $1`, promptId); err != nil {
		return err
	}
	for _, code := range countryCodes {
		if code == "" {
			continue
		}
		if _, err := tx.Exec(
			`INSERT INTO prompt_countries (prompt_id, country_code) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			promptId, code,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r promptRepositoryDB) SetActive(tx *sqlx.Tx, id int, active bool) error {
	_, err := tx.Exec(`UPDATE prompts SET active=$1 WHERE id=$2`, active, id)
	return err
}

func (r promptRepositoryDB) CountActive(companyId int) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM prompts WHERE company_id = $1 AND active = TRUE`, companyId)
	return count, err
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
			WHERE p.id = $1 AND (
				c.user_id = $2
				OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id = c.id AND cm.user_id = $2 AND cm.status = 'active')
			)
		)`, id, userId)
	return owned, err
}
