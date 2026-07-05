package repository

import "github.com/jmoiron/sqlx"

type ApiKey struct {
	Id         int     `db:"id"`
	CompanyId  int     `db:"company_id"`
	Name       string  `db:"name"`
	KeyPrefix  string  `db:"key_prefix"`
	KeyValue   string  `db:"key_value"`
	Status     string  `db:"status"`
	LastUsedAt *string `db:"last_used_at"`
	CreatedAt  string  `db:"created_at"`
}

type ApiKeyRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(companyId int) ([]ApiKey, error)
	Create(tx *sqlx.Tx, k ApiKey) (int, error)
	Delete(tx *sqlx.Tx, id int) error
	BelongsToUser(id, userId int) (bool, error)
}
