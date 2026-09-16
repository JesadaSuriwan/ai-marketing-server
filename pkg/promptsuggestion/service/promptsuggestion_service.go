package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	categoryRepository "github.com/ai-marketing/ai-marketing-server/pkg/category/repository"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	dashboardRepository "github.com/ai-marketing/ai-marketing-server/pkg/dashboard/repository"
	promptRepository "github.com/ai-marketing/ai-marketing-server/pkg/prompt/repository"
	promptRunService "github.com/ai-marketing/ai-marketing-server/pkg/promptrun/service"
	"github.com/ai-marketing/ai-marketing-server/pkg/promptsuggestion/repository"
	usageService "github.com/ai-marketing/ai-marketing-server/pkg/usage/service"
	"github.com/jmoiron/sqlx"
)

const systemPrompt = `You are an AI-search visibility strategist for a marketing agency. You are given a summary of a company's current tracked prompts and their performance across AI chatbots (ChatGPT, Gemini, Perplexity, etc), including brand coverage, competitor rankings, and cited domains.

Suggest exactly 5 new prompts (search queries) worth tracking that would fill coverage gaps, target categories the company is weak in, or capture opportunities where competitors are currently winning. Each suggestion must be a realistic, natural-language question a real user might ask an AI chatbot.

Respond with ONLY a JSON array, no prose, no markdown code fences, matching exactly this shape:
[{"title": "short label", "content": "the full natural-language prompt text", "rationale": "one sentence on why this prompt matters", "category": "a short category name"}]`

type claudeSuggestion struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Rationale string `json:"rationale"`
	Category  string `json:"category"`
}

type promptSuggestionService struct {
	suggestionRepository repository.PromptSuggestionRepository
	dashboardRepository  dashboardRepository.DashboardRepository
	promptRepository     promptRepository.PromptRepository
	categoryRepository   categoryRepository.CategoryRepository
	companyRepository    companyRepository.CompanyRepository
	promptRunService     promptRunService.PromptRunService
	aiProvider           AIProvider
	usageService         usageService.UsageService
}

func NewPromptSuggestionService(
	suggestionRepository repository.PromptSuggestionRepository,
	dashboardRepository dashboardRepository.DashboardRepository,
	promptRepository promptRepository.PromptRepository,
	categoryRepository categoryRepository.CategoryRepository,
	companyRepository companyRepository.CompanyRepository,
	promptRunService promptRunService.PromptRunService,
	aiProvider AIProvider,
	usageService usageService.UsageService,
) PromptSuggestionService {
	return promptSuggestionService{
		suggestionRepository, dashboardRepository, promptRepository,
		categoryRepository, companyRepository, promptRunService, aiProvider, usageService,
	}
}

func toData(ps repository.PromptSuggestion) PromptSuggestionData {
	rationale := ""
	if ps.Rationale != nil {
		rationale = *ps.Rationale
	}
	category := ""
	if ps.Category != nil {
		category = *ps.Category
	}
	return PromptSuggestionData{
		Id: ps.Id, CompanyId: ps.CompanyId, Title: ps.Title, Content: ps.Content,
		Rationale: rationale, Category: category, Status: ps.Status,
		CreatedPromptId: ps.CreatedPromptId, CreatedAt: ps.CreatedAt,
	}
}

// buildContext summarizes the company's current tracked-prompt performance
// data into a compact text block for Claude to analyze.
func (s promptSuggestionService) buildContext(companyId int) (string, error) {
	var b strings.Builder

	company, err := s.companyRepository.GetById(companyId)
	if err != nil {
		return "", err
	}
	industry := "unknown"
	if company.Industry != nil {
		industry = *company.Industry
	}
	fmt.Fprintf(&b, "Company: %s (industry: %s)\n\n", company.Name, industry)

	categories, err := s.categoryRepository.GetAll(companyId)
	if err != nil {
		return "", err
	}
	names := []string{}
	for _, c := range categories {
		names = append(names, c.Name)
	}
	fmt.Fprintf(&b, "Existing categories: %s\n\n", strings.Join(names, ", "))

	prompts, err := s.promptRepository.GetAll(companyId)
	if err != nil {
		return "", err
	}
	b.WriteString("Currently tracked prompts (do not suggest duplicates of these):\n")
	for _, p := range prompts {
		fmt.Fprintf(&b, "- %s\n", p.Title)
	}
	b.WriteString("\n")

	overview, err := s.dashboardRepository.GetPromptsOverview(companyId, "", "")
	if err != nil {
		return "", err
	}
	b.WriteString("Per-prompt brand coverage (brand_coverage% = share of AI answers mentioning the company's own brand):\n")
	for _, o := range overview {
		fmt.Fprintf(&b, "- %q [%s]: brand_coverage=%d%%, brand_mentions=%d/%d, competitors=%s\n",
			o.Title, o.Tag, o.BrandCoverage, o.BrandMentions, o.TotalBrandMentions, o.Competitors)
	}
	b.WriteString("\n")

	ranking, err := s.dashboardRepository.GetBrandRanking(companyId)
	if err != nil {
		return "", err
	}
	b.WriteString("Brand ranking (own brand vs competitors, by mentions):\n")
	for _, r := range ranking {
		own := ""
		if r.IsOwn {
			own = " (own brand)"
		}
		fmt.Fprintf(&b, "- #%d %s%s: mentions=%d, coverage=%.1f%%, share_of_voice=%.1f%%\n",
			r.Rank, r.Name, own, r.Mentions, r.BrandCoverage, r.ShareOfVoice)
	}
	b.WriteString("\n")

	domains, err := s.dashboardRepository.GetTopDomains(companyId)
	if err != nil {
		return "", err
	}
	b.WriteString("Top cited domains across all tracked prompts:\n")
	for _, d := range domains {
		fmt.Fprintf(&b, "- %s: %d mentions\n", d.Domain, d.MentionCount)
	}

	return b.String(), nil
}

// extractJSONArray trims any surrounding prose/markdown fences Claude may add
// despite instructions, keeping only the bracketed JSON array.
func extractJSONArray(s string) string {
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start == -1 || end == -1 || end < start {
		return s
	}
	return s[start : end+1]
}

func (s promptSuggestionService) Generate(companyId, userId int) (*PromptSuggestionListResponse, error) {
	context, err := s.buildContext(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	raw, model, tokenUsage, err := s.aiProvider.Complete(systemPrompt, context)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err := s.usageService.Log(&companyId, "claude", "prompt_suggestion", model, tokenUsage); err != nil {
		logs.Error(fmt.Errorf("failed to log usage for prompt suggestion generation (company %d): %w", companyId, err))
	}

	var parsed []claudeSuggestion
	if err := json.Unmarshal([]byte(extractJSONArray(raw)), &parsed); err != nil {
		logs.Error(err)
		logs.Info("prompt suggestion: unparseable claude response: " + raw)
		return nil, errs.NewUnexpectedError()
	}

	tx, err := s.suggestionRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	data := []PromptSuggestionData{}
	for _, p := range parsed {
		if p.Title == "" || p.Content == "" {
			continue
		}
		rationale := p.Rationale
		category := p.Category
		id, err := s.suggestionRepository.Create(tx, repository.PromptSuggestion{
			CompanyId: companyId, Title: p.Title, Content: p.Content,
			Rationale: &rationale, Category: &category,
		})
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}
		data = append(data, PromptSuggestionData{
			Id: id, CompanyId: companyId, Title: p.Title, Content: p.Content,
			Rationale: rationale, Category: category, Status: "pending",
		})
	}

	if err := tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &PromptSuggestionListResponse{Status: true, Desc: "Suggestions generated successfully", Data: data}, nil
}

func (s promptSuggestionService) List(companyId int) (*PromptSuggestionListResponse, error) {
	rows, err := s.suggestionRepository.GetByCompanyId(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []PromptSuggestionData{}
	for _, r := range rows {
		data = append(data, toData(r))
	}

	return &PromptSuggestionListResponse{Status: true, Desc: "Get suggestions successful", Data: data}, nil
}

// resolveCategory finds an existing category by case-insensitive name match,
// creating one if none matches. Returns nil if name is blank.
func (s promptSuggestionService) resolveCategory(tx *sqlx.Tx, companyId int, name string) (*int, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}

	categories, err := s.categoryRepository.GetAll(companyId)
	if err != nil {
		return nil, err
	}
	for _, c := range categories {
		if strings.EqualFold(c.Name, name) {
			id := c.Id
			return &id, nil
		}
	}

	id, err := s.categoryRepository.Create(tx, categoryRepository.Category{CompanyId: companyId, Name: name})
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (s promptSuggestionService) UpdateStatus(id, userId int, status string) (*SimpleResponse, error) {
	if status != "accepted" && status != "dismissed" {
		return nil, errs.NewBadRequestError("status must be 'accepted' or 'dismissed'")
	}

	owned, err := s.suggestionRepository.BelongsToUser(id, userId)
	if err != nil || !owned {
		return nil, errs.NewForbiddenError("access denied")
	}

	tx, err := s.suggestionRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	var createdPromptId *int
	if status == "accepted" {
		suggestion, err := s.suggestionRepository.GetById(id)
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}

		categoryName := ""
		if suggestion.Category != nil {
			categoryName = *suggestion.Category
		}
		categoryId, err := s.resolveCategory(tx, suggestion.CompanyId, categoryName)
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}

		// Accepting a suggestion creates a real prompt too — it must respect
		// the same active-prompt limit as manually created ones.
		company, err := s.companyRepository.GetById(suggestion.CompanyId)
		if err != nil {
			return nil, errs.NewNotFoundError("company not found")
		}
		activeCount, err := s.promptRepository.CountActive(suggestion.CompanyId)
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}
		if activeCount >= company.PromptLimit {
			return nil, errs.NewBadRequestError(fmt.Sprintf(
				"you've reached your active prompt limit of %d — deactivate a prompt or raise the limit in Settings",
				company.PromptLimit,
			))
		}

		newId, err := s.promptRepository.Create(tx, promptRepository.Prompt{
			CompanyId: suggestion.CompanyId, TagId: categoryId,
			Title: suggestion.Title, Content: suggestion.Content,
		})
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}
		createdPromptId = &newId
	}

	if err := s.suggestionRepository.UpdateStatus(tx, id, status, createdPromptId); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err := tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	// Best-effort, fire-and-forget: same auto first-run as manually created
	// prompts get — see pkg/prompt/service's Create for why this is async.
	if createdPromptId != nil {
		promptId := *createdPromptId
		go func() {
			if _, err := s.promptRunService.RunSystem(promptId); err != nil {
				logs.Error(fmt.Errorf("auto first-run failed for prompt %d: %w", promptId, err))
			}
		}()
	}

	return &SimpleResponse{Status: true, Desc: "Suggestion updated successfully"}, nil
}
