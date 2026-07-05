package repository

import "github.com/jmoiron/sqlx"

type Member struct {
	Id        int    `db:"id"`
	CompanyId int    `db:"company_id"`
	Email     string `db:"email"`
	Name      string `db:"name"`
	Role      string `db:"role"`
	JoinedAt  string `db:"joined_at"`
}

type MemberRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(companyId int) ([]Member, error)
	Create(tx *sqlx.Tx, m Member) (int, error)
	Delete(tx *sqlx.Tx, id int) error
	BelongsToUser(id, userId int) (bool, error)
}
