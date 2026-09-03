import type { Metadata } from 'next';
import Link from 'next/link';
import { ArrowRight, Zap, Search, TrendingUp, Shield, Globe, Star } from 'lucide-react';
import { jobs, categories } from '@/lib/api';
import { JobCard } from '@/components/jobs/JobCard';
import type { Job, Category } from '@/lib/types';

export const metadata: Metadata = {
  title: 'RemoteScope — Find Your Remote Dream Job',
};

export const revalidate = 300; // 5 min ISR

async function getHomeData(): Promise<{ latestJobs: Job[]; trendingJobs: Job[]; allCategories: Category[] }> {
  try {
    const [latestRes, trendingRes, catsRes] = await Promise.all([
      jobs.latest(),
      jobs.trending(),
      categories.list(),
    ]);
    return {
      latestJobs:    (latestRes.data ?? []).slice(0, 8),
      trendingJobs:  (trendingRes.data ?? []).slice(0, 6),
      allCategories: catsRes.data ?? [],
    };
  } catch {
    return { latestJobs: [], trendingJobs: [], allCategories: [] };
  }
}

export default async function HomePage() {
  const { latestJobs, trendingJobs, allCategories } = await getHomeData();

  return (
    <div className="bg-glow-top">
      {/* ── Hero ─────────────────────────────────────────── */}
      <section className="relative pt-32 pb-20 px-4 sm:px-6 lg:px-8 overflow-hidden">
        {/* Background orbs */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute top-20 left-1/4 w-96 h-96 rounded-full bg-brand-500/8 blur-3xl" />
          <div className="absolute top-40 right-1/4 w-64 h-64 rounded-full bg-accent-violet/6 blur-3xl" />
          <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-full h-px bg-gradient-to-r from-transparent via-white/10 to-transparent" />
        </div>

        <div className="relative max-w-4xl mx-auto text-center">
          {/* Badge */}
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass-card border-brand-500/20 text-sm mb-8 animate-fade-in">
            <span className="dot-live" />
            <span className="text-surface-300">
              <span className="text-white font-semibold">12,400+</span> remote jobs updated daily
            </span>
          </div>

          {/* Heading */}
          <h1 className="text-5xl sm:text-6xl lg:text-7xl font-bold tracking-tight mb-6 animate-slide-up leading-[1.1]">
            Find your{' '}
            <span className="gradient-text">perfect remote</span>
            <br />job in seconds
          </h1>

          <p className="text-lg sm:text-xl text-surface-300 max-w-2xl mx-auto mb-10 animate-fade-in leading-relaxed">
            AI-powered job aggregation from 50+ sources. Semantic search, instant filters,
            and smart recommendations — built for remote-first careers.
          </p>

          {/* CTA */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 animate-slide-up">
            <Link href="/jobs" className="btn-primary text-base px-8 py-3 shadow-glow-lg hover:shadow-glow-lg animate-pulse-glow">
              Browse Jobs
              <ArrowRight className="w-5 h-5" />
            </Link>
            <Link href="/search" className="btn-ghost text-base px-8 py-3">
              <Search className="w-5 h-5" />
              Semantic Search
            </Link>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-3 gap-6 max-w-lg mx-auto mt-16 animate-fade-in">
            {[
              { value: '50+',   label: 'Job Sources' },
              { value: '200K+', label: 'Jobs Indexed' },
              { value: '10K+',  label: 'Companies' },
            ].map(({ value, label }) => (
              <div key={label} className="text-center">
                <p className="text-2xl font-bold text-white mb-1">{value}</p>
                <p className="text-xs text-surface-400">{label}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Categories ───────────────────────────────────── */}
      {allCategories.length > 0 && (
        <section className="py-12 px-4 sm:px-6 lg:px-8">
          <div className="max-w-7xl mx-auto">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-bold text-white">Browse by Category</h2>
              <Link href="/jobs" className="text-sm text-brand-400 hover:text-brand-300 transition-colors flex items-center gap-1">
                View all <ArrowRight className="w-4 h-4" />
              </Link>
            </div>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
              {allCategories.slice(0, 12).map((cat) => (
                <Link
                  key={cat.slug}
                  href={`/jobs?category=${cat.slug}`}
                  className="glass-card p-3 text-center hover:border-white/[0.12] hover:shadow-card transition-all group"
                  style={{ '--cat-color': cat.color } as React.CSSProperties}
                >
                  <div className="text-2xl mb-2 group-hover:scale-110 transition-transform">{cat.icon}</div>
                  <p className="text-xs font-medium text-surface-300 group-hover:text-white transition-colors leading-tight">
                    {cat.name}
                  </p>
                </Link>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* ── Trending Jobs ─────────────────────────────────── */}
      {trendingJobs.length > 0 && (
        <section className="py-12 px-4 sm:px-6 lg:px-8">
          <div className="max-w-7xl mx-auto">
            <div className="flex items-center gap-3 mb-6">
              <div className="p-1.5 rounded-lg bg-accent-amber/10 text-accent-amber">
                <TrendingUp className="w-4 h-4" />
              </div>
              <h2 className="text-xl font-bold text-white">Trending Now</h2>
            </div>
            <div className="grid gap-3">
              {trendingJobs.map((job) => (
                <JobCard key={job.id} job={job} variant="compact" />
              ))}
            </div>
          </div>
        </section>
      )}

      {/* ── Latest Jobs ───────────────────────────────────── */}
      {latestJobs.length > 0 && (
        <section className="py-12 px-4 sm:px-6 lg:px-8">
          <div className="max-w-7xl mx-auto">
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-3">
                <div className="p-1.5 rounded-lg bg-brand-500/10 text-brand-400">
                  <Zap className="w-4 h-4" />
                </div>
                <h2 className="text-xl font-bold text-white">Latest Jobs</h2>
              </div>
              <Link href="/jobs" className="text-sm text-brand-400 hover:text-brand-300 transition-colors flex items-center gap-1">
                View all <ArrowRight className="w-4 h-4" />
              </Link>
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              {latestJobs.map((job) => (
                <JobCard key={job.id} job={job} />
              ))}
            </div>
            <div className="text-center mt-8">
              <Link href="/jobs" className="btn-primary px-8 py-3 text-base">
                Browse All Jobs
                <ArrowRight className="w-5 h-5" />
              </Link>
            </div>
          </div>
        </section>
      )}

      {/* ── Features ─────────────────────────────────────── */}
      <section className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-14">
            <p className="section-label mb-3">Why RemoteScope</p>
            <h2 className="text-3xl sm:text-4xl font-bold text-white">
              Built for serious remote job seekers
            </h2>
          </div>
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {FEATURES.map(({ icon: Icon, title, desc, color }) => (
              <div key={title} className="glass-card p-6 hover:border-white/[0.12] transition-all group">
                <div
                  className="w-10 h-10 rounded-xl flex items-center justify-center mb-4 transition-transform group-hover:scale-110"
                  style={{ background: color + '18' }}
                >
                  <Icon className="w-5 h-5" style={{ color }} />
                </div>
                <h3 className="font-semibold text-white mb-2">{title}</h3>
                <p className="text-sm text-surface-300 leading-relaxed">{desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── CTA Banner ───────────────────────────────────── */}
      <section className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-4xl mx-auto">
          <div className="gradient-border p-px rounded-2xl">
            <div className="glass-card rounded-2xl p-10 text-center">
              <Star className="w-8 h-8 text-accent-amber mx-auto mb-4 animate-float" />
              <h2 className="text-3xl sm:text-4xl font-bold text-white mb-4">
                Land your remote job today
              </h2>
              <p className="text-surface-300 mb-8 max-w-md mx-auto">
                Join thousands of engineers, designers, and product managers who found
                their dream remote role through RemoteScope.
              </p>
              <Link href="/auth/signup" className="btn-primary text-base px-8 py-3 shadow-glow-lg">
                Create Free Account
                <ArrowRight className="w-5 h-5" />
              </Link>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}

const FEATURES = [
  {
    icon: Search,
    title: 'AI Semantic Search',
    desc: 'Describe what you want in plain English. Our vector search understands context, not just keywords.',
    color: '#4355f5',
  },
  {
    icon: TrendingUp,
    title: 'Real-Time Aggregation',
    desc: 'Scraped from 50+ job boards every 6 hours — Remote OK, Remotive, LinkedIn, Greenhouse, and more.',
    color: '#10b981',
  },
  {
    icon: Shield,
    title: 'Fake Job Detection',
    desc: 'Our AI pipeline flags suspicious listings before they reach you. Browse with confidence.',
    color: '#f59e0b',
  },
  {
    icon: Globe,
    title: 'Timezone Filtering',
    desc: 'Find jobs that actually work for your schedule. Filter by timezone overlap and working hours.',
    color: '#00d4ff',
  },
  {
    icon: Zap,
    title: 'Smart Recommendations',
    desc: 'The more you use RemoteScope, the smarter it gets. Personalized picks based on your profile.',
    color: '#a855f7',
  },
  {
    icon: Star,
    title: 'Application Tracking',
    desc: 'Save jobs, track applications from bookmark to offer — all in one clean dashboard.',
    color: '#f43f5e',
  },
];
