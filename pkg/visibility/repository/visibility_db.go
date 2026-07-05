package repository

import (
	"strconv"

	"github.com/jmoiron/sqlx"
)

type visibilityRepositoryDB struct {
	db *sqlx.DB
}

func NewVisibilityRepositoryDB(db *sqlx.DB) VisibilityRepository {
	return visibilityRepositoryDB{db}
}

func (r visibilityRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r visibilityRepositoryDB) GetAll(brandId int, from, to string) ([]Visibility, error) {
	list := []Visibility{}
	query := `SELECT id, brand_id, prompt_id, platform, score, mentions, date, created_at FROM visibility_data WHERE brand_id = $1`
	args := []interface{}{brandId}
	if from != "" {
		args = append(args, from)
		query += ` AND date >= $` + strconv.Itoa(len(args))
	}
	if to != "" {
		args = append(args, to)
		query += ` AND date <= $` + strconv.Itoa(len(args))
	}
	query += ` ORDER BY date DESC`
	err := r.db.Select(&list, query, args...)
	return list, err
}

func (r visibilityRepositoryDB) Create(tx *sqlx.Tx, v Visibility) (int, error) {
	var id int
	query := `INSERT INTO visibility_data (brand_id, prompt_id, platform, score, mentions, date) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`
	err := tx.QueryRowx(query, v.BrandId, v.PromptId, v.Platform, v.Score, v.Mentions, v.Date).Scan(&id)
	return id, err
}

func (r visibilityRepositoryDB) GetBreakdown(brandId int) ([]PlatformBreakdown, error) {
	list := []PlatformBreakdown{}
	query := `SELECT platform, AVG(score) as score, SUM(mentions) as mentions FROM visibility_data WHERE brand_id = $1 GROUP BY platform`
	err := r.db.Select(&list, query, brandId)
	return list, err
}
