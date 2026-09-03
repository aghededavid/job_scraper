import type {
  APIResponse,
  AdminStats,
  Job,
  JobFilters,
  Notification,
  PaginatedResponse,
  SavedJob,
  SavedSearch,
  ScraperSource,
  TokenPair,
  User,
  UserPreferences,
  Category,
  Company,
} from './types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL
  ? `${process.env.NEXT_PUBLIC_API_URL}/api/v1`
  : '/api/v1';

// ── Fetch helper ─────────────────────────────────────────────────────

async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<APIResponse<T>> {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
    credentials: 'include', // send cookies
  });

  const json = await res.json();

  if (!res.ok) {
    return { success: false, error: json.error || `HTTP ${res.status}` };
  }

  return json;
}

// ── Auth ─────────────────────────────────────────────────────────────

export const auth = {
  signup: (name: string, email: string, password: string) =>
    apiFetch<{ user: User; tokens: TokenPair }>('/auth/signup', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    }),

  login: (email: string, password: string) =>
    apiFetch<{ user: User; tokens: TokenPair }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  logout: () => apiFetch('/auth/logout', { method: 'POST' }),

  refresh: (refreshToken?: string) =>
    apiFetch<TokenPair>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    }),

  oauthUrl: (provider: 'google' | 'github' | 'linkedin') =>
    `${API_BASE}/auth/oauth/${provider}`,
};

// ── Jobs ─────────────────────────────────────────────────────────────

export const jobs = {
  list: (filters: JobFilters = {}) => {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([k, v]) => {
      if (v !== undefined && v !== null && v !== '') {
        if (Array.isArray(v)) v.forEach((i) => params.append(k, i));
        else params.set(k, String(v));
      }
    });
    return apiFetch<PaginatedResponse<Job>>(`/jobs?${params.toString()}`);
  },

  get: (id: string) => apiFetch<Job>(`/jobs/${id}`),

  latest: () => apiFetch<Job[]>('/jobs/latest'),

  trending: () => apiFetch<Job[]>('/jobs/trending'),

  recommendations: () => apiFetch<Job[]>('/recommendations'),

  save: (id: string) => apiFetch(`/jobs/${id}/save`, { method: 'POST' }),

  unsave: (id: string) => apiFetch(`/jobs/${id}/save`, { method: 'DELETE' }),

  hide: (id: string, reason?: string) =>
    apiFetch(`/jobs/${id}/hide`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
    }),

  apply: (id: string) => apiFetch(`/jobs/${id}/apply`, { method: 'POST' }),
};

// ── Search ───────────────────────────────────────────────────────────

export const search = {
  query: (q: string, filters: JobFilters = {}) =>
    apiFetch<{ results: Job[]; total: number; query: string }>('/search', {
      method: 'POST',
      body: JSON.stringify({ q, ...filters }),
    }),

  suggest: (q: string) => {
    const params = new URLSearchParams({ q });
    return apiFetch<string[]>(`/search/suggest?${params}`);
  },
};

// ── Companies ────────────────────────────────────────────────────────

export const companies = {
  list: (q?: string) => {
    const params = q ? `?q=${encodeURIComponent(q)}` : '';
    return apiFetch<{ data: Company[]; total: number }>(`/companies${params}`);
  },
  get: (slug: string) => apiFetch<Company>(`/companies/${slug}`),
};

// ── Categories ───────────────────────────────────────────────────────

export const categories = {
  list: () => apiFetch<Category[]>('/categories'),
};

// ── User ─────────────────────────────────────────────────────────────

export const user = {
  me: () => apiFetch<User>('/users/me'),

  update: (data: { name?: string; avatar_url?: string }) =>
    apiFetch<User>('/users/me', { method: 'PUT', body: JSON.stringify(data) }),

  updatePreferences: (prefs: Partial<UserPreferences>) =>
    apiFetch<UserPreferences>('/users/me/preferences', {
      method: 'PUT',
      body: JSON.stringify(prefs),
    }),

  completeOnboarding: () =>
    apiFetch('/users/me/complete-onboarding', { method: 'POST' }),
};

// ── Saved Jobs ───────────────────────────────────────────────────────

export const savedJobs = {
  list: () =>
    apiFetch<{ items: SavedJob[]; total: number }>('/saved-jobs'),

  applied: () =>
    apiFetch<{ items: SavedJob[]; total: number }>('/applied-jobs'),
};

// ── Notifications ────────────────────────────────────────────────────

export const notifications = {
  list: () => apiFetch<{ items: Notification[]; unread: number }>('/notifications'),

  markRead: (id: string) =>
    apiFetch(`/notifications/${id}/read`, { method: 'PUT' }),

  markAllRead: () => apiFetch('/notifications/read-all', { method: 'PUT' }),

  savedSearches: () => apiFetch<SavedSearch[]>('/saved-searches'),

  createSavedSearch: (name: string, query: Record<string, unknown>) =>
    apiFetch<SavedSearch>('/saved-searches', {
      method: 'POST',
      body: JSON.stringify({ name, query }),
    }),

  deleteSavedSearch: (id: string) =>
    apiFetch(`/saved-searches/${id}`, { method: 'DELETE' }),
};

// ── Admin ────────────────────────────────────────────────────────────

export const admin = {
  stats: () => apiFetch<AdminStats>('/admin/stats'),

  scrapers: () => apiFetch<ScraperSource[]>('/admin/scrapers'),

  triggerScraper: (id: string) =>
    apiFetch(`/admin/scrapers/${id}/run`, { method: 'POST' }),

  scraperRuns: () => apiFetch<unknown[]>('/admin/scraper-runs'),

  users: () => apiFetch<{ data: User[]; total: number }>('/admin/users'),

  updateUserRole: (id: string, role: string) =>
    apiFetch(`/admin/users/${id}/role`, {
      method: 'PUT',
      body: JSON.stringify({ role }),
    }),

  deleteJob: (id: string) => apiFetch(`/admin/jobs/${id}`, { method: 'DELETE' }),

  flagJob: (id: string) =>
    apiFetch(`/admin/jobs/${id}/flag-fake`, { method: 'POST' }),

  blacklistCompany: (id: string, reason: string) =>
    apiFetch(`/admin/companies/${id}/blacklist`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
    }),
};
