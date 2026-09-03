'use client';

import { useState, useRef, useCallback, useEffect } from 'react';
import { Search, Sparkles, Loader2, X, Clock, ArrowRight } from 'lucide-react';
import { JobCard } from '@/components/jobs/JobCard';
import { search as searchApi } from '@/lib/api';
import { cn, debounce } from '@/lib/utils';
import type { Job } from '@/lib/types';

const SAMPLE_QUERIES = [
  'Senior React developer with GraphQL experience',
  'Go backend engineer at a startup',
  'Remote DevOps with Kubernetes expertise',
  'ML engineer working on NLP models',
  'Product manager for a fintech company',
  'Full-stack engineer building data pipelines',
];

export default function SearchPage() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Job[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [recentSearches, setRecentSearches] = useState<string[]>([]);
  const inputRef = useRef<HTMLInputElement>(null);

  // Load recent searches from localStorage
  useEffect(() => {
    try {
      const stored = JSON.parse(localStorage.getItem('rs-recent-searches') || '[]');
      setRecentSearches(stored.slice(0, 5));
    } catch { /* ignore */ }
  }, []);

  const saveRecentSearch = (q: string) => {
    try {
      const stored: string[] = JSON.parse(localStorage.getItem('rs-recent-searches') || '[]');
      const updated = [q, ...stored.filter((s) => s !== q)].slice(0, 5);
      localStorage.setItem('rs-recent-searches', JSON.stringify(updated));
      setRecentSearches(updated);
    } catch { /* ignore */ }
  };

  const doSearch = useCallback(async (q: string) => {
    if (!q.trim()) return;
    setLoading(true);
    setHasSearched(true);
    setShowSuggestions(false);
    saveRecentSearch(q.trim());
    try {
      const res = await searchApi.query(q.trim());
      if (res.success && res.data) {
        setResults(res.data.results ?? []);
        setTotal(res.data.total ?? 0);
      } else {
        setResults([]);
        setTotal(0);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  // Suggestions
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const fetchSuggestions = useCallback(
    debounce(async (q: string) => {
      if (q.length < 2) { setSuggestions([]); return; }
      const res = await searchApi.suggest(q);
      if (res.success && res.data) setSuggestions(res.data.slice(0, 6));
    }, 300),
    []
  );

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setQuery(val);
    if (val.trim()) {
      fetchSuggestions(val);
      setShowSuggestions(true);
    } else {
      setShowSuggestions(false);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    doSearch(query);
  };

  const handleSuggestionClick = (s: string) => {
    setQuery(s);
    doSearch(s);
  };

  return (
    <div className="min-h-screen bg-glow-top">
      {/* Hero search bar */}
      <div className={cn(
        'transition-all duration-500',
        hasSearched ? 'py-8' : 'pt-24 pb-12'
      )}>
        <div className="max-w-3xl mx-auto px-4 sm:px-6">
          {!hasSearched && (
            <div className="text-center mb-8 animate-fade-in">
              <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass-card border-brand-500/20 text-sm mb-6">
                <Sparkles className="w-4 h-4 text-brand-400" />
                <span className="text-surface-300">AI-Powered Semantic Search</span>
              </div>
              <h1 className="text-4xl sm:text-5xl font-bold text-white mb-4">
                Find jobs with{' '}
                <span className="gradient-text">natural language</span>
              </h1>
              <p className="text-surface-300 text-lg">
                Describe what you&apos;re looking for in plain English — our vector search understands context.
              </p>
            </div>
          )}

          {/* Search form */}
          <div className="relative">
            <form onSubmit={handleSubmit}>
              <div className="relative">
                <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-surface-400 pointer-events-none" />
                <input
                  ref={inputRef}
                  id="semantic-search-input"
                  type="text"
                  value={query}
                  onChange={handleInputChange}
                  onFocus={() => query.length >= 2 && setShowSuggestions(true)}
                  onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
                  placeholder="e.g. Senior React engineer with TypeScript at a startup…"
                  className="input pl-12 pr-32 py-4 text-base"
                  autoFocus
                />
                {query && (
                  <button
                    type="button"
                    onClick={() => { setQuery(''); setSuggestions([]); inputRef.current?.focus(); }}
                    className="absolute right-24 top-1/2 -translate-y-1/2 text-surface-400 hover:text-white transition-colors"
                    aria-label="Clear"
                  >
                    <X className="w-4 h-4" />
                  </button>
                )}
                <button
                  id="search-submit"
                  type="submit"
                  disabled={!query.trim() || loading}
                  className="absolute right-2 top-1/2 -translate-y-1/2 btn-primary py-2 px-4 text-sm disabled:opacity-50"
                >
                  {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Search'}
                </button>
              </div>
            </form>

            {/* Suggestions dropdown */}
            {showSuggestions && suggestions.length > 0 && (
              <div className="absolute top-full left-0 right-0 mt-2 glass-card p-2 z-50">
                {suggestions.map((s) => (
                  <button
                    key={s}
                    onClick={() => handleSuggestionClick(s)}
                    className="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-surface-300 hover:text-white hover:bg-white/[0.05] transition-all text-left"
                  >
                    <Search className="w-3.5 h-3.5 flex-shrink-0" />
                    {s}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Recent + sample queries */}
          {!hasSearched && (
            <div className="mt-8 space-y-6 animate-fade-in">
              {recentSearches.length > 0 && (
                <div>
                  <p className="section-label mb-3 flex items-center gap-2">
                    <Clock className="w-3 h-3" /> Recent searches
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {recentSearches.map((s) => (
                      <button
                        key={s}
                        onClick={() => handleSuggestionClick(s)}
                        className="tag hover:border-brand-500/40 hover:text-brand-300 hover:bg-brand-950/40"
                      >
                        {s}
                      </button>
                    ))}
                  </div>
                </div>
              )}
              <div>
                <p className="section-label mb-3 flex items-center gap-2">
                  <Sparkles className="w-3 h-3" /> Try these searches
                </p>
                <div className="flex flex-wrap gap-2">
                  {SAMPLE_QUERIES.map((q) => (
                    <button
                      key={q}
                      onClick={() => { setQuery(q); doSearch(q); }}
                      className="tag hover:border-brand-500/40 hover:text-brand-300 hover:bg-brand-950/40 transition-all"
                    >
                      {q}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Results */}
      {hasSearched && (
        <div className="max-w-3xl mx-auto px-4 sm:px-6 pb-16">
          {loading ? (
            <div className="flex flex-col items-center gap-4 py-16 text-surface-400">
              <div className="relative">
                <Loader2 className="w-10 h-10 animate-spin text-brand-400" />
                <Sparkles className="w-5 h-5 absolute -top-1 -right-1 text-accent-violet animate-pulse" />
              </div>
              <p className="text-sm">Searching with AI embeddings…</p>
            </div>
          ) : results.length === 0 ? (
            <div className="text-center py-16 glass-card">
              <div className="text-5xl mb-4">🤔</div>
              <h3 className="text-xl font-semibold text-white mb-2">No results found</h3>
              <p className="text-surface-300 mb-6">
                Try rephrasing or being more specific about the role and skills
              </p>
              <div className="flex flex-wrap justify-center gap-2">
                {SAMPLE_QUERIES.slice(0, 3).map((q) => (
                  <button
                    key={q}
                    onClick={() => { setQuery(q); doSearch(q); }}
                    className="tag hover:border-brand-500/40 hover:text-brand-300"
                  >
                    {q}
                    <ArrowRight className="w-3 h-3" />
                  </button>
                ))}
              </div>
            </div>
          ) : (
            <>
              <div className="flex items-center justify-between mb-4">
                <p className="text-surface-300 text-sm">
                  <span className="text-white font-semibold">{total}</span> results for &ldquo;{query}&rdquo;
                </p>
              </div>
              <div className="space-y-4">
                {results.map((job) => (
                  <JobCard key={job.id} job={job} />
                ))}
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
}
