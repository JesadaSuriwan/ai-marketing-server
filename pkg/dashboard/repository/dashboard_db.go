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

func (r dashboardRepositoryDB) GetPromptRankings(promptId, companyId int, from, to string) ([]PromptRanking, error) {
	list := []PromptRanking{}
	// One row per real cited URL exists per brand (chatgpt/gemini/perplexity
	// each contribute their own), plus rows with no brand match at all
	// (brand_id IS NULL) — the INNER JOIN to brands drops those, and the
	// GROUP BY collapses the rest to one ranking row per brand, picking the
	// best-ranked citation's sentiment/positioning as representative and
	// summing citation_frequency across all of that brand's citations.
	query := `
		WITH brand_citations AS (
			SELECT c.ranking, b.name AS brand, c.sentiment, c.brand_positioning, c.is_competitor, c.citation_frequency
			FROM citations c
			JOIN brands b ON c.brand_id = b.id
			JOIN prompts p ON c.prompt_id = p.id
			WHERE c.prompt_id = $1
			  AND p.company_id = $2
			  AND ($3 = '' OR c.last_checked >= $3::date)
			  AND ($4 = '' OR c.last_checked <= $4::date)
		)
		SELECT
			MIN(ranking) AS ranking,
			brand,
			(ARRAY_AGG(sentiment ORDER BY ranking ASC))[1] AS sentiment,
			(ARRAY_AGG(brand_positioning ORDER BY ranking ASC))[1] AS brand_positioning,
			BOOL_OR(is_competitor) AS is_competitor,
			SUM(citation_frequency)::INT AS visibility
		FROM brand_citations
		GROUP BY brand
		ORDER BY MIN(ranking) ASC
		LIMIT 10
	`
	err := r.db.Select(&list, query, promptId, companyId, from, to)
	return list, err
}

func (r dashboardRepositoryDB) GetPromptsOverview(companyId int, from, to string) ([]PromptOverview, error) {
	list := []PromptOverview{}
	query := `
		WITH own_brand_id AS (
			SELECT COALESCE(MIN(id), 0) AS id FROM brands WHERE company_id = $1 AND is_own = TRUE
		)
		SELECT
			p.id AS prompt_id,
			p.title,
			COALESCE(pc.name, '') AS tag,
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
				   AND ($2 = '' OR c2.last_checked >= $2::date)
				   AND ($3 = '' OR c2.last_checked <= $3::date)
				),
				''
			) AS competitors,
			COALESCE(
				(SELECT STRING_AGG(pco.country_code, ',' ORDER BY pco.country_code)
				 FROM prompt_countries pco WHERE pco.prompt_id = p.id),
				''
			) AS countries,
			p.active
		FROM prompts p
		CROSS JOIN own_brand_id ob
		LEFT JOIN citations c ON c.prompt_id = p.id
			AND ($2 = '' OR c.last_checked >= $2::date)
			AND ($3 = '' OR c.last_checked <= $3::date)
		LEFT JOIN prompt_categories pc ON pc.id = p.tag_id
		WHERE p.company_id = $1
		GROUP BY p.id, p.title, ob.id, pc.name, p.active
		ORDER BY brand_mentions DESC, p.id ASC
	`
	err := r.db.Select(&list, query, companyId, from, to)
	return list, err
}

// GetBrandRanking now takes the same date/engine/tag/country filters
// GetBrandCoverageTrend already did — previously it always computed over
// every citation ever recorded for the company, completely ignoring
// whatever the Overview page's filter bar was set to, while the trend chart
// right next to it already respected those filters. That meant "Your Brand
// Mentions" and "Your Average Brand Position" silently never changed when
// someone changed the date range, even though they sit directly beside a
// chart that visibly does. Empty string/0 for any filter means "no filter
// on this dimension", same convention as GetBrandCoverageTrend.
func (r dashboardRepositoryDB) GetBrandRanking(companyId int, from, to, engine string, tagId int, country string) ([]BrandRankingRow, error) {
	list := []BrandRankingRow{}
	query := `
		WITH filtered_prompts AS (
			SELECT DISTINCT p.id
			FROM prompts p
			LEFT JOIN prompt_countries pco ON pco.prompt_id = p.id
			WHERE p.company_id = $1
			  AND ($5 = 0 OR p.tag_id = $5)
			  AND ($6 = '' OR pco.country_code = $6)
		),
		company_citations AS (
			SELECT c.brand_id, c.prompt_id, c.ranking, c.sentiment
			FROM citations c
			JOIN filtered_prompts fp ON fp.id = c.prompt_id
			WHERE ($2 = '' OR c.last_checked >= $2::date)
			  AND ($3 = '' OR c.last_checked <= $3::date)
			  AND ($4 = '' OR c.ai_platform = $4)
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
		prompt_count AS (SELECT NULLIF(COUNT(*), 0) AS total FROM filtered_prompts),
		total_mentions AS (SELECT NULLIF(SUM(mentions), 0) AS total FROM brand_stats)
		SELECT
			bs.id,
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
	err := r.db.Select(&list, query, companyId, from, to, engine, tagId, country)
	return list, err
}

// GetBrandCoverageTrend computes each brand's day-by-day coverage —
// COUNT(DISTINCT covered prompt) / COUNT(DISTINCT tracked prompt) — from the
// raw per-(prompt,brand,engine,day) citation facts, filtered by engine, tag
// and/or country at query time. The denominator (which prompts count as
// "tracked") uses each prompt's CURRENT tag/country, not what it was on that
// historical day — the same simplification other historical reports in this
// app already make, and unavoidable since tag/country aren't versioned.
func (r dashboardRepositoryDB) GetBrandCoverageTrend(companyId int, from, to, engine string, tagId int, country string) ([]BrandCoverageTrendRow, error) {
	list := []BrandCoverageTrendRow{}
	query := `
		WITH filtered_prompts AS (
			SELECT DISTINCT p.id
			FROM prompts p
			LEFT JOIN prompt_countries pco ON pco.prompt_id = p.id
			WHERE p.company_id = $1
			  AND ($4 = 0 OR p.tag_id = $4)
			  AND ($5 = '' OR pco.country_code = $5)
		),
		total AS (SELECT COUNT(*) AS n FROM filtered_prompts)
		SELECT bpcd.brand_id,
			TO_CHAR(bpcd.stat_date, 'YYYY-MM-DD') AS date,
			COALESCE(ROUND(COUNT(DISTINCT bpcd.prompt_id)::NUMERIC / NULLIF((SELECT n FROM total), 0) * 100, 1), 0)::FLOAT AS coverage
		FROM brand_prompt_coverage_daily bpcd
		JOIN brands b ON b.id = bpcd.brand_id AND b.company_id = $1
		JOIN filtered_prompts fp ON fp.id = bpcd.prompt_id
		WHERE ($2 = '' OR bpcd.stat_date >= $2::date)
		  AND ($3 = '' OR bpcd.stat_date <= $3::date)
		  AND ($6 = '' OR bpcd.engine = $6)
		GROUP BY bpcd.brand_id, bpcd.stat_date
		ORDER BY bpcd.stat_date ASC
	`
	err := r.db.Select(&list, query, companyId, from, to, tagId, country, engine)
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

func (r dashboardRepositoryDB) GetCitationURLs(companyId int, from, to string) ([]CitationURLDetail, error) {
	list := []CitationURLDetail{}
	query := `
		WITH own_brand_id AS (
			SELECT COALESCE(MIN(id), 0) AS id FROM brands WHERE company_id = $1 AND is_own = TRUE
		)
		SELECT
			c.url,
			COALESCE(MIN(c.content), '') AS title,
			COALESCE(BOOL_OR(c.brand_id = ob.id AND ob.id != 0), FALSE) AS brand_mentioned,
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
			COALESCE(MIN(c.source_type), 'other') AS source_type,
			COUNT(DISTINCT c.prompt_id)::INT AS cited,
			COALESCE(STRING_AGG(DISTINCT c.ai_platform, ','), '') AS engines,
			COALESCE(STRING_AGG(DISTINCT pc.name, ','), '') AS tags,
			COALESCE(MAX(c.target_country), '') AS target_country
		FROM citations c
		CROSS JOIN own_brand_id ob
		JOIN prompts p ON p.id = c.prompt_id
		LEFT JOIN brands b ON b.id = c.brand_id AND b.company_id = $1
		LEFT JOIN prompt_categories pc ON pc.id = p.tag_id
		WHERE p.company_id = $1
		  AND c.url IS NOT NULL
		  AND c.url != ''
		  AND ($2 = '' OR c.last_checked >= $2::date)
		  AND ($3 = '' OR c.last_checked <= $3::date)
		GROUP BY c.url, ob.id
		ORDER BY cited DESC
	`
	// No LIMIT here (there used to be one, capped at 200) — GetCitationURLChanges
	// (Gaining/Losing ground) considers every citation with no such cap, so a
	// domain capped out of this query but still surfaced there would be
	// completely unfindable in the KPI cards and table it's supposed to relate
	// to. The table itself paginates client-side, so a larger result set here
	// doesn't cost a bigger render — just a bigger response payload.
	err := r.db.Select(&list, query, companyId, from, to)
	return list, err
}

// GetDomainCoverageTrend turns the existing per-URL daily citation counts
// (citation_url_daily_stats, already bumped on every prompt run) into each
// domain's day-by-day share of that day's citations.
func (r dashboardRepositoryDB) GetDomainCoverageTrend(companyId int) ([]DomainCoverageTrendRow, error) {
	list := []DomainCoverageTrendRow{}
	query := `
		WITH per_domain_day AS (
			SELECT
				REGEXP_REPLACE(url, '^(?:https?://)?(?:www\.)?([^/?#]*).*$', '\1') AS domain,
				stat_date,
				SUM(citation_count) AS cnt
			FROM citation_url_daily_stats
			WHERE company_id = $1
			GROUP BY domain, stat_date
		),
		day_totals AS (
			SELECT stat_date, SUM(cnt) AS total FROM per_domain_day GROUP BY stat_date
		)
		SELECT pd.domain,
			TO_CHAR(pd.stat_date, 'YYYY-MM-DD') AS date,
			COALESCE(ROUND(pd.cnt::NUMERIC / NULLIF(dt.total, 0) * 100, 1), 0)::FLOAT AS coverage
		FROM per_domain_day pd
		JOIN day_totals dt ON dt.stat_date = pd.stat_date
		ORDER BY pd.stat_date ASC
	`
	err := r.db.Select(&list, query, companyId)
	return list, err
}

// GetPromptBrandCoverageTrend is the same idea as GetBrandCoverageTrend but
// scoped to one prompt: for each day, what share of the engines that ran
// this prompt that day cited each brand. Built from brand_prompt_coverage_daily
// (recorded per prompt+brand+engine+day since that table was introduced —
// see its migration comment for why there's no earlier history) joined
// against prompt_runs for the day's real engine count.
func (r dashboardRepositoryDB) GetPromptBrandCoverageTrend(promptId, companyId int) ([]PromptBrandCoverageTrendRow, error) {
	list := []PromptBrandCoverageTrendRow{}
	query := `
		WITH runs_per_day AS (
			SELECT created_at::date AS stat_date, COUNT(DISTINCT ai_platform) AS n
			FROM prompt_runs
			WHERE prompt_id = $1
			GROUP BY created_at::date
		)
		SELECT b.name AS brand,
			TO_CHAR(bpcd.stat_date, 'YYYY-MM-DD') AS date,
			COALESCE(ROUND(COUNT(DISTINCT bpcd.engine)::NUMERIC / NULLIF((SELECT n FROM runs_per_day WHERE stat_date = bpcd.stat_date), 0) * 100, 1), 0)::FLOAT AS coverage
		FROM brand_prompt_coverage_daily bpcd
		JOIN brands b ON b.id = bpcd.brand_id
		WHERE bpcd.prompt_id = $1 AND b.company_id = $2
		GROUP BY b.name, bpcd.stat_date
		ORDER BY bpcd.stat_date ASC
	`
	err := r.db.Select(&list, query, promptId, companyId)
	return list, err
}

// GetCitationURLChanges scopes the current-vs-previous-period comparison to
// only the URLs that match the page's filter bar — computed the same way
// GetCitationURLs computes domain_category/source_type/engines/tags/country
// per URL (aggregated across that URL's citations, current state — the daily
// snapshot table itself doesn't carry these dimensions, only a per-day
// count), so a domain filtered out of the table is filtered out of
// Gaining/Losing ground too. Empty string ("" for category/sourceType,
// tags/engines/countries) means "no filter on this dimension", matching the
// page's own "no filter selected" convention.
func (r dashboardRepositoryDB) GetCitationURLChanges(companyId int, currentFrom, currentTo, previousFrom, previousTo, category, sourceType, tags, engines, countries string) ([]CitationURLChange, error) {
	list := []CitationURLChange{}
	query := `
		WITH own_brand_id AS (
			SELECT COALESCE(MIN(id), 0) AS id FROM brands WHERE company_id = $1 AND is_own = TRUE
		),
		url_attrs AS (
			SELECT
				c.url,
				CASE
					WHEN BOOL_OR(b.id IS NOT NULL AND b.is_own = TRUE) THEN 'Brand'
					WHEN BOOL_OR(b.id IS NOT NULL AND b.is_own = FALSE) THEN 'Competitor'
					ELSE 'Others'
				END AS category,
				COALESCE(MIN(c.source_type), 'other') AS source_type,
				COALESCE(STRING_AGG(DISTINCT c.ai_platform, ','), '') AS engines,
				COALESCE(STRING_AGG(DISTINCT pc.name, ','), '') AS tags,
				COALESCE(MAX(c.target_country), '') AS target_country
			FROM citations c
			CROSS JOIN own_brand_id ob
			JOIN prompts p ON p.id = c.prompt_id AND p.company_id = $1
			LEFT JOIN brands b ON b.id = c.brand_id AND b.company_id = $1
			LEFT JOIN prompt_categories pc ON pc.id = p.tag_id
			WHERE c.url IS NOT NULL AND c.url != ''
			GROUP BY c.url, ob.id
		),
		matching_urls AS (
			SELECT url FROM url_attrs
			WHERE ($6 = '' OR category = $6)
			  AND ($7 = '' OR source_type = $7)
			  AND ($8 = '' OR EXISTS (SELECT 1 FROM unnest(string_to_array($8, ',')) t WHERE t = ANY(string_to_array(tags, ','))))
			  AND ($9 = '' OR EXISTS (SELECT 1 FROM unnest(string_to_array($9, ',')) t WHERE t = ANY(string_to_array(engines, ','))))
			  AND ($10 = '' OR target_country = ANY(string_to_array($10, ',')))
		),
		current_period AS (
			SELECT url, SUM(citation_count) AS cnt
			FROM citation_url_daily_stats
			WHERE company_id = $1 AND stat_date >= $2 AND stat_date <= $3
			  AND url IN (SELECT url FROM matching_urls)
			GROUP BY url
		),
		previous_period AS (
			SELECT url, SUM(citation_count) AS cnt
			FROM citation_url_daily_stats
			WHERE company_id = $1 AND stat_date >= $4 AND stat_date <= $5
			  AND url IN (SELECT url FROM matching_urls)
			GROUP BY url
		),
		combined AS (
			SELECT COALESCE(c.url, p.url) AS url,
				COALESCE(c.cnt, 0)::INT AS current_count,
				COALESCE(p.cnt, 0)::INT AS previous_count
			FROM current_period c
			FULL OUTER JOIN previous_period p ON c.url = p.url
		)
		SELECT
			combined.url,
			COALESCE(MIN(cit.content), '') AS title,
			combined.current_count,
			combined.previous_count
		FROM combined
		LEFT JOIN citations cit ON cit.url = combined.url
		LEFT JOIN prompts pr ON pr.id = cit.prompt_id AND pr.company_id = $1
		GROUP BY combined.url, combined.current_count, combined.previous_count
	`
	err := r.db.Select(&list, query, companyId, currentFrom, currentTo, previousFrom, previousTo, category, sourceType, tags, engines, countries)
	return list, err
}

func (r dashboardRepositoryDB) GetBrandCitations(companyId, brandId int) ([]BrandCitation, error) {
	list := []BrandCitation{}
	query := `
		SELECT
			c.url,
			COALESCE(MIN(c.content), '') AS title,
			REGEXP_REPLACE(c.url, '^(?:https?://)?(?:www\.)?([^/?#]*).*$', '\1') AS domain,
			COALESCE(STRING_AGG(DISTINCT c.ai_platform, ','), '') AS engines,
			COUNT(DISTINCT c.prompt_id)::INT AS cited,
			COALESCE(MAX(c.last_checked)::TEXT, '') AS last_seen
		FROM citations c
		JOIN prompts p ON p.id = c.prompt_id
		WHERE p.company_id = $1 AND c.brand_id = $2 AND c.url IS NOT NULL AND c.url != ''
		GROUP BY c.url
		ORDER BY cited DESC
		LIMIT 200
	`
	err := r.db.Select(&list, query, companyId, brandId)
	return list, err
}

func (r dashboardRepositoryDB) GetCitationURLPrompts(url string, companyId int) ([]CitationURLPrompt, error) {
	list := []CitationURLPrompt{}
	// One citation row per engine can exist for the same (prompt, url) — the
	// GROUP BY collapses those to one row per prompt, same
	// pick-the-top-ranked-row-as-representative idiom GetPromptRankings uses
	// for sentiment, plus a comma-joined engine list and a summed
	// citation_frequency across all of that prompt's engines for this URL.
	query := `
		SELECT
			p.id AS prompt_id,
			p.title,
			STRING_AGG(DISTINCT c.ai_platform, ',') AS engines,
			(ARRAY_AGG(c.sentiment ORDER BY c.ranking ASC))[1] AS sentiment,
			MIN(c.ranking) AS ranking,
			SUM(c.citation_frequency)::INT AS citation_frequency
		FROM citations c
		JOIN prompts p ON p.id = c.prompt_id
		WHERE c.url = $1 AND p.company_id = $2
		GROUP BY p.id, p.title
		ORDER BY MIN(c.ranking) ASC
	`
	err := r.db.Select(&list, query, url, companyId)
	return list, err
}

func (r dashboardRepositoryDB) GetPromptDomains(promptId, companyId int, from, to string) ([]PromptDomain, error) {
	list := []PromptDomain{}
	query := `
		SELECT
			REGEXP_REPLACE(url, '^(?:https?://)?([^/?#]*).*$', '\1') as domain,
			COUNT(*) as mention_count,
			COALESCE(AVG(citation_frequency), 0) as avg_citation,
			BOOL_OR(is_competitor) as is_competitor,
			COALESCE(MIN(snippet), MIN(c.content), '') as snippet,
			COALESCE(MIN(c.source_type), 'other') as source_type
		FROM citations c
		JOIN prompts p ON c.prompt_id = p.id
		WHERE c.prompt_id = $1
		  AND p.company_id = $2
		  AND c.url IS NOT NULL
		  AND c.url != ''
		  AND ($3 = '' OR c.last_checked >= $3::date)
		  AND ($4 = '' OR c.last_checked <= $4::date)
		GROUP BY REGEXP_REPLACE(url, '^(?:https?://)?([^/?#]*).*$', '\1')
		ORDER BY mention_count DESC
		LIMIT 10
	`
	err := r.db.Select(&list, query, promptId, companyId, from, to)
	return list, err
}
