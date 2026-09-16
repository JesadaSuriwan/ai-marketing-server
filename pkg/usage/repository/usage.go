package repository

type UsageLog struct {
	Id           int     `db:"id"`
	CompanyId    *int    `db:"company_id"`
	Engine       string  `db:"engine"`
	Purpose      string  `db:"purpose"`
	Model        string  `db:"model"`
	InputTokens  int     `db:"input_tokens"`
	OutputTokens int     `db:"output_tokens"`
	CostUsd      float64 `db:"cost_usd"`
	CreatedAt    string  `db:"created_at"`
}

// EngineBreakdown is one (engine, purpose) bucket, summed across every
// company the requesting user owns or is an accepted member of.
type EngineBreakdown struct {
	Engine       string  `db:"engine"`
	Purpose      string  `db:"purpose"`
	Calls        int     `db:"calls"`
	InputTokens  int     `db:"input_tokens"`
	OutputTokens int     `db:"output_tokens"`
	CostUsd      float64 `db:"cost_usd"`
}

type UsageRepository interface {
	Log(l UsageLog) error
	// from/to are optional "YYYY-MM-DD" bounds on created_at; empty means unbounded.
	GetBreakdownForUser(userId int, from, to string) ([]EngineBreakdown, error)
}
