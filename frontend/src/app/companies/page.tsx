import type { Metadata } from 'next';
import Link from 'next/link';
import { Search, Building2, ArrowRight, Globe, Users, CheckCircle2 } from 'lucide-react';
import { companies as companiesApi } from '@/lib/api';
import type { Company } from '@/lib/types';

export const metadata: Metadata = {
  title: 'Remote Companies',
  description: 'Browse top remote-friendly companies hiring on RemoteScope.',
};

export const revalidate = 600;

async function getCompanies(): Promise<Company[]> {
  try {
    const res = await companiesApi.list();
    return res.data?.data ?? [];
  } catch {
    return [];
  }
}

function CompanyCard({ company }: { company: Company }) {
  return (
    <Link
      href={`/companies/${company.slug}`}
      className="glass-card p-5 flex items-start gap-4 hover:border-white/[0.12] hover:shadow-card transition-all group"
    >
      {/* Logo */}
      <div className="w-12 h-12 rounded-xl bg-surface-600 flex items-center justify-center flex-shrink-0 border border-white/[0.06] overflow-hidden">
        {company.logo_url ? (
          <img src={company.logo_url} alt={company.name} className="w-full h-full object-contain p-1" />
        ) : (
          <span className="text-xl font-bold text-surface-300">{company.name[0]}</span>
        )}
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5 mb-1">
          <h2 className="font-semibold text-white group-hover:text-brand-300 transition-colors truncate">
            {company.name}
          </h2>
          {company.is_verified && (
            <CheckCircle2 className="w-3.5 h-3.5 text-accent-cyan flex-shrink-0" />
          )}
        </div>

        {company.description && (
          <p className="text-sm text-surface-400 line-clamp-2 mb-3 leading-relaxed">
            {company.description}
          </p>
        )}

        <div className="flex flex-wrap items-center gap-3 text-xs text-surface-500">
          {company.is_remote_friendly && (
            <span className="flex items-center gap-1 text-accent-emerald">
              <Globe className="w-3 h-3" />
              Remote-friendly
            </span>
          )}
          {company.employee_count && (
            <span className="flex items-center gap-1">
              <Users className="w-3 h-3" />
              {company.employee_count.toLocaleString()}+ employees
            </span>
          )}
          {company.website && (
            <span className="flex items-center gap-1">
              <Globe className="w-3 h-3" />
              {company.website.replace(/^https?:\/\//, '').split('/')[0]}
            </span>
          )}
        </div>
      </div>

      <ArrowRight className="w-4 h-4 text-surface-500 group-hover:text-brand-400 flex-shrink-0 mt-1 transition-colors" />
    </Link>
  );
}

export default async function CompaniesPage() {
  const companiesList = await getCompanies();

  return (
    <div className="bg-glow-top min-h-screen">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Header */}
        <div className="text-center mb-12">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full glass-card text-sm text-surface-300 mb-4">
            <Building2 className="w-4 h-4 text-brand-400" />
            <span>{companiesList.length}+ companies hiring</span>
          </div>
          <h1 className="text-4xl font-bold text-white mb-3">
            Remote-First Companies
          </h1>
          <p className="text-surface-300 max-w-xl mx-auto">
            Discover top companies with strong remote cultures — from startups to Fortune 500s.
          </p>
        </div>

        {/* Search */}
        <div className="relative mb-8 max-w-lg mx-auto">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-400" />
          <input
            id="companies-search"
            type="text"
            placeholder="Search companies…"
            className="input pl-11"
          />
        </div>

        {/* Grid */}
        {companiesList.length === 0 ? (
          <div className="text-center py-20 glass-card">
            <Building2 className="w-12 h-12 text-surface-500 mx-auto mb-3" />
            <h3 className="text-white font-semibold mb-1">No companies found</h3>
            <p className="text-surface-400 text-sm">Check back soon as we add more companies</p>
          </div>
        ) : (
          <div className="grid gap-4 md:grid-cols-2">
            {companiesList.map((company) => (
              <CompanyCard key={company.id} company={company} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
