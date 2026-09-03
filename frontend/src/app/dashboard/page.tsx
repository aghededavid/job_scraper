'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  Bookmark, CheckCircle2, Clock, Briefcase, TrendingUp,
  ArrowRight, Bell, Search, Star, RefreshCw
} from 'lucide-react';
import { JobCard } from '@/components/jobs/JobCard';
import { useAuth } from '@/contexts/AuthContext';
import { savedJobs as savedJobsApi, jobs as jobsApi, notifications } from '@/lib/api';
import { timeAgo, formatDate } from '@/lib/utils';
import { cn } from '@/lib/utils';
import type { SavedJob, Job, Notification } from '@/lib/types';

type Tab = 'saved' | 'applied' | 'recommendations';

const APPLICATION_STATUS_STYLES: Record<string, { label: string; color: string; bg: string }> = {
  bookmarked:    { label: 'Bookmarked',    color: 'text-surface-300',    bg: 'bg-surface-600' },
  applied:       { label: 'Applied',       color: 'text-brand-300',      bg: 'bg-brand-950/60' },
  phone_screen:  { label: 'Phone Screen',  color: 'text-accent-cyan',    bg: 'bg-accent-cyan/10' },
  interview:     { label: 'Interview',     color: 'text-accent-violet',   bg: 'bg-accent-violet/10' },
  offer:         { label: 'Offer',         color: 'text-accent-emerald',  bg: 'bg-accent-emerald/10' },
  rejected:      { label: 'Rejected',      color: 'text-accent-rose',     bg: 'bg-accent-rose/10' },
  accepted:      { label: 'Accepted ✓',    color: 'text-accent-emerald',  bg: 'bg-accent-emerald/20' },
  withdrawn:     { label: 'Withdrawn',     color: 'text-surface-400',     bg: 'bg-surface-700' },
};

function StatusBadge({ status }: { status: string }) {
  const s = APPLICATION_STATUS_STYLES[status] || APPLICATION_STATUS_STYLES.bookmarked;
  return (
    <span className={cn('px-2 py-0.5 rounded-md text-xs font-medium', s.color, s.bg)}>
      {s.label}
    </span>
  );
}

function SavedJobRow({ saved }: { saved: SavedJob }) {
  const job = saved.job;
  if (!job) return null;
  return (
    <div className="flex items-center gap-4 p-4 glass-card hover:border-white/[0.12] transition-all group">
      <div className="w-10 h-10 rounded-lg bg-surface-600 flex items-center justify-center flex-shrink-0 overflow-hidden">
        {job.company_logo ? (
          <img src={job.company_logo} alt={job.company_name} className="w-full h-full object-contain p-1" />
        ) : (
          <span className="text-base font-bold text-surface-300">{job.company_name[0]}</span>
        )}
      </div>
      <div className="flex-1 min-w-0">
        <Link href={`/jobs/${job.id}`} className="text-sm font-semibold text-white hover:text-brand-300 transition-colors truncate block">
          {job.title}
        </Link>
        <p className="text-xs text-surface-400 truncate">{job.company_name}</p>
      </div>
      <div className="flex items-center gap-3 flex-shrink-0">
        <StatusBadge status={saved.status} />
        <span className="text-xs text-surface-500 hidden sm:block">{timeAgo(saved.created_at)}</span>
        <Link href={`/jobs/${job.id}`} className="text-surface-400 hover:text-white transition-colors opacity-0 group-hover:opacity-100">
          <ArrowRight className="w-4 h-4" />
        </Link>
      </div>
    </div>
  );
}

export default function DashboardPage() {
  const { user, isLoading, isAuthenticated } = useAuth();
  const router = useRouter();

  const [activeTab, setActiveTab] = useState<Tab>('saved');
  const [savedJobs, setSavedJobs] = useState<SavedJob[]>([]);
  const [appliedJobs, setAppliedJobs] = useState<SavedJob[]>([]);
  const [recommendedJobs, setRecommendedJobs] = useState<Job[]>([]);
  const [unreadNotifs, setUnreadNotifs] = useState<Notification[]>([]);
  const [dataLoaded, setDataLoaded] = useState(false);

  // Redirect if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.replace('/auth/login');
    }
  }, [isLoading, isAuthenticated, router]);

  // Load dashboard data
  useEffect(() => {
    if (!isAuthenticated) return;
    Promise.all([
      savedJobsApi.list(),
      savedJobsApi.applied(),
      jobsApi.recommendations(),
      notifications.list(),
    ]).then(([savedRes, appliedRes, recRes, notifRes]) => {
      if (savedRes.success && savedRes.data) setSavedJobs(savedRes.data.items ?? []);
      if (appliedRes.success && appliedRes.data) setAppliedJobs(appliedRes.data.items ?? []);
      if (recRes.success && recRes.data) setRecommendedJobs((recRes.data as unknown as Job[]).slice(0, 6));
      if (notifRes.success && notifRes.data) {
        setUnreadNotifs((notifRes.data.items ?? []).filter((n) => !n.read).slice(0, 5));
      }
    }).finally(() => setDataLoaded(true));
  }, [isAuthenticated]);

  if (isLoading || !isAuthenticated) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <RefreshCw className="w-6 h-6 animate-spin text-brand-400" />
      </div>
    );
  }

  const stats = [
    { label: 'Saved Jobs', value: savedJobs.length, icon: Bookmark, color: '#4355f5' },
    { label: 'Applied', value: appliedJobs.length, icon: CheckCircle2, color: '#10b981' },
    { label: 'In Progress', value: appliedJobs.filter((j) => ['phone_screen', 'interview', 'offer'].includes(j.status)).length, icon: TrendingUp, color: '#f59e0b' },
  ];

  const tabs: { id: Tab; label: string; icon: typeof Bookmark; count?: number }[] = [
    { id: 'saved', label: 'Saved Jobs', icon: Bookmark, count: savedJobs.length },
    { id: 'applied', label: 'Applied', icon: Briefcase, count: appliedJobs.length },
    { id: 'recommendations', label: 'For You', icon: Star },
  ];

  return (
    <div className="min-h-screen bg-glow-top">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold text-white">
              Welcome back, {user?.name?.split(' ')[0]} 👋
            </h1>
            <p className="text-surface-400 mt-1">Your job search at a glance</p>
          </div>
          <div className="flex gap-3">
            <Link href="/search" className="btn-ghost text-sm">
              <Search className="w-4 h-4" />
              Search
            </Link>
            <Link href="/jobs" className="btn-primary text-sm">
              Browse Jobs
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 mb-8">
          {stats.map(({ label, value, icon: Icon, color }) => (
            <div key={label} className="glass-card p-5">
              <div className="flex items-center justify-between mb-3">
                <p className="text-sm text-surface-400">{label}</p>
                <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: color + '20' }}>
                  <Icon className="w-4 h-4" style={{ color }} />
                </div>
              </div>
              <p className="text-2xl font-bold text-white">{dataLoaded ? value : '—'}</p>
            </div>
          ))}
        </div>

        <div className="grid lg:grid-cols-3 gap-6">
          {/* Main content */}
          <div className="lg:col-span-2">
            {/* Tabs */}
            <div className="flex gap-1 p-1 bg-surface-700/40 rounded-xl border border-white/[0.06] mb-4">
              {tabs.map(({ id, label, icon: Icon, count }) => (
                <button
                  key={id}
                  onClick={() => setActiveTab(id)}
                  className={cn(
                    'flex-1 flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all',
                    activeTab === id
                      ? 'bg-brand-500 text-white shadow-glow-sm'
                      : 'text-surface-300 hover:text-white hover:bg-white/[0.05]'
                  )}
                >
                  <Icon className="w-4 h-4" />
                  {label}
                  {count !== undefined && count > 0 && (
                    <span className={cn(
                      'text-xs px-1.5 py-0.5 rounded-md font-semibold',
                      activeTab === id ? 'bg-white/20' : 'bg-white/10'
                    )}>
                      {count}
                    </span>
                  )}
                </button>
              ))}
            </div>

            {/* Tab content */}
            <div className="space-y-3">
              {activeTab === 'saved' && (
                <>
                  {!dataLoaded ? (
                    Array.from({ length: 4 }).map((_, i) => (
                      <div key={i} className="glass-card p-4 animate-pulse">
                        <div className="flex gap-3">
                          <div className="w-10 h-10 skeleton rounded-lg flex-shrink-0" />
                          <div className="flex-1 space-y-2">
                            <div className="h-4 skeleton rounded w-3/4" />
                            <div className="h-3 skeleton rounded w-1/2" />
                          </div>
                        </div>
                      </div>
                    ))
                  ) : savedJobs.length === 0 ? (
                    <div className="text-center py-12 glass-card">
                      <Bookmark className="w-12 h-12 text-surface-500 mx-auto mb-3" />
                      <h3 className="text-white font-semibold mb-1">No saved jobs yet</h3>
                      <p className="text-surface-400 text-sm mb-4">Start bookmarking jobs to track them here</p>
                      <Link href="/jobs" className="btn-primary text-sm">Browse Jobs</Link>
                    </div>
                  ) : (
                    savedJobs.map((saved) => <SavedJobRow key={saved.id} saved={saved} />)
                  )}
                </>
              )}

              {activeTab === 'applied' && (
                <>
                  {!dataLoaded ? (
                    Array.from({ length: 3 }).map((_, i) => (
                      <div key={i} className="glass-card p-4 animate-pulse">
                        <div className="h-4 skeleton rounded w-3/4 mb-2" />
                        <div className="h-3 skeleton rounded w-1/2" />
                      </div>
                    ))
                  ) : appliedJobs.length === 0 ? (
                    <div className="text-center py-12 glass-card">
                      <Briefcase className="w-12 h-12 text-surface-500 mx-auto mb-3" />
                      <h3 className="text-white font-semibold mb-1">No applications yet</h3>
                      <p className="text-surface-400 text-sm mb-4">Track your applications and their status here</p>
                      <Link href="/jobs" className="btn-primary text-sm">Find Jobs</Link>
                    </div>
                  ) : (
                    appliedJobs.map((saved) => <SavedJobRow key={saved.id} saved={saved} />)
                  )}
                </>
              )}

              {activeTab === 'recommendations' && (
                <>
                  {!dataLoaded ? (
                    <div className="grid gap-4 md:grid-cols-2">
                      {Array.from({ length: 4 }).map((_, i) => (
                        <div key={i} className="glass-card p-5 animate-pulse h-40">
                          <div className="h-4 skeleton rounded w-3/4 mb-2" />
                          <div className="h-3 skeleton rounded w-1/2" />
                        </div>
                      ))}
                    </div>
                  ) : recommendedJobs.length === 0 ? (
                    <div className="text-center py-12 glass-card">
                      <Star className="w-12 h-12 text-surface-500 mx-auto mb-3" />
                      <h3 className="text-white font-semibold mb-1">No recommendations yet</h3>
                      <p className="text-surface-400 text-sm mb-4">
                        Complete your <Link href="/profile" className="text-brand-400 hover:text-brand-300">profile preferences</Link> to get personalized picks
                      </p>
                    </div>
                  ) : (
                    <div className="grid gap-4 md:grid-cols-2">
                      {recommendedJobs.map((job) => <JobCard key={job.id} job={job} />)}
                    </div>
                  )}
                </>
              )}
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-4">
            {/* Notifications */}
            <div className="glass-card p-5">
              <div className="flex items-center justify-between mb-4">
                <h2 className="font-semibold text-white flex items-center gap-2">
                  <Bell className="w-4 h-4 text-brand-400" />
                  Notifications
                </h2>
                <Link href="#" className="text-xs text-brand-400 hover:text-brand-300 transition-colors">
                  View all
                </Link>
              </div>
              {unreadNotifs.length === 0 ? (
                <p className="text-surface-400 text-sm text-center py-4">All caught up! 🎉</p>
              ) : (
                <div className="space-y-3">
                  {unreadNotifs.map((n) => (
                    <div key={n.id} className="p-3 rounded-xl bg-brand-950/40 border border-brand-500/20">
                      <p className="text-sm font-medium text-white mb-0.5">{n.title}</p>
                      <p className="text-xs text-surface-400">{n.body}</p>
                      <p className="text-xs text-surface-500 mt-1">{timeAgo(n.created_at)}</p>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Quick links */}
            <div className="glass-card p-5">
              <h2 className="font-semibold text-white mb-4 flex items-center gap-2">
                <Clock className="w-4 h-4 text-surface-400" />
                Quick Actions
              </h2>
              <div className="space-y-2">
                {[
                  { label: 'Update preferences', href: '/profile', icon: '⚙️' },
                  { label: 'Semantic search', href: '/search', icon: '🔍' },
                  { label: 'Browse by category', href: '/jobs', icon: '📂' },
                ].map(({ label, href, icon }) => (
                  <Link
                    key={href}
                    href={href}
                    className="flex items-center gap-3 p-2.5 rounded-lg text-sm text-surface-300 hover:text-white hover:bg-white/[0.05] transition-all group"
                  >
                    <span className="text-base">{icon}</span>
                    {label}
                    <ArrowRight className="w-3.5 h-3.5 ml-auto opacity-0 group-hover:opacity-100 transition-opacity" />
                  </Link>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
