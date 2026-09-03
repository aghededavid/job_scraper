'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  BarChart3, Users, Briefcase, Building2, Zap, Clock,
  Play, RefreshCw, Shield, Trash2, Flag, ChevronRight, AlertTriangle
} from 'lucide-react';
import { useAuth } from '@/contexts/AuthContext';
import { admin as adminApi } from '@/lib/api';
import { timeAgo, formatDate } from '@/lib/utils';
import { cn } from '@/lib/utils';
import type { AdminStats, ScraperSource, User } from '@/lib/types';

type AdminTab = 'overview' | 'scrapers' | 'users' | 'jobs';

function StatCard({ label, value, icon: Icon, color, sub }: {
  label: string; value: number | string; icon: typeof BarChart3; color: string; sub?: string;
}) {
  return (
    <div className="glass-card p-5">
      <div className="flex items-center justify-between mb-3">
        <p className="text-sm text-surface-400">{label}</p>
        <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: color + '20' }}>
          <Icon className="w-4 h-4" style={{ color }} />
        </div>
      </div>
      <p className="text-2xl font-bold text-white">{typeof value === 'number' ? value.toLocaleString() : value}</p>
      {sub && <p className="text-xs text-surface-500 mt-1">{sub}</p>}
    </div>
  );
}

function ScraperRow({ scraper, onRun, running }: {
  scraper: ScraperSource;
  onRun: (id: string) => void;
  running: boolean;
}) {
  const statusColor = scraper.last_run_status === 'success'
    ? 'text-accent-emerald'
    : scraper.last_run_status === 'error'
      ? 'text-accent-rose'
      : 'text-surface-400';

  return (
    <div className="flex items-center gap-4 p-4 rounded-xl border border-white/[0.06] bg-surface-700/30 hover:bg-surface-700/50 transition-all">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-0.5">
          <p className="text-sm font-semibold text-white">{scraper.name}</p>
          <span className={cn('text-xs font-medium', statusColor)}>
            {scraper.last_run_status || 'never run'}
          </span>
        </div>
        <div className="flex items-center gap-3 text-xs text-surface-500">
          <span>{scraper.fetch_method.toUpperCase()}</span>
          <span>{scraper.schedule_cron}</span>
          {scraper.last_run_at && <span>Last run: {timeAgo(scraper.last_run_at)}</span>}
          <span className="text-accent-emerald">{scraper.jobs_scraped_total.toLocaleString()} jobs</span>
        </div>
      </div>
      <button
        onClick={() => onRun(scraper.id)}
        disabled={running || !scraper.is_active}
        className={cn(
          'flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-all',
          running
            ? 'bg-brand-500/20 text-brand-300 cursor-wait'
            : scraper.is_active
              ? 'bg-brand-500/20 text-brand-300 hover:bg-brand-500/30 border border-brand-500/30'
              : 'bg-surface-600 text-surface-500 cursor-not-allowed'
        )}
      >
        {running ? <RefreshCw className="w-3.5 h-3.5 animate-spin" /> : <Play className="w-3.5 h-3.5" />}
        {running ? 'Running…' : 'Run Now'}
      </button>
    </div>
  );
}

function UserRow({ u, onRoleChange }: { u: User; onRoleChange: (id: string, role: string) => void }) {
  return (
    <div className="flex items-center gap-4 p-3 rounded-xl border border-white/[0.05] hover:bg-white/[0.02] transition-all">
      <div className="w-8 h-8 rounded-full bg-brand-500/20 flex items-center justify-center text-sm font-bold text-brand-300 flex-shrink-0">
        {u.name?.[0] || u.email[0]}
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-white truncate">{u.name || 'Unnamed'}</p>
        <p className="text-xs text-surface-400 truncate">{u.email}</p>
      </div>
      <div className="flex items-center gap-2 flex-shrink-0">
        <span className={cn(
          'text-xs px-2 py-0.5 rounded-md font-medium',
          u.role === 'admin' ? 'bg-accent-amber/20 text-accent-amber' :
          u.role === 'premium' ? 'bg-brand-500/20 text-brand-300' :
          'bg-surface-600 text-surface-400'
        )}>
          {u.role}
        </span>
        <p className="text-xs text-surface-500 hidden sm:block">{formatDate(u.created_at)}</p>
        <select
          value={u.role}
          onChange={(e) => onRoleChange(u.id, e.target.value)}
          className="text-xs bg-surface-700 border border-white/10 text-white rounded-lg px-2 py-1 cursor-pointer hover:border-white/20 transition-colors"
          title="Change role"
        >
          <option value="user">User</option>
          <option value="premium">Premium</option>
          <option value="admin">Admin</option>
        </select>
      </div>
    </div>
  );
}

export default function AdminPage() {
  const { isLoading, isAuthenticated, isAdmin } = useAuth();
  const router = useRouter();

  const [activeTab, setActiveTab] = useState<AdminTab>('overview');
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [scrapers, setScrapers] = useState<ScraperSource[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [runningScrapers, setRunningScrapers] = useState<Set<string>>(new Set());
  const [dataLoaded, setDataLoaded] = useState(false);

  useEffect(() => {
    if (!isLoading && (!isAuthenticated || !isAdmin)) {
      router.replace('/');
    }
  }, [isLoading, isAuthenticated, isAdmin, router]);

  useEffect(() => {
    if (!isAdmin) return;
    Promise.all([
      adminApi.stats(),
      adminApi.scrapers(),
      adminApi.users(),
    ]).then(([statsRes, scrapersRes, usersRes]) => {
      if (statsRes.success && statsRes.data) setStats(statsRes.data);
      if (scrapersRes.success && scrapersRes.data) setScrapers(scrapersRes.data);
      if (usersRes.success && usersRes.data) setUsers(usersRes.data.data ?? []);
    }).finally(() => setDataLoaded(true));
  }, [isAdmin]);

  const handleRunScraper = async (id: string) => {
    setRunningScrapers((s) => new Set([...s, id]));
    try {
      await adminApi.triggerScraper(id);
    } finally {
      setTimeout(() => {
        setRunningScrapers((s) => { const n = new Set(s); n.delete(id); return n; });
      }, 3000);
    }
  };

  const handleRoleChange = async (id: string, role: string) => {
    await adminApi.updateUserRole(id, role);
    setUsers((u) => u.map((usr) => usr.id === id ? { ...usr, role: role as User['role'] } : usr));
  };

  if (isLoading || !isAuthenticated || !isAdmin) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <RefreshCw className="w-6 h-6 animate-spin text-brand-400" />
      </div>
    );
  }

  const tabs: { id: AdminTab; label: string; icon: typeof BarChart3 }[] = [
    { id: 'overview', label: 'Overview', icon: BarChart3 },
    { id: 'scrapers', label: 'Scrapers', icon: Zap },
    { id: 'users', label: 'Users', icon: Users },
  ];

  return (
    <div className="min-h-screen bg-glow-top">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center gap-3 mb-8">
          <div className="w-10 h-10 rounded-xl bg-accent-amber/20 flex items-center justify-center">
            <Shield className="w-5 h-5 text-accent-amber" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-white">Admin Panel</h1>
            <p className="text-surface-400 text-sm">System management &amp; monitoring</p>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-1 p-1 bg-surface-700/40 rounded-xl border border-white/[0.06] mb-6 w-fit">
          {tabs.map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              onClick={() => setActiveTab(id)}
              className={cn(
                'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all',
                activeTab === id
                  ? 'bg-brand-500 text-white shadow-glow-sm'
                  : 'text-surface-300 hover:text-white hover:bg-white/[0.05]'
              )}
            >
              <Icon className="w-4 h-4" />
              {label}
            </button>
          ))}
        </div>

        {/* Overview */}
        {activeTab === 'overview' && (
          <div className="animate-fade-in space-y-6">
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <StatCard label="Total Users" value={stats?.total_users ?? '—'} icon={Users} color="#4355f5" />
              <StatCard label="Active Jobs" value={stats?.active_jobs ?? '—'} icon={Briefcase} color="#10b981" sub={`of ${stats?.total_jobs ?? 0} total`} />
              <StatCard label="Companies" value={stats?.total_companies ?? '—'} icon={Building2} color="#f59e0b" />
              <StatCard label="Jobs Today" value={stats?.jobs_today ?? '—'} icon={Zap} color="#a855f7" sub={`${stats?.jobs_this_hour ?? 0} this hour`} />
            </div>

            <div className="glass-card p-6">
              <h2 className="font-semibold text-white mb-4 flex items-center gap-2">
                <Zap className="w-4 h-4 text-brand-400" />
                Scraper Sources
              </h2>
              <div className="grid grid-cols-2 gap-3 text-center">
                <div className="p-4 rounded-xl bg-accent-emerald/10 border border-accent-emerald/20">
                  <p className="text-2xl font-bold text-accent-emerald">{stats?.active_sources ?? '—'}</p>
                  <p className="text-xs text-surface-400 mt-1">Active Sources</p>
                </div>
                <div className="p-4 rounded-xl bg-surface-700/40 border border-white/[0.06]">
                  <p className="text-2xl font-bold text-white">{stats?.total_sources ?? '—'}</p>
                  <p className="text-xs text-surface-400 mt-1">Total Sources</p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Scrapers */}
        {activeTab === 'scrapers' && (
          <div className="animate-fade-in">
            <div className="flex items-center justify-between mb-4">
              <h2 className="font-semibold text-white">Scraper Sources ({scrapers.length})</h2>
              <div className="flex items-center gap-2 text-xs text-surface-400">
                <Clock className="w-3.5 h-3.5" />
                Schedules run automatically via cron
              </div>
            </div>
            {!dataLoaded ? (
              <div className="space-y-3">
                {Array.from({ length: 4 }).map((_, i) => (
                  <div key={i} className="h-16 skeleton rounded-xl" />
                ))}
              </div>
            ) : scrapers.length === 0 ? (
              <div className="text-center py-12 glass-card">
                <AlertTriangle className="w-10 h-10 text-surface-500 mx-auto mb-2" />
                <p className="text-surface-400">No scrapers configured. Run <code className="text-brand-400">make seed</code> first.</p>
              </div>
            ) : (
              <div className="space-y-3">
                {scrapers.map((scraper) => (
                  <ScraperRow
                    key={scraper.id}
                    scraper={scraper}
                    onRun={handleRunScraper}
                    running={runningScrapers.has(scraper.id)}
                  />
                ))}
              </div>
            )}
          </div>
        )}

        {/* Users */}
        {activeTab === 'users' && (
          <div className="animate-fade-in">
            <div className="flex items-center justify-between mb-4">
              <h2 className="font-semibold text-white">Users ({users.length})</h2>
            </div>
            {!dataLoaded ? (
              <div className="space-y-2">
                {Array.from({ length: 6 }).map((_, i) => (
                  <div key={i} className="h-14 skeleton rounded-xl" />
                ))}
              </div>
            ) : users.length === 0 ? (
              <div className="text-center py-12 glass-card">
                <Users className="w-10 h-10 text-surface-500 mx-auto mb-2" />
                <p className="text-surface-400">No users found</p>
              </div>
            ) : (
              <div className="glass-card p-4 space-y-2">
                {users.map((u) => (
                  <UserRow key={u.id} u={u} onRoleChange={handleRoleChange} />
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
