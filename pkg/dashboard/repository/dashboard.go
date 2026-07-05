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
	Ranking         int     `db:"ranking"`
	Brand           string  `db:"brand"`
	Sentiment       string  `db:"sentiment"`
	BrandPositioning string  `db:"brand_positioning"`
	IsCompetitor    bool    `db:"is_competitor"`
	Visibility      int     `db:"visibility"`
}

type PromptDomain struct {
	Domain       string  `db:"domain"`
	MentionCount int     `db:"mention_count"`
	AvgCitation  float64 `db:"avg_citation"`
	IsCompetitor bool    `db:"is_competitor"`
	Snippet      string  `db:"snippet"`
}

type PromptOverview struct {
	PromptId             int    `db:"prompt_id"`
	Title                string `db:"title"`
	Category             string `db:"category"`
	BrandCoverage        int    `db:"brand_coverage"`
	BrandSentiment       int    `db:"brand_sentiment"`
	BrandMentions        int    `db:"brand_mentions"`
	TotalBrandMentions   int    `db:"total_brand_mentions"`
	DomainCitations      int    `db:"domain_citations"`
	TotalDomainCitations int    `db:"total_domain_citations"`
	Competitors          string `db:"competitors"`
}

type BrandRankingRow struct {
	Rank           int     `db:"rank"`
	Name           string  `db:"name"`
	IsOwn          bool    `db:"is_own"`
	SentimentScore int     `db:"sentiment_score"`
	Mentions       int     `db:"mentions"`
	BrandCoverage  float64 `db:"brand_coverage"`
	ShareOfVoice   float64 `db:"share_of_voice"`
	AvgPosition    float64 `db:"avg_position"`
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
	Cited          int    `db:"cited"`
}

type CitationURLPrompt struct {
	PromptId int    `db:"prompt_id"`
	Title    string `db:"title"`
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
	GetPromptRankings(promptId, companyId int) ([]PromptRanking, error)
	GetPromptDomains(promptId, companyId int) ([]PromptDomain, error)
	GetBrandRanking(companyId int) ([]BrandRankingRow, error)
	GetTopPromptsByBrand(companyId int) ([]TopPromptByBrand, error)
	GetTopCitationURLs(companyId int) ([]CitationURL, error)
	GetPromptsOverview(companyId int) ([]PromptOverview, error)
	GetCitationURLs(companyId int) ([]CitationURLDetail, error)
	GetCitationURLPrompts(url string, companyId int) ([]CitationURLPrompt, error)
}
