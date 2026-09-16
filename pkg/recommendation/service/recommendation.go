package service

import "github.com/ai-marketing/ai-marketing-server/providers/usage"

// AIProvider is implemented by an AI platform client capable of a single
// system+user prompt completion. Same shape as promptsuggestion.AIProvider —
// both are satisfied by providers/anthropic.Client, passed in separately by
// the router so each package stays independent.
type AIProvider interface {
	Complete(systemPrompt, userPrompt string) (response string, model string, tokenUsage usage.Usage, err error)
}

type RecommendationData struct {
	Id        int      `json:"id"`
	CompanyId int      `json:"company_id"`
	Type      string   `json:"type"`
	Placement string   `json:"placement"`
	Impact    string   `json:"impact"`
	Engine    string   `json:"engine"`
	Text      string   `json:"text"`
	Link      *string  `json:"link"`
	LinkText  *string  `json:"link_text"`
	Why       string   `json:"why"`
	Steps     []string `json:"steps"`
	PromptId  *int     `json:"prompt_id"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
}

type RecommendationListResponse struct {
	Status bool                 `json:"status"`
	Desc   string               `json:"desc"`
	Data   []RecommendationData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type RecommendationService interface {
	// guidance is optional free text the user typed to steer generation —
	// see buildCandidates/Generate for how it's constrained so it can only
	// reorder/reframe the real candidates, never invent new ones.
	Generate(companyId, userId int, guidance string) (*RecommendationListResponse, error)
	List(companyId int) (*RecommendationListResponse, error)
	UpdateStatus(id, userId int, status string) (*SimpleResponse, error)
}
