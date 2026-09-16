package repository

import "github.com/jmoiron/sqlx"

type recommendationRuleRepositoryDB struct {
	db *sqlx.DB
}

func NewRecommendationRuleRepositoryDB(db *sqlx.DB) RecommendationRuleRepository {
	return recommendationRuleRepositoryDB{db}
}

func (r recommendationRuleRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r recommendationRuleRepositoryDB) GetAll(companyId int) ([]RecommendationRule, error) {
	list := []RecommendationRule{}
	err := r.db.Select(&list, `SELECT id, company_id, text, active, created_at, updated_at FROM recommendation_rules WHERE company_id = $1 ORDER BY created_at DESC`, companyId)
	return list, err
}

func (r recommendationRuleRepositoryDB) GetActiveTexts(companyId int) ([]string, error) {
	list := []string{}
	err := r.db.Select(&list, `SELECT text FROM recommendation_rules WHERE company_id = $1 AND active = TRUE ORDER BY created_at ASC`, companyId)
	return list, err
}

func (r recommendationRuleRepositoryDB) Create(tx *sqlx.Tx, rule RecommendationRule) (int, error) {
	var id int
	err := tx.QueryRowx(`INSERT INTO recommendation_rules (company_id, text) VALUES ($1,$2) RETURNING id`, rule.CompanyId, rule.Text).Scan(&id)
	return id, err
}

func (r recommendationRuleRepositoryDB) Update(tx *sqlx.Tx, id int, text *string, active *bool) error {
	_, err := tx.Exec(
		`UPDATE recommendation_rules SET text=COALESCE($1, text), active=COALESCE($2, active), updated_at=NOW() WHERE id=$3`,
		text, active, id,
	)
	return err
}

func (r recommendationRuleRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`DELETE FROM recommendation_rules WHERE id = $1`, id)
	return err
}

func (r recommendationRuleRepositoryDB) BelongsToUser(id, userId int) (bool, error) {
	var owned bool
	err := r.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM recommendation_rules rr
			JOIN companies c ON rr.company_id = c.id
			WHERE rr.id = $1 AND (
				c.user_id = $2
				OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id = c.id AND cm.user_id = $2 AND cm.status = 'active')
			)
		)`, id, userId)
	return owned, err
}
