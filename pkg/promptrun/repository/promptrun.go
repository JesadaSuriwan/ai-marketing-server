package repository

type PromptRun struct {
	Id          int    `db:"id"`
	PromptId    int    `db:"prompt_id"`
	AiPlatform  string `db:"ai_platform"`
	Model       string `db:"model"`
	RawResponse string `db:"raw_response"`
	CreatedAt   string `db:"created_at"`
}

type PromptRunRepository interface {
	Create(promptId int, aiPlatform, model, rawResponse string) (PromptRun, error)
	GetByPromptId(promptId int) ([]PromptRun, error)
}
