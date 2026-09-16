package main

import (
	"github.com/ai-marketing/ai-marketing-server/config"
	"github.com/ai-marketing/ai-marketing-server/databases"
	"github.com/joho/godotenv"
	"log"
)

var schema = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    initials TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
-- Set on accounts created with a temp password (e.g. by a Team Lead/Admin
-- adding a member) so the frontend can force a real password to be chosen
-- before letting the user do anything else.
ALTER TABLE users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS companies (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    initials TEXT NOT NULL DEFAULT '',
    logo_url TEXT,
    industry TEXT,
    website TEXT,
    location TEXT,
    size TEXT,
    founded TEXT,
    phone TEXT,
    email TEXT,
    about TEXT,
    plan TEXT DEFAULT 'Free',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
-- Caps how many ACTIVE prompts (see prompts.active) this workspace can track
-- at once; adjustable by Admin/Team Lead only.
ALTER TABLE companies ADD COLUMN IF NOT EXISTS prompt_limit INT NOT NULL DEFAULT 50;
-- Nullable: existing rows predate this feature and have no contract on file.
-- New workspaces are required (at the application layer) to set one at
-- creation — once it passes, prompt runs for that company are paused (see
-- the check in promptrun_service.go's run()) until renewed.
ALTER TABLE companies ADD COLUMN IF NOT EXISTS contract_end_date DATE;

CREATE TABLE IF NOT EXISTS company_tags (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    tag TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS company_markets (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    market TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS company_social_links (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE UNIQUE,
    linkedin TEXT,
    twitter TEXT,
    facebook TEXT
);

CREATE TABLE IF NOT EXISTS company_key_contacts (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    email TEXT,
    phone TEXT
);

CREATE TABLE IF NOT EXISTS brands (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    domain TEXT NOT NULL,
    industry TEXT,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    is_own BOOLEAN NOT NULL DEFAULT FALSE,
    logo_url TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Additional domains for a brand (e.g. a competitor's regional sites)
-- beyond its primary brands.domain — same pattern as the subdomains table
-- used for a company's own additional domains.
CREATE TABLE IF NOT EXISTS brand_domains (
    id SERIAL PRIMARY KEY,
    brand_id INT NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    domain TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS prompt_categories (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS prompts (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    tag_id INT REFERENCES prompt_categories(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
-- Renamed from category_id — "Tags" is the only user-facing name for this
-- concept now, the old column name was a leftover source of confusion.
-- Guarded so re-running the migration against a database that's already
-- been renamed doesn't error (a bare RENAME COLUMN isn't idempotent).
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'prompts' AND column_name = 'category_id'
    ) THEN
        ALTER TABLE prompts RENAME COLUMN category_id TO tag_id;
    END IF;
END $$;
-- Inactive prompts are paused (no scheduled or manual runs) without deleting
-- their history — the counterpart to companies.prompt_limit, which only
-- counts active prompts.
ALTER TABLE prompts ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT TRUE;

-- Which countries a prompt is scoped to for reporting/filtering (Prompts
-- overview's country filter) — a display/filter dimension only, it doesn't
-- change what's actually sent to any AI engine. Many-to-many since a prompt
-- can target multiple countries at once. country_code is ISO 3166-1 alpha-2
-- (e.g. "TH", "US") so the frontend can derive the flag emoji without a
-- lookup table.
CREATE TABLE IF NOT EXISTS prompt_countries (
    id SERIAL PRIMARY KEY,
    prompt_id INT NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    country_code TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (prompt_id, country_code)
);

CREATE TABLE IF NOT EXISTS prompt_runs (
    id SERIAL PRIMARY KEY,
    prompt_id INT NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    ai_platform TEXT NOT NULL,
    model TEXT NOT NULL,
    raw_response TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS prompt_suggestions (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    rationale TEXT,
    category TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    created_prompt_id INT REFERENCES prompts(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS visibility_data (
    id SERIAL PRIMARY KEY,
    brand_id INT NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    prompt_id INT REFERENCES prompts(id) ON DELETE SET NULL,
    platform TEXT NOT NULL,
    score FLOAT NOT NULL DEFAULT 0,
    mentions INT NOT NULL DEFAULT 0,
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS citations (
    id SERIAL PRIMARY KEY,
    prompt_id INT NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    brand_id INT REFERENCES brands(id) ON DELETE SET NULL,
    prompt_run_id INT REFERENCES prompt_runs(id) ON DELETE SET NULL,
    ranking INT NOT NULL,
    content TEXT NOT NULL,
    url TEXT,
    ai_platform TEXT,
    date_discovered DATE,
    last_checked DATE,
    sentiment TEXT DEFAULT 'neutral',
    brand_positioning TEXT DEFAULT 'mentioned',
    citation_frequency INT DEFAULT 0,
    snippet TEXT,
    is_competitor BOOLEAN DEFAULT FALSE,
    notes TEXT,
    is_archived BOOLEAN DEFAULT FALSE,
    source_type TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- CREATE TABLE IF NOT EXISTS is a no-op on tables that already exist, so
-- columns added to the citations definition above need an explicit ALTER
-- for databases migrated before this column existed.
ALTER TABLE citations ADD COLUMN IF NOT EXISTS prompt_run_id INT REFERENCES prompt_runs(id) ON DELETE SET NULL;
-- source_type classifies the citing domain: editorial | directory | reference
-- | social | news | marketplace | official_site | other — see the extraction
-- prompt in pkg/promptrun/service for the exact taxonomy.
ALTER TABLE citations ADD COLUMN IF NOT EXISTS source_type TEXT;
-- ISO 3166-1 alpha-2 code (e.g. "TH") for the market the cited URL appears
-- to target — NULL when undetermined. Derived deterministically from the
-- domain's ccTLD where possible, falling back to the extraction model's
-- guess for generic TLDs (.com, .org, ...); see detectCountryFromTLD in
-- pkg/promptrun/service. Purely informational/filterable — has no effect on
-- what's actually searched or how a prompt's own country tag is set.
ALTER TABLE citations ADD COLUMN IF NOT EXISTS target_country TEXT;

-- Superseded by citations_prompt_brand_url_platform_uniq below — dropped so
-- the same URL can get a separate row per AI platform (e.g. ChatGPT and
-- Gemini both citing the same page for the same prompt/brand should show as
-- two distinct citations, not silently collapse into one).
DROP INDEX IF EXISTS citations_prompt_brand_url_uniq;

CREATE UNIQUE INDEX IF NOT EXISTS citations_prompt_brand_url_platform_uniq
    ON citations (prompt_id, brand_id, url, ai_platform) NULLS NOT DISTINCT;

CREATE TABLE IF NOT EXISTS citation_competitors (
    id SERIAL PRIMARY KEY,
    citation_id INT NOT NULL REFERENCES citations(id) ON DELETE CASCADE,
    competitor_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS citation_ranking_history (
    id SERIAL PRIMARY KEY,
    citation_id INT NOT NULL REFERENCES citations(id) ON DELETE CASCADE,
    date_label TEXT NOT NULL,
    rank INT NOT NULL
);

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_value TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'Active',
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS company_members (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'Viewer',
    joined_at TIMESTAMP DEFAULT NOW()
);

-- Real email-invitation flow: a member starts 'pending' with a one-time
-- token, and becomes 'active' (linked to a real users row) once they accept.
-- Rows created before this existed have no token, so they're backfilled to
-- 'active' below rather than left stuck pending.
ALTER TABLE company_members ADD COLUMN IF NOT EXISTS user_id INT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE company_members ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE company_members ADD COLUMN IF NOT EXISTS invite_token TEXT;
ALTER TABLE company_members ADD COLUMN IF NOT EXISTS invite_token_expires_at TIMESTAMP;
UPDATE company_members SET status = 'active' WHERE status = 'pending' AND invite_token IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS company_members_invite_token_uniq ON company_members (invite_token) WHERE invite_token IS NOT NULL;

-- Legacy invite roles (Editor/Viewer/Admin) replaced by the 4-role model
-- (admin/team_lead/specialist/customer) — "admin" itself is never stored
-- here, it's always derived from companies.user_id (the creator), so no
-- legacy value maps to it.
-- Case-insensitive: the old invite dropdowns actually submitted lowercase
-- values ("admin"/"editor"/"viewer") despite the backend's default being
-- capitalized ("Viewer") — both forms need to be caught here, since a row
-- with role="admin" would otherwise be silently treated as real Admin by
-- the new permission system (which only ever computes "admin" from company
-- ownership, never expects to find it stored in this column).
UPDATE company_members SET role = 'team_lead' WHERE LOWER(role) = 'admin';
UPDATE company_members SET role = 'specialist' WHERE LOWER(role) = 'editor';
UPDATE company_members SET role = 'customer' WHERE LOWER(role) = 'viewer';
ALTER TABLE company_members ALTER COLUMN role SET DEFAULT 'specialist';

CREATE TABLE IF NOT EXISTS subdomains (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    subdomain TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'Active',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recommendations (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    placement TEXT NOT NULL,
    impact TEXT NOT NULL,
    engine TEXT NOT NULL DEFAULT 'chatgpt',
    text TEXT NOT NULL,
    link TEXT,
    link_text TEXT,
    why TEXT NOT NULL,
    steps TEXT NOT NULL DEFAULT '[]',
    prompt_id INT REFERENCES prompts(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'suggested',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS recommendations_dedupe_uniq
    ON recommendations (company_id, type, prompt_id, link) NULLS NOT DISTINCT;

-- Standing, company-authored rules that steer every future recommendation
-- generation (unlike the one-off "guidance" text on the generate call,
-- these persist). active lets a user toggle a rule off without losing it —
-- e.g. to experiment with different combinations before settling on one.
CREATE TABLE IF NOT EXISTS recommendation_rules (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS company_ai_engines (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    platform TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (company_id, platform)
);

-- A Specialist's request to turn Claude on for a workspace — approved via a
-- one-click emailed link (same token+expiry pattern as company_members'
-- invite flow), sent to the workspace creator and any active Team Leads.
CREATE TABLE IF NOT EXISTS claude_approval_requests (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    requested_by_user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    token TEXT NOT NULL,
    token_expires_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS claude_approval_requests_token_uniq ON claude_approval_requests (token);

-- One row per (company, cited URL, day) — bumped once for every citation
-- event recorded that day, so Top Winners/Losers can diff two real periods
-- instead of comparing against a running lifetime total. Starts empty; there
-- is no historical backfill possible since prior runs never recorded a date
-- breakdown.
CREATE TABLE IF NOT EXISTS citation_url_daily_stats (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    stat_date DATE NOT NULL,
    citation_count INT NOT NULL DEFAULT 0,
    UNIQUE (company_id, url, stat_date)
);

CREATE TABLE IF NOT EXISTS company_notification_prefs (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE UNIQUE,
    email_reports BOOLEAN NOT NULL DEFAULT TRUE,
    visibility_alerts BOOLEAN NOT NULL DEFAULT TRUE,
    member_activity BOOLEAN NOT NULL DEFAULT FALSE,
    weekly_digest BOOLEAN NOT NULL DEFAULT TRUE
);

-- One row per real AI provider call. Token counts come straight from each
-- provider's own API response; cost_usd is computed at write time from
-- providers/usage's price table (see that package for why it can drift).
CREATE TABLE IF NOT EXISTS api_usage_log (
    id SERIAL PRIMARY KEY,
    company_id INT REFERENCES companies(id) ON DELETE SET NULL,
    engine TEXT NOT NULL,
    purpose TEXT NOT NULL,
    model TEXT NOT NULL,
    input_tokens INT NOT NULL DEFAULT 0,
    output_tokens INT NOT NULL DEFAULT 0,
    cost_usd NUMERIC(12,6) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);
`

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env variables")
	}

	conf := config.GetConfig()
	db := databases.NewPostgresDatabase(conf.Database)

	db.GetConnection().MustExec(schema)

	log.Println("Migration completed successfully")
}
