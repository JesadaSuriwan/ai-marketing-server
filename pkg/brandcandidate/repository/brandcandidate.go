package repository

import "github.com/jmoiron/sqlx"

type BrandCandidate struct {
	Id           int    `db:"id"`
	CompanyId    int    `db:"company_id"`
	Name         string `db:"name"`
	MentionCount int    `db:"mention_count"`
	Status       string `db:"status"`
	FirstSeen    string `db:"first_seen"`
	LastSeen     string `db:"last_seen"`
}

type BrandCandidateRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	// Upsert records a sighting of name for companyId — first time creates
	// a pending row, a later sighting bumps mention_count/last_seen. Once a
	// candidate has been resolved or dismissed, further sightings are a
	// no-op: it won't silently flip back to pending and re-appear.
	Upsert(tx *sqlx.Tx, companyId int, name string) error
	GetPending(companyId int) ([]BrandCandidate, error)
	GetById(id int) (*BrandCandidate, error)
	SetStatus(tx *sqlx.Tx, id int, status string) error
	BelongsToCompany(id, companyId int) (bool, error)
}
