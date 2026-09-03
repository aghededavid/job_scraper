import { notFound } from 'next/navigation';
import Link from 'next/link';
import type { Metadata } from 'next';
import {
  MapPin, Clock, DollarSign, Briefcase, Globe,
  ExternalLink, ArrowLeft, Star, CheckCircle2, AlertTriangle
} from 'lucide-react';
import { JobCard } from '@/components/jobs/JobCard';
import { jobs as jobsApi } from '@/lib/api';
import { timeAgo, formatSalary, formatDate, REMOTE_TYPE_LABELS, EXPERIENCE_LABELS, EMPLOYMENT_LABELS, getCategoryColor } from '@/lib/utils';
import type { Job } from '@/lib/types';

interface Props { params: { id: string } }

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  try {
    const res = await jobsApi.get(params.id);
    if (!res.success || !res.data) return { title: 'Job Not Found' };
    const job = res.data;
    return {
      title: `${job.title} at ${job.company_name}`,
      description: job.summary || job.description.slice(0, 160),
    };
  } catch {
    return { title: 'Job | RemoteScope' };
  }
}

export const revalidate = 300;

async function getJob(id: string): Promise<Job | null> {
  try {
    const res = await jobsApi.get(id);
    return res.success && res.data ? res.data : null;
  } catch {
    return null;
  }
}

async function getRelated(job: Job): Promise<Job[]> {
  try {
    const res = await jobsApi.list({
      category: job.category,
      page: 1,
      per_page: 4,
    });
    const items = res.data?.data ?? [];
    return items.filter((j) => j.id !== job.id).slice(0, 3);
  } catch {
    return [];
  }
}

function JobDetailSkeleton() {
  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8 animate-pulse">
      <div className="h-4 w-24 skeleton rounded mb-6" />
      <div className="glass-card p-8 mb-6">
        <div className="flex gap-4 mb-6">
          <div className="w-16 h-16 skeleton rounded-2xl flex-shrink-0" />
          <div className="flex-1 space-y-2">
            <div className="h-7 skeleton rounded w-2/3" />
            <div className="h-4 skeleton rounded w-1/3" />
          </div>
        </div>
        <div className="flex gap-2 mb-6">
          {[80, 100, 70, 90].map((w) => (
            <div key={w} className="h-7 skeleton rounded-lg" style={{ width: w }} />
          ))}
        </div>
        <div className="space-y-3">
          {[100, 90, 95, 80, 85].map((w) => (
            <div key={w} className={`h-4 skeleton rounded`} style={{ width: `${w}%` }} />
          ))}
        </div>
      </div>
    </div>
  );
}

export default async function JobDetailPage({ params }: Props) {
  const job = await getJob(params.id);
  if (!job) return notFound();

  const related = await getRelated(job);
  const salary = formatSalary(job.salary_min, job.salary_max, job.salary_currency, job.salary_period);
  const categoryColor = getCategoryColor(job.category);

  return (
    <div className="bg-glow-top min-h-screen">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Back */}
        <Link
          href="/jobs"
          className="inline-flex items-center gap-2 text-sm text-surface-400 hover:text-white transition-colors mb-6 group"
        >
          <ArrowLeft className="w-4 h-4 group-hover:-translate-x-0.5 transition-transform" />
          Back to jobs
        </Link>

        {/* Main card */}
        <div className="glass-card p-8 mb-6 animate-fade-in">
          {/* Fake job warning */}
          {job.is_fake_suspected && (
            <div className="flex items-start gap-3 p-4 rounded-xl bg-accent-amber/10 border border-accent-amber/30 text-accent-amber text-sm mb-6">
              <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" />
              <span>Our AI flagged this listing as potentially suspicious. Exercise caution before applying.</span>
            </div>
          )}

          {/* Header */}
          <div className="flex items-start gap-5 mb-6">
            {/* Logo */}
            <div className="w-16 h-16 rounded-2xl bg-surface-600 flex items-center justify-center flex-shrink-0 overflow-hidden border border-white/[0.07]">
              {job.company_logo ? (
                <img src={job.company_logo} alt={job.company_name} className="w-full h-full object-contain p-2" />
              ) : (
                <span className="text-2xl font-bold text-surface-300">{job.company_name[0]}</span>
              )}
            </div>

            <div className="flex-1 min-w-0">
              {job.is_featured && (
                <div className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-brand-950/80 border border-brand-500/30 text-brand-300 text-xs font-medium mb-2">
                  <Star className="w-3 h-3" fill="currentColor" />
                  Featured
                </div>
              )}
              <h1 className="text-2xl font-bold text-white leading-tight mb-1">{job.title}</h1>
              <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-surface-300">
                <Link
                  href={`/companies/${job.company?.slug || job.company_name.toLowerCase().replace(/\s+/g, '-')}`}
                  className="font-semibold hover:text-white transition-colors flex items-center gap-1"
                >
                  {job.company_name}
                  {job.company?.is_verified && (
                    <CheckCircle2 className="w-3.5 h-3.5 text-accent-cyan" />
                  )}
                </Link>
                <span className="flex items-center gap-1">
                  <MapPin className="w-3.5 h-3.5" />
                  {job.location || REMOTE_TYPE_LABELS[job.remote_type] || 'Remote'}
                </span>
                <span className="flex items-center gap-1">
                  <Clock className="w-3.5 h-3.5" />
                  {timeAgo(job.posted_at || job.first_seen_at)}
                </span>
              </div>
            </div>

            {/* Apply CTA */}
            <a
              href={job.apply_url}
              target="_blank"
              rel="noopener noreferrer"
              id={`apply-job-detail-${job.id}`}
              className="btn-primary text-sm px-6 py-3 shadow-glow-md flex-shrink-0 hidden sm:inline-flex"
            >
              Apply Now
              <ExternalLink className="w-4 h-4" />
            </a>
          </div>

          {/* Tags */}
          <div className="flex flex-wrap gap-2 mb-6">
            <span className="tag">{REMOTE_TYPE_LABELS[job.remote_type] || job.remote_type}</span>
            {job.employment_type && (
              <span className="tag">
                <Briefcase className="w-3 h-3" />
                {EMPLOYMENT_LABELS[job.employment_type] || job.employment_type}
              </span>
            )}
            {job.experience_level && (
              <span className="tag">{EXPERIENCE_LABELS[job.experience_level] || job.experience_level}</span>
            )}
            {job.category && (
              <span
                className="tag font-medium"
                style={{ color: categoryColor, borderColor: categoryColor + '40', backgroundColor: categoryColor + '15' }}
              >
                {job.category}
              </span>
            )}
            {salary && (
              <span className="tag text-accent-emerald border-accent-emerald/30 bg-accent-emerald/10">
                <DollarSign className="w-3 h-3" />
                {salary}
              </span>
            )}
          </div>

          {/* Skills */}
          {job.skills?.length > 0 && (
            <div className="mb-6">
              <p className="section-label mb-3">Required Skills</p>
              <div className="flex flex-wrap gap-2">
                {job.skills.map((skill) => (
                  <span key={skill} className="tag-brand">{skill}</span>
                ))}
              </div>
            </div>
          )}

          {/* Mobile apply */}
          <a
            href={job.apply_url}
            target="_blank"
            rel="noopener noreferrer"
            className="btn-primary w-full justify-center py-3 text-base sm:hidden mb-6"
          >
            Apply Now
            <ExternalLink className="w-4 h-4" />
          </a>

          {/* Divider */}
          <div className="h-px bg-white/[0.06] my-6" />

          {/* Description */}
          <div>
            <h2 className="text-lg font-semibold text-white mb-4">Job Description</h2>
            {job.description_html ? (
              <div
                className="prose-dark"
                dangerouslySetInnerHTML={{ __html: job.description_html }}
              />
            ) : (
              <div className="prose-dark whitespace-pre-line text-surface-300 leading-relaxed">
                {job.description}
              </div>
            )}
          </div>
        </div>

        {/* Company card */}
        {job.company && (
          <div className="glass-card p-6 mb-6 animate-fade-in">
            <h2 className="text-base font-semibold text-white mb-4 flex items-center gap-2">
              About {job.company_name}
              {job.company.is_verified && (
                <CheckCircle2 className="w-4 h-4 text-accent-cyan" />
              )}
            </h2>
            <div className="flex flex-wrap gap-4 text-sm text-surface-300">
              {job.company.website && (
                <a
                  href={job.company.website}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-1.5 hover:text-white transition-colors"
                >
                  <Globe className="w-4 h-4" />
                  {job.company.website.replace(/^https?:\/\//, '')}
                </a>
              )}
              {job.company.employee_count && (
                <span>{job.company.employee_count.toLocaleString()}+ employees</span>
              )}
              {job.company.is_remote_friendly && (
                <span className="text-accent-emerald">🌍 Remote-friendly</span>
              )}
            </div>
            {job.company.description && (
              <p className="text-surface-300 text-sm mt-3 leading-relaxed">{job.company.description}</p>
            )}
            <Link
              href={`/companies/${job.company.slug}`}
              className="inline-flex items-center gap-1 text-sm text-brand-400 hover:text-brand-300 transition-colors mt-3"
            >
              View all jobs at {job.company_name}
              <ExternalLink className="w-3 h-3" />
            </Link>
          </div>
        )}

        {/* Meta */}
        <div className="glass-card p-6 mb-8 animate-fade-in">
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-4 text-sm">
            <div>
              <p className="section-label mb-1">Posted</p>
              <p className="text-white">{formatDate(job.posted_at || job.first_seen_at)}</p>
            </div>
            {job.expires_at && (
              <div>
                <p className="section-label mb-1">Expires</p>
                <p className="text-white">{formatDate(job.expires_at)}</p>
              </div>
            )}
            <div>
              <p className="section-label mb-1">Views</p>
              <p className="text-white">{job.view_count.toLocaleString()}</p>
            </div>
          </div>
        </div>

        {/* Related jobs */}
        {related.length > 0 && (
          <div className="animate-fade-in">
            <h2 className="text-xl font-bold text-white mb-4">Similar Jobs</h2>
            <div className="grid gap-4 md:grid-cols-2">
              {related.map((j) => (
                <JobCard key={j.id} job={j} />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
