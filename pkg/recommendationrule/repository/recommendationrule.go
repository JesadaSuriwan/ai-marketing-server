package repository

import "github.com/jmoiron/sqlx"

type RecommendationRule struct {
	Id        int    `db:"id"`
	CompanyId int    `db:"company_id"`
	Text      string `db:"text"`
	Active    bool   `db:"active"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type RecommendationRuleRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(companyId int) ([]RecommendationRule, error)
	// GetActiveTexts returns just the text of every active rule — the only
	// shape recommendation generation actually needs.
	GetActiveTexts(companyId int) ([]string, error)
	Create(tx *sqlx.Tx, r RecommendationRule) (int, error)
	// Update is a partial update — a nil field is left untouched, so toggling
	// active doesn't require resending text (and vice versa).
	Update(tx *sqlx.Tx, id int, text *string, active *bool) error
	Delete(tx *sqlx.Tx, id int) error
	BelongsToUser(id, userId int) (bool, error)
}
