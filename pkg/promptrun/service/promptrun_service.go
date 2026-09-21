package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	brandRepository "github.com/ai-marketing/ai-marketing-server/pkg/brand/repository"
	citationRepository "github.com/ai-marketing/ai-marketing-server/pkg/citation/repository"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	promptRepository "github.com/ai-marketing/ai-marketing-server/pkg/prompt/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/promptrun/repository"
	subdomainRepository "github.com/ai-marketing/ai-marketing-server/pkg/subdomain/repository"
	usageService "github.com/ai-marketing/ai-marketing-server/pkg/usage/service"
	visibilityRepository "github.com/ai-marketing/ai-marketing-server/pkg/visibility/repository"
	"github.com/ai-marketing/ai-marketing-server/providers/citation"
)

const extractionSystemPrompt = `You are a citation classification assistant. You are given the text of an AI chatbot's answer to a search query, a list of EVERY real source URL the AI actually cited, and a list of brand names being tracked (including which one is the company's own brand).

You must return exactly one entry for EVERY URL in the "Real cited source URLs" list — do not skip any, and do not invent URLs that aren't in that list. Most cited URLs will have nothing to do with any tracked brand, and that's expected — classify them anyway.

For each URL, determine:
- "source_type": what KIND of site it is, using your knowledge of the domain plus its title/context:
  - "editorial": an independent blog, review site, "best of" listicle, or niche publication — content someone wrote and could realistically be pitched for placement (this is the most actionable category for marketing outreach)
  - "directory": a structured business listing/map site (Google Maps, Yelp-style directories) — not editorial content
  - "reference": an encyclopedic/reference site (Wikipedia, Britannica-style)
  - "social": a forum, social media platform, or Q&A site (Reddit, Quora, Facebook)
  - "news": a news publication or press release
  - "marketplace": an e-commerce marketplace listing (Amazon, Lazada, Shopee)
  - "official_site": the brand's own official website/domain
  - "other": anything that doesn't clearly fit the above — including sites unrelated to any tracked brand (educational, government, generic reference pages, etc.)
- whether the answer text discusses this URL in connection with one of the TRACKED brands specifically. If so:
  - "brand_name": the exact tracked brand name
  - "ranking": its approximate ranking/position among brand mentions (1 = mentioned first or most prominently)
  - "sentiment": "positive", "neutral", or "negative"
  - "snippet": a short one-sentence quote or paraphrase of the relevant part of the text
  If this URL is NOT connected to any tracked brand, set "brand_name" to an empty string, "ranking" to 0, "sentiment" to "neutral", and "snippet" to a short one-sentence description of what the page appears to be about instead.
- "target_country": the ISO 3166-1 alpha-2 code (e.g. "TH", "US", "JP") of the market this specific page is targeting, ONLY when you have a real signal for it — a locale segment in the URL path or subdomain (e.g. "/th/", "th.example.com"), the page title/snippet being in a specific country's language, or content that's explicitly about that country. This is only asked for URLs on generic domains (.com, .org, .io, etc. — a country-coded domain's market is already known from the domain itself). If you have no genuine signal either way, return an empty string — do NOT guess from the brand's general market or default to any particular country.

Respond with ONLY a JSON array, no prose, no markdown fences, matching exactly:
[{"url": "one of the given cited URLs, copied exactly", "source_type": "editorial"|"directory"|"reference"|"social"|"news"|"marketplace"|"official_site"|"other", "brand_name": "exact tracked brand name, or empty string", "ranking": int, "sentiment": "positive"|"neutral"|"negative", "snippet": "short quote or description", "target_country": "ISO 3166-1 alpha-2 code, or empty string"}]`

type extractedMention struct {
	Url           string `json:"url"`
	SourceType    string `json:"source_type"`
	BrandName     string `json:"brand_name"`
	Ranking       int    `json:"ranking"`
	Sentiment     string `json:"sentiment"`
	Snippet       string `json:"snippet"`
	TargetCountry string `json:"target_country"`
}

var validSourceTypes = map[string]bool{
	"editorial": true, "directory": true, "reference": true, "social": true,
	"news": true, "marketplace": true, "official_site": true, "other": true,
}

type promptRunService struct {
	promptRunRepository  repository.PromptRunRepository
	promptRepository     promptRepository.PromptRepository
	brandRepository      brandRepository.BrandRepository
	citationRepository   citationRepository.CitationRepository
	visibilityRepository visibilityRepository.VisibilityRepository
	companyRepository    companyRepository.CompanyRepository
	subdomainRepository  subdomainRepository.SubdomainRepository
	engines              []Engine
	extractionProvider   ExtractionProvider
	usageService         usageService.UsageService
}

func NewPromptRunService(
	promptRunRepository repository.PromptRunRepository,
	promptRepository promptRepository.PromptRepository,
	brandRepository brandRepository.BrandRepository,
	citationRepository citationRepository.CitationRepository,
	visibilityRepository visibilityRepository.VisibilityRepository,
	companyRepository companyRepository.CompanyRepository,
	subdomainRepository subdomainRepository.SubdomainRepository,
	engines []Engine,
	extractionProvider ExtractionProvider,
	usageService usageService.UsageService,
) PromptRunService {
	return promptRunService{
		promptRunRepository, promptRepository, brandRepository, citationRepository,
		visibilityRepository, companyRepository, subdomainRepository, engines, extractionProvider, usageService,
	}
}

func toPromptRunData(r repository.PromptRun) PromptRunData {
	return PromptRunData{
		Id:          r.Id,
		PromptId:    r.PromptId,
		AiPlatform:  r.AiPlatform,
		Model:       r.Model,
		RawResponse: r.RawResponse,
		CreatedAt:   r.CreatedAt,
	}
}

// enabledEngines filters the fully-configured engine list down to whichever
// ones this company has turned on in AI Setting. A platform with no row yet
// in company_ai_engines defaults to enabled, so existing companies keep
// running every engine until they explicitly opt out.
func (s promptRunService) enabledEngines(companyId int) ([]Engine, error) {
	rows, err := s.companyRepository.GetAiEngines(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	disabled := map[string]bool{}
	for _, r := range rows {
		if !r.Enabled {
			disabled[r.Platform] = true
		}
	}

	filtered := []Engine{}
	for _, eng := range s.engines {
		if !disabled[eng.Platform] {
			filtered = append(filtered, eng)
		}
	}
	return filtered, nil
}

func (s promptRunService) verifyOwnership(promptId, userId int) error {
	owned, err := s.promptRepository.BelongsToUser(promptId, userId)
	if err != nil || !owned {
		return errs.NewForbiddenError("access denied")
	}
	return nil
}

func (s promptRunService) Run(promptId, userId int) (*PromptRunListResponse, error) {
	if err := s.verifyOwnership(promptId, userId); err != nil {
		return nil, err
	}
	return s.run(promptId)
}

func (s promptRunService) RunSystem(promptId int) (*PromptRunListResponse, error) {
	return s.run(promptId)
}

// run fans out the prompt across every configured engine (e.g. ChatGPT and
// Gemini run one after another). Each engine is best-effort: one failing
// doesn't stop the others — the result includes whichever runs succeeded.
// promptCountry is the prompt's single country code. Prompts saved before
// the one-country rule may still list several; the first (alphabetical) wins.
func promptCountry(countries string) string {
	code, _, _ := strings.Cut(countries, ",")
	return strings.ToUpper(strings.TrimSpace(code))
}

func (s promptRunService) run(promptId int) (*PromptRunListResponse, error) {
	prompt, err := s.promptRepository.GetById(promptId)
	if err != nil {
		return nil, errs.NewNotFoundError("prompt not found")
	}
	if !prompt.Active {
		return nil, errs.NewBadRequestError("this prompt is paused — reactivate it in Prompts to run it")
	}

	company, err := s.companyRepository.GetById(prompt.CompanyId)
	if err != nil {
		return nil, errs.NewNotFoundError("company not found")
	}
	if company.ContractEndDate != nil {
		if end, err := time.Parse("2006-01-02", *company.ContractEndDate); err == nil && time.Now().After(end) {
			return nil, errs.NewBadRequestError("this workspace's contract has ended — renew it in Settings to resume prompt runs")
		}
	}

	if len(s.engines) == 0 {
		return nil, errs.NewUnexpectedError()
	}

	engines, err := s.enabledEngines(prompt.CompanyId)
	if err != nil {
		return nil, err
	}
	if len(engines) == 0 {
		return nil, errs.NewBadRequestError("no AI engines are enabled for this company — turn at least one on in AI Setting")
	}

	data := []PromptRunData{}
	for _, eng := range engines {
		response, model, citations, tokenUsage, err := eng.Provider.Complete(prompt.Content, promptCountry(prompt.Countries))
		if err != nil {
			logs.Error(fmt.Errorf("%s run failed for prompt %d: %w", eng.Platform, promptId, err))
			continue
		}

		if err := s.usageService.Log(&prompt.CompanyId, eng.Platform, "prompt_run", model, tokenUsage); err != nil {
			logs.Error(fmt.Errorf("failed to log usage for %s run on prompt %d: %w", eng.Platform, promptId, err))
		}

		run, err := s.promptRunRepository.Create(promptId, eng.Platform, model, response)
		if err != nil {
			logs.Error(err)
			continue
		}

		// Extraction is best-effort: the raw response is already saved, so a
		// failure here shouldn't fail the whole run — just log and move on.
		if err := s.extractCitations(prompt, run, eng.Platform, response, citations); err != nil {
			logs.Error(fmt.Errorf("citation extraction failed for prompt %d (%s): %w", promptId, eng.Platform, err))
		}

		data = append(data, toPromptRunData(run))
	}

	if len(data) == 0 {
		return nil, errs.NewUnexpectedError()
	}

	if err := s.citationRepository.SnapshotBrandCoverage(prompt.CompanyId); err != nil {
		logs.Error(fmt.Errorf("failed to snapshot brand coverage for company %d: %w", prompt.CompanyId, err))
	}

	return &PromptRunListResponse{Status: true, Desc: "Prompt run successful", Data: data}, nil
}

func (s promptRunService) extractCitations(
	prompt *promptRepository.Prompt,
	run repository.PromptRun,
	platform string,
	responseText string,
	sourceCitations []citation.Citation,
) error {
	if len(sourceCitations) == 0 {
		return nil
	}

	brands, err := s.brandRepository.GetAll(prompt.CompanyId)
	if err != nil {
		return err
	}

	// Best-effort: a lookup failure here shouldn't fail extraction, it just
	// means the official_site domain check below is skipped. ownDomains
	// combines the primary website with any additional domains (regional
	// sites like ikea.co.id that aren't subdomains of the primary one) so a
	// citation on any of them gets recognized as official too.
	var ownDomains []string
	if company, err := s.companyRepository.GetById(prompt.CompanyId); err != nil {
		logs.Error(fmt.Errorf("failed to load company %d for citation domain matching: %w", prompt.CompanyId, err))
	} else if company.Website != nil && *company.Website != "" {
		ownDomains = append(ownDomains, *company.Website)
	}
	if extra, err := s.subdomainRepository.GetAll(prompt.CompanyId); err != nil {
		logs.Error(fmt.Errorf("failed to load additional domains for company %d: %w", prompt.CompanyId, err))
	} else {
		for _, d := range extra {
			ownDomains = append(ownDomains, d.Subdomain)
		}
	}

	userPrompt := buildExtractionPrompt(responseText, sourceCitations, brands, ownDomains)
	raw, extractionModel, tokenUsage, err := s.extractionProvider.Complete(extractionSystemPrompt, userPrompt)
	if err != nil {
		return err
	}

	if err := s.usageService.Log(&prompt.CompanyId, "claude", "citation_extraction", extractionModel, tokenUsage); err != nil {
		logs.Error(fmt.Errorf("failed to log usage for citation extraction on prompt %d: %w", prompt.Id, err))
	}

	var parsed []extractedMention
	if err := json.Unmarshal([]byte(extractJSONArray(raw)), &parsed); err != nil {
		return fmt.Errorf("unparseable extraction response: %w (raw: %s)", err, raw)
	}

	byUrl := map[string]extractedMention{}
	for _, m := range parsed {
		byUrl[strings.TrimSpace(m.Url)] = m
	}

	tx, err := s.citationRepository.NewTransaction()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	today := time.Now().Format("2006-01-02")
	runId := run.Id

	// Drive the loop off sourceCitations — the real, verified URL list
	// straight from OpenAI's web_search annotations — not off Claude's
	// output. That way every real citation gets a row even if Claude's
	// classification pass drops or mismatches one; Claude only supplies the
	// source_type/brand classification per URL, it can't cause a citation to
	// go missing.
	for i, sc := range sourceCitations {
		url := sc.URL
		if strings.TrimSpace(url) == "" {
			continue
		}
		m, matched := byUrl[strings.TrimSpace(url)]

		ranking := i + 1
		sentiment := "neutral"
		snippet := sc.Title
		sourceType := "other"
		var brand *brandRepository.Brand
		if matched {
			if m.Ranking > 0 {
				ranking = m.Ranking
			}
			s := strings.ToLower(strings.TrimSpace(m.Sentiment))
			if s == "positive" || s == "negative" {
				sentiment = s
			}
			if strings.TrimSpace(m.Snippet) != "" {
				snippet = m.Snippet
			}
			st := strings.ToLower(strings.TrimSpace(m.SourceType))
			if validSourceTypes[st] {
				sourceType = st
			}
			brand = matchBrand(brands, m.BrandName)
		}

		// Domain match is verified ground truth (the company's own saved
		// website vs. the real cited URL) — it overrides the model's guess
		// rather than deferring to it, same as the rest of this pipeline
		// trusts real provider data over LLM output wherever both exist.
		if isOwnDomain(ownDomains, url) {
			sourceType = "official_site"
			if own := ownBrand(brands); own != nil {
				brand = own
			}
		}

		// Country: ccTLD is a deterministic, verified signal and always wins
		// when present (same "real data beats model guess" rule as above).
		// Only defer to the model's target_country guess for generic TLDs,
		// where there's no domain-level signal to check.
		targetCountry := detectCountryFromTLD(url)
		if targetCountry == "" && matched {
			if tc := strings.ToUpper(strings.TrimSpace(m.TargetCountry)); isValidCountryCode(tc) {
				targetCountry = tc
			}
		}
		var targetCountryPtr *string
		if targetCountry != "" {
			targetCountryPtr = &targetCountry
		}

		var brandIdPtr *int
		isCompetitor := false
		brandPositioning := "not_mentioned"
		if brand != nil {
			bid := brand.Id
			brandIdPtr = &bid
			isCompetitor = !brand.IsOwn
			brandPositioning = "mentioned"
		}

		if _, err := s.citationRepository.Upsert(tx, citationRepository.Citation{
			PromptId:          prompt.Id,
			BrandId:           brandIdPtr,
			PromptRunId:       &runId,
			Ranking:           ranking,
			Content:           responseText,
			Url:               &url,
			AiPlatform:        &platform,
			DateDiscovered:    &today,
			LastChecked:       &today,
			Sentiment:         sentiment,
			BrandPositioning:  brandPositioning,
			CitationFrequency: 1,
			Snippet:           &snippet,
			IsCompetitor:      isCompetitor,
			SourceType:        &sourceType,
			TargetCountry:     targetCountryPtr,
		}); err != nil {
			return err
		}

		if err := s.citationRepository.BumpDailyStat(tx, prompt.CompanyId, url, today); err != nil {
			return err
		}

		if brand != nil {
			if _, err := s.visibilityRepository.Create(tx, visibilityRepository.Visibility{
				BrandId:  brand.Id,
				PromptId: &prompt.Id,
				Platform: platform,
				Score:    sentimentScore(sentiment),
				Mentions: 1,
				Date:     today,
			}); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func buildExtractionPrompt(responseText string, citations []citation.Citation, brands []brandRepository.Brand, ownDomains []string) string {
	var b strings.Builder
	b.WriteString("AI answer text:\n")
	b.WriteString(responseText)

	b.WriteString("\n\nReal cited source URLs:\n")
	if len(citations) == 0 {
		b.WriteString("(none)\n")
	}
	for _, c := range citations {
		fmt.Fprintf(&b, "- %s (%s)\n", c.URL, c.Title)
	}

	b.WriteString("\nTracked brands:\n")
	for _, br := range brands {
		own := ""
		if br.IsOwn {
			own = " (this is the company's own brand)"
		}
		fmt.Fprintf(&b, "- %s%s\n", br.Name, own)
	}

	if len(ownDomains) > 0 {
		fmt.Fprintf(&b, "\nThe company's own official domain(s): %s — any cited URL on one of these domains (or a subdomain of one) must be classified as \"official_site\".\n", strings.Join(ownDomains, ", "))
	}

	return b.String()
}

// normalizeDomain strips scheme, "www.", path/query, and port so different
// ways of writing the same site (https://www.ikea.com/th/en vs ikea.com)
// compare equal.
func normalizeDomain(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return ""
	}
	return strings.ToLower(strings.TrimPrefix(u.Hostname(), "www."))
}

// isOwnDomain reports whether citedURL's host matches one of the company's
// saved domains (primary website or an additional domain), or a subdomain of
// one of them.
func isOwnDomain(ownDomains []string, citedURL string) bool {
	cited := normalizeDomain(citedURL)
	if cited == "" {
		return false
	}
	for _, raw := range ownDomains {
		domain := normalizeDomain(raw)
		if domain == "" {
			continue
		}
		if cited == domain || strings.HasSuffix(cited, "."+domain) {
			return true
		}
	}
	return false
}

func ownBrand(brands []brandRepository.Brand) *brandRepository.Brand {
	for i := range brands {
		if brands[i].IsOwn {
			return &brands[i]
		}
	}
	return nil
}

func matchBrand(brands []brandRepository.Brand, name string) *brandRepository.Brand {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return nil
	}
	for i := range brands {
		if strings.ToLower(brands[i].Name) == name {
			return &brands[i]
		}
	}
	return nil
}

func sentimentScore(sentiment string) float64 {
	switch sentiment {
	case "positive":
		return 100
	case "negative":
		return 20
	default:
		return 60
	}
}

// extractJSONArray trims any surrounding prose/markdown fences the model may
// add despite instructions, keeping only the bracketed JSON array.
func extractJSONArray(s string) string {
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start == -1 || end == -1 || end < start {
		return s
	}
	return s[start : end+1]
}

func (s promptRunService) GetHistory(promptId, userId int) (*PromptRunListResponse, error) {
	if err := s.verifyOwnership(promptId, userId); err != nil {
		return nil, err
	}

	runs, err := s.promptRunRepository.GetByPromptId(promptId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []PromptRunData{}
	for _, r := range runs {
		data = append(data, toPromptRunData(r))
	}

	return &PromptRunListResponse{Status: true, Desc: "Get prompt run history successful", Data: data}, nil
}

func (s promptRunService) RunAllForCompany(companyId int) (*RunAllResponse, error) {
	prompts, err := s.promptRepository.GetAll(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	ran, failed := 0, 0
	for i, p := range prompts {
		if !p.Active {
			continue
		}
		if _, err := s.run(p.Id); err != nil {
			logs.Error(fmt.Errorf("scheduled run failed for prompt %d: %w", p.Id, err))
			failed++
		} else {
			ran++
		}
		if i < len(prompts)-1 {
			time.Sleep(2 * time.Second)
		}
	}

	return &RunAllResponse{Status: true, Desc: "Run all completed", Ran: ran, Failed: failed}, nil
}
