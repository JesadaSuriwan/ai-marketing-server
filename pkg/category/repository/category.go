package repository

import "github.com/jmoiron/sqlx"

type Category struct {
	Id        int    `db:"id"`
	CompanyId int    `db:"company_id"`
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
}

type CategoryRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(companyId int) ([]Category, error)
	Create(tx *sqlx.Tx, c Category) (int, error)
	Update(tx *sqlx.Tx, id int, name string) error
	Delete(tx *sqlx.Tx, id int) error
	BelongsToUser(id, userId int) (bool, error)
}
