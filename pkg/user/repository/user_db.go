package repository

import "github.com/jmoiron/sqlx"

type userRepositoryDB struct {
	db *sqlx.DB
}

func NewUserRepositoryDB(db *sqlx.DB) UserRepository {
	return userRepositoryDB{db}
}

func (r userRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r userRepositoryDB) GetById(id int) (*User, error) {
	user := User{}
	query := `SELECT id, email, password, name, initials, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r userRepositoryDB) Update(tx *sqlx.Tx, id int, name, initials string) error {
	query := `UPDATE users SET name = $1, initials = $2, updated_at = NOW() WHERE id = $3`
	_, err := tx.Exec(query, name, initials, id)
	return err
}
