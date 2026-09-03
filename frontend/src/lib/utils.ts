import { type ClassValue, clsx } from 'clsx';
import { formatDistanceToNow, format } from 'date-fns';
import type { Job } from './types';

// ── Class name utility ────────────────────────────────────────────
export function cn(...inputs: ClassValue[]) {
  return clsx(inputs);
}

// ── Date formatting ───────────────────────────────────────────────
export function timeAgo(date: string | Date) {
  return formatDistanceToNow(new Date(date), { addSuffix: true });
}

export function formatDate(date: string | Date, fmt = 'MMM d, yyyy') {
  return format(new Date(date), fmt);
}

// ── Salary formatting ─────────────────────────────────────────────
export function formatSalary(
  min?: number,
  max?: number,
  currency = 'USD',
  period = 'yearly'
): string {
  if (!min && !max) return '';

  const fmt = (n: number) =>
    new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency,
      notation: 'compact',
      maximumFractionDigits: 0,
    }).format(n);

  const suffix = period === 'yearly' ? '/yr' : period === 'monthly' ? '/mo' : '/hr';

  if (min && max) return `${fmt(min)} – ${fmt(max)}${suffix}`;
  if (min) return `${fmt(min)}+${suffix}`;
  if (max) return `Up to ${fmt(max)}${suffix}`;
  return '';
}

// ── Remote type label ─────────────────────────────────────────────
export const REMOTE_TYPE_LABELS: Record<string, string> = {
  'fully-remote':    '🌍 Fully Remote',
  'remote-friendly': '🔀 Remote-Friendly',
  'hybrid':          '🏢 Hybrid',
};

export const EXPERIENCE_LABELS: Record<string, string> = {
  entry:     'Entry Level',
  mid:       'Mid Level',
  senior:    'Senior',
  lead:      'Lead / Staff',
  executive: 'Executive',
};

export const EMPLOYMENT_LABELS: Record<string, string> = {
  'full-time':   'Full-Time',
  'part-time':   'Part-Time',
  contract:      'Contract',
  freelance:     'Freelance',
  internship:    'Internship',
};

// ── Posted within labels ──────────────────────────────────────────
export const POSTED_WITHIN_OPTIONS = [
  { value: '',    label: 'Any time' },
  { value: '1h',  label: 'Last hour' },
  { value: '24h', label: 'Last 24h' },
  { value: '7d',  label: 'Last 7 days' },
  { value: '30d', label: 'Last 30 days' },
];

// ── Category colors ───────────────────────────────────────────────
export function getCategoryColor(category?: string): string {
  const map: Record<string, string> = {
    'Frontend':             '#3b82f6',
    'Backend':              '#8b5cf6',
    'Full Stack':           '#06b6d4',
    'DevOps & Cloud':       '#10b981',
    'AI & Machine Learning':'#f59e0b',
    'Data Science':         '#ef4444',
    'Mobile':               '#ec4899',
    'UI/UX Design':         '#f97316',
    'Cybersecurity':        '#dc2626',
    'Product Management':   '#7c3aed',
  };
  return map[category || ''] || '#6680ff';
}

// ── Skill popularity (used for sizing tags) ───────────────────────
export function truncateSkills(skills: string[], max = 5) {
  if (!skills) return { visible: [], hidden: 0 };
  const visible = skills.slice(0, max);
  const hidden = Math.max(0, skills.length - max);
  return { visible, hidden };
}

// ── Job apply link ────────────────────────────────────────────────
export function buildApplyUrl(job: Job): string {
  return job.apply_url || job.company_website || '#';
}

// ── URL search params ─────────────────────────────────────────────
export function buildQueryString(params: Record<string, unknown>): string {
  const sp = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return;
    if (Array.isArray(v)) v.forEach((i) => sp.append(k, String(i)));
    else sp.set(k, String(v));
  });
  return sp.toString();
}

// ── Debounce ──────────────────────────────────────────────────────
export function debounce<T extends (...args: unknown[]) => unknown>(
  fn: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timer: ReturnType<typeof setTimeout>;
  return (...args: Parameters<T>) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  };
}
