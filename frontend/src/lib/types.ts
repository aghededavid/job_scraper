// ── Types matching the Go backend models ───────────────────────────

export type UserRole = 'user' | 'premium' | 'admin';
export type AuthProvider = 'email' | 'google' | 'github' | 'linkedin';
export type ExperienceLevel = 'entry' | 'mid' | 'senior' | 'lead' | 'executive';
export type EmploymentType = 'full-time' | 'part-time' | 'contract' | 'freelance' | 'internship';
export type RemoteType = 'fully-remote' | 'remote-friendly' | 'hybrid';
export type SalaryPeriod = 'hourly' | 'monthly' | 'yearly';
export type ApplicationStatus = 'bookmarked' | 'applied' | 'phone_screen' | 'interview' | 'offer' | 'rejected' | 'accepted' | 'withdrawn';

export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url?: string;
  role: UserRole;
  email_verified: boolean;
  auth_provider: AuthProvider;
  created_at: string;
}

export interface UserPreferences {
  id: string;
  user_id: string;
  preferred_categories: string[];
  preferred_skills: string[];
  preferred_experience_levels: ExperienceLevel[];
  preferred_employment_types: EmploymentType[];
  preferred_remote_types: RemoteType[];
  salary_min?: number;
  salary_max?: number;
  preferred_timezones: string[];
  email_notifications: boolean;
  push_notifications: boolean;
  notification_frequency: string;
  onboarding_completed: boolean;
}

export interface Company {
  id: string;
  name: string;
  slug: string;
  website?: string;
  logo_url?: string;
  description?: string;
  employee_count?: number;
  is_remote_friendly: boolean;
  is_verified: boolean;
  is_blacklisted: boolean;
}

export interface Job {
  id: string;
  source_id: string;
  external_id: string;
  title: string;
  normalized_title?: string;
  company_id?: string;
  company?: Company;
  company_name: string;
  company_logo?: string;
  company_website?: string;
  description: string;
  description_html?: string;
  summary?: string;
  location: string;
  remote_type: RemoteType;
  experience_level?: ExperienceLevel;
  employment_type?: EmploymentType;
  skills: string[];
  tags: string[];
  category?: string;
  salary_min?: number;
  salary_max?: number;
  salary_currency?: string;
  salary_period?: SalaryPeriod;
  apply_url: string;
  is_active: boolean;
  is_featured: boolean;
  is_fake_suspected: boolean;
  view_count: number;
  apply_count: number;
  first_seen_at: string;
  posted_at: string;
  expires_at?: string;
  // user-context
  is_saved?: boolean;
  is_hidden?: boolean;
  is_applied?: boolean;
}

export interface SavedJob {
  id: string;
  user_id: string;
  job_id: string;
  job?: Job;
  status: ApplicationStatus;
  notes?: string;
  hide_reason?: string;
  applied_at?: string;
  created_at: string;
}

export interface Notification {
  id: string;
  user_id: string;
  type: string;
  title: string;
  body: string;
  data?: Record<string, unknown>;
  read: boolean;
  created_at: string;
}

export interface SavedSearch {
  id: string;
  user_id: string;
  name: string;
  query: Record<string, unknown>;
  created_at: string;
}

export interface Category {
  name: string;
  slug: string;
  icon: string;
  color: string;
}

export interface JobFilters {
  q?: string;
  category?: string;
  experience?: ExperienceLevel[];
  employment_type?: EmploymentType[];
  remote_type?: RemoteType;
  salary_min?: number;
  salary_max?: number;
  tech?: string[];
  posted_within?: string;
  page?: number;
  per_page?: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface APIResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface ScraperSource {
  id: string;
  name: string;
  slug: string;
  adapter_class: string;
  base_url: string;
  fetch_method: string;
  schedule_cron: string;
  is_active: boolean;
  last_run_at?: string;
  last_run_status?: string;
  jobs_scraped_total: number;
}

export interface AdminStats {
  total_users: number;
  total_jobs: number;
  active_jobs: number;
  total_companies: number;
  jobs_today: number;
  jobs_this_hour: number;
  total_sources: number;
  active_sources: number;
}
