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

func (r citationRepositoryDB) GetAll(promptId int, from, to string) ([]Citation, error) {
	list := []Citation{}
	query := `SELECT c.id, c.prompt_id, c.brand_id, c.ranking, c.content, c.url, c.ai_platform, c.date_discovered, c.last_checked, c.sentiment, c.brand_positioning, c.citation_frequency, c.snippet, c.is_competitor, c.notes, c.is_archived, c.source_type, c.target_country, c.created_at, b.name AS brand_name
		FROM citations c
		LEFT JOIN brands b ON b.id = c.brand_id
		WHERE c.prompt_id = $1
		  AND ($2 = '' OR c.last_checked >= $2::date)
		  AND ($3 = '' OR c.last_checked <= $3::date)
		ORDER BY c.ranking ASC`
	err := r.db.Select(&list, query, promptId, from, to)
	return list, err
}

func (r citationRepositoryDB) Create(tx *sqlx.Tx, c Citation) (int, error) {
	var id int
	query := `INSERT INTO citations (prompt_id, brand_id, ranking, content, url, ai_platform, date_discovered, last_checked, sentiment, brand_positioning, citation_frequency, snippet, is_competitor, notes, is_archived, source_type, target_country)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) RETURNING id`
	err := tx.QueryRowx(query, c.PromptId, c.BrandId, c.Ranking, c.Content, c.Url, c.AiPlatform, c.DateDiscovered, c.LastChecked, c.Sentiment, c.BrandPositioning, c.CitationFrequency, c.Snippet, c.IsCompetitor, c.Notes, c.IsArchived, c.SourceType, c.TargetCountry).Scan(&id)
	return id, err
}

// Upsert inserts a citation, or if one already exists for the same
// (prompt_id, brand_id, url) — the extraction pipeline's natural identity for
// "this brand cited at this source for this prompt" — updates it in place and
// bumps citation_frequency instead of creating a duplicate row.
func (r citationRepositoryDB) Upsert(tx *sqlx.Tx, c Citation) (int, error) {
	var id int
	query := `INSERT INTO citations (prompt_id, brand_id, prompt_run_id, ranking, content, url, ai_platform, date_discovered, last_checked, sentiment, brand_positioning, citation_frequency, snippet, is_competitor, notes, is_archived, source_type, target_country)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		ON CONFLICT (prompt_id, brand_id, url, ai_platform) DO UPDATE SET
			prompt_run_id = EXCLUDED.prompt_run_id,
			ranking = EXCLUDED.ranking,
			content = EXCLUDED.content,
			ai_platform = EXCLUDED.ai_platform,
			last_checked = EXCLUDED.last_checked,
			sentiment = EXCLUDED.sentiment,
			brand_positioning = EXCLUDED.brand_positioning,
			citation_frequency = citations.citation_frequency + 1,
			snippet = EXCLUDED.snippet,
			is_competitor = EXCLUDED.is_competitor,
			source_type = EXCLUDED.source_type,
			target_country = EXCLUDED.target_country
		RETURNING id`
	err := tx.QueryRowx(query, c.PromptId, c.BrandId, c.PromptRunId, c.Ranking, c.Content, c.Url, c.AiPlatform, c.DateDiscovered, c.LastChecked, c.Sentiment, c.BrandPositioning, c.CitationFrequency, c.Snippet, c.IsCompetitor, c.Notes, c.IsArchived, c.SourceType, c.TargetCountry).Scan(&id)
	return id, err
}

func (r citationRepositoryDB) BumpDailyStat(tx *sqlx.Tx, companyId int, url, date string) error {
	query := `INSERT INTO citation_url_daily_stats (company_id, url, stat_date, citation_count) VALUES ($1,$2,$3,1)
		ON CONFLICT (company_id, url, stat_date) DO UPDATE SET citation_count = citation_url_daily_stats.citation_count + 1`
	_, err := tx.Exec(query, companyId, url, date)
	return err
}

func (r citationRepositoryDB) GetById(id int) (*Citation, error) {
	c := Citation{}
	query := `SELECT c.id, c.prompt_id, c.brand_id, c.ranking, c.content, c.url, c.ai_platform, c.date_discovered, c.last_checked, c.sentiment, c.brand_positioning, c.citation_frequency, c.snippet, c.is_competitor, c.notes, c.is_archived, c.source_type, c.target_country, c.created_at, b.name AS brand_name
		FROM citations c
		LEFT JOIN brands b ON b.id = c.brand_id
		WHERE c.id = $1`
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
			WHERE ci.id = $1 AND (
				c.user_id = $2
				OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id = c.id AND cm.user_id = $2 AND cm.status = 'active')
			)
		)`, id, userId)
	return owned, err
}
