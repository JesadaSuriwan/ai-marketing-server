package repository

import "github.com/jmoiron/sqlx"

type Visibility struct {
	Id        int     `db:"id"`
	BrandId   int     `db:"brand_id"`
	PromptId  *int    `db:"prompt_id"`
	Platform  string  `db:"platform"`
	Score     float64 `db:"score"`
	Mentions  int     `db:"mentions"`
	Date      string  `db:"date"`
	CreatedAt string  `db:"created_at"`
}

type PlatformBreakdown struct {
	Platform string  `db:"platform"`
	Score    float64 `db:"score"`
	Mentions int     `db:"mentions"`
}

type VisibilityRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(brandId int, from, to string) ([]Visibility, error)
	Create(tx *sqlx.Tx, v Visibility) (int, error)
	GetBreakdown(brandId int) ([]PlatformBreakdown, error)
}
