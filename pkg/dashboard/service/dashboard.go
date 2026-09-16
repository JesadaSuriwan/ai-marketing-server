package service

type StatsData struct {
	TotalVisibility   float64 `json:"total_visibility"`
	SearchImpressions int     `json:"search_impressions"`
	BrandMentions     int     `json:"brand_mentions"`
	VisibilityScore   float64 `json:"visibility_score"`
}

type TopPromptData struct {
	PromptId      int    `json:"prompt_id"`
	Title         string `json:"title"`
	CitationCount int    `json:"citation_count"`
}

type TopDomainData struct {
	Domain       string `json:"domain"`
	MentionCount int    `json:"mention_count"`
}

type RecentCitationData struct {
	Id         int     `json:"id"`
	PromptId   int     `json:"prompt_id"`
	BrandId    *int    `json:"brand_id"`
	Ranking    int     `json:"ranking"`
	Content    string  `json:"content"`
	Url        *string `json:"url"`
	AiPlatform *string `json:"ai_platform"`
	CreatedAt  string  `json:"created_at"`
}

type StatsResponse struct {
	Status bool      `json:"status"`
	Desc   string    `json:"desc"`
	Data   StatsData `json:"data"`
}

type TopPromptsResponse struct {
	Status bool            `json:"status"`
	Desc   string          `json:"desc"`
	Data   []TopPromptData `json:"data"`
}

type TopDomainsResponse struct {
	Status bool            `json:"status"`
	Desc   string          `json:"desc"`
	Data   []TopDomainData `json:"data"`
}

type RecentCitationsResponse struct {
	Status bool                 `json:"status"`
	Desc   string               `json:"desc"`
	Data   []RecentCitationData `json:"data"`
}

type VisibilityTrendPoint struct {
	Period   string  `json:"period"`
	Score    float64 `json:"score"`
	Mentions int     `json:"mentions"`
}

type PlatformBreakdownData struct {
	Platform string  `json:"platform"`
	Score    float64 `json:"score"`
	Mentions int     `json:"mentions"`
}

type VisibilityTrendResponse struct {
	Status bool                   `json:"status"`
	Desc   string                 `json:"desc"`
	Data   []VisibilityTrendPoint `json:"data"`
}

type PlatformBreakdownResponse struct {
	Status bool                    `json:"status"`
	Desc   string                  `json:"desc"`
	Data   []PlatformBreakdownData `json:"data"`
}

type CompanyMetricsData struct {
	BrandCount  int `json:"brand_count"`
	PromptCount int `json:"prompt_count"`
}

type CompanyMetricsResponse struct {
	Status bool               `json:"status"`
	Desc   string             `json:"desc"`
	Data   CompanyMetricsData `json:"data"`
}

type PromptRankingData struct {
	Ranking          int    `json:"ranking"`
	Brand            string `json:"brand"`
	Sentiment        string `json:"sentiment"`
	BrandPositioning string `json:"brand_positioning"`
	IsCompetitor     bool   `json:"is_competitor"`
	Visibility       int    `json:"visibility"`
}

type PromptDomainData struct {
	Domain       string  `json:"domain"`
	MentionCount int     `json:"mention_count"`
	AvgCitation  float64 `json:"avg_citation"`
	IsCompetitor bool    `json:"is_competitor"`
	Snippet      string  `json:"snippet"`
	SourceType   string  `json:"source_type"`
}

type PromptRankingsResponse struct {
	Status bool                `json:"status"`
	Desc   string              `json:"desc"`
	Data   []PromptRankingData `json:"data"`
}

type PromptDomainsResponse struct {
	Status bool               `json:"status"`
	Desc   string             `json:"desc"`
	Data   []PromptDomainData `json:"data"`
}

type PromptOverviewData struct {
	PromptId             int    `json:"prompt_id"`
	Title                string `json:"title"`
	Tag                  string `json:"tag"`
	BrandCoverage        int    `json:"brand_coverage"`
	BrandSentiment       int    `json:"brand_sentiment"`
	BrandMentions        int    `json:"brand_mentions"`
	TotalBrandMentions   int    `json:"total_brand_mentions"`
	DomainCitations      int    `json:"domain_citations"`
	TotalDomainCitations int    `json:"total_domain_citations"`
	Competitors          string `json:"competitors"`
	Countries            string `json:"countries"`
	Active               bool   `json:"active"`
}

type PromptsOverviewResponse struct {
	Status bool                 `json:"status"`
	Desc   string               `json:"desc"`
	Data   []PromptOverviewData `json:"data"`
}

type BrandRankingData struct {
	Id             int     `json:"id"`
	Rank           int     `json:"rank"`
	Name           string  `json:"name"`
	IsOwn          bool    `json:"is_own"`
	SentimentScore int     `json:"sentiment_score"`
	Mentions       int     `json:"mentions"`
	BrandCoverage  float64 `json:"brand_coverage"`
	ShareOfVoice   float64 `json:"share_of_voice"`
	AvgPosition    float64 `json:"avg_position"`
}

type BrandCitationData struct {
	Url      string `json:"url"`
	Title    string `json:"title"`
	Domain   string `json:"domain"`
	Engines  string `json:"engines"`
	Cited    int    `json:"cited"`
	LastSeen string `json:"last_seen"`
}

type BrandCitationsResponse struct {
	Status bool                `json:"status"`
	Desc   string              `json:"desc"`
	Data   []BrandCitationData `json:"data"`
}

type BrandRankingResponse struct {
	Status bool               `json:"status"`
	Desc   string             `json:"desc"`
	Data   []BrandRankingData `json:"data"`
}

type TopPromptByBrandData struct {
	PromptId        int    `json:"prompt_id"`
	Title           string `json:"title"`
	MyBrandMentions int    `json:"my_brand_mentions"`
}

type TopPromptsByBrandResponse struct {
	Status bool                   `json:"status"`
	Desc   string                 `json:"desc"`
	Data   []TopPromptByBrandData `json:"data"`
}

type CitationURLData struct {
	Rank          int     `json:"rank"`
	Url           string  `json:"url"`
	CitationShare float64 `json:"citation_share"`
	CitationCount int     `json:"citation_count"`
}

type TopCitationURLsResponse struct {
	Status bool              `json:"status"`
	Desc   string            `json:"desc"`
	Data   []CitationURLData `json:"data"`
}

type CitationURLDetailData struct {
	Url            string `json:"url"`
	Title          string `json:"title"`
	BrandMentioned bool   `json:"brand_mentioned"`
	Competitors    string `json:"competitors"`
	Domain         string `json:"domain"`
	DomainCategory string `json:"domain_category"`
	SourceType     string `json:"source_type"`
	Cited          int    `json:"cited"`
	Engines        string `json:"engines"`
	Tags           string `json:"tags"`
	TargetCountry  string `json:"target_country"`
}

type CitationURLsResponse struct {
	Status bool                    `json:"status"`
	Desc   string                  `json:"desc"`
	Data   []CitationURLDetailData `json:"data"`
}

type CitationChangeData struct {
	Url           string  `json:"url"`
	Title         string  `json:"title"`
	CurrentCount  int     `json:"current_count"`
	PreviousCount int     `json:"previous_count"`
	ChangePct     float64 `json:"change_pct"`
	IsNew         bool    `json:"is_new"`
	IsDropped     bool    `json:"is_dropped"`
}

type CitationWinnersLosersData struct {
	Winners []CitationChangeData `json:"winners"`
	Losers  []CitationChangeData `json:"losers"`
}

type CitationWinnersLosersResponse struct {
	Status bool                      `json:"status"`
	Desc   string                    `json:"desc"`
	Data   CitationWinnersLosersData `json:"data"`
}

type CitationURLPromptData struct {
	PromptId          int    `json:"prompt_id"`
	Title             string `json:"title"`
	Engines           string `json:"engines"`
	Sentiment         string `json:"sentiment"`
	Ranking           int    `json:"ranking"`
	CitationFrequency int    `json:"citation_frequency"`
}

type CitationURLPromptsResponse struct {
	Status bool                    `json:"status"`
	Desc   string                  `json:"desc"`
	Data   []CitationURLPromptData `json:"data"`
}

type DashboardService interface {
	GetStats(companyId int) (*StatsResponse, error)
	GetTopPrompts(companyId int) (*TopPromptsResponse, error)
	GetTopDomains(companyId int) (*TopDomainsResponse, error)
	GetRecentCitations(companyId int) (*RecentCitationsResponse, error)
	GetVisibilityTrend(companyId int, interval string) (*VisibilityTrendResponse, error)
	GetPlatformBreakdown(companyId int) (*PlatformBreakdownResponse, error)
	GetPromptTrend(promptId, companyId int, interval, from, to string) (*VisibilityTrendResponse, error)
	GetCompanyMetrics(companyId int) (*CompanyMetricsResponse, error)
	GetPromptRankings(promptId, companyId int, from, to string) (*PromptRankingsResponse, error)
	GetPromptDomains(promptId, companyId int, from, to string) (*PromptDomainsResponse, error)
	GetBrandRanking(companyId int) (*BrandRankingResponse, error)
	GetTopPromptsByBrand(companyId int) (*TopPromptsByBrandResponse, error)
	GetTopCitationURLs(companyId int) (*TopCitationURLsResponse, error)
	GetPromptsOverview(companyId int, from, to string) (*PromptsOverviewResponse, error)
	GetCitationURLs(companyId int) (*CitationURLsResponse, error)
	GetCitationURLPrompts(url string, companyId int) (*CitationURLPromptsResponse, error)
	GetCitationWinnersLosers(companyId int) (*CitationWinnersLosersResponse, error)
	GetBrandCitations(companyId, brandId int) (*BrandCitationsResponse, error)
}
