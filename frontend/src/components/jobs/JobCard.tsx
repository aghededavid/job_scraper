'use client';

import { useState } from 'react';
import Link from 'next/link';
import {
  Bookmark, BookmarkCheck, MapPin, Clock, DollarSign,
  ExternalLink, Star, Eye, Briefcase, EyeOff, CheckCircle2
} from 'lucide-react';
import { cn, timeAgo, formatSalary, REMOTE_TYPE_LABELS, EXPERIENCE_LABELS, truncateSkills, getCategoryColor } from '@/lib/utils';
import { jobs as jobsApi } from '@/lib/api';
import type { Job } from '@/lib/types';

interface JobCardProps {
  job: Job;
  variant?: 'default' | 'compact' | 'featured';
  onUpdate?: (id: string, updates: Partial<Job>) => void;
}

export function JobCard({ job, variant = 'default', onUpdate }: JobCardProps) {
  const [saved, setSaved] = useState(job.is_saved ?? false);
  const [hidden, setHidden] = useState(job.is_hidden ?? false);
  const [applied, setApplied] = useState(job.is_applied ?? false);
  const [saving, setSaving] = useState(false);

  const { visible: visibleSkills, hidden: hiddenSkillCount } = truncateSkills(job.skills, variant === 'compact' ? 3 : 6);
  const salary = formatSalary(job.salary_min, job.salary_max, job.salary_currency, job.salary_period);
  const categoryColor = getCategoryColor(job.category);

  const handleSave = async (e: React.MouseEvent) => {
    e.preventDefault();
    if (saving) return;
    setSaving(true);
    try {
      if (saved) {
        await jobsApi.unsave(job.id);
        setSaved(false);
        onUpdate?.(job.id, { is_saved: false });
      } else {
        await jobsApi.save(job.id);
        setSaved(true);
        onUpdate?.(job.id, { is_saved: true });
      }
    } finally {
      setSaving(false);
    }
  };

  const handleHide = async (e: React.MouseEvent) => {
    e.preventDefault();
    await jobsApi.hide(job.id);
    setHidden(true);
    onUpdate?.(job.id, { is_hidden: true });
  };

  if (hidden) return null;

  if (variant === 'compact') {
    return (
      <Link
        href={`/jobs/${job.id}`}
        className="glass-card p-4 flex items-center gap-4 hover:border-white/[0.12] hover:shadow-card-hover transition-all group animate-fade-in"
      >
        {/* Logo */}
        <div className="w-10 h-10 rounded-lg bg-surface-600 flex items-center justify-center flex-shrink-0 overflow-hidden">
          {job.company_logo ? (
            <img src={job.company_logo} alt={job.company_name} className="w-full h-full object-contain p-1" />
          ) : (
            <span className="text-lg font-bold text-surface-300">{job.company_name[0]}</span>
          )}
        </div>
        {/* Info */}
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-white truncate group-hover:text-brand-300 transition-colors">{job.title}</p>
          <p className="text-xs text-surface-300 truncate">{job.company_name}</p>
        </div>
        {/* Meta */}
        <div className="hidden sm:flex items-center gap-2 flex-shrink-0">
          {salary && <span className="text-xs text-accent-emerald font-medium">{salary}</span>}
          <span className="text-xs text-surface-400">{timeAgo(job.posted_at || job.first_seen_at)}</span>
        </div>
      </Link>
    );
  }

  return (
    <article
      className={cn(
        'glass-card p-5 flex flex-col gap-4 hover:border-white/[0.12] hover:shadow-card-hover transition-all duration-200 animate-fade-in group relative',
        job.is_featured && 'ring-1 ring-brand-500/40'
      )}
    >
      {/* Featured badge */}
      {job.is_featured && (
        <div className="absolute top-4 right-4 flex items-center gap-1 px-2 py-1 rounded-lg bg-brand-950/80 border border-brand-500/30 text-brand-300 text-xs font-medium">
          <Star className="w-3 h-3" fill="currentColor" />
          Featured
        </div>
      )}

      {/* Header */}
      <div className="flex items-start gap-3">
        {/* Company logo */}
        <Link href={`/companies/${job.company?.slug || job.company_name.toLowerCase()}`}>
          <div className="w-12 h-12 rounded-xl bg-surface-600 flex items-center justify-center flex-shrink-0 overflow-hidden border border-white/[0.06] hover:border-white/10 transition-colors">
            {job.company_logo ? (
              <img src={job.company_logo} alt={job.company_name} className="w-full h-full object-contain p-1.5" />
            ) : (
              <span className="text-xl font-bold text-surface-300">{job.company_name[0]}</span>
            )}
          </div>
        </Link>

        <div className="flex-1 min-w-0">
          {/* Job title */}
          <Link href={`/jobs/${job.id}`}>
            <h3 className="font-semibold text-white text-base leading-tight group-hover:text-brand-300 transition-colors mb-1 pr-16">
              {job.title}
            </h3>
          </Link>
          {/* Company + location */}
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-surface-300">
            <Link
              href={`/companies/${job.company?.slug || job.company_name.toLowerCase()}`}
              className="hover:text-white transition-colors font-medium"
            >
              {job.company_name}
            </Link>
            {job.company?.is_verified && (
              <CheckCircle2 className="w-3.5 h-3.5 text-accent-cyan flex-shrink-0" />
            )}
            <span className="flex items-center gap-1">
              <MapPin className="w-3 h-3" />
              {job.location || REMOTE_TYPE_LABELS[job.remote_type] || 'Remote'}
            </span>
          </div>
        </div>
      </div>

      {/* Tags row */}
      <div className="flex flex-wrap items-center gap-2">
        {/* Remote type */}
        <span className="tag">
          {REMOTE_TYPE_LABELS[job.remote_type] || job.remote_type}
        </span>

        {/* Employment type */}
        {job.employment_type && (
          <span className="tag">
            <Briefcase className="w-3 h-3" />
            {job.employment_type.replace('-', ' ')}
          </span>
        )}

        {/* Experience */}
        {job.experience_level && (
          <span className="tag">
            {EXPERIENCE_LABELS[job.experience_level] || job.experience_level}
          </span>
        )}

        {/* Category */}
        {job.category && (
          <span
            className="tag font-medium"
            style={{ color: categoryColor, borderColor: categoryColor + '40', backgroundColor: categoryColor + '15' }}
          >
            {job.category}
          </span>
        )}
      </div>

      {/* Skills */}
      {visibleSkills.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {visibleSkills.map((skill) => (
            <span key={skill} className="tag-brand text-xs">
              {skill}
            </span>
          ))}
          {hiddenSkillCount > 0 && (
            <span className="tag text-xs text-surface-400">+{hiddenSkillCount} more</span>
          )}
        </div>
      )}

      {/* Footer row */}
      <div className="flex items-center justify-between mt-auto pt-1 border-t border-white/[0.04]">
        {/* Meta */}
        <div className="flex items-center gap-3 text-xs text-surface-400">
          <span className="flex items-center gap-1">
            <Clock className="w-3 h-3" />
            {timeAgo(job.posted_at || job.first_seen_at)}
          </span>
          {job.view_count > 0 && (
            <span className="flex items-center gap-1">
              <Eye className="w-3 h-3" />
              {job.view_count.toLocaleString()}
            </span>
          )}
          {salary && (
            <span className="flex items-center gap-1 text-accent-emerald font-medium">
              <DollarSign className="w-3 h-3" />
              {salary}
            </span>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1">
          <button
            id={`hide-job-${job.id}`}
            onClick={handleHide}
            title="Hide job"
            className="p-1.5 rounded-lg text-surface-400 hover:text-surface-300 hover:bg-white/[0.04] transition-all opacity-0 group-hover:opacity-100"
          >
            <EyeOff className="w-4 h-4" />
          </button>

          <button
            id={`save-job-${job.id}`}
            onClick={handleSave}
            disabled={saving}
            title={saved ? 'Unsave' : 'Save'}
            className={cn(
              'p-1.5 rounded-lg transition-all',
              saved
                ? 'text-brand-400 bg-brand-950/60 hover:bg-brand-950'
                : 'text-surface-400 hover:text-white hover:bg-white/[0.06]'
            )}
          >
            {saved
              ? <BookmarkCheck className="w-4 h-4" />
              : <Bookmark className="w-4 h-4" />
            }
          </button>

          <a
            href={job.apply_url}
            target="_blank"
            rel="noopener noreferrer"
            id={`apply-job-${job.id}`}
            className="btn-primary py-1.5 px-3 text-xs"
            onClick={(e) => e.stopPropagation()}
          >
            Apply
            <ExternalLink className="w-3 h-3" />
          </a>
        </div>
      </div>
    </article>
  );
}
