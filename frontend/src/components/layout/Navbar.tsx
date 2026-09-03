'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Search, Bell, User, Menu, X, Zap, BookmarkCheck,
  LayoutDashboard, Settings, LogOut, Shield
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useAuth } from '@/contexts/AuthContext';

const NAV_LINKS = [
  { href: '/jobs',    label: 'Jobs' },
  { href: '/search',  label: 'Search' },
  { href: '/companies', label: 'Companies' },
];

export function Navbar() {
  const pathname = usePathname();
  const { user, isAuthenticated, isAdmin, logout } = useAuth();
  const [scrolled, setScrolled] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [profileOpen, setProfileOpen] = useState(false);

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 10);
    window.addEventListener('scroll', handleScroll, { passive: true });
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  useEffect(() => {
    setMobileOpen(false);
    setProfileOpen(false);
  }, [pathname]);

  return (
    <header
      className={cn(
        'fixed top-0 left-0 right-0 z-50 transition-all duration-300',
        scrolled
          ? 'bg-surface-900/90 backdrop-blur-xl border-b border-white/[0.06] shadow-glass'
          : 'bg-transparent'
      )}
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link href="/" className="flex items-center gap-2.5 group">
            <div className="w-8 h-8 rounded-lg bg-gradient-brand flex items-center justify-center shadow-glow-sm group-hover:shadow-glow-md transition-all">
              <Zap className="w-4 h-4 text-white" strokeWidth={2.5} />
            </div>
            <span className="font-bold text-lg tracking-tight">
              Remote<span className="text-brand-400">Scope</span>
            </span>
          </Link>

          {/* Desktop Nav */}
          <nav className="hidden md:flex items-center gap-1">
            {NAV_LINKS.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className={cn(
                  'px-4 py-2 rounded-lg text-sm font-medium transition-all',
                  pathname === link.href || pathname.startsWith(link.href + '/')
                    ? 'text-white bg-white/[0.08]'
                    : 'text-surface-300 hover:text-white hover:bg-white/[0.05]'
                )}
              >
                {link.label}
              </Link>
            ))}
          </nav>

          {/* Right side */}
          <div className="flex items-center gap-2">
            {/* Search button */}
            <Link
              href="/search"
              className="hidden sm:flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-surface-300 hover:text-white
                         bg-surface-700/60 border border-white/[0.06] hover:border-white/10 transition-all"
            >
              <Search className="w-3.5 h-3.5" />
              <span className="text-xs text-surface-400">Search jobs...</span>
              <kbd className="ml-2 px-1.5 py-0.5 rounded text-[10px] bg-surface-600 text-surface-400 font-mono">⌘K</kbd>
            </Link>

            {isAuthenticated ? (
              <>
                {/* Notifications */}
                <Link
                  href="/dashboard"
                  className="relative p-2 rounded-lg text-surface-300 hover:text-white hover:bg-white/[0.05] transition-all"
                >
                  <Bell className="w-5 h-5" />
                  <span className="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-accent-rose border border-surface-900" />
                </Link>

                {/* Profile dropdown */}
                <div className="relative">
                  <button
                    id="profile-menu-btn"
                    onClick={() => setProfileOpen(!profileOpen)}
                    className="flex items-center gap-2 p-1.5 rounded-xl hover:bg-white/[0.05] transition-all"
                  >
                    <div className="w-8 h-8 rounded-full bg-gradient-brand flex items-center justify-center text-sm font-semibold">
                      {user?.avatar_url ? (
                        <img src={user.avatar_url} alt={user.name} className="w-full h-full rounded-full object-cover" />
                      ) : (
                        user?.name?.[0]?.toUpperCase() || 'U'
                      )}
                    </div>
                  </button>

                  {profileOpen && (
                    <div className="absolute right-0 top-12 w-56 glass-card p-1.5 animate-slide-down">
                      <div className="px-3 py-2 border-b border-white/[0.06] mb-1">
                        <p className="text-sm font-medium text-white truncate">{user?.name}</p>
                        <p className="text-xs text-surface-400 truncate">{user?.email}</p>
                      </div>
                      <ProfileMenuItem href="/dashboard" icon={<LayoutDashboard className="w-4 h-4" />} label="Dashboard" />
                      <ProfileMenuItem href="/dashboard" icon={<BookmarkCheck className="w-4 h-4" />} label="Saved Jobs" />
                      <ProfileMenuItem href="/profile" icon={<User className="w-4 h-4" />} label="Profile" />
                      <ProfileMenuItem href="/profile/settings" icon={<Settings className="w-4 h-4" />} label="Settings" />
                      {isAdmin && (
                        <ProfileMenuItem href="/admin" icon={<Shield className="w-4 h-4" />} label="Admin Panel" />
                      )}
                      <div className="mt-1 pt-1 border-t border-white/[0.06]">
                        <button
                          onClick={() => logout()}
                          className="w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm text-surface-300
                                     hover:text-accent-rose hover:bg-accent-rose/10 transition-all text-left"
                        >
                          <LogOut className="w-4 h-4" />
                          Sign Out
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              </>
            ) : (
              <>
                <Link href="/auth/login" className="btn-ghost text-sm py-2 px-4 hidden sm:inline-flex">
                  Sign In
                </Link>
                <Link href="/auth/signup" className="btn-primary text-sm py-2 px-4">
                  Get Started
                </Link>
              </>
            )}

            {/* Mobile menu toggle */}
            <button
              id="mobile-menu-btn"
              className="md:hidden p-2 rounded-lg text-surface-300 hover:text-white hover:bg-white/[0.05] transition-all"
              onClick={() => setMobileOpen(!mobileOpen)}
            >
              {mobileOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>
        </div>

        {/* Mobile menu */}
        {mobileOpen && (
          <div className="md:hidden py-4 border-t border-white/[0.06] animate-slide-down">
            {NAV_LINKS.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="block px-4 py-3 rounded-lg text-sm font-medium text-surface-300 hover:text-white hover:bg-white/[0.05] transition-all"
              >
                {link.label}
              </Link>
            ))}
            {!isAuthenticated && (
              <div className="flex gap-2 mt-4 pt-4 border-t border-white/[0.06]">
                <Link href="/auth/login" className="btn-ghost flex-1 justify-center text-sm">Sign In</Link>
                <Link href="/auth/signup" className="btn-primary flex-1 justify-center text-sm">Get Started</Link>
              </div>
            )}
          </div>
        )}
      </div>
    </header>
  );
}

function ProfileMenuItem({
  href,
  icon,
  label,
}: {
  href: string;
  icon: React.ReactNode;
  label: string;
}) {
  return (
    <Link
      href={href}
      className="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm text-surface-300
                 hover:text-white hover:bg-white/[0.05] transition-all"
    >
      {icon}
      {label}
    </Link>
  );
}
