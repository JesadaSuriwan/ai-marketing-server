package repository

import "github.com/jmoiron/sqlx"

type authRepositoryDB struct {
	db *sqlx.DB
}

func NewAuthRepositoryDB(db *sqlx.DB) AuthRepository {
	return authRepositoryDB{db}
}

func (r authRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r authRepositoryDB) CreateUser(tx *sqlx.Tx, u User) (int, error) {
	var id int
	query := `INSERT INTO users (email, password, name, initials) VALUES ($1, $2, $3, $4) RETURNING id`
	err := tx.QueryRowx(query, u.Email, u.Password, u.Name, u.Initials).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r authRepositoryDB) GetUserByEmail(email string) (*User, error) {
	user := User{}
	query := `SELECT id, email, password, name, initials, created_at, updated_at FROM users WHERE email = $1`
	err := r.db.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r authRepositoryDB) GetUserById(id int) (*User, error) {
	user := User{}
	query := `SELECT id, email, password, name, initials, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
