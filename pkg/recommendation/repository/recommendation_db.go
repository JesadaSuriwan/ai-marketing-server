package repository

import "github.com/jmoiron/sqlx"

type recommendationRepositoryDB struct {
	db *sqlx.DB
}

func NewRecommendationRepositoryDB(db *sqlx.DB) RecommendationRepository {
	return recommendationRepositoryDB{db}
}

func (r recommendationRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

// Upsert inserts a recommendation, or refreshes the copy (why/steps/impact) of
// an existing one identified by (company_id, type, prompt_id, link) — the
// natural identity of "this opportunity". If the user already moved it to
// todo/archived, a re-generate leaves it alone rather than resetting it back
// to 'suggested'.
func (r recommendationRepositoryDB) Upsert(tx *sqlx.Tx, rec Recommendation) (int, error) {
	var id int
	query := `INSERT INTO recommendations (company_id, type, placement, impact, engine, text, link, link_text, why, steps, prompt_id, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'suggested')
		ON CONFLICT (company_id, type, prompt_id, link) DO UPDATE SET
			placement = EXCLUDED.placement,
			impact = EXCLUDED.impact,
			engine = EXCLUDED.engine,
			text = EXCLUDED.text,
			link_text = EXCLUDED.link_text,
			why = EXCLUDED.why,
			steps = EXCLUDED.steps,
			updated_at = NOW()
		WHERE recommendations.status = 'suggested'
		RETURNING id`
	err := tx.QueryRowx(query, rec.CompanyId, rec.Type, rec.Placement, rec.Impact, rec.Engine, rec.Text, rec.Link, rec.LinkText, rec.Why, rec.Steps, rec.PromptId).Scan(&id)
	if err != nil {
		// Conflict existed but the WHERE clause excluded it (status != 'suggested') —
		// RETURNING produces no row rather than an error in that case, but some
		// drivers surface it as sql.ErrNoRows; treat as a non-fatal no-op.
		if err.Error() == "sql: no rows in result set" {
			return 0, nil
		}
		return 0, err
	}
	return id, nil
}

func (r recommendationRepositoryDB) GetByCompanyId(companyId int) ([]Recommendation, error) {
	list := []Recommendation{}
	query := `SELECT id, company_id, type, placement, impact, engine, text, link, link_text, why, steps, prompt_id, status, created_at, updated_at
		FROM recommendations WHERE company_id = $1 ORDER BY
		CASE impact WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END, created_at DESC`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r recommendationRepositoryDB) UpdateStatus(tx *sqlx.Tx, id int, status string) error {
	_, err := tx.Exec(`UPDATE recommendations SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

func (r recommendationRepositoryDB) BelongsToUser(id, userId int) (bool, error) {
	var owned bool
	err := r.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM recommendations r
			JOIN companies c ON r.company_id = c.id
			WHERE r.id = $1 AND (
				c.user_id = $2
				OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id = c.id AND cm.user_id = $2 AND cm.status = 'active')
			)
		)`, id, userId)
	return owned, err
}
