package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	brandRepository "github.com/ai-marketing/ai-marketing-server/pkg/brand/repository"
	citationRepository "github.com/ai-marketing/ai-marketing-server/pkg/citation/repository"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	dashboardRepository "github.com/ai-marketing/ai-marketing-server/pkg/dashboard/repository"
	promptRepository "github.com/ai-marketing/ai-marketing-server/pkg/prompt/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/recommendation/repository"
	recommendationRuleRepository "github.com/ai-marketing/ai-marketing-server/pkg/recommendationrule/repository"
	usageService "github.com/ai-marketing/ai-marketing-server/pkg/usage/service"
)

const recommendationSystemPrompt = `You are an AI-search visibility strategist writing recommendations for a marketing agency's client. You are given a numbered list of REAL, verified opportunities already detected from the client's AI-citation tracking data — each one already has a real URL/domain, competitor name, and/or tracked prompt attached.

Your ONLY job is to write persuasive, concise copy for each numbered candidate. Do NOT invent, add, or modify any URL, domain, number, brand name, or platform — use only what is given to you in each candidate's data.

You may also be given the company's "Standing rules" (persistent preferences they've saved, applying to every generation) and/or one-off "User guidance" for this run specifically. Use both only to decide which of the given candidates to emphasize and how to angle/frame their copy (tone, which detail to lead with). Never let either cause you to invent a URL, number, or opportunity that isn't already in the numbered list — if a rule or guidance references something not present in the candidates, ignore that part of it.

For each candidate, produce:
- "text": one sentence describing the recommended action
- "why": one or two sentences explaining why this matters, referencing the real data given (competitor name, domain, coverage percentage, etc.)
- "steps": exactly 4 short, actionable steps to execute it
- "headline": ONLY for [content_gap] candidates — a short, compelling page title written in the SAME LANGUAGE as the tracked prompt/query quoted in that candidate's data (match that query's language exactly, do not translate it to English) — the actual title a content team could publish verbatim. For every other candidate kind, set "headline" to an empty string.

Respond with ONLY a JSON array, no prose, no markdown fences, matching exactly this shape:
[{"id": <candidate number>, "text": "...", "why": "...", "steps": ["...", "...", "...", "..."], "headline": "..."}]`

// candidate is a deterministically-identified real opportunity, built from
// existing citation/prompt data before Claude ever sees it. Claude only
// writes copy for candidates that already exist — it never originates a URL,
// prompt, or competitor name, so a hallucination can't turn into bad outreach
// advice.
type candidate struct {
	Id        int
	Kind      string // editorial_gap | existing_mention | content_gap
	Type      string // content | content_partnership
	Placement string // on_page | off_page
	Impact    string // high | medium
	Link      *string
	LinkText  *string
	PromptId  *int
}

type claudeCopy struct {
	Id       int      `json:"id"`
	Text     string   `json:"text"`
	Why      string   `json:"why"`
	Steps    []string `json:"steps"`
	Headline string   `json:"headline"`
}

// portfolioRec is a recommendation about the shape of the tracked-prompt list
// itself (funnel-stage mix, branded-vs-unbranded ratio) rather than any single
// citation. It's pure arithmetic over prompts/brands already in the database,
// so — unlike the citation-grounded candidates above — its copy is written
// directly in Go rather than handed to Claude; there's no room for a
// hallucinated number here since Go already computed the exact one.
type portfolioRec struct {
	Type   string
	Impact string
	Text   string
	Why    string
	Steps  []string
}

type recommendationService struct {
	recommendationRepository     repository.RecommendationRepository
	dashboardRepository          dashboardRepository.DashboardRepository
	companyRepository            companyRepository.CompanyRepository
	promptRepository             promptRepository.PromptRepository
	brandRepository              brandRepository.BrandRepository
	citationRepository           citationRepository.CitationRepository
	recommendationRuleRepository recommendationRuleRepository.RecommendationRuleRepository
	aiProvider                   AIProvider
	usageService                 usageService.UsageService
}

func NewRecommendationService(
	recommendationRepository repository.RecommendationRepository,
	dashboardRepository dashboardRepository.DashboardRepository,
	companyRepository companyRepository.CompanyRepository,
	promptRepository promptRepository.PromptRepository,
	brandRepository brandRepository.BrandRepository,
	citationRepository citationRepository.CitationRepository,
	recommendationRuleRepository recommendationRuleRepository.RecommendationRuleRepository,
	aiProvider AIProvider,
	usageService usageService.UsageService,
) RecommendationService {
	return recommendationService{
		recommendationRepository, dashboardRepository, companyRepository,
		promptRepository, brandRepository, citationRepository, recommendationRuleRepository,
		aiProvider, usageService,
	}
}

// buildPortfolioCandidates flags branded prompts (containing the company's
// own brand name — these inflate coverage stats since you're likely to
// already dominate them) and reports the top-of-funnel share of the tracked
// prompt list (approximated as "does not mention the own brand" — a
// deliberately simple heuristic: branded/purchase-intent queries are
// considered further down the funnel, everything else top-of-funnel).
func (s recommendationService) buildPortfolioCandidates(companyId int) ([]portfolioRec, error) {
	prompts, err := s.promptRepository.GetAll(companyId)
	if err != nil {
		return nil, err
	}
	if len(prompts) == 0 {
		return nil, nil
	}

	brands, err := s.brandRepository.GetAll(companyId)
	if err != nil {
		return nil, err
	}
	ownNames := []string{}
	for _, b := range brands {
		if b.IsOwn && strings.TrimSpace(b.Name) != "" {
			ownNames = append(ownNames, strings.ToLower(b.Name))
		}
	}

	brandedCount := 0
	topOfFunnelCount := 0
	for _, p := range prompts {
		text := strings.ToLower(p.Title + " " + p.Content)
		branded := false
		for _, n := range ownNames {
			if strings.Contains(text, n) {
				branded = true
				break
			}
		}
		if branded {
			brandedCount++
		} else {
			topOfFunnelCount++
		}
	}

	recs := []portfolioRec{}

	if brandedCount > 0 {
		recs = append(recs, portfolioRec{
			Type: "branded_prompts", Impact: "medium",
			Text: fmt.Sprintf("Remove branded prompts for more objective recommendations. You have %d branded prompt(s).", brandedCount),
			Why:  "Prompts that already include your own brand name tend to show inflated visibility, since you're more likely to already dominate them. Replacing them with unbranded, category-level queries gives a more objective read on real competitive visibility.",
			Steps: []string{
				"Review your tracked prompts list",
				"Identify prompts that include your brand name",
				"Replace them with generic, unbranded search queries in the same category",
				"Re-run tracking to get an objective baseline",
			},
		})
	}

	if len(prompts) >= 3 {
		pct := int(float64(topOfFunnelCount) / float64(len(prompts)) * 100)
		recs = append(recs, portfolioRec{
			Type: "prompt_mix", Impact: "low",
			Text: fmt.Sprintf("Rebalance your prompt mix — %d%% of your prompts are top-of-funnel.", pct),
			Why: fmt.Sprintf(
				"A healthy mix of top-, middle-, and bottom-of-funnel prompts gives a more complete picture of your AI visibility across the full buyer journey. Currently %d%% of your %d tracked prompts are broad, unbranded (top-of-funnel) queries.",
				pct, len(prompts),
			),
			Steps: []string{
				"Audit your current prompt list by funnel stage",
				"Add middle-of-funnel prompts (comparison or feature-specific queries)",
				"Add bottom-of-funnel prompts (branded, purchase-intent queries)",
				"Aim for a more even split across funnel stages",
			},
		})
	}

	return recs, nil
}

func toData(r repository.Recommendation) RecommendationData {
	steps := []string{}
	_ = json.Unmarshal([]byte(r.Steps), &steps)
	return RecommendationData{
		Id: r.Id, CompanyId: r.CompanyId, Type: r.Type, Placement: r.Placement,
		Impact: r.Impact, Engine: r.Engine, Text: r.Text, Link: r.Link, LinkText: r.LinkText,
		Why: r.Why, Steps: steps, PromptId: r.PromptId, Status: r.Status, CreatedAt: r.CreatedAt,
	}
}

// buildCandidates derives real, groundable opportunities from data that
// already exists — no new joins, reuses the dashboard repository the same
// way pkg/promptsuggestion does.
func (s recommendationService) buildCandidates(companyId int) ([]candidate, map[int]string, error) {
	// contexts holds, per candidate id, the plain-text description handed to
	// Claude — kept separate from the candidate struct so the struct only
	// carries what actually gets persisted.
	contexts := map[int]string{}
	nextId := 1

	urls, err := s.dashboardRepository.GetCitationURLs(companyId)
	if err != nil {
		return nil, nil, err
	}

	editorialGaps := []candidate{}
	existingMentions := []candidate{}
	for _, u := range urls {
		if u.SourceType != "editorial" {
			continue
		}
		url := u.Url
		// u.Title is actually the full raw AI-response text this citation was
		// extracted from (a known quirk of GetCitationURLs, elsewhere always
		// display-truncated) — never usable as a short label. The domain is
		// the only genuinely short, meaningful thing to show as link text.
		linkText := u.Domain
		if !u.BrandMentioned && u.Competitors != "" {
			c := candidate{Id: nextId, Kind: "editorial_gap", Type: "content_partnership", Placement: "off_page", Link: &url, LinkText: &linkText}
			nCompetitors := len(strings.Split(u.Competitors, ","))
			if nCompetitors >= 2 || u.Cited >= 2 {
				c.Impact = "high"
			} else {
				c.Impact = "medium"
			}
			contexts[nextId] = fmt.Sprintf(
				"[editorial_gap] An editorial article at domain %q (url: %s) currently cites our competitor(s) %s but does not mention our brand. It has appeared in %d tracked prompt(s). Recommend pitching this publication for a mention or feature.",
				u.Domain, url, u.Competitors, u.Cited,
			)
			editorialGaps = append(editorialGaps, c)
			nextId++
		} else if u.BrandMentioned {
			c := candidate{Id: nextId, Kind: "existing_mention", Type: "content_partnership", Placement: "off_page", Impact: "medium", Link: &url, LinkText: &linkText}
			contexts[nextId] = fmt.Sprintf(
				"[existing_mention] Our brand is already cited in an editorial article at domain %q (url: %s), appearing in %d tracked prompt(s). Recommend deepening this relationship (e.g. a co-branded feature, updated quote, or sponsorship) to increase future citation frequency.",
				u.Domain, url, u.Cited,
			)
			existingMentions = append(existingMentions, c)
			nextId++
		}
	}
	sort.Slice(editorialGaps, func(i, j int) bool { return editorialGaps[i].Impact == "high" && editorialGaps[j].Impact != "high" })

	overview, err := s.dashboardRepository.GetPromptsOverview(companyId, "", "")
	if err != nil {
		return nil, nil, err
	}
	contentGaps := []candidate{}
	for _, o := range overview {
		// Zero coverage only: real competitive activity exists for this prompt
		// (TotalBrandMentions > 0) but our brand is cited nowhere in it.
		if o.TotalBrandMentions == 0 || o.BrandCoverage > 0 {
			continue
		}
		promptId := o.PromptId
		c := candidate{Id: nextId, Kind: "content_gap", Type: "content", Placement: "on_page", PromptId: &promptId, Impact: "high"}
		competitors := o.Competitors
		if competitors == "" {
			competitors = "competitors"
		}

		// Ground this in a real competitor citation for the prompt, if one
		// exists with a URL — gives Claude (and the resulting recommendation)
		// an actual page to reference and link to, not just a competitor name.
		competitorNote := ""
		if cits, err := s.citationRepository.GetAll(promptId, "", ""); err == nil {
			var withURL []citationRepository.Citation
			for _, ct := range cits {
				if ct.IsCompetitor && ct.Url != nil && strings.TrimSpace(*ct.Url) != "" {
					withURL = append(withURL, ct)
				}
			}
			if len(withURL) > 0 {
				sort.Slice(withURL, func(i, j int) bool { return withURL[i].CitationFrequency > withURL[j].CitationFrequency })
				top := withURL[0]
				citedURL := *top.Url
				domain := extractDomain(citedURL)
				c.Link = &citedURL
				c.LinkText = &domain
				competitorNote = fmt.Sprintf(" Based on %d competitor citation(s) found for this prompt, including %s (%s).", len(withURL), domain, citedURL)
			}
		}

		contexts[nextId] = fmt.Sprintf(
			"[content_gap] The tracked prompt %q (tag: %s) has only %d%% brand coverage despite %d total brand mentions recorded across AI answers, with %s currently appearing instead.%s Recommend publishing on-page content that directly targets this search intent, and write a headline for it in the same language as the prompt quoted above.",
			o.Title, o.Tag, o.BrandCoverage, o.TotalBrandMentions, competitors, competitorNote,
		)
		contentGaps = append(contentGaps, c)
		nextId++
	}
	sort.Slice(contentGaps, func(i, j int) bool { return contentGaps[i].Impact == "high" && contentGaps[j].Impact != "high" })

	all := []candidate{}
	all = append(all, capList(editorialGaps, 5)...)
	all = append(all, capList(existingMentions, 3)...)
	all = append(all, capList(contentGaps, 20)...)

	return all, contexts, nil
}

// extractDomain returns the lowercased host of a URL (stripping "www."),
// falling back to the raw string if it doesn't parse as a URL.
func extractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Hostname() == "" {
		return rawURL
	}
	return strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
}

func capList(c []candidate, n int) []candidate {
	if len(c) > n {
		return c[:n]
	}
	return c
}

func (s recommendationService) Generate(companyId, userId int, guidance string) (*RecommendationListResponse, error) {
	candidates, contexts, err := s.buildCandidates(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	portfolioRecs, err := s.buildPortfolioCandidates(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	if len(candidates) == 0 && len(portfolioRecs) == 0 {
		return &RecommendationListResponse{Status: true, Desc: "No new opportunities found", Data: []RecommendationData{}}, nil
	}

	copyById := map[int]claudeCopy{}
	if len(candidates) > 0 {
		company, err := s.companyRepository.GetById(companyId)
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}

		var b strings.Builder
		fmt.Fprintf(&b, "Client company: %s\n\n", company.Name)
		for _, c := range candidates {
			fmt.Fprintf(&b, "%d. %s\n", c.Id, contexts[c.Id])
		}

		activeRules, err := s.recommendationRuleRepository.GetActiveTexts(companyId)
		if err != nil {
			logs.Error(fmt.Errorf("failed to load recommendation rules for company %d: %w", companyId, err))
		} else if len(activeRules) > 0 {
			b.WriteString("\nStanding rules (apply to every generation):\n")
			for _, rule := range activeRules {
				fmt.Fprintf(&b, "- %s\n", rule)
			}
		}

		if guidance = strings.TrimSpace(guidance); guidance != "" {
			fmt.Fprintf(&b, "\nUser guidance for this run: %s\n", guidance)
		}

		raw, model, tokenUsage, err := s.aiProvider.Complete(recommendationSystemPrompt, b.String())
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}

		if err := s.usageService.Log(&companyId, "claude", "recommendation", model, tokenUsage); err != nil {
			logs.Error(fmt.Errorf("failed to log usage for recommendation generation (company %d): %w", companyId, err))
		}

		var copies []claudeCopy
		if err := json.Unmarshal([]byte(extractJSONArray(raw)), &copies); err != nil {
			logs.Error(err)
			logs.Info("recommendation: unparseable claude response: " + raw)
			return nil, errs.NewUnexpectedError()
		}
		for _, cp := range copies {
			copyById[cp.Id] = cp
		}
	}

	tx, err := s.recommendationRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	data := []RecommendationData{}

	// Portfolio-level recs are written directly in Go — no Claude round trip,
	// nothing to match by id.
	for _, pr := range portfolioRecs {
		stepsJSON, err := json.Marshal(pr.Steps)
		if err != nil {
			continue
		}
		id, err := s.recommendationRepository.Upsert(tx, repository.Recommendation{
			CompanyId: companyId, Type: pr.Type, Placement: "general", Impact: pr.Impact,
			Engine: "", Text: pr.Text, Why: pr.Why, Steps: string(stepsJSON),
		})
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}
		if id == 0 {
			continue
		}
		data = append(data, RecommendationData{
			Id: id, CompanyId: companyId, Type: pr.Type, Placement: "general", Impact: pr.Impact,
			Engine: "", Text: pr.Text, Why: pr.Why, Steps: pr.Steps, Status: "suggested",
		})
	}

	for _, c := range candidates {
		cp, ok := copyById[c.Id]
		if !ok || cp.Text == "" || len(cp.Steps) == 0 {
			continue
		}
		// For content_gap, lead with the actual publishable headline (in the
		// prompt's own language) instead of Claude's generic action sentence
		// — this is the specific, copy-pasteable title a content team can use.
		text := cp.Text
		if c.Kind == "content_gap" && strings.TrimSpace(cp.Headline) != "" {
			text = fmt.Sprintf("Publish a new page — \"%s\" — on your site.", strings.TrimSpace(cp.Headline))
		}
		stepsJSON, err := json.Marshal(cp.Steps)
		if err != nil {
			continue
		}
		id, err := s.recommendationRepository.Upsert(tx, repository.Recommendation{
			CompanyId: companyId, Type: c.Type, Placement: c.Placement, Impact: c.Impact,
			Engine: "chatgpt", Text: text, Link: c.Link, LinkText: c.LinkText,
			Why: cp.Why, Steps: string(stepsJSON), PromptId: c.PromptId,
		})
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}
		if id == 0 {
			// Existing row was already moved out of 'suggested' — left untouched.
			continue
		}
		data = append(data, RecommendationData{
			Id: id, CompanyId: companyId, Type: c.Type, Placement: c.Placement, Impact: c.Impact,
			Engine: "chatgpt", Text: text, Link: c.Link, LinkText: c.LinkText,
			Why: cp.Why, Steps: cp.Steps, PromptId: c.PromptId, Status: "suggested",
		})
	}

	if err := tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &RecommendationListResponse{Status: true, Desc: "Recommendations generated successfully", Data: data}, nil
}

func (s recommendationService) List(companyId int) (*RecommendationListResponse, error) {
	rows, err := s.recommendationRepository.GetByCompanyId(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	data := []RecommendationData{}
	for _, r := range rows {
		data = append(data, toData(r))
	}
	return &RecommendationListResponse{Status: true, Desc: "Get recommendations successful", Data: data}, nil
}

func (s recommendationService) UpdateStatus(id, userId int, status string) (*SimpleResponse, error) {
	if status != "suggested" && status != "todo" && status != "archived" {
		return nil, errs.NewBadRequestError("status must be 'suggested', 'todo', or 'archived'")
	}

	owned, err := s.recommendationRepository.BelongsToUser(id, userId)
	if err != nil || !owned {
		return nil, errs.NewForbiddenError("access denied")
	}

	tx, err := s.recommendationRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err := s.recommendationRepository.UpdateStatus(tx, id, status); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err := tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Recommendation updated successfully"}, nil
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
