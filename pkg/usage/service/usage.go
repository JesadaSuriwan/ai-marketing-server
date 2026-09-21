package service

import "github.com/ai-marketing/ai-marketing-server/providers/usage"

type UsageBreakdownData struct {
	Engine       string  `json:"engine"`
	Purpose      string  `json:"purpose"`
	Calls        int     `json:"calls"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CostUsd      float64 `json:"cost_usd"`
}

type UsageSummaryResponse struct {
	Status bool                 `json:"status"`
	Desc   string               `json:"desc"`
	Data   []UsageBreakdownData `json:"data"`
}

type UsageService interface {
	// Log is best-effort by design — callers treat a logging failure as
	// non-fatal (the real AI call already succeeded; losing a cost-tracking
	// row shouldn't fail the user-facing request), so this only returns an
	// error for the caller to log, never to propagate.
	Log(companyId *int, engine, purpose, model string, u usage.Usage) error
	// from/to are optional "YYYY-MM-DD" bounds; empty means all-time.
	GetSummaryForCompany(companyId int, from, to string) (*UsageSummaryResponse, error)
}
