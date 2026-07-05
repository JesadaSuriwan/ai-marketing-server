package repository

import "github.com/jmoiron/sqlx"

type Citation struct {
	Id                int     `db:"id"`
	PromptId          int     `db:"prompt_id"`
	BrandId           *int    `db:"brand_id"`
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
	CreatedAt         string  `db:"created_at"`
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

type CitationRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(promptId int) ([]Citation, error)
	Create(tx *sqlx.Tx, c Citation) (int, error)
	GetById(id int) (*Citation, error)
	GetCompetitors(citationId int) ([]CitationCompetitor, error)
	GetRankingHistory(citationId int) ([]CitationRankingHistory, error)
	Update(tx *sqlx.Tx, c Citation) error
	Delete(tx *sqlx.Tx, id int) error
	UpdateNotes(tx *sqlx.Tx, id int, notes string) error
	UpdateArchive(tx *sqlx.Tx, id int, isArchived bool) error
	BelongsToUser(id, userId int) (bool, error)
}
