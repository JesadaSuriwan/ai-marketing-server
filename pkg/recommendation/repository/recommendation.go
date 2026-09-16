package repository

import "github.com/jmoiron/sqlx"

type Recommendation struct {
	Id        int     `db:"id"`
	CompanyId int     `db:"company_id"`
	Type      string  `db:"type"`
	Placement string  `db:"placement"`
	Impact    string  `db:"impact"`
	Engine    string  `db:"engine"`
	Text      string  `db:"text"`
	Link      *string `db:"link"`
	LinkText  *string `db:"link_text"`
	Why       string  `db:"why"`
	Steps     string  `db:"steps"`
	PromptId  *int    `db:"prompt_id"`
	Status    string  `db:"status"`
	CreatedAt string  `db:"created_at"`
	UpdatedAt string  `db:"updated_at"`
}

type RecommendationRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	Upsert(tx *sqlx.Tx, r Recommendation) (int, error)
	GetByCompanyId(companyId int) ([]Recommendation, error)
	UpdateStatus(tx *sqlx.Tx, id int, status string) error
	BelongsToUser(id, userId int) (bool, error)
}
