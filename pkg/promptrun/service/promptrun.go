package service

import (
	"github.com/ai-marketing/ai-marketing-server/providers/citation"
	"github.com/ai-marketing/ai-marketing-server/providers/usage"
)

// Trigger values recorded on every prompt_run_logs row. TriggerScheduled is
// only ever set by the cron sweep (pkg/scheduler) — every other path,
// including the auto first-run on prompt create/suggestion-approve (which
// also goes through RunSystem, since those have no authenticated user in
// context), counts as TriggerManual: a real person's action caused it, even
// if indirectly.
const (
	TriggerScheduled = "scheduled"
	TriggerManual    = "manual"
)

const (
	RunStatusSuccess = "success"
	RunStatusFailure = "failure"
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

type RunLogData struct {
	Id           int     `json:"id"`
	PromptId     *int    `json:"prompt_id"`
	PromptTitle  string  `json:"prompt_title"`
	TagId        *int    `json:"tag_id"`
	TagName      *string `json:"tag_name"`
	Country      string  `json:"country"`
	AiPlatform   string  `json:"ai_platform"`
	Model        string  `json:"model"`
	Status       string  `json:"status"`
	ErrorMessage *string `json:"error_message"`
	TriggerType  string  `json:"trigger_type"`
	DurationMs   int     `json:"duration_ms"`
	CreatedAt    string  `json:"created_at"`
}

type RunLogsResponse struct {
	Status bool         `json:"status"`
	Desc   string       `json:"desc"`
	Data   []RunLogData `json:"data"`
}

// RunLogFilters: comma-separated lists, empty means no filter — same
// convention as the dashboard package's filter params.
type RunLogFilters struct {
	From, To                   string
	Engines, TagIds, Countries string
}

type PromptRunService interface {
	// Run fans out across every configured engine (e.g. ChatGPT and Gemini),
	// returning one PromptRunData per engine that succeeded.
	Run(promptId, userId int) (*PromptRunListResponse, error)
	// RunSystem runs a prompt without a requesting-user ownership check —
	// used by the scheduler and by auto-triggered runs (prompt create,
	// suggestion approve), none of which have an authenticated user in
	// context. triggerType is what gets recorded on the resulting log rows —
	// pass TriggerScheduled or TriggerManual.
	RunSystem(promptId int, triggerType string) (*PromptRunListResponse, error)
	GetHistory(promptId, userId int) (*PromptRunListResponse, error)
	// RunAllForCompany runs every prompt belonging to companyId, used by the
	// manual "run all" trigger and reusable by the scheduler per-company too.
	RunAllForCompany(companyId int) (*RunAllResponse, error)
	// GetRunLogs returns every prompt_run_logs row for companyId matching
	// filters — every engine attempt, success or failure, regardless of what
	// triggered it.
	GetRunLogs(companyId int, filters RunLogFilters) (*RunLogsResponse, error)
}
