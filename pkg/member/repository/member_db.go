package repository

import "github.com/jmoiron/sqlx"

type memberRepositoryDB struct {
	db *sqlx.DB
}

func NewMemberRepositoryDB(db *sqlx.DB) MemberRepository {
	return memberRepositoryDB{db}
}

func (r memberRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r memberRepositoryDB) GetAll(companyId int) ([]Member, error) {
	list := []Member{}
	err := r.db.Select(&list, `SELECT id, company_id, email, name, role, joined_at, user_id, status FROM company_members WHERE company_id = $1 ORDER BY joined_at DESC`, companyId)
	return list, err
}

func (r memberRepositoryDB) GetById(id int) (*Member, error) {
	m := Member{}
	err := r.db.Get(&m, `SELECT id, company_id, email, name, role, joined_at, user_id, status FROM company_members WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r memberRepositoryDB) Create(tx *sqlx.Tx, m Member) (int, error) {
	var id int
	err := tx.QueryRowx(
		`INSERT INTO company_members (company_id, email, name, role, status, user_id)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		m.CompanyId, m.Email, m.Name, m.Role, m.Status, m.UserId,
	).Scan(&id)
	return id, err
}

func (r memberRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`DELETE FROM company_members WHERE id=$1`, id)
	return err
}
