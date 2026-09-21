package service

import (
	"github.com/ai-marketing/ai-marketing-server/providers/citation"
	"github.com/ai-marketing/ai-marketing-server/providers/usage"
)

// AIProvider is implemented by an AI platform client that actually runs the
// prompt (search-grounded — returns the real source URLs it cited alongside
// the response text). Satisfied by providers/openai.Client and
// providers/gemini.Client.
type AIProvider interface {
	Complete(prompt, country string) (response string, model string, citations []citation.Citation, tokenUsage usage.Usage, err error)
}

// ExtractionProvider is implemented by the model used to turn a raw run
// response into structured brand mentions (Claude, via providers/anthropic).
type ExtractionProvider interface {
	Complete(systemPrompt, userPrompt string) (response string, model string, tokenUsage usage.Usage, err error)
}

// Engine pairs an AIProvider with the platform name its runs/citations get
// tagged with (e.g. "chatgpt", "gemini"). A prompt run fans out across every
// configured engine.
type Engine struct {
	Platform string
	Provider AIProvider
}

type PromptRunData struct {
	Id          int    `json:"id"`
	PromptId    int    `json:"prompt_id"`
	AiPlatform  string `json:"ai_platform"`
	Model       string `json:"model"`
	RawResponse string `json:"raw_response"`
	CreatedAt   string `json:"created_at"`
}

type PromptRunListResponse struct {
	Status bool            `json:"status"`
	Desc   string          `json:"desc"`
	Data   []PromptRunData `json:"data"`
}

type RunAllResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
	Ran    int    `json:"ran"`
	Failed int    `json:"failed"`
}

type PromptRunService interface {
	// Run fans out across every configured engine (e.g. ChatGPT and Gemini),
	// returning one PromptRunData per engine that succeeded.
	Run(promptId, userId int) (*PromptRunListResponse, error)
	// RunSystem runs a prompt without a requesting-user ownership check —
	// used by the scheduler, which has no authenticated user in context.
	RunSystem(promptId int) (*PromptRunListResponse, error)
	GetHistory(promptId, userId int) (*PromptRunListResponse, error)
	// RunAllForCompany runs every prompt belonging to companyId, used by the
	// manual "run all" trigger and reusable by the scheduler per-company too.
	RunAllForCompany(companyId int) (*RunAllResponse, error)
}
