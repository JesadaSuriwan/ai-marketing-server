package repository

import "github.com/jmoiron/sqlx"

type Prompt struct {
	Id        int     `db:"id"`
	CompanyId int     `db:"company_id"`
	TagId     *int    `db:"tag_id"`
	TagName   *string `db:"tag_name"`
	// Countries is a comma-joined list of ISO 3166-1 alpha-2 codes (e.g.
	// "TH,US") — read-only display data, populated via a correlated
	// subquery against prompt_countries. Write access is SetCountries.
	Countries string `db:"countries"`
	Title     string `db:"title"`
	Content   string `db:"content"`
	Active    bool   `db:"active"`
	CreatedAt string `db:"created_at"`
}

type PromptRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(companyId int) ([]Prompt, error)
	// GetAllAcrossCompanies returns only active prompts — the scheduler's
	// sweep source, so a paused prompt is never auto-run.
	GetAllAcrossCompanies() ([]Prompt, error)
	Create(tx *sqlx.Tx, p Prompt) (int, error)
	GetById(id int) (*Prompt, error)
	// Update is a partial update — a nil field is left untouched rather than
	// overwritten, so e.g. changing just the tag doesn't require
	// resending title/content (and can't accidentally blank them).
	Update(tx *sqlx.Tx, id int, tagId *int, title, content *string) error
	// SetCountries replaces the full set of countries a prompt is scoped to
	// (delete-then-insert — simplest correct approach for a small
	// multi-select set, and countryCodes is never large enough for that to
	// matter). A nil/empty slice clears every country.
	SetCountries(tx *sqlx.Tx, promptId int, countryCodes []string) error
	SetActive(tx *sqlx.Tx, id int, active bool) error
	// CountActive is the denominator prompt_limit enforcement checks against.
	CountActive(companyId int) (int, error)
	Delete(tx *sqlx.Tx, id int) error
	BelongsToUser(id, userId int) (bool, error)
}
