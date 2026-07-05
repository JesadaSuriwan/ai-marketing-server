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

type AuthRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	CreateUser(tx *sqlx.Tx, u User) (int, error)
	GetUserByEmail(email string) (*User, error)
	GetUserById(id int) (*User, error)
}
