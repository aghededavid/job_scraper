'use client';

import { useState } from 'react';
import { ChevronDown, X, SlidersHorizontal } from 'lucide-react';
import { cn, REMOTE_TYPE_LABELS, EXPERIENCE_LABELS, EMPLOYMENT_LABELS, POSTED_WITHIN_OPTIONS } from '@/lib/utils';
import type { JobFilters, ExperienceLevel, EmploymentType, RemoteType } from '@/lib/types';

interface FilterPanelProps {
  filters: JobFilters;
  onChange: (filters: JobFilters) => void;
  categories?: { name: string; slug: string; icon: string; color: string }[];
}

const POPULAR_SKILLS = [
  'TypeScript', 'Python', 'React', 'Node.js', 'Go', 'Rust',
  'AWS', 'Docker', 'Kubernetes', 'PostgreSQL', 'GraphQL', 'Next.js',
];

export function FilterPanel({ filters, onChange, categories = [] }: FilterPanelProps) {
  const [collapsed, setCollapsed] = useState(false);

  const update = (patch: Partial<JobFilters>) =>
    onChange({ ...filters, ...patch, page: 1 });

  const toggleArray = <T extends string>(
    arr: T[] | undefined,
    val: T
  ): T[] => {
    const current = arr ?? [];
    return current.includes(val)
      ? current.filter((v) => v !== val)
      : [...current, val];
  };

  const hasActiveFilters =
    filters.category ||
    (filters.experience?.length ?? 0) > 0 ||
    (filters.employment_type?.length ?? 0) > 0 ||
    filters.remote_type ||
    filters.salary_min ||
    filters.salary_max ||
    (filters.tech?.length ?? 0) > 0 ||
    filters.posted_within;

  const clearAll = () =>
    onChange({ page: 1, per_page: filters.per_page, q: filters.q });

  return (
    <aside className="glass-card p-4 space-y-5 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <SlidersHorizontal className="w-4 h-4 text-brand-400" />
          <span className="font-semibold text-sm text-white">Filters</span>
          {hasActiveFilters && (
            <span className="px-1.5 py-0.5 rounded-md bg-brand-500/20 text-brand-300 text-xs font-medium">
              Active
            </span>
          )}
        </div>
        <div className="flex items-center gap-1">
          {hasActiveFilters && (
            <button
              onClick={clearAll}
              className="flex items-center gap-1 text-xs text-surface-400 hover:text-accent-rose transition-colors px-2 py-1 rounded-lg hover:bg-accent-rose/10"
            >
              <X className="w-3 h-3" />
              Clear
            </button>
          )}
          <button
            onClick={() => setCollapsed(!collapsed)}
            className="p-1.5 rounded-lg text-surface-400 hover:text-white hover:bg-white/[0.05] transition-all"
          >
            <ChevronDown className={cn('w-4 h-4 transition-transform', collapsed && 'rotate-180')} />
          </button>
        </div>
      </div>

      {!collapsed && (
        <>
          {/* Category */}
          {categories.length > 0 && (
            <FilterSection title="Category">
              <div className="space-y-1">
                {categories.map((cat) => (
                  <button
                    key={cat.slug}
                    onClick={() => update({ category: filters.category === cat.slug ? undefined : cat.slug })}
                    className={cn(
                      'w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-all text-left',
                      filters.category === cat.slug
                        ? 'bg-brand-500/20 text-white border border-brand-500/30'
                        : 'text-surface-300 hover:text-white hover:bg-white/[0.05]'
                    )}
                  >
                    <span>{cat.icon}</span>
                    <span>{cat.name}</span>
                  </button>
                ))}
              </div>
            </FilterSection>
          )}

          {/* Remote Type */}
          <FilterSection title="Remote Type">
            <div className="space-y-1">
              {(Object.entries(REMOTE_TYPE_LABELS) as [RemoteType, string][]).map(([val, label]) => (
                <FilterCheckbox
                  key={val}
                  label={label}
                  checked={filters.remote_type === val}
                  onChange={() => update({ remote_type: filters.remote_type === val ? undefined : val })}
                />
              ))}
            </div>
          </FilterSection>

          {/* Experience Level */}
          <FilterSection title="Experience">
            <div className="space-y-1">
              {(Object.entries(EXPERIENCE_LABELS) as [ExperienceLevel, string][]).map(([val, label]) => (
                <FilterCheckbox
                  key={val}
                  label={label}
                  checked={(filters.experience ?? []).includes(val)}
                  onChange={() => update({ experience: toggleArray(filters.experience, val) })}
                />
              ))}
            </div>
          </FilterSection>

          {/* Employment Type */}
          <FilterSection title="Employment Type">
            <div className="space-y-1">
              {(Object.entries(EMPLOYMENT_LABELS) as [EmploymentType, string][]).map(([val, label]) => (
                <FilterCheckbox
                  key={val}
                  label={label}
                  checked={(filters.employment_type ?? []).includes(val)}
                  onChange={() => update({ employment_type: toggleArray(filters.employment_type, val) })}
                />
              ))}
            </div>
          </FilterSection>

          {/* Salary Range */}
          <FilterSection title="Salary Range (USD/yr)">
            <div className="grid grid-cols-2 gap-2">
              <input
                type="number"
                placeholder="Min"
                value={filters.salary_min ?? ''}
                onChange={(e) => update({ salary_min: e.target.value ? Number(e.target.value) : undefined })}
                className="input text-xs py-2"
              />
              <input
                type="number"
                placeholder="Max"
                value={filters.salary_max ?? ''}
                onChange={(e) => update({ salary_max: e.target.value ? Number(e.target.value) : undefined })}
                className="input text-xs py-2"
              />
            </div>
          </FilterSection>

          {/* Posted Within */}
          <FilterSection title="Posted Within">
            <div className="space-y-1">
              {POSTED_WITHIN_OPTIONS.map((opt) => (
                <FilterCheckbox
                  key={opt.value}
                  label={opt.label}
                  checked={(filters.posted_within ?? '') === opt.value}
                  onChange={() => update({ posted_within: opt.value || undefined })}
                />
              ))}
            </div>
          </FilterSection>

          {/* Skills */}
          <FilterSection title="Skills">
            <div className="flex flex-wrap gap-1.5">
              {POPULAR_SKILLS.map((skill) => (
                <button
                  key={skill}
                  onClick={() => update({ tech: toggleArray(filters.tech, skill) })}
                  className={cn(
                    'px-2.5 py-1 rounded-lg text-xs font-medium transition-all border',
                    (filters.tech ?? []).includes(skill)
                      ? 'bg-brand-500/20 text-brand-300 border-brand-500/40'
                      : 'text-surface-300 border-white/[0.06] hover:border-white/10 hover:text-white bg-surface-700/60'
                  )}
                >
                  {skill}
                </button>
              ))}
            </div>
          </FilterSection>
        </>
      )}
    </aside>
  );
}

function FilterSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <p className="section-label mb-2.5">{title}</p>
      {children}
    </div>
  );
}

function FilterCheckbox({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: () => void;
}) {
  return (
    <label className="flex items-center gap-2.5 px-2 py-1.5 rounded-lg cursor-pointer group hover:bg-white/[0.03] transition-all">
      <div
        className={cn(
          'w-4 h-4 rounded flex items-center justify-center border transition-all flex-shrink-0',
          checked
            ? 'bg-brand-500 border-brand-500'
            : 'border-white/20 group-hover:border-white/30'
        )}
        onClick={onChange}
      >
        {checked && (
          <svg className="w-2.5 h-2.5 text-white" fill="none" viewBox="0 0 10 10">
            <path d="M1.5 5l2.5 2.5 4.5-4.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        )}
      </div>
      <span
        className={cn('text-sm transition-colors', checked ? 'text-white' : 'text-surface-300 group-hover:text-white')}
        onClick={onChange}
      >
        {label}
      </span>
    </label>
  );
}
