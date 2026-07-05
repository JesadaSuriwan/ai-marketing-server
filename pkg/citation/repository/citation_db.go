package repository

import "github.com/jmoiron/sqlx"

type citationRepositoryDB struct {
	db *sqlx.DB
}

func NewCitationRepositoryDB(db *sqlx.DB) CitationRepository {
	return citationRepositoryDB{db}
}

func (r citationRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r citationRepositoryDB) GetAll(promptId int) ([]Citation, error) {
	list := []Citation{}
	query := `SELECT id, prompt_id, brand_id, ranking, content, url, ai_platform, date_discovered, last_checked, sentiment, brand_positioning, citation_frequency, snippet, is_competitor, notes, is_archived, created_at FROM citations WHERE prompt_id = $1 ORDER BY ranking ASC`
	err := r.db.Select(&list, query, promptId)
	return list, err
}

func (r citationRepositoryDB) Create(tx *sqlx.Tx, c Citation) (int, error) {
	var id int
	query := `INSERT INTO citations (prompt_id, brand_id, ranking, content, url, ai_platform, date_discovered, last_checked, sentiment, brand_positioning, citation_frequency, snippet, is_competitor, notes, is_archived)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id`
	err := tx.QueryRowx(query, c.PromptId, c.BrandId, c.Ranking, c.Content, c.Url, c.AiPlatform, c.DateDiscovered, c.LastChecked, c.Sentiment, c.BrandPositioning, c.CitationFrequency, c.Snippet, c.IsCompetitor, c.Notes, c.IsArchived).Scan(&id)
	return id, err
}

func (r citationRepositoryDB) GetById(id int) (*Citation, error) {
	c := Citation{}
	query := `SELECT id, prompt_id, brand_id, ranking, content, url, ai_platform, date_discovered, last_checked, sentiment, brand_positioning, citation_frequency, snippet, is_competitor, notes, is_archived, created_at FROM citations WHERE id = $1`
	err := r.db.Get(&c, query, id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r citationRepositoryDB) GetCompetitors(citationId int) ([]CitationCompetitor, error) {
	list := []CitationCompetitor{}
	err := r.db.Select(&list, `SELECT id, citation_id, competitor_name FROM citation_competitors WHERE citation_id = $1`, citationId)
	return list, err
}

func (r citationRepositoryDB) GetRankingHistory(citationId int) ([]CitationRankingHistory, error) {
	list := []CitationRankingHistory{}
	err := r.db.Select(&list, `SELECT id, citation_id, date_label, rank FROM citation_ranking_history WHERE citation_id = $1 ORDER BY id ASC`, citationId)
	return list, err
}

func (r citationRepositoryDB) Update(tx *sqlx.Tx, c Citation) error {
	query := `UPDATE citations SET brand_id=$1, ranking=$2, content=$3, url=$4, ai_platform=$5, date_discovered=$6, last_checked=$7, sentiment=$8, brand_positioning=$9, citation_frequency=$10, snippet=$11, is_competitor=$12, notes=$13, is_archived=$14 WHERE id=$15`
	_, err := tx.Exec(query, c.BrandId, c.Ranking, c.Content, c.Url, c.AiPlatform, c.DateDiscovered, c.LastChecked, c.Sentiment, c.BrandPositioning, c.CitationFrequency, c.Snippet, c.IsCompetitor, c.Notes, c.IsArchived, c.Id)
	return err
}

func (r citationRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`DELETE FROM citations WHERE id = $1`, id)
	return err
}

func (r citationRepositoryDB) UpdateNotes(tx *sqlx.Tx, id int, notes string) error {
	_, err := tx.Exec(`UPDATE citations SET notes=$1 WHERE id=$2`, notes, id)
	return err
}

func (r citationRepositoryDB) UpdateArchive(tx *sqlx.Tx, id int, isArchived bool) error {
	_, err := tx.Exec(`UPDATE citations SET is_archived=$1 WHERE id=$2`, isArchived, id)
	return err
}

func (r citationRepositoryDB) BelongsToUser(id, userId int) (bool, error) {
	var owned bool
	err := r.db.Get(&owned, `
		SELECT EXISTS(
			SELECT 1 FROM citations ci
			JOIN prompts p ON ci.prompt_id = p.id
			JOIN companies c ON p.company_id = c.id
			WHERE ci.id = $1 AND c.user_id = $2
		)`, id, userId)
	return owned, err
}
