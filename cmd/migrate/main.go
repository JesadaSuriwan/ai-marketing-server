package main

import (
	"github.com/joho/godotenv"
	"log"
	"github.com/ai-marketing/ai-marketing-server/config"
	"github.com/ai-marketing/ai-marketing-server/databases"
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

CREATE TABLE IF NOT EXISTS prompt_categories (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS prompts (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    category_id INT REFERENCES prompt_categories(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
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
    created_at TIMESTAMP DEFAULT NOW()
);

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

CREATE TABLE IF NOT EXISTS subdomains (
    id SERIAL PRIMARY KEY,
    company_id INT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    subdomain TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'Active',
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
