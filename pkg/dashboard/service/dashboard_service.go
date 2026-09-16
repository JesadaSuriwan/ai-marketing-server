package service

import (
	"sort"
	"time"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/dashboard/repository"
)

type dashboardService struct {
	dashboardRepository repository.DashboardRepository
}

func NewDashboardService(dashboardRepository repository.DashboardRepository) DashboardService {
	return dashboardService{dashboardRepository}
}

func (s dashboardService) GetStats(companyId int) (*StatsResponse, error) {
	stats, err := s.dashboardRepository.GetStats(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &StatsResponse{
		Status: true,
		Desc:   "Get stats successful",
		Data: StatsData{
			TotalVisibility:   stats.TotalVisibility,
			SearchImpressions: stats.SearchImpressions,
			BrandMentions:     stats.BrandMentions,
			VisibilityScore:   stats.VisibilityScore,
		},
	}, nil
}

func (s dashboardService) GetTopPrompts(companyId int) (*TopPromptsResponse, error) {
	prompts, err := s.dashboardRepository.GetTopPrompts(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []TopPromptData{}
	for _, p := range prompts {
		data = append(data, TopPromptData{PromptId: p.PromptId, Title: p.Title, CitationCount: p.CitationCount})
	}

	return &TopPromptsResponse{Status: true, Desc: "Get top prompts successful", Data: data}, nil
}

func (s dashboardService) GetTopDomains(companyId int) (*TopDomainsResponse, error) {
	domains, err := s.dashboardRepository.GetTopDomains(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []TopDomainData{}
	for _, d := range domains {
		data = append(data, TopDomainData{Domain: d.Domain, MentionCount: d.MentionCount})
	}

	return &TopDomainsResponse{Status: true, Desc: "Get top domains successful", Data: data}, nil
}

func (s dashboardService) GetRecentCitations(companyId int) (*RecentCitationsResponse, error) {
	citations, err := s.dashboardRepository.GetRecentCitations(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []RecentCitationData{}
	for _, c := range citations {
		data = append(data, RecentCitationData{
			Id: c.Id, PromptId: c.PromptId, BrandId: c.BrandId, Ranking: c.Ranking,
			Content: c.Content, Url: c.Url, AiPlatform: c.AiPlatform, CreatedAt: c.CreatedAt,
		})
	}

	return &RecentCitationsResponse{Status: true, Desc: "Get recent citations successful", Data: data}, nil
}

func (s dashboardService) GetVisibilityTrend(companyId int, interval string) (*VisibilityTrendResponse, error) {
	points, err := s.dashboardRepository.GetVisibilityTrend(companyId, interval)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []VisibilityTrendPoint{}
	for _, p := range points {
		data = append(data, VisibilityTrendPoint{Period: p.Period, Score: p.Score, Mentions: p.Mentions})
	}

	return &VisibilityTrendResponse{Status: true, Desc: "Get visibility trend successful", Data: data}, nil
}

func (s dashboardService) GetPlatformBreakdown(companyId int) (*PlatformBreakdownResponse, error) {
	platforms, err := s.dashboardRepository.GetPlatformBreakdown(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []PlatformBreakdownData{}
	for _, p := range platforms {
		data = append(data, PlatformBreakdownData{Platform: p.Platform, Score: p.Score, Mentions: p.Mentions})
	}

	return &PlatformBreakdownResponse{Status: true, Desc: "Get platform breakdown successful", Data: data}, nil
}

func (s dashboardService) GetPromptTrend(promptId, companyId int, interval, from, to string) (*VisibilityTrendResponse, error) {
	points, err := s.dashboardRepository.GetPromptTrend(promptId, companyId, interval, from, to)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []VisibilityTrendPoint{}
	for _, p := range points {
		data = append(data, VisibilityTrendPoint{Period: p.Period, Score: p.Score, Mentions: p.Mentions})
	}

	return &VisibilityTrendResponse{Status: true, Desc: "Get prompt trend successful", Data: data}, nil
}

func (s dashboardService) GetCompanyMetrics(companyId int) (*CompanyMetricsResponse, error) {
	m, err := s.dashboardRepository.GetCompanyMetrics(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &CompanyMetricsResponse{
		Status: true,
		Desc:   "Get company metrics successful",
		Data:   CompanyMetricsData{BrandCount: m.BrandCount, PromptCount: m.PromptCount},
	}, nil
}

func (s dashboardService) GetPromptRankings(promptId, companyId int, from, to string) (*PromptRankingsResponse, error) {
	rows, err := s.dashboardRepository.GetPromptRankings(promptId, companyId, from, to)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	data := []PromptRankingData{}
	for _, r := range rows {
		data = append(data, PromptRankingData{
			Ranking: r.Ranking, Brand: r.Brand, Sentiment: r.Sentiment,
			BrandPositioning: r.BrandPositioning, IsCompetitor: r.IsCompetitor, Visibility: r.Visibility,
		})
	}
	return &PromptRankingsResponse{Status: true, Desc: "Get prompt rankings successful", Data: data}, nil
}

func (s dashboardService) GetPromptsOverview(companyId int, from, to string) (*PromptsOverviewResponse, error) {
	rows, err := s.dashboardRepository.GetPromptsOverview(companyId, from, to)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	data := []PromptOverviewData{}
	for _, r := range rows {
		data = append(data, PromptOverviewData{
			PromptId: r.PromptId, Title: r.Title, Tag: r.Tag,
			BrandCoverage: r.BrandCoverage, BrandSentiment: r.BrandSentiment,
			BrandMentions: r.BrandMentions, TotalBrandMentions: r.TotalBrandMentions,
			DomainCitations: r.DomainCitations, TotalDomainCitations: r.TotalDomainCitations,
			Competitors: r.Competitors, Countries: r.Countries, Active: r.Active,
		})
	}
	return &PromptsOverviewResponse{Status: true, Desc: "Get prompts overview successful", Data: data}, nil
}

func (s dashboardService) GetBrandRanking(companyId int) (*BrandRankingResponse, error) {
	rows, err := s.dashboardRepository.GetBrandRanking(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	data := []BrandRankingData{}
	for _, r := range rows {
		data = append(data, BrandRankingData{
			Id: r.Id, Rank: r.Rank, Name: r.Name, IsOwn: r.IsOwn,
			SentimentScore: r.SentimentScore, Mentions: r.Mentions,
			BrandCoverage: r.BrandCoverage, ShareOfVoice: r.ShareOfVoice, AvgPosition: r.AvgPosition,
		})
	}
	return &BrandRankingResponse{Status: true, Desc: "Get brand ranking successful", Data: data}, nil
}

func (s dashboardService) GetTopPromptsByBrand(companyId int) (*TopPromptsByBrandResponse, error) {
	rows, err := s.dashboardRepository.GetTopPromptsByBrand(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	data := []TopPromptByBrandData{}
	for _, r := range rows {
		data = append(data, TopPromptByBrandData{PromptId: r.PromptId, Title: r.Title, MyBrandMentions: r.MyBrandMentions})
	}
	return &TopPromptsByBrandResponse{Status: true, Desc: "Get top prompts by brand successful", Data: data}, nil
}

func (s dashboardService) GetTopCitationURLs(companyId int) (*TopCitationURLsResponse, error) {
	rows, err := s.dashboardRepository.GetTopCitationURLs(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	data := []CitationURLData{}
	for _, r := range rows {
		data = append(data, CitationURLData{Rank: r.Rank, Url: r.Url, CitationShare: r.CitationShare, CitationCount: r.CitationCount})
	}
	return &TopCitationURLsResponse{Status: true, Desc: "Get top citation URLs successful", Data: data}, nil
}

func (s dashboardService) GetCitationURLs(companyId int) (*CitationURLsResponse, error) {
	rows, err := s.dashboardRepository.GetCitationURLs(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	data := []CitationURLDetailData{}
	for _, r := range rows {
		data = append(data, CitationURLDetailData{
			Url: r.Url, Title: r.Title, BrandMentioned: r.BrandMentioned,
			Competitors: r.Competitors, Domain: r.Domain,
			DomainCategory: r.DomainCategory, SourceType: r.SourceType, Cited: r.Cited,
			Engines: r.Engines, Tags: r.Tags, TargetCountry: r.TargetCountry,
		})
	}
	return &CitationURLsResponse{Status: true, Desc: "Get citation URLs successful", Data: data}, nil
}

func (s dashboardService) GetCitationURLPrompts(url string, companyId int) (*CitationURLPromptsResponse, error) {
	rows, err := s.dashboardRepository.GetCitationURLPrompts(url, companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	data := []CitationURLPromptData{}
	for _, r := range rows {
		data = append(data, CitationURLPromptData{
			PromptId: r.PromptId, Title: r.Title, Engines: r.Engines,
			Sentiment: r.Sentiment, Ranking: r.Ranking, CitationFrequency: r.CitationFrequency,
		})
	}
	return &CitationURLPromptsResponse{Status: true, Desc: "Get citation URL prompts successful", Data: data}, nil
}

func (s dashboardService) GetPromptDomains(promptId, companyId int, from, to string) (*PromptDomainsResponse, error) {
	rows, err := s.dashboardRepository.GetPromptDomains(promptId, companyId, from, to)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	data := []PromptDomainData{}
	for _, d := range rows {
		data = append(data, PromptDomainData{
			Domain: d.Domain, MentionCount: d.MentionCount, AvgCitation: d.AvgCitation,
			IsCompetitor: d.IsCompetitor, Snippet: d.Snippet, SourceType: d.SourceType,
		})
	}
	return &PromptDomainsResponse{Status: true, Desc: "Get prompt domains successful", Data: data}, nil
}

// GetCitationWinnersLosers compares each cited URL's citation count over the
// last 7 days against the 7 days before that, using real daily snapshots
// (citation_url_daily_stats) rather than the citations table's running
// lifetime total, which has no period boundaries to diff against.
func (s dashboardService) GetCitationWinnersLosers(companyId int) (*CitationWinnersLosersResponse, error) {
	now := time.Now()
	currentTo := now.Format("2006-01-02")
	currentFrom := now.AddDate(0, 0, -6).Format("2006-01-02")
	previousTo := now.AddDate(0, 0, -7).Format("2006-01-02")
	previousFrom := now.AddDate(0, 0, -13).Format("2006-01-02")

	rows, err := s.dashboardRepository.GetCitationURLChanges(companyId, currentFrom, currentTo, previousFrom, previousTo)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	winners := []CitationChangeData{}
	losers := []CitationChangeData{}

	for _, r := range rows {
		if r.CurrentCount == 0 && r.PreviousCount == 0 {
			continue
		}

		d := CitationChangeData{
			Url: r.Url, Title: r.Title,
			CurrentCount: r.CurrentCount, PreviousCount: r.PreviousCount,
		}

		switch {
		case r.PreviousCount == 0 && r.CurrentCount > 0:
			d.IsNew = true
			d.ChangePct = 100
		case r.PreviousCount > 0 && r.CurrentCount == 0:
			d.IsDropped = true
			d.ChangePct = -100
		case r.PreviousCount == r.CurrentCount:
			continue // no real change — not a winner or loser
		default:
			d.ChangePct = (float64(r.CurrentCount) - float64(r.PreviousCount)) / float64(r.PreviousCount) * 100
		}

		if d.ChangePct > 0 || d.IsNew {
			winners = append(winners, d)
		} else if d.ChangePct < 0 || d.IsDropped {
			losers = append(losers, d)
		}
	}

	sort.Slice(winners, func(i, j int) bool {
		return winnerSortKey(winners[i]) > winnerSortKey(winners[j])
	})
	sort.Slice(losers, func(i, j int) bool {
		return loserSortKey(losers[i]) < loserSortKey(losers[j])
	})

	if len(winners) > 3 {
		winners = winners[:3]
	}
	if len(losers) > 3 {
		losers = losers[:3]
	}

	return &CitationWinnersLosersResponse{
		Status: true, Desc: "Get citation winners/losers successful",
		Data: CitationWinnersLosersData{Winners: winners, Losers: losers},
	}, nil
}

// winnerSortKey ranks brand-new citations above every numeric percentage —
// there is no finite "% increase" from zero, so they're ordered by raw
// volume instead, ahead of any bounded percentage gain.
func winnerSortKey(d CitationChangeData) float64 {
	if d.IsNew {
		return 1e9 + float64(d.CurrentCount)
	}
	return d.ChangePct
}

func loserSortKey(d CitationChangeData) float64 {
	if d.IsDropped {
		return -1e9 - float64(d.PreviousCount)
	}
	return d.ChangePct
}

func (s dashboardService) GetBrandCitations(companyId, brandId int) (*BrandCitationsResponse, error) {
	rows, err := s.dashboardRepository.GetBrandCitations(companyId, brandId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	data := []BrandCitationData{}
	for _, r := range rows {
		data = append(data, BrandCitationData{
			Url: r.Url, Title: r.Title, Domain: r.Domain,
			Engines: r.Engines, Cited: r.Cited, LastSeen: r.LastSeen,
		})
	}
	return &BrandCitationsResponse{Status: true, Desc: "Get brand citations successful", Data: data}, nil
}
