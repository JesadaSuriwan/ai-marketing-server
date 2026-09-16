package server

import (
	"net/http"
	"strconv"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	brandRepository "github.com/ai-marketing/ai-marketing-server/pkg/brand/repository"
	citationRepository "github.com/ai-marketing/ai-marketing-server/pkg/citation/repository"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/prompt/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/prompt/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/prompt/service"
	promptRunHandler "github.com/ai-marketing/ai-marketing-server/pkg/promptrun/handler"
	promptRunRepository "github.com/ai-marketing/ai-marketing-server/pkg/promptrun/repository"
	promptRunService "github.com/ai-marketing/ai-marketing-server/pkg/promptrun/service"
	subdomainRepository "github.com/ai-marketing/ai-marketing-server/pkg/subdomain/repository"
	visibilityRepository "github.com/ai-marketing/ai-marketing-server/pkg/visibility/repository"
	"github.com/ai-marketing/ai-marketing-server/providers/anthropic"
	"github.com/ai-marketing/ai-marketing-server/providers/claude"
	"github.com/ai-marketing/ai-marketing-server/providers/gemini"
	"github.com/ai-marketing/ai-marketing-server/providers/openai"
	"github.com/ai-marketing/ai-marketing-server/providers/perplexity"
	"github.com/ai-marketing/ai-marketing-server/providers/serpapi"
	"github.com/gin-gonic/gin"
)

func (s *ginServer) initPromptRouter() {
	promptRepository := repository.NewPromptRepositoryDB(s.db)
	companyRepoForPrompts := companyRepository.NewCompanyRepositoryDB(s.db)

	if s.conf.Env.OpenAIAPIKey == "" {
		logs.Info("OPENAI_API_KEY not configured: ChatGPT prompt runs will fail until it is set")
	}
	if s.conf.Env.GeminiAPIKey == "" {
		logs.Info("GEMINI_API_KEY not configured: Gemini prompt runs will fail until it is set")
	}
	if s.conf.Env.PerplexityAPIKey == "" {
		logs.Info("PERPLEXITY_API_KEY not configured: Perplexity prompt runs will fail until it is set")
	}
	if s.conf.Env.ClaudeAPIKey == "" {
		logs.Info("CLAUDE_API_KEY not configured: Claude prompt runs will fail until it is set")
	}
	if s.conf.Env.AnthropicAPIKey == "" {
		logs.Info("ANTHROPIC_API_KEY not configured: citation extraction will be skipped on prompt runs until it is set")
	}
	if s.conf.Env.SerpApiAPIKey == "" {
		logs.Info("SERPAPI_API_KEY not configured: Google AI Overview prompt runs will fail until it is set")
	}
	openaiClient := openai.NewClient(s.conf.Env.OpenAIAPIKey)
	geminiClient := gemini.NewClient(s.conf.Env.GeminiAPIKey)
	perplexityClient := perplexity.NewClient(s.conf.Env.PerplexityAPIKey)
	claudeClient := claude.NewClient(s.conf.Env.ClaudeAPIKey)
	serpApiClient := serpapi.NewClient(s.conf.Env.SerpApiAPIKey)
	// Separate from claudeClient above: this key/client is used only for the
	// citation-extraction step (turning a raw run response into structured
	// brand mentions), not for running prompts.
	anthropicClient := anthropic.NewClient(s.conf.Env.AnthropicAPIKey)

	// Every "Run Prompt" fans out across all engines here — add a new
	// provider's client to this list to have it run alongside the others.
	engines := []promptRunService.Engine{
		{Platform: "chatgpt", Provider: openaiClient},
		{Platform: "gemini", Provider: geminiClient},
		{Platform: "perplexity", Provider: perplexityClient},
		{Platform: "claude", Provider: claudeClient},
		{Platform: "google-ai", Provider: serpApiClient},
	}

	runRepository := promptRunRepository.NewPromptRunRepositoryDB(s.db)
	brandRepo := brandRepository.NewBrandRepositoryDB(s.db)
	citationRepo := citationRepository.NewCitationRepositoryDB(s.db)
	visibilityRepo := visibilityRepository.NewVisibilityRepositoryDB(s.db)
	companyRepo := companyRepository.NewCompanyRepositoryDB(s.db)
	subdomainRepo := subdomainRepository.NewSubdomainRepositoryDB(s.db)

	runService := promptRunService.NewPromptRunService(
		runRepository, promptRepository, brandRepo, citationRepo, visibilityRepo, companyRepo, subdomainRepo,
		engines, anthropicClient, s.usageService,
	)
	// Exposed so the scheduler (started separately) can reuse the exact same
	// run/extraction pipeline for its automatic sweep.
	s.promptRunService = runService
	runHandler := promptRunHandler.NewPromptRunHandler(runService)

	// runService also powers the auto first-run fired when a prompt is
	// created — that's why prompt/service construction sits after it here.
	promptService := service.NewPromptService(promptRepository, companyRepoForPrompts, runService, s.db)
	promptHandler := handler.NewPromptHandler(promptService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/prompt")
	router.Use(authMiddleware.AuthRequired)

	router.GET("", companyMiddleware.OwnerByQuery, promptHandler.GetAll)
	router.POST("", companyMiddleware.OwnerByBodyCompanyId, promptHandler.Create)

	// Detail routes: ownership verified inside service via BelongsToUser
	router.GET("/:id", promptHandler.GetById)
	router.PUT("/:id", promptHandler.Update)
	router.PUT("/:id/active", promptHandler.SetActive)
	router.DELETE("/:id", promptHandler.Delete)

	// Run routes: ownership verified inside service via BelongsToUser
	router.POST("/:id/run", runHandler.Run)
	router.GET("/:id/runs", runHandler.GetHistory)

	// Manual "run all my prompts now" — same pipeline the scheduler uses,
	// triggerable on demand instead of waiting for the next cron tick.
	router.POST("/run-all", companyMiddleware.OwnerByQuery, func(c *gin.Context) {
		companyId, err := strconv.Atoi(c.Query("company_id"))
		if err != nil {
			errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
			return
		}
		result, err := runService.RunAllForCompany(companyId)
		if err != nil {
			errs.HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	})
}
