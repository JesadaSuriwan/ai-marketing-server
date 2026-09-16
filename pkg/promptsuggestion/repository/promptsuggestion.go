package repository

import "github.com/jmoiron/sqlx"

type PromptSuggestion struct {
	Id              int     `db:"id"`
	CompanyId       int     `db:"company_id"`
	Title           string  `db:"title"`
	Content         string  `db:"content"`
	Rationale       *string `db:"rationale"`
	Category        *string `db:"category"`
	Status          string  `db:"status"`
	CreatedPromptId *int    `db:"created_prompt_id"`
	CreatedAt       string  `db:"created_at"`
}

type PromptSuggestionRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	Create(tx *sqlx.Tx, ps PromptSuggestion) (int, error)
	GetById(id int) (*PromptSuggestion, error)
	GetByCompanyId(companyId int) ([]PromptSuggestion, error)
	UpdateStatus(tx *sqlx.Tx, id int, status string, createdPromptId *int) error
	BelongsToUser(id, userId int) (bool, error)
}
