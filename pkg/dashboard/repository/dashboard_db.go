package repository

import (
	"strconv"

	"github.com/jmoiron/sqlx"
)

type dashboardRepositoryDB struct {
	db *sqlx.DB
}

func NewDashboardRepositoryDB(db *sqlx.DB) DashboardRepository {
	return dashboardRepositoryDB{db}
}

func (r dashboardRepositoryDB) GetStats(companyId int) (*DashboardStats, error) {
	stats := DashboardStats{}
	query := `
		SELECT
			COALESCE(SUM(vd.score), 0) as total_visibility,
			COALESCE(SUM(vd.mentions), 0) as search_impressions,
			COALESCE(SUM(vd.mentions), 0) as brand_mentions,
			COALESCE(AVG(vd.score), 0) as visibility_score
		FROM visibility_data vd
		JOIN brands b ON vd.brand_id = b.id
		WHERE b.company_id = $1
	`
	err := r.db.Get(&stats, query, companyId)
	if err != nil {
		return &DashboardStats{}, nil
	}
	return &stats, nil
}

func (r dashboardRepositoryDB) GetTopPrompts(companyId int) ([]TopPrompt, error) {
	list := []TopPrompt{}
	query := `
		SELECT p.id as prompt_id, p.title, COUNT(c.id) as citation_count
		FROM prompts p
		LEFT JOIN citations c ON c.prompt_id = p.id
		WHERE p.company_id = $1
		GROUP BY p.id, p.title
		ORDER BY citation_count DESC
		LIMIT 5
	`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetTopDomains(companyId int) ([]TopDomain, error) {
	list := []TopDomain{}
	query := `
		SELECT b.domain, COUNT(c.id) as mention_count
		FROM brands b
		LEFT JOIN citations c ON c.brand_id = b.id
		WHERE b.company_id = $1
		GROUP BY b.domain
		ORDER BY mention_count DESC
		LIMIT 7
	`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetRecentCitations(companyId int) ([]RecentCitation, error) {
	list := []RecentCitation{}
	query := `
		SELECT c.id, c.prompt_id, c.brand_id, c.ranking, c.content, c.url, c.ai_platform, c.created_at
		FROM citations c
		JOIN prompts p ON c.prompt_id = p.id
		WHERE p.company_id = $1
		ORDER BY c.created_at DESC
		LIMIT 5
	`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetVisibilityTrend(companyId int, interval string) ([]VisibilityTrendPoint, error) {
	list := []VisibilityTrendPoint{}
	trunc := "month"
	switch interval {
	case "day", "week", "month", "year":
		trunc = interval
	}
	query := `
		SELECT
			TO_CHAR(DATE_TRUNC('` + trunc + `', vd.date::date), 'YYYY-MM-DD') as period,
			COALESCE(SUM(vd.score), 0) as score,
			COALESCE(SUM(vd.mentions), 0) as mentions
		FROM visibility_data vd
		JOIN brands b ON vd.brand_id = b.id
		WHERE b.company_id = $1
		GROUP BY DATE_TRUNC('` + trunc + `', vd.date::date)
		ORDER BY DATE_TRUNC('` + trunc + `', vd.date::date) ASC
	`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetPlatformBreakdown(companyId int) ([]PlatformBreakdown, error) {
	list := []PlatformBreakdown{}
	query := `
		SELECT
			vd.platform,
			COALESCE(AVG(vd.score), 0) as score,
			COALESCE(SUM(vd.mentions), 0) as mentions
		FROM visibility_data vd
		JOIN brands b ON vd.brand_id = b.id
		WHERE b.company_id = $1
		GROUP BY vd.platform
		ORDER BY score DESC
	`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetPromptTrend(promptId, companyId int, interval, from, to string) ([]VisibilityTrendPoint, error) {
	list := []VisibilityTrendPoint{}
	trunc := "month"
	switch interval {
	case "day", "week", "month", "year":
		trunc = interval
	}
	query := `
		SELECT
			TO_CHAR(DATE_TRUNC('` + trunc + `', vd.date::date), 'YYYY-MM-DD') as period,
			COALESCE(AVG(vd.score), 0) as score,
			COALESCE(SUM(vd.mentions), 0) as mentions
		FROM visibility_data vd
		JOIN prompts p ON vd.prompt_id = p.id
		WHERE vd.prompt_id = $1
		  AND p.company_id = $2
	`
	args := []interface{}{promptId, companyId}
	if from != "" {
		args = append(args, from)
		query += ` AND vd.date >= $` + strconv.Itoa(len(args))
	}
	if to != "" {
		args = append(args, to)
		query += ` AND vd.date <= $` + strconv.Itoa(len(args))
	}
	query += ` GROUP BY DATE_TRUNC('` + trunc + `', vd.date::date) ORDER BY DATE_TRUNC('` + trunc + `', vd.date::date) ASC`
	err := r.db.Select(&list, query, args...)
	return list, err
}

func (r dashboardRepositoryDB) GetCompanyMetrics(companyId int) (*CompanyMetrics, error) {
	m := CompanyMetrics{}
	query := `
		SELECT
			(SELECT COUNT(*) FROM brands WHERE company_id = $1 AND status = 'Active') as brand_count,
			(SELECT COUNT(*) FROM prompts WHERE company_id = $1) as prompt_count
	`
	err := r.db.Get(&m, query, companyId)
	if err != nil {
		return &CompanyMetrics{}, nil
	}
	return &m, nil
}

func (r dashboardRepositoryDB) GetPromptRankings(promptId, companyId int) ([]PromptRanking, error) {
	list := []PromptRanking{}
	query := `
		SELECT
			c.ranking,
			COALESCE(b.name, c.brand_positioning, 'Unknown') as brand,
			COALESCE(c.sentiment, '') as sentiment,
			COALESCE(c.brand_positioning, '') as brand_positioning,
			c.is_competitor,
			c.citation_frequency as visibility
		FROM citations c
		LEFT JOIN brands b ON c.brand_id = b.id
		JOIN prompts p ON c.prompt_id = p.id
		WHERE c.prompt_id = $1
		  AND p.company_id = $2
		ORDER BY c.ranking ASC
		LIMIT 10
	`
	err := r.db.Select(&list, query, promptId, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetPromptsOverview(companyId int) ([]PromptOverview, error) {
	list := []PromptOverview{}
	query := `
		WITH own_brand_id AS (
			SELECT COALESCE(MIN(id), 0) AS id FROM brands WHERE company_id = $1 AND is_own = TRUE
		)
		SELECT
			p.id AS prompt_id,
			p.title,
			COALESCE(pc.name, '') AS category,
			COUNT(DISTINCT CASE WHEN c.brand_id = ob.id AND ob.id != 0 THEN c.id END)::INT AS brand_mentions,
			COUNT(DISTINCT c.id)::INT AS total_brand_mentions,
			COALESCE(SUM(CASE
				WHEN c.brand_id = ob.id AND ob.id != 0 AND LOWER(COALESCE(c.sentiment,'')) = 'positive' THEN 1
				WHEN c.brand_id = ob.id AND ob.id != 0 AND LOWER(COALESCE(c.sentiment,'')) = 'negative' THEN -1
				ELSE 0 END), 0)::INT AS brand_sentiment,
			CASE
				WHEN COUNT(DISTINCT c.id) > 0
				THEN ROUND(COUNT(DISTINCT CASE WHEN c.brand_id = ob.id AND ob.id != 0 THEN c.id END)::NUMERIC
					/ COUNT(DISTINCT c.id) * 100, 0)::INT
				ELSE 0::INT
			END AS brand_coverage,
			COUNT(DISTINCT CASE WHEN c.brand_id = ob.id AND ob.id != 0 AND c.url IS NOT NULL AND c.url != '' THEN c.url END)::INT AS domain_citations,
			COUNT(DISTINCT CASE WHEN c.url IS NOT NULL AND c.url != '' THEN c.url END)::INT AS total_domain_citations,
			COALESCE(
				(SELECT STRING_AGG(DISTINCT b2.name, ',')
				 FROM citations c2
				 JOIN brands b2 ON b2.id = c2.brand_id
				 WHERE c2.prompt_id = p.id
				   AND b2.is_own = FALSE
				   AND b2.company_id = $1
				),
				''
			) AS competitors
		FROM prompts p
		CROSS JOIN own_brand_id ob
		LEFT JOIN citations c ON c.prompt_id = p.id
		LEFT JOIN prompt_categories pc ON pc.id = p.category_id
		WHERE p.company_id = $1
		GROUP BY p.id, p.title, ob.id, pc.name
		ORDER BY brand_mentions DESC, p.id ASC
	`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetBrandRanking(companyId int) ([]BrandRankingRow, error) {
	list := []BrandRankingRow{}
	query := `
		WITH company_citations AS (
			SELECT c.brand_id, c.prompt_id, c.ranking, c.sentiment
			FROM citations c
			JOIN prompts p ON c.prompt_id = p.id
			WHERE p.company_id = $1
		),
		brand_stats AS (
			SELECT
				b.id, b.name, b.is_own,
				COALESCE(COUNT(DISTINCT cc.prompt_id), 0) AS covered_prompts,
				COALESCE(COUNT(cc.brand_id), 0) AS mentions,
				COALESCE(AVG(CASE WHEN cc.ranking > 0 THEN cc.ranking::FLOAT ELSE NULL END), 99) AS avg_position,
				COALESCE(SUM(CASE
					WHEN LOWER(COALESCE(cc.sentiment,'')) = 'positive' THEN 1
					WHEN LOWER(COALESCE(cc.sentiment,'')) = 'negative' THEN -1
					ELSE 0 END), 0) AS sentiment_score
			FROM brands b
			LEFT JOIN company_citations cc ON cc.brand_id = b.id
			WHERE b.company_id = $1
			GROUP BY b.id, b.name, b.is_own
		),
		prompt_count AS (SELECT NULLIF(COUNT(*), 0) AS total FROM prompts WHERE company_id = $1),
		total_mentions AS (SELECT NULLIF(SUM(mentions), 0) AS total FROM brand_stats)
		SELECT
			ROW_NUMBER() OVER (ORDER BY bs.mentions DESC)::INT AS rank,
			bs.name,
			bs.is_own,
			bs.sentiment_score::INT AS sentiment_score,
			bs.mentions::INT,
			COALESCE(ROUND((bs.covered_prompts::NUMERIC / pc.total * 100), 1), 0)::FLOAT AS brand_coverage,
			COALESCE(ROUND((bs.mentions::NUMERIC / tm.total * 100), 1), 0)::FLOAT AS share_of_voice,
			ROUND(bs.avg_position::NUMERIC, 2)::FLOAT AS avg_position
		FROM brand_stats bs, prompt_count pc, total_mentions tm
		ORDER BY bs.mentions DESC
		LIMIT 20
	`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetTopPromptsByBrand(companyId int) ([]TopPromptByBrand, error) {
	list := []TopPromptByBrand{}
	query := `
		SELECT p.id AS prompt_id, p.title, COUNT(c.id)::INT AS my_brand_mentions
		FROM prompts p
		JOIN citations c ON c.prompt_id = p.id
		JOIN brands b ON c.brand_id = b.id
		WHERE p.company_id = $1
		  AND b.company_id = $1
		  AND b.is_own = TRUE
		GROUP BY p.id, p.title
		ORDER BY my_brand_mentions DESC
		LIMIT 10
	`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetTopCitationURLs(companyId int) ([]CitationURL, error) {
	list := []CitationURL{}
	query := `
		WITH url_counts AS (
			SELECT c.url, COUNT(c.id) AS cnt
			FROM citations c
			JOIN prompts p ON c.prompt_id = p.id
			WHERE p.company_id = $1 AND c.url IS NOT NULL AND c.url != ''
			GROUP BY c.url
		),
		total AS (SELECT NULLIF(SUM(cnt), 0) AS total FROM url_counts)
		SELECT
			ROW_NUMBER() OVER (ORDER BY uc.cnt DESC)::INT AS rank,
			uc.url,
			uc.cnt::INT AS citation_count,
			COALESCE(ROUND(uc.cnt::NUMERIC / t.total * 100, 1), 0)::FLOAT AS citation_share
		FROM url_counts uc, total t
		ORDER BY uc.cnt DESC
		LIMIT 10
	`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetCitationURLs(companyId int) ([]CitationURLDetail, error) {
	list := []CitationURLDetail{}
	query := `
		WITH own_brand_id AS (
			SELECT COALESCE(MIN(id), 0) AS id FROM brands WHERE company_id = $1 AND is_own = TRUE
		)
		SELECT
			c.url,
			COALESCE(MIN(c.content), '') AS title,
			BOOL_OR(c.brand_id = ob.id AND ob.id != 0) AS brand_mentioned,
			COALESCE(
				STRING_AGG(DISTINCT CASE WHEN b.id IS NOT NULL AND b.is_own = FALSE THEN b.name END, ','),
				''
			) AS competitors,
			REGEXP_REPLACE(c.url, '^(?:https?://)?(?:www\.)?([^/?#]*).*$', '\1') AS domain,
			CASE
				WHEN BOOL_OR(b.id IS NOT NULL AND b.is_own = TRUE) THEN 'Brand'
				WHEN BOOL_OR(b.id IS NOT NULL AND b.is_own = FALSE) THEN 'Competitor'
				ELSE 'Others'
			END AS domain_category,
			COUNT(DISTINCT c.prompt_id)::INT AS cited
		FROM citations c
		CROSS JOIN own_brand_id ob
		JOIN prompts p ON p.id = c.prompt_id
		LEFT JOIN brands b ON b.id = c.brand_id AND b.company_id = $1
		WHERE p.company_id = $1
		  AND c.url IS NOT NULL
		  AND c.url != ''
		GROUP BY c.url, ob.id
		ORDER BY cited DESC
		LIMIT 200
	`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetCitationURLPrompts(url string, companyId int) ([]CitationURLPrompt, error) {
	list := []CitationURLPrompt{}
	query := `
		SELECT DISTINCT p.id AS prompt_id, p.title
		FROM citations c
		JOIN prompts p ON p.id = c.prompt_id
		WHERE c.url = $1 AND p.company_id = $2
		ORDER BY p.title
	`
	err := r.db.Select(&list, query, url, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetPromptDomains(promptId, companyId int) ([]PromptDomain, error) {
	list := []PromptDomain{}
	query := `
		SELECT
			REGEXP_REPLACE(url, '^(?:https?://)?([^/?#]*).*$', '\1') as domain,
			COUNT(*) as mention_count,
			COALESCE(AVG(citation_frequency), 0) as avg_citation,
			BOOL_OR(is_competitor) as is_competitor,
			COALESCE(MIN(snippet), MIN(content), '') as snippet
		FROM citations c
		JOIN prompts p ON c.prompt_id = p.id
		WHERE c.prompt_id = $1
		  AND p.company_id = $2
		  AND c.url IS NOT NULL
		  AND c.url != ''
		GROUP BY REGEXP_REPLACE(url, '^(?:https?://)?([^/?#]*).*$', '\1')
		ORDER BY mention_count DESC
		LIMIT 10
	`
	err := r.db.Select(&list, query, promptId, companyId)
	return list, err
}
