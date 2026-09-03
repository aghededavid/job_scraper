-- =====================================================================
-- RemoteScope — 002 Seed Data Migration
-- Runs automatically on first `docker-compose up` via initdb.d
-- Safe to re-run (uses ON CONFLICT DO NOTHING)
-- =====================================================================

-- ── Categories ────────────────────────────────────────────────────────
INSERT INTO categories (id, name, slug, icon, color, sort_order) VALUES
  (uuid_generate_v4(), 'Frontend',            'frontend',           '🖥️',  '#3b82f6', 1),
  (uuid_generate_v4(), 'Backend',             'backend',            '⚙️',  '#8b5cf6', 2),
  (uuid_generate_v4(), 'Full Stack',          'full-stack',         '🔗',  '#06b6d4', 3),
  (uuid_generate_v4(), 'DevOps & Cloud',      'devops-cloud',       '☁️',  '#10b981', 4),
  (uuid_generate_v4(), 'AI & Machine Learning','ai-ml',             '🤖',  '#f59e0b', 5),
  (uuid_generate_v4(), 'Data Science',        'data-science',       '📊',  '#ef4444', 6),
  (uuid_generate_v4(), 'Mobile',              'mobile',             '📱',  '#ec4899', 7),
  (uuid_generate_v4(), 'UI/UX Design',        'ui-ux-design',       '🎨',  '#f97316', 8),
  (uuid_generate_v4(), 'Cybersecurity',       'cybersecurity',      '🔒',  '#dc2626', 9),
  (uuid_generate_v4(), 'Product Management',  'product-management', '📋',  '#7c3aed', 10),
  (uuid_generate_v4(), 'QA & Testing',        'qa-testing',         '🧪',  '#0ea5e9', 11),
  (uuid_generate_v4(), 'Blockchain & Web3',   'blockchain-web3',    '⛓️',  '#d97706', 12)
ON CONFLICT (slug) DO NOTHING;

-- ── Job Sources ───────────────────────────────────────────────────────
INSERT INTO job_sources (id, name, slug, adapter_class, base_url, fetch_method, schedule_cron, is_active) VALUES
  (uuid_generate_v4(), 'Remote OK',         'remoteok',      'RemoteOKScraper',   'https://remoteok.com',          'api',  '0 */6 * * *',  true),
  (uuid_generate_v4(), 'Remotive',          'remotive',      'RemotiveScraper',   'https://remotive.com',          'api',  '30 */6 * * *', true),
  (uuid_generate_v4(), 'Jobicy',            'jobicy',        'JobicyScraper',     'https://jobicy.com',            'api',  '0 */8 * * *',  true),
  (uuid_generate_v4(), 'We Work Remotely',  'weworkremotely','WWRScraper',        'https://weworkremotely.com',    'rss',  '15 */6 * * *', true),
  (uuid_generate_v4(), 'Greenhouse',        'greenhouse',    'GreenhouseScraper', 'https://boards.greenhouse.io',  'api',  '45 */12 * * *',true),
  (uuid_generate_v4(), 'Lever',             'lever',         'LeverScraper',      'https://jobs.lever.co',         'api',  '0 */12 * * *', true),
  (uuid_generate_v4(), 'Workable',          'workable',      'WorkableScraper',   'https://apply.workable.com',    'api',  '30 */12 * * *',false),
  (uuid_generate_v4(), 'LinkedIn Remote',   'linkedin',      'LinkedInScraper',   'https://www.linkedin.com/jobs', 'html', '0 */24 * * *', false)
ON CONFLICT (slug) DO NOTHING;

-- ── Demo Companies ────────────────────────────────────────────────────
-- A handful of well-known remote-friendly companies for local dev/testing
INSERT INTO companies (id, name, slug, website, logo_url, description, employee_count, is_remote_friendly, is_verified, is_blacklisted) VALUES
  (uuid_generate_v4(), 'GitHub',         'github',         'https://github.com',         'https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png',                    'Where the world builds software.',                    10000, true, true, false),
  (uuid_generate_v4(), 'Stripe',         'stripe',         'https://stripe.com',          'https://stripe.com/img/v3/home/social.png',                                                    'Payments infrastructure for the internet.',           8000,  true, true, false),
  (uuid_generate_v4(), 'Shopify',        'shopify',        'https://shopify.com',         'https://cdn.shopify.com/s/files/1/0070/7032/files/shopify-logo.png',                            'Commerce platform for businesses of all sizes.',      10000, true, true, false),
  (uuid_generate_v4(), 'Automattic',     'automattic',     'https://automattic.com',      'https://automattic.com/wp-content/uploads/2020/09/automattic-logo.png',                         'The people behind WordPress.com.',                    2000,  true, true, false),
  (uuid_generate_v4(), 'Gitlab',         'gitlab',         'https://gitlab.com',          'https://about.gitlab.com/images/press/logo/png/gitlab-logo-200.png',                            'The DevOps Platform.',                                2000,  true, true, false),
  (uuid_generate_v4(), 'Hashicorp',      'hashicorp',      'https://hashicorp.com',       'https://www.hashicorp.com/img/logo.svg',                                                         'Infrastructure automation software.',                 2000,  true, true, false),
  (uuid_generate_v4(), 'Vercel',         'vercel',         'https://vercel.com',          'https://assets.vercel.com/image/upload/front/favicon/vercel/favicon.ico',                       'Frontend cloud platform.',                            500,   true, true, false),
  (uuid_generate_v4(), 'Supabase',       'supabase',       'https://supabase.com',        'https://supabase.com/favicon/favicon-196x196.png',                                              'The open-source Firebase alternative.',               200,   true, true, false)
ON CONFLICT (slug) DO NOTHING;
