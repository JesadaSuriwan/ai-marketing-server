package repository

import "github.com/jmoiron/sqlx"

type Citation struct {
	Id                int     `db:"id"`
	PromptId          int     `db:"prompt_id"`
	BrandId           *int    `db:"brand_id"`
	PromptRunId       *int    `db:"prompt_run_id"`
	Ranking           int     `db:"ranking"`
	Content           string  `db:"content"`
	Url               *string `db:"url"`
	AiPlatform        *string `db:"ai_platform"`
	DateDiscovered    *string `db:"date_discovered"`
	LastChecked       *string `db:"last_checked"`
	Sentiment         string  `db:"sentiment"`
	BrandPositioning  string  `db:"brand_positioning"`
	CitationFrequency int     `db:"citation_frequency"`
	Snippet           *string `db:"snippet"`
	IsCompetitor      bool    `db:"is_competitor"`
	Notes             *string `db:"notes"`
	IsArchived        bool    `db:"is_archived"`
	// SourceType classifies the citing domain: editorial | directory |
	// reference | social | news | marketplace | official_site | other.
	SourceType *string `db:"source_type"`
	// TargetCountry is an ISO 3166-1 alpha-2 code for the market this URL
	// targets, or nil when undetermined — see detectCountryFromTLD in
	// pkg/promptrun/service for how it's derived.
	TargetCountry *string `db:"target_country"`
	CreatedAt     string  `db:"created_at"`
	// BrandName is joined in from brands.name — read-only, only populated by
	// GetAll/GetById, never written by Create/Upsert/Update.
	BrandName *string `db:"brand_name"`
}

type CitationCompetitor struct {
	Id             int    `db:"id"`
	CitationId     int    `db:"citation_id"`
	CompetitorName string `db:"competitor_name"`
}

type CitationRankingHistory struct {
	Id         int    `db:"id"`
	CitationId int    `db:"citation_id"`
	DateLabel  string `db:"date_label"`
	Rank       int    `db:"rank"`
}

// CitationUrlDailyStat is one (company, url, day) bucket, incremented once
// per citation event — the real, non-fabricated basis for period-over-period
// comparisons like Top Winners/Losers.
type CitationUrlDailyStat struct {
	Id            int    `db:"id"`
	CompanyId     int    `db:"company_id"`
	Url           string `db:"url"`
	StatDate      string `db:"stat_date"`
	CitationCount int    `db:"citation_count"`
}

type CitationRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(promptId int, from, to string) ([]Citation, error)
	Create(tx *sqlx.Tx, c Citation) (int, error)
	Upsert(tx *sqlx.Tx, c Citation) (int, error)
	GetById(id int) (*Citation, error)
	GetCompetitors(citationId int) ([]CitationCompetitor, error)
	GetRankingHistory(citationId int) ([]CitationRankingHistory, error)
	Update(tx *sqlx.Tx, c Citation) error
	Delete(tx *sqlx.Tx, id int) error
	UpdateNotes(tx *sqlx.Tx, id int, notes string) error
	UpdateArchive(tx *sqlx.Tx, id int, isArchived bool) error
	BelongsToUser(id, userId int) (bool, error)
	BumpDailyStat(tx *sqlx.Tx, companyId int, url, date string) error
}
