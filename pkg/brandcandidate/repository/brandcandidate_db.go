package repository

import "github.com/jmoiron/sqlx"

type brandCandidateRepositoryDB struct {
	db *sqlx.DB
}

func NewBrandCandidateRepositoryDB(db *sqlx.DB) BrandCandidateRepository {
	return brandCandidateRepositoryDB{db}
}

func (r brandCandidateRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r brandCandidateRepositoryDB) Upsert(tx *sqlx.Tx, companyId int, name string) error {
	_, err := tx.Exec(`
		INSERT INTO brand_candidates (company_id, name) VALUES ($1, $2)
		ON CONFLICT (company_id, name) DO UPDATE SET
			mention_count = brand_candidates.mention_count + 1,
			last_seen = NOW()
		WHERE brand_candidates.status = 'pending'`, companyId, name)
	return err
}

func (r brandCandidateRepositoryDB) GetPending(companyId int) ([]BrandCandidate, error) {
	list := []BrandCandidate{}
	query := `SELECT id, company_id, name, mention_count, status, first_seen, last_seen
		FROM brand_candidates WHERE company_id = $1 AND status = 'pending' ORDER BY mention_count DESC, last_seen DESC`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r brandCandidateRepositoryDB) GetById(id int) (*BrandCandidate, error) {
	c := BrandCandidate{}
	query := `SELECT id, company_id, name, mention_count, status, first_seen, last_seen FROM brand_candidates WHERE id = $1`
	if err := r.db.Get(&c, query, id); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r brandCandidateRepositoryDB) SetStatus(tx *sqlx.Tx, id int, status string) error {
	_, err := tx.Exec(`UPDATE brand_candidates SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r brandCandidateRepositoryDB) BelongsToCompany(id, companyId int) (bool, error) {
	var owned bool
	err := r.db.Get(&owned, `SELECT EXISTS(SELECT 1 FROM brand_candidates WHERE id = $1 AND company_id = $2)`, id, companyId)
	return owned, err
}
