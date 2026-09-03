-- ============================================================
-- RemoteScope — Database Migrations
-- Migration 001: Initial Schema
-- ============================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- For trigram similarity search
CREATE EXTENSION IF NOT EXISTS "unaccent";  -- For accent-insensitive search

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE user_role AS ENUM ('user', 'admin', 'moderator');
CREATE TYPE auth_provider AS ENUM ('email', 'google', 'github', 'linkedin');
CREATE TYPE experience_level AS ENUM ('internship', 'entry', 'junior', 'mid', 'senior', 'staff', 'principal', 'any');
CREATE TYPE employment_type AS ENUM ('full_time', 'part_time', 'contract', 'freelance', 'temporary', 'any');
CREATE TYPE remote_type AS ENUM ('worldwide', 'region_restricted', 'any');
CREATE TYPE notification_freq AS ENUM ('instant', 'hourly', 'daily', 'off');
CREATE TYPE fetch_method AS ENUM ('api', 'rss', 'html', 'webhook', 'graphql');
CREATE TYPE scraper_status AS ENUM ('running', 'success', 'failed', 'partial');
CREATE TYPE notification_channel AS ENUM ('email', 'push', 'sms', 'telegram', 'discord', 'slack', 'browser');
CREATE TYPE notification_type AS ENUM ('new_job', 'saved_search_match', 'application_update', 'system', 'marketing');
CREATE TYPE event_type AS ENUM ('view', 'click', 'save', 'apply', 'hide', 'share', 'search');
CREATE TYPE application_status AS ENUM ('saved', 'applied', 'phone_screen', 'interview', 'offer', 'rejected', 'withdrawn');

-- ============================================================
-- USERS
-- ============================================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           TEXT NOT NULL UNIQUE,
    password_hash   TEXT,                          -- NULL for OAuth-only users
    name            TEXT NOT NULL DEFAULT '',
    avatar_url      TEXT,
    auth_provider   auth_provider NOT NULL DEFAULT 'email',
    auth_provider_id TEXT,                         -- OAuth provider user ID
    email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    role            user_role NOT NULL DEFAULT 'user',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_auth_provider ON users(auth_provider, auth_provider_id);
CREATE INDEX idx_users_role ON users(role) WHERE role != 'user';

-- ============================================================
-- USER PREFERENCES (Onboarding)
-- ============================================================

CREATE TABLE user_preferences (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id                 UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    job_types               TEXT[] NOT NULL DEFAULT '{}',      -- e.g. ['software_engineering', 'devops']
    experience_levels       experience_level[] NOT NULL DEFAULT '{}',
    employment_types        employment_type[] NOT NULL DEFAULT '{}',
    technologies            TEXT[] NOT NULL DEFAULT '{}',      -- e.g. ['python', 'react', 'go']
    timezones               TEXT[] NOT NULL DEFAULT '{}',      -- e.g. ['worldwide', 'north_america']
    salary_min              INTEGER,                           -- Annual, in USD
    salary_max              INTEGER,
    salary_currency         CHAR(3) NOT NULL DEFAULT 'USD',
    remote_type             remote_type NOT NULL DEFAULT 'any',
    work_authorization      TEXT,                             -- e.g. 'us_citizen', 'eu_citizen'
    notification_freq       notification_freq NOT NULL DEFAULT 'daily',
    notification_channels   notification_channel[] NOT NULL DEFAULT '{email}',
    onboarding_completed    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- COMPANIES
-- ============================================================

CREATE TABLE companies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    website         TEXT,
    logo_url        TEXT,
    description     TEXT,
    size            TEXT,                  -- 'startup', '1-10', '11-50', '51-200', '201-500', '500+'
    funding_stage   TEXT,                  -- 'bootstrapped', 'seed', 'series_a', 'public', etc.
    tech_stack      TEXT[] NOT NULL DEFAULT '{}',
    glassdoor_url   TEXT,
    linkedin_url    TEXT,
    twitter_url     TEXT,
    is_verified     BOOLEAN NOT NULL DEFAULT FALSE,
    is_blacklisted  BOOLEAN NOT NULL DEFAULT FALSE,
    blacklist_reason TEXT,
    total_jobs      INTEGER NOT NULL DEFAULT 0,
    active_jobs     INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_companies_slug ON companies(slug);
CREATE INDEX idx_companies_name ON companies USING GIN(to_tsvector('english', name));
CREATE INDEX idx_companies_verified ON companies(is_verified) WHERE is_verified = TRUE;
CREATE INDEX idx_companies_active ON companies(is_blacklisted) WHERE is_blacklisted = FALSE;

-- ============================================================
-- JOB SOURCES (Adapter Registry)
-- ============================================================

CREATE TABLE job_sources (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    adapter_class   TEXT NOT NULL,         -- Python class name e.g. 'RemoteOKScraper'
    base_url        TEXT NOT NULL,
    fetch_method    fetch_method NOT NULL DEFAULT 'api',
    schedule_cron   TEXT NOT NULL DEFAULT '*/15 * * * *',  -- Every 15 min
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    respects_robots BOOLEAN NOT NULL DEFAULT TRUE,
    rate_limit_rpm  INTEGER NOT NULL DEFAULT 30,           -- Requests per minute
    last_run_at     TIMESTAMPTZ,
    last_success_at TIMESTAMPTZ,
    consecutive_failures INTEGER NOT NULL DEFAULT 0,
    config          JSONB NOT NULL DEFAULT '{}',           -- Source-specific config
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- CATEGORIES
-- ============================================================

CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL UNIQUE,
    slug        TEXT NOT NULL UNIQUE,
    parent_id   UUID REFERENCES categories(id),
    icon        TEXT,
    color       TEXT,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent ON categories(parent_id);

-- ============================================================
-- SKILLS (Normalized Taxonomy)
-- ============================================================

CREATE TABLE skills (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL UNIQUE,
    slug        TEXT NOT NULL UNIQUE,
    category    TEXT,                  -- 'language', 'framework', 'tool', 'cloud', 'database', etc.
    aliases     TEXT[] NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_skills_slug ON skills(slug);
CREATE INDEX idx_skills_aliases ON skills USING GIN(aliases);

-- ============================================================
-- JOBS (Canonical)
-- ============================================================

CREATE TABLE jobs (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title                   TEXT NOT NULL,
    normalized_title        TEXT,                              -- AI-normalized title
    company_id              UUID REFERENCES companies(id) ON DELETE SET NULL,
    description             TEXT NOT NULL DEFAULT '',
    description_summary     TEXT,                             -- AI-generated 2-sentence summary
    location                TEXT,
    is_remote               BOOLEAN NOT NULL DEFAULT TRUE,
    remote_type             remote_type NOT NULL DEFAULT 'worldwide',
    timezone_restrictions   TEXT[] NOT NULL DEFAULT '{}',
    salary_min              INTEGER,
    salary_max              INTEGER,
    salary_currency         CHAR(3) NOT NULL DEFAULT 'USD',
    salary_estimated        BOOLEAN NOT NULL DEFAULT FALSE,   -- TRUE if AI estimated
    experience_level        experience_level NOT NULL DEFAULT 'any',
    employment_type         employment_type NOT NULL DEFAULT 'full_time',
    skills                  TEXT[] NOT NULL DEFAULT '{}',
    technologies            TEXT[] NOT NULL DEFAULT '{}',
    benefits                TEXT[] NOT NULL DEFAULT '{}',
    visa_sponsorship        BOOLEAN,
    category                TEXT,                             -- AI-classified category
    subcategory             TEXT,
    keywords                TEXT[] NOT NULL DEFAULT '{}',
    apply_url               TEXT NOT NULL,
    canonical_url           TEXT,
    fingerprint             TEXT UNIQUE,                      -- For deduplication
    first_seen_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    posted_at               TIMESTAMPTZ,                      -- Original post date from source
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at              TIMESTAMPTZ,
    freshness_score         FLOAT NOT NULL DEFAULT 1.0,        -- 0.0 (stale) to 1.0 (fresh)
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    is_fake_suspected       BOOLEAN NOT NULL DEFAULT FALSE,
    fake_score              FLOAT,                            -- 0.0 (legit) to 1.0 (fake)
    ai_processed            BOOLEAN NOT NULL DEFAULT FALSE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Full-text search index
CREATE INDEX idx_jobs_fts ON jobs USING GIN(
    to_tsvector('english',
        coalesce(title, '') || ' ' ||
        coalesce(description, '') || ' ' ||
        array_to_string(skills, ' ') || ' ' ||
        array_to_string(technologies, ' ')
    )
);
CREATE INDEX idx_jobs_active ON jobs(is_active, created_at DESC) WHERE is_active = TRUE;
CREATE INDEX idx_jobs_company ON jobs(company_id);
CREATE INDEX idx_jobs_category ON jobs(category);
CREATE INDEX idx_jobs_experience ON jobs(experience_level);
CREATE INDEX idx_jobs_employment ON jobs(employment_type);
CREATE INDEX idx_jobs_remote_type ON jobs(remote_type);
CREATE INDEX idx_jobs_posted_at ON jobs(posted_at DESC NULLS LAST);
CREATE INDEX idx_jobs_first_seen ON jobs(first_seen_at DESC);
CREATE INDEX idx_jobs_salary ON jobs(salary_min, salary_max) WHERE salary_min IS NOT NULL;
CREATE INDEX idx_jobs_skills ON jobs USING GIN(skills);
CREATE INDEX idx_jobs_technologies ON jobs USING GIN(technologies);
CREATE INDEX idx_jobs_keywords ON jobs USING GIN(keywords);
CREATE INDEX idx_jobs_fingerprint ON jobs(fingerprint) WHERE fingerprint IS NOT NULL;

-- ============================================================
-- JOB EMBEDDINGS (pgvector)
-- ============================================================

CREATE TABLE job_embeddings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id          UUID NOT NULL UNIQUE REFERENCES jobs(id) ON DELETE CASCADE,
    embedding       vector(1536) NOT NULL,           -- OpenAI text-embedding-3-small
    model_version   TEXT NOT NULL DEFAULT 'text-embedding-3-small',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- HNSW index for fast ANN search
CREATE INDEX idx_job_embeddings_hnsw ON job_embeddings
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- ============================================================
-- JOB SOURCE URLS (Deduplication tracking)
-- ============================================================

CREATE TABLE job_source_urls (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id          UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    source_id       UUID NOT NULL REFERENCES job_sources(id),
    source_url      TEXT NOT NULL,
    source_job_id   TEXT,                      -- ID as used by the source
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source_id, source_url)
);

CREATE INDEX idx_job_source_urls_job ON job_source_urls(job_id);
CREATE INDEX idx_job_source_urls_source ON job_source_urls(source_id);

-- ============================================================
-- SCRAPER RUNS
-- ============================================================

CREATE TABLE scraper_runs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_id       UUID NOT NULL REFERENCES job_sources(id),
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    status          scraper_status NOT NULL DEFAULT 'running',
    jobs_found      INTEGER NOT NULL DEFAULT 0,
    jobs_new        INTEGER NOT NULL DEFAULT 0,
    jobs_updated    INTEGER NOT NULL DEFAULT 0,
    jobs_removed    INTEGER NOT NULL DEFAULT 0,
    error_message   TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scraper_runs_source ON scraper_runs(source_id, started_at DESC);
CREATE INDEX idx_scraper_runs_status ON scraper_runs(status);

-- ============================================================
-- SAVED JOBS
-- ============================================================

CREATE TABLE saved_jobs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id      UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, job_id)
);

CREATE INDEX idx_saved_jobs_user ON saved_jobs(user_id, created_at DESC);

-- ============================================================
-- APPLIED JOBS
-- ============================================================

CREATE TABLE applied_jobs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id      UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    status      application_status NOT NULL DEFAULT 'applied',
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, job_id)
);

CREATE INDEX idx_applied_jobs_user ON applied_jobs(user_id, applied_at DESC);

-- ============================================================
-- HIDDEN JOBS
-- ============================================================

CREATE TABLE hidden_jobs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id      UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    reason      TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, job_id)
);

CREATE INDEX idx_hidden_jobs_user ON hidden_jobs(user_id);

-- ============================================================
-- SEARCH HISTORY
-- ============================================================

CREATE TABLE search_history (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,  -- NULL for guests
    session_id      TEXT,
    query           TEXT NOT NULL,
    filters         JSONB NOT NULL DEFAULT '{}',
    result_count    INTEGER,
    clicked_job_ids UUID[] NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_search_history_user ON search_history(user_id, created_at DESC);

-- ============================================================
-- SAVED SEARCHES
-- ============================================================

CREATE TABLE saved_searches (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                    TEXT NOT NULL,
    query                   TEXT,
    filters                 JSONB NOT NULL DEFAULT '{}',
    notify                  BOOLEAN NOT NULL DEFAULT TRUE,
    notification_channels   notification_channel[] NOT NULL DEFAULT '{email}',
    last_notified_at        TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_saved_searches_user ON saved_searches(user_id);

-- ============================================================
-- NOTIFICATIONS
-- ============================================================

CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        notification_type NOT NULL,
    title       TEXT NOT NULL,
    body        TEXT NOT NULL,
    data        JSONB NOT NULL DEFAULT '{}',
    channel     notification_channel NOT NULL DEFAULT 'browser',
    is_read     BOOLEAN NOT NULL DEFAULT FALSE,
    is_sent     BOOLEAN NOT NULL DEFAULT FALSE,
    sent_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = FALSE;
CREATE INDEX idx_notifications_pending ON notifications(is_sent) WHERE is_sent = FALSE;

-- ============================================================
-- USER EVENTS (ML signal collection)
-- ============================================================

CREATE TABLE user_events (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,   -- NULL for guests
    session_id  TEXT,
    job_id      UUID REFERENCES jobs(id) ON DELETE CASCADE,
    event_type  event_type NOT NULL,
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_events_user ON user_events(user_id, created_at DESC);
CREATE INDEX idx_user_events_job ON user_events(job_id);
CREATE INDEX idx_user_events_type ON user_events(event_type, created_at DESC);

-- ============================================================
-- JOB ANALYTICS (aggregated)
-- ============================================================

CREATE TABLE job_analytics (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id              UUID NOT NULL UNIQUE REFERENCES jobs(id) ON DELETE CASCADE,
    views               INTEGER NOT NULL DEFAULT 0,
    clicks              INTEGER NOT NULL DEFAULT 0,
    saves               INTEGER NOT NULL DEFAULT 0,
    applications        INTEGER NOT NULL DEFAULT 0,
    last_activity_at    TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- AUDIT LOG
-- ============================================================

CREATE TABLE audit_log (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    action          TEXT NOT NULL,
    resource_type   TEXT NOT NULL,
    resource_id     UUID,
    old_value       JSONB,
    new_value       JSONB,
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_user ON audit_log(user_id, created_at DESC);
CREATE INDEX idx_audit_log_resource ON audit_log(resource_type, resource_id);

-- ============================================================
-- AUTO-UPDATE updated_at TRIGGER
-- ============================================================

CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_user_prefs_updated_at BEFORE UPDATE ON user_preferences FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_companies_updated_at BEFORE UPDATE ON companies FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_job_sources_updated_at BEFORE UPDATE ON job_sources FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_jobs_updated_at BEFORE UPDATE ON jobs FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_applied_jobs_updated_at BEFORE UPDATE ON applied_jobs FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_saved_searches_updated_at BEFORE UPDATE ON saved_searches FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_job_analytics_updated_at BEFORE UPDATE ON job_analytics FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- FRESHNESS SCORE: Auto-decay via view
-- (Updated by backend on a schedule)
-- ============================================================

CREATE OR REPLACE FUNCTION calculate_freshness_score(posted_at TIMESTAMPTZ)
RETURNS FLOAT AS $$
DECLARE
    hours_old FLOAT;
BEGIN
    IF posted_at IS NULL THEN RETURN 0.5; END IF;
    hours_old := EXTRACT(EPOCH FROM (NOW() - posted_at)) / 3600.0;
    -- Score decays from 1.0 to ~0.0 over 30 days
    RETURN GREATEST(0.0, 1.0 - (hours_old / 720.0));
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ============================================================
-- SEED: Job Sources
-- ============================================================

INSERT INTO job_sources (name, slug, adapter_class, base_url, fetch_method, schedule_cron, rate_limit_rpm) VALUES
('RemoteOK',         'remoteok',       'RemoteOKScraper',    'https://remoteok.com/remote-jobs.json',          'api',  '*/5 * * * *',  10),
('Remotive',         'remotive',       'RemotiveScraper',    'https://remotive.com/api/remote-jobs',           'api',  '*/10 * * * *', 30),
('Jobicy',           'jobicy',         'JobicyScraper',      'https://jobicy.com/api/v2/remote-jobs',          'api',  '*/10 * * * *', 30),
('We Work Remotely', 'weworkremotely', 'WWRScraper',         'https://weworkremotely.com',                     'rss',  '*/15 * * * *', 30),
('Greenhouse',       'greenhouse',     'GreenhouseScraper',  'https://boards-api.greenhouse.io/v1/boards',     'api',  '*/30 * * * *', 60),
('Lever',            'lever',          'LeverScraper',       'https://api.lever.co/v0/postings',               'api',  '*/30 * * * *', 60),
('Ashby',            'ashby',          'AshbyScraper',       'https://jobs.ashbyhq.com/api/non-user-graphql',  'api',  '*/30 * * * *', 60),
('Himalayas',        'himalayas',      'HimalayasScraper',   'https://himalayas.app/jobs/api',                 'api',  '*/15 * * * *', 30),
('Arc.dev',          'arc',            'ArcScraper',         'https://arc.dev/api',                            'api',  '0 * * * *',    20),
('JS Remotely',      'jsremotely',     'JSRemotelyScraper',  'https://jsremotely.com/feed',                    'rss',  '0 * * * *',    60),
('YC Companies',     'ycombinator',    'YCScraper',          'https://www.workatastartup.com/jobs',            'api',  '0 */2 * * *',  20);

-- ============================================================
-- SEED: Categories
-- ============================================================

INSERT INTO categories (name, slug, icon, color) VALUES
('Software Engineering', 'software-engineering',  '💻', '#3B82F6'),
('Frontend',             'frontend',              '🎨', '#8B5CF6'),
('Backend',              'backend',               '⚙️', '#06B6D4'),
('Full Stack',           'full-stack',            '🔧', '#10B981'),
('DevOps & Cloud',       'devops-cloud',          '☁️', '#F59E0B'),
('AI & Machine Learning','ai-ml',                 '🤖', '#EF4444'),
('Data Science',         'data-science',          '📊', '#6366F1'),
('Cybersecurity',        'cybersecurity',         '🔒', '#DC2626'),
('Mobile',               'mobile',                '📱', '#7C3AED'),
('Blockchain',           'blockchain',            '⛓️', '#D97706'),
('Embedded Systems',     'embedded-systems',      '🔌', '#374151'),
('Product Management',   'product-management',    '📋', '#0891B2'),
('UI/UX Design',         'ui-ux-design',          '✏️', '#EC4899'),
('Graphic Design',       'graphic-design',        '🖌️', '#F472B6'),
('Marketing',            'marketing',             '📣', '#FB923C'),
('Sales',                'sales',                 '💼', '#A16207'),
('Customer Support',     'customer-support',      '🎧', '#16A34A'),
('Finance',              'finance',               '💰', '#CA8A04'),
('HR & Recruiting',      'hr-recruiting',         '👥', '#9333EA'),
('Legal',                'legal',                 '⚖️', '#4B5563'),
('Writing & Content',    'writing-content',       '✍️', '#0284C7'),
('Healthcare',           'healthcare',            '🏥', '#059669');
