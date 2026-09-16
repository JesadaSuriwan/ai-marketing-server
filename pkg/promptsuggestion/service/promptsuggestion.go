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
	Generate(companyId, userId int) (*PromptSuggestionListResponse, error)
	List(companyId int) (*PromptSuggestionListResponse, error)
	UpdateStatus(id, userId int, status string) (*SimpleResponse, error)
}
