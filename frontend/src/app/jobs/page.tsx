'use client';

import { useState, useEffect, useCallback, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { Search, Loader2, SlidersHorizontal, ChevronLeft, ChevronRight } from 'lucide-react';
import { FilterPanel } from '@/components/jobs/FilterPanel';
import { JobCard } from '@/components/jobs/JobCard';
import { jobs as jobsApi, categories } from '@/lib/api';
import { debounce, buildQueryString } from '@/lib/utils';
import type { Job, JobFilters, Category } from '@/lib/types';

function JobsSkeleton() {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      {Array.from({ length: 8 }).map((_, i) => (
        <div key={i} className="glass-card p-5 animate-pulse">
          <div className="flex gap-3 mb-4">
            <div className="w-12 h-12 rounded-xl skeleton flex-shrink-0" />
            <div className="flex-1 space-y-2">
              <div className="h-4 skeleton rounded w-3/4" />
              <div className="h-3 skeleton rounded w-1/2" />
            </div>
          </div>
          <div className="flex gap-2 mb-3">
            <div className="h-6 w-20 skeleton rounded-lg" />
            <div className="h-6 w-16 skeleton rounded-lg" />
            <div className="h-6 w-24 skeleton rounded-lg" />
          </div>
          <div className="flex gap-1.5 mb-4">
            {[60, 80, 50, 70].map((w) => (
              <div key={w} className={`h-5 skeleton rounded-lg`} style={{ width: w }} />
            ))}
          </div>
          <div className="h-px bg-white/[0.04] mt-3 mb-3" />
          <div className="flex justify-between">
            <div className="h-4 w-24 skeleton rounded" />
            <div className="h-7 w-16 skeleton rounded-lg" />
          </div>
        </div>
      ))}
    </div>
  );
}

function Pagination({ page, totalPages, onChange }: { page: number; totalPages: number; onChange: (p: number) => void }) {
  if (totalPages <= 1) return null;
  const pages = Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
    if (totalPages <= 5) return i + 1;
    if (page <= 3) return i + 1;
    if (page >= totalPages - 2) return totalPages - 4 + i;
    return page - 2 + i;
  });
  return (
    <div className="flex items-center justify-center gap-2 mt-8">
      <button
        onClick={() => onChange(page - 1)}
        disabled={page === 1}
        className="p-2 rounded-lg text-surface-400 hover:text-white hover:bg-white/[0.06] disabled:opacity-30 disabled:cursor-not-allowed transition-all"
        aria-label="Previous page"
      >
        <ChevronLeft className="w-4 h-4" />
      </button>
      {pages.map((p) => (
        <button
          key={p}
          onClick={() => onChange(p)}
          className={`w-9 h-9 rounded-lg text-sm font-medium transition-all ${
            p === page
              ? 'bg-brand-500 text-white shadow-glow-sm'
              : 'text-surface-300 hover:text-white hover:bg-white/[0.06]'
          }`}
        >
          {p}
        </button>
      ))}
      <button
        onClick={() => onChange(page + 1)}
        disabled={page === totalPages}
        className="p-2 rounded-lg text-surface-400 hover:text-white hover:bg-white/[0.06] disabled:opacity-30 disabled:cursor-not-allowed transition-all"
        aria-label="Next page"
      >
        <ChevronRight className="w-4 h-4" />
      </button>
    </div>
  );
}

function JobsContent() {
  const searchParams = useSearchParams();
  const router = useRouter();

  const [allCategories, setAllCategories] = useState<Category[]>([]);
  const [jobsList, setJobsList] = useState<Job[]>([]);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);
  const [showFilters, setShowFilters] = useState(true);

  const [filters, setFilters] = useState<JobFilters>(() => ({
    q: searchParams.get('q') || undefined,
    category: searchParams.get('category') || undefined,
    remote_type: (searchParams.get('remote_type') as JobFilters['remote_type']) || undefined,
    page: Number(searchParams.get('page') || 1),
    per_page: 20,
  }));

  // Sync filters to URL
  const syncUrl = useCallback((f: JobFilters) => {
    const qs = buildQueryString(f);
    router.replace(`/jobs${qs ? '?' + qs : ''}`, { scroll: false });
  }, [router]);

  // Load categories once
  useEffect(() => {
    categories.list().then((res) => {
      if (res.success && res.data) setAllCategories(res.data);
    });
  }, []);

  // Fetch jobs when filters change
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const fetchJobs = useCallback(
    debounce(async (f: JobFilters) => {
      setLoading(true);
      try {
        const res = await jobsApi.list(f);
        if (res.success && res.data) {
          setJobsList(res.data.data ?? []);
          setTotal(res.data.total ?? 0);
          setTotalPages(res.data.total_pages ?? 1);
        }
      } finally {
        setLoading(false);
      }
    }, 300),
    []
  );

  useEffect(() => {
    fetchJobs(filters);
    syncUrl(filters);
  }, [filters, fetchJobs, syncUrl]);

  const handleFilterChange = (newFilters: JobFilters) => {
    setFilters({ ...newFilters, page: 1, per_page: 20 });
  };

  const handlePageChange = (p: number) => {
    setFilters((f) => ({ ...f, page: p }));
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleSearchChange = debounce((q: string) => {
    setFilters((f) => ({ ...f, q: q || undefined, page: 1 }));
  }, 400);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-2">Browse Remote Jobs</h1>
        <p className="text-surface-300">
          {loading ? 'Loading...' : `${total.toLocaleString()} remote jobs found`}
        </p>
      </div>

      {/* Search bar */}
      <div className="relative mb-6">
        <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-surface-400" />
        <input
          id="jobs-search"
          type="text"
          defaultValue={filters.q || ''}
          onChange={(e) => handleSearchChange(e.target.value)}
          placeholder="Search jobs, skills, companies…"
          className="input pl-12 pr-12 py-3.5 text-base"
        />
        <button
          onClick={() => setShowFilters(!showFilters)}
          className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm text-surface-300 hover:text-white hover:bg-white/[0.06] transition-all lg:hidden"
          aria-label="Toggle filters"
        >
          <SlidersHorizontal className="w-4 h-4" />
          Filters
        </button>
      </div>

      <div className="flex gap-6">
        {/* Sidebar */}
        <aside className={`w-64 flex-shrink-0 ${showFilters ? 'block' : 'hidden'} lg:block`}>
          <div className="sticky top-24">
            <FilterPanel
              filters={filters}
              onChange={handleFilterChange}
              categories={allCategories}
            />
          </div>
        </aside>

        {/* Main content */}
        <div className="flex-1 min-w-0">
          {loading ? (
            <JobsSkeleton />
          ) : jobsList.length === 0 ? (
            <div className="text-center py-20 glass-card">
              <div className="text-5xl mb-4">🔍</div>
              <h3 className="text-xl font-semibold text-white mb-2">No jobs found</h3>
              <p className="text-surface-300 mb-6">Try adjusting your filters or search terms</p>
              <button
                onClick={() => setFilters({ page: 1, per_page: 20 })}
                className="btn-ghost"
              >
                Clear all filters
              </button>
            </div>
          ) : (
            <>
              <div className="grid gap-4 md:grid-cols-2">
                {jobsList.map((job) => (
                  <JobCard key={job.id} job={job} />
                ))}
              </div>
              <Pagination
                page={filters.page ?? 1}
                totalPages={totalPages}
                onChange={handlePageChange}
              />
            </>
          )}
        </div>
      </div>
    </div>
  );
}

export default function JobsPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center min-h-[60vh]">
        <Loader2 className="w-8 h-8 animate-spin text-brand-400" />
      </div>
    }>
      <JobsContent />
    </Suspense>
  );
}
