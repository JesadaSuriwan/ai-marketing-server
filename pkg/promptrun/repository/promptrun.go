package repository

type PromptRun struct {
	Id          int    `db:"id"`
	PromptId    int    `db:"prompt_id"`
	AiPlatform  string `db:"ai_platform"`
	Model       string `db:"model"`
	RawResponse string `db:"raw_response"`
	CreatedAt   string `db:"created_at"`
}

// PromptRunLog is one (prompt, engine) attempt, success or failure. PromptId
// and TagId are nullable — ON DELETE SET NULL when the source prompt/tag is
// later deleted — which is why PromptTitle and Country are snapshotted onto
// the row itself instead of only ever being read via a join. BatchId is
// shared by every engine's row from the same run() call, so the frontend can
// group a prompt's whole run (one row per engine) into a single expandable entry.
type PromptRunLog struct {
	Id           int     `db:"id"`
	CompanyId    int     `db:"company_id"`
	PromptId     *int    `db:"prompt_id"`
	PromptTitle  string  `db:"prompt_title"`
	TagId        *int    `db:"tag_id"`
	TagName      *string `db:"tag_name"`
	Country      string  `db:"country"`
	AiPlatform   string  `db:"ai_platform"`
	Model        string  `db:"model"`
	Status       string  `db:"status"`
	ErrorMessage *string `db:"error_message"`
	TriggerType  string  `db:"trigger_type"`
	DurationMs   int     `db:"duration_ms"`
	BatchId      string  `db:"batch_id"`
	CreatedAt    string  `db:"created_at"`
}

// CreateRunLogParams.PromptId is always a real id at the point a log is
// written (it only becomes NULL later, if that prompt is deleted) — plain
// int here, not *int, so every caller can't forget to set it.
type CreateRunLogParams struct {
	CompanyId    int
	PromptId     int
	PromptTitle  string
	TagId        *int
	Country      string
	AiPlatform   string
	Model        string
	Status       string
	ErrorMessage *string
	TriggerType  string
	DurationMs   int
	BatchId      string
}

// RunLogFilters mirrors the dashboard package's filter convention:
// comma-separated lists, empty string means no filter on that dimension.
type RunLogFilters struct {
	From, To                   string
	Engines, TagIds, Countries string
}

type PromptRunRepository interface {
	Create(promptId int, aiPlatform, model, rawResponse string) (PromptRun, error)
	GetByPromptId(promptId int) ([]PromptRun, error)
	CreateRunLog(p CreateRunLogParams) error
	GetRunLogs(companyId int, filters RunLogFilters) ([]PromptRunLog, error)
}
