package repository

import "github.com/jmoiron/sqlx"

type promptRunRepositoryDB struct {
	db *sqlx.DB
}

func NewPromptRunRepositoryDB(db *sqlx.DB) PromptRunRepository {
	return promptRunRepositoryDB{db}
}

func (r promptRunRepositoryDB) Create(promptId int, aiPlatform, model, rawResponse string) (PromptRun, error) {
	run := PromptRun{PromptId: promptId, AiPlatform: aiPlatform, Model: model, RawResponse: rawResponse}
	err := r.db.QueryRowx(
		`INSERT INTO prompt_runs (prompt_id, ai_platform, model, raw_response) VALUES ($1,$2,$3,$4) RETURNING id, created_at`,
		promptId, aiPlatform, model, rawResponse,
	).Scan(&run.Id, &run.CreatedAt)
	return run, err
}

func (r promptRunRepositoryDB) GetByPromptId(promptId int) ([]PromptRun, error) {
	list := []PromptRun{}
	err := r.db.Select(&list, `SELECT id, prompt_id, ai_platform, model, raw_response, created_at FROM prompt_runs WHERE prompt_id = $1 ORDER BY created_at DESC`, promptId)
	return list, err
}

func (r promptRunRepositoryDB) CreateRunLog(p CreateRunLogParams) error {
	_, err := r.db.Exec(
		`INSERT INTO prompt_run_logs
			(company_id, prompt_id, prompt_title, tag_id, country, ai_platform, model, status, error_message, trigger_type, duration_ms, batch_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		p.CompanyId, p.PromptId, p.PromptTitle, p.TagId, p.Country, p.AiPlatform, p.Model, p.Status, p.ErrorMessage, p.TriggerType, p.DurationMs, p.BatchId,
	)
	return err
}

// GetRunLogs: from/to compare against created_at by calendar day (the whole
// "to" day is included, not just its midnight instant) — same convention
// the rest of the app uses for date-range filters.
func (r promptRunRepositoryDB) GetRunLogs(companyId int, f RunLogFilters) ([]PromptRunLog, error) {
	list := []PromptRunLog{}
	query := `
		SELECT
			l.id, l.company_id, l.prompt_id, l.prompt_title, l.tag_id, pc.name AS tag_name,
			l.country, l.ai_platform, l.model, l.status, l.error_message, l.trigger_type, l.duration_ms, l.batch_id,
			to_char(l.created_at, 'YYYY-MM-DD"T"HH24:MI:SS') || '+07:00' AS created_at
		FROM prompt_run_logs l
		LEFT JOIN prompt_categories pc ON pc.id = l.tag_id
		WHERE l.company_id = $1
		  AND ($2 = '' OR l.created_at >= $2::date)
		  AND ($3 = '' OR l.created_at < ($3::date + INTERVAL '1 day'))
		  AND ($4 = '' OR l.ai_platform = ANY(string_to_array($4, ',')))
		  AND ($5 = '' OR l.tag_id = ANY(string_to_array($5, ',')::int[]))
		  AND ($6 = '' OR l.country = ANY(string_to_array($6, ',')))
		ORDER BY l.created_at DESC
	`
	err := r.db.Select(&list, query, companyId, f.From, f.To, f.Engines, f.TagIds, f.Countries)
	return list, err
}
