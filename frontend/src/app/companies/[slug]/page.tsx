import { notFound } from 'next/navigation';
import Link from 'next/link';
import type { Metadata } from 'next';
import {
  Globe, Users, CheckCircle2, ArrowLeft, ExternalLink, Briefcase
} from 'lucide-react';
import { JobCard } from '@/components/jobs/JobCard';
import { companies as companiesApi, jobs as jobsApi } from '@/lib/api';
import type { Company, Job } from '@/lib/types';

interface Props { params: { slug: string } }

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  try {
    const res = await companiesApi.get(params.slug);
    if (!res.success || !res.data) return { title: 'Company Not Found' };
    return {
      title: `${res.data.name} — Remote Jobs`,
      description: res.data.description || `Browse remote jobs at ${res.data.name}`,
    };
  } catch {
    return { title: 'Company | RemoteScope' };
  }
}

export const revalidate = 600;

async function getData(slug: string): Promise<{ company: Company | null; jobs: Job[] }> {
  try {
    const [compRes, jobsRes] = await Promise.all([
      companiesApi.get(slug),
      jobsApi.list({ per_page: 12 }),
    ]);
    const company = compRes.success ? compRes.data ?? null : null;
    // Filter jobs by company name (since we don't have a direct company jobs endpoint)
    const allJobs = jobsRes.data?.data ?? [];
    const companyJobs = company
      ? allJobs.filter((j) => j.company_name.toLowerCase() === company.name.toLowerCase())
      : [];
    return { company, jobs: companyJobs };
  } catch {
    return { company: null, jobs: [] };
  }
}

export default async function CompanyDetailPage({ params }: Props) {
  const { company, jobs } = await getData(params.slug);
  if (!company) return notFound();

  return (
    <div className="bg-glow-top min-h-screen">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Back */}
        <Link
          href="/companies"
          className="inline-flex items-center gap-2 text-sm text-surface-400 hover:text-white transition-colors mb-6 group"
        >
          <ArrowLeft className="w-4 h-4 group-hover:-translate-x-0.5 transition-transform" />
          All Companies
        </Link>

        {/* Company header */}
        <div className="glass-card p-8 mb-6 animate-fade-in">
          <div className="flex items-start gap-6">
            {/* Logo */}
            <div className="w-20 h-20 rounded-2xl bg-surface-600 flex items-center justify-center flex-shrink-0 border border-white/[0.07] overflow-hidden">
              {company.logo_url ? (
                <img src={company.logo_url} alt={company.name} className="w-full h-full object-contain p-2" />
              ) : (
                <span className="text-3xl font-bold text-surface-300">{company.name[0]}</span>
              )}
            </div>

            <div className="flex-1">
              <div className="flex items-center gap-2 mb-1">
                <h1 className="text-2xl font-bold text-white">{company.name}</h1>
                {company.is_verified && (
                  <CheckCircle2 className="w-5 h-5 text-accent-cyan" />
                )}
              </div>

              <div className="flex flex-wrap items-center gap-4 text-sm text-surface-300 mb-4">
                {company.is_remote_friendly && (
                  <span className="flex items-center gap-1 text-accent-emerald font-medium">
                    <Globe className="w-4 h-4" />
                    Remote-friendly
                  </span>
                )}
                {company.employee_count && (
                  <span className="flex items-center gap-1">
                    <Users className="w-4 h-4" />
                    {company.employee_count.toLocaleString()}+ employees
                  </span>
                )}
              </div>

              {company.description && (
                <p className="text-surface-300 leading-relaxed mb-4">{company.description}</p>
              )}

              {company.website && (
                <a
                  href={company.website}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-2 text-sm text-brand-400 hover:text-brand-300 transition-colors"
                >
                  <Globe className="w-4 h-4" />
                  {company.website.replace(/^https?:\/\//, '')}
                  <ExternalLink className="w-3 h-3" />
                </a>
              )}
            </div>
          </div>
        </div>

        {/* Jobs at this company */}
        <div className="animate-fade-in">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-1.5 rounded-lg bg-brand-500/10">
              <Briefcase className="w-4 h-4 text-brand-400" />
            </div>
            <h2 className="text-xl font-bold text-white">
              Open Positions
              {jobs.length > 0 && (
                <span className="ml-2 text-sm font-normal text-surface-400">({jobs.length})</span>
              )}
            </h2>
          </div>

          {jobs.length === 0 ? (
            <div className="text-center py-16 glass-card">
              <Briefcase className="w-12 h-12 text-surface-500 mx-auto mb-3" />
              <h3 className="text-white font-semibold mb-2">No open positions right now</h3>
              <p className="text-surface-400 text-sm mb-4">
                Check back later or browse all remote jobs
              </p>
              <Link href="/jobs" className="btn-primary text-sm">
                Browse All Jobs
              </Link>
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2">
              {jobs.map((job) => (
                <JobCard key={job.id} job={job} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
