package repository

import "github.com/jmoiron/sqlx"

type User struct {
	Id        int    `db:"id"`
	Email     string `db:"email"`
	Password  string `db:"password"`
	Name      string `db:"name"`
	Initials  string `db:"initials"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type UserRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetById(id int) (*User, error)
	Update(tx *sqlx.Tx, id int, name, initials string) error
}
