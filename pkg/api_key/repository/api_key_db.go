package repository

import "github.com/jmoiron/sqlx"

type apiKeyRepositoryDB struct {
	db *sqlx.DB
}

func NewApiKeyRepositoryDB(db *sqlx.DB) ApiKeyRepository {
	return apiKeyRepositoryDB{db}
}

func (r apiKeyRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r apiKeyRepositoryDB) GetAll(companyId int) ([]ApiKey, error) {
	list := []ApiKey{}
	err := r.db.Select(&list,
		`SELECT id, company_id, name, key_prefix, key_value, status, last_used_at, created_at
		 FROM api_keys
		 WHERE company_id = $1 AND status != 'Revoked'
		 ORDER BY created_at DESC`,
		companyId,
	)
	return list, err
}

func (r apiKeyRepositoryDB) Create(tx *sqlx.Tx, k ApiKey) (int, error) {
	var id int
	err := tx.QueryRowx(
		`INSERT INTO api_keys (company_id, name, key_prefix, key_value, status) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		k.CompanyId, k.Name, k.KeyPrefix, k.KeyValue, k.Status,
	).Scan(&id)
	return id, err
}

func (r apiKeyRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`UPDATE api_keys SET status='Revoked' WHERE id=$1`, id)
	return err
}

func (r apiKeyRepositoryDB) BelongsToUser(id, userId int) (bool, error) {
	var exists bool
	err := r.db.QueryRowx(
		`SELECT EXISTS(SELECT 1 FROM api_keys ak JOIN companies c ON ak.company_id = c.id WHERE ak.id = $1 AND (
			c.user_id = $2
			OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id = c.id AND cm.user_id = $2 AND cm.status = 'active')
		))`,
		id, userId,
	).Scan(&exists)
	return exists, err
}
