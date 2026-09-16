package repository

type DashboardStats struct {
	TotalVisibility   float64 `db:"total_visibility"`
	SearchImpressions int     `db:"search_impressions"`
	BrandMentions     int     `db:"brand_mentions"`
	VisibilityScore   float64 `db:"visibility_score"`
}

type TopPrompt struct {
	PromptId      int    `db:"prompt_id"`
	Title         string `db:"title"`
	CitationCount int    `db:"citation_count"`
}

type TopDomain struct {
	Domain       string `db:"domain"`
	MentionCount int    `db:"mention_count"`
}

type RecentCitation struct {
	Id         int     `db:"id"`
	PromptId   int     `db:"prompt_id"`
	BrandId    *int    `db:"brand_id"`
	Ranking    int     `db:"ranking"`
	Content    string  `db:"content"`
	Url        *string `db:"url"`
	AiPlatform *string `db:"ai_platform"`
	CreatedAt  string  `db:"created_at"`
}

type VisibilityTrendPoint struct {
	Period   string  `db:"period"`
	Score    float64 `db:"score"`
	Mentions int     `db:"mentions"`
}

type PlatformBreakdown struct {
	Platform string  `db:"platform"`
	Score    float64 `db:"score"`
	Mentions int     `db:"mentions"`
}

type CompanyMetrics struct {
	BrandCount  int `db:"brand_count"`
	PromptCount int `db:"prompt_count"`
}

type PromptRanking struct {
	Ranking          int    `db:"ranking"`
	Brand            string `db:"brand"`
	Sentiment        string `db:"sentiment"`
	BrandPositioning string `db:"brand_positioning"`
	IsCompetitor     bool   `db:"is_competitor"`
	Visibility       int    `db:"visibility"`
}

type PromptDomain struct {
	Domain       string  `db:"domain"`
	MentionCount int     `db:"mention_count"`
	AvgCitation  float64 `db:"avg_citation"`
	IsCompetitor bool    `db:"is_competitor"`
	Snippet      string  `db:"snippet"`
	SourceType   string  `db:"source_type"`
}

type PromptOverview struct {
	PromptId             int    `db:"prompt_id"`
	Title                string `db:"title"`
	Tag                  string `db:"tag"`
	BrandCoverage        int    `db:"brand_coverage"`
	BrandSentiment       int    `db:"brand_sentiment"`
	BrandMentions        int    `db:"brand_mentions"`
	TotalBrandMentions   int    `db:"total_brand_mentions"`
	DomainCitations      int    `db:"domain_citations"`
	TotalDomainCitations int    `db:"total_domain_citations"`
	Competitors          string `db:"competitors"`
	Countries            string `db:"countries"`
	Active               bool   `db:"active"`
}

type BrandRankingRow struct {
	Id             int     `db:"id"`
	Rank           int     `db:"rank"`
	Name           string  `db:"name"`
	IsOwn          bool    `db:"is_own"`
	SentimentScore int     `db:"sentiment_score"`
	Mentions       int     `db:"mentions"`
	BrandCoverage  float64 `db:"brand_coverage"`
	ShareOfVoice   float64 `db:"share_of_voice"`
	AvgPosition    float64 `db:"avg_position"`
}

// BrandCitation is one URL that cited a specific brand — the real
// replacement for BrandDetail's old hardcoded "Visibility"/"Content" mock
// data. Engines is comma-joined since a URL can be cited by more than one.
type BrandCitation struct {
	Url      string `db:"url"`
	Title    string `db:"title"`
	Domain   string `db:"domain"`
	Engines  string `db:"engines"`
	Cited    int    `db:"cited"`
	LastSeen string `db:"last_seen"`
}

type TopPromptByBrand struct {
	PromptId        int    `db:"prompt_id"`
	Title           string `db:"title"`
	MyBrandMentions int    `db:"my_brand_mentions"`
}

type CitationURL struct {
	Rank          int     `db:"rank"`
	Url           string  `db:"url"`
	CitationShare float64 `db:"citation_share"`
	CitationCount int     `db:"citation_count"`
}

type CitationURLDetail struct {
	Url            string `db:"url"`
	Title          string `db:"title"`
	BrandMentioned bool   `db:"brand_mentioned"`
	Competitors    string `db:"competitors"`
	Domain         string `db:"domain"`
	DomainCategory string `db:"domain_category"`
	SourceType     string `db:"source_type"`
	Cited          int    `db:"cited"`
	// Engines is a comma-joined list of every ai_platform that has cited
	// this URL (e.g. "chatgpt,claude") — a URL can be cited by more than one.
	Engines string `db:"engines"`
	// Tags is a comma-joined list of the distinct prompt category/tag names
	// across every prompt that cited this URL.
	Tags string `db:"tags"`
	// TargetCountry is an ISO 3166-1 alpha-2 code, empty when undetermined —
	// see citations.target_country / detectCountryFromTLD.
	TargetCountry string `db:"target_country"`
}

type CitationURLPrompt struct {
	PromptId int    `db:"prompt_id"`
	Title    string `db:"title"`
	// Engines is a comma-joined list of the distinct ai_platform values that
	// cited this URL when this prompt ran.
	Engines string `db:"engines"`
	// Sentiment and Ranking are the best-ranked citation row's own values for
	// this (prompt, url) pair — same "pick the top-ranked row as
	// representative" idiom GetPromptRankings uses, not a synthetic average.
	Sentiment string `db:"sentiment"`
	Ranking   int    `db:"ranking"`
	// CitationFrequency sums citation_frequency across every engine's
	// citation row for this (prompt, url) pair.
	CitationFrequency int `db:"citation_frequency"`
}

// CitationURLChange is one URL's citation count in two comparison windows —
// the raw input to Top Winners/Losers. Percent-change and New/Dropped
// labeling happens in the service layer, not here.
type CitationURLChange struct {
	Url           string `db:"url"`
	Title         string `db:"title"`
	CurrentCount  int    `db:"current_count"`
	PreviousCount int    `db:"previous_count"`
}

type DashboardRepository interface {
	GetStats(companyId int) (*DashboardStats, error)
	GetTopPrompts(companyId int) ([]TopPrompt, error)
	GetTopDomains(companyId int) ([]TopDomain, error)
	GetRecentCitations(companyId int) ([]RecentCitation, error)
	GetVisibilityTrend(companyId int, interval string) ([]VisibilityTrendPoint, error)
	GetPlatformBreakdown(companyId int) ([]PlatformBreakdown, error)
	GetPromptTrend(promptId, companyId int, interval, from, to string) ([]VisibilityTrendPoint, error)
	GetCompanyMetrics(companyId int) (*CompanyMetrics, error)
	GetPromptRankings(promptId, companyId int, from, to string) ([]PromptRanking, error)
	GetPromptDomains(promptId, companyId int, from, to string) ([]PromptDomain, error)
	GetBrandRanking(companyId int) ([]BrandRankingRow, error)
	GetTopPromptsByBrand(companyId int) ([]TopPromptByBrand, error)
	GetTopCitationURLs(companyId int) ([]CitationURL, error)
	GetPromptsOverview(companyId int, from, to string) ([]PromptOverview, error)
	GetCitationURLs(companyId int) ([]CitationURLDetail, error)
	GetCitationURLPrompts(url string, companyId int) ([]CitationURLPrompt, error)
	GetCitationURLChanges(companyId int, currentFrom, currentTo, previousFrom, previousTo string) ([]CitationURLChange, error)
	GetBrandCitations(companyId, brandId int) ([]BrandCitation, error)
}
