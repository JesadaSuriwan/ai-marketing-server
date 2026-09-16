package repository

import (
	"github.com/jmoiron/sqlx"
)

type Member struct {
	Id        int    `db:"id"`
	CompanyId int    `db:"company_id"`
	Email     string `db:"email"`
	Name      string `db:"name"`
	Role      string `db:"role"`
	JoinedAt  string `db:"joined_at"`
	UserId    *int   `db:"user_id"`
	Status    string `db:"status"`
}

type MemberRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(companyId int) ([]Member, error)
	GetById(id int) (*Member, error)
	Create(tx *sqlx.Tx, m Member) (int, error)
	Delete(tx *sqlx.Tx, id int) error
}
