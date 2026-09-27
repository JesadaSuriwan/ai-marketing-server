package service

import "github.com/ai-marketing/ai-marketing-server/providers/usage"

// AIProvider is implemented by an AI platform client capable of a single
// system+user prompt completion (used here for analysis, not prompt runs).
type AIProvider interface {
	Complete(systemPrompt, userPrompt string) (response string, model string, tokenUsage usage.Usage, err error)
}

type PromptSuggestionData struct {
	Id              int    `json:"id"`
	CompanyId       int    `json:"company_id"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	Rationale       string `json:"rationale"`
	Category        string `json:"category"`
	Intent          string `json:"intent"`
	Status          string `json:"status"`
	CreatedPromptId *int   `json:"created_prompt_id"`
	CreatedAt       string `json:"created_at"`
}

type PromptSuggestionListResponse struct {
	Status bool                   `json:"status"`
	Desc   string                 `json:"desc"`
	Data   []PromptSuggestionData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type PromptSuggestionService interface {
	// seed is an optional topic to focus the suggestions on — empty means
	// the general gap-analysis behavior (unchanged from before).
	Generate(companyId, userId int, seed string) (*PromptSuggestionListResponse, error)
	List(companyId int) (*PromptSuggestionListResponse, error)
	// countryCode is required when status is "accepted" — every prompt needs
	// exactly one tracked country, same rule as creating one manually.
	UpdateStatus(id, userId int, status, countryCode string) (*SimpleResponse, error)
}
