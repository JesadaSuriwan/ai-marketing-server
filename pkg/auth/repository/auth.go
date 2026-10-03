package repository

import "github.com/jmoiron/sqlx"

type User struct {
	Id                 int    `db:"id"`
	Email              string `db:"email"`
	Password           string `db:"password"`
	Name               string `db:"name"`
	Initials           string `db:"initials"`
	MustChangePassword bool   `db:"must_change_password"`
	// EncryptedPassword is NULL for everyone except Customer role members —
	// see the encrypted_password column comment in the migration.
	EncryptedPassword *string `db:"encrypted_password"`
	CreatedAt         string  `db:"created_at"`
	UpdatedAt         string  `db:"updated_at"`
}

type AuthRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	CreateUser(tx *sqlx.Tx, u User) (int, error)
	GetUserByEmail(email string) (*User, error)
	GetUserById(id int) (*User, error)
	UpdatePassword(tx *sqlx.Tx, userId int, hashedPassword string) error
	// UpdateEncryptedPassword keeps the admin-viewable copy in sync — called
	// alongside UpdatePassword (same transaction) by every path that changes
	// a Customer role member's password, whether that's an Admin/Team Lead
	// setting it directly or the member changing it themselves.
	UpdateEncryptedPassword(tx *sqlx.Tx, userId int, encryptedPassword string) error
}
