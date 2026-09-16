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
