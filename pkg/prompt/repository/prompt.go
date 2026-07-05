package repository

import "github.com/jmoiron/sqlx"

type Prompt struct {
	Id         int     `db:"id"`
	CompanyId  int     `db:"company_id"`
	CategoryId *int    `db:"category_id"`
	Title      string  `db:"title"`
	Content    string  `db:"content"`
	CreatedAt  string  `db:"created_at"`
}

type PromptRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(companyId int) ([]Prompt, error)
	Create(tx *sqlx.Tx, p Prompt) (int, error)
	GetById(id int) (*Prompt, error)
	Update(tx *sqlx.Tx, p Prompt) error
	Delete(tx *sqlx.Tx, id int) error
	BelongsToUser(id, userId int) (bool, error)
}
