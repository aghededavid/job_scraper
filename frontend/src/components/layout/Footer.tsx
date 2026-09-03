import Link from 'next/link';
import { Zap, Github, Twitter, Linkedin, Heart } from 'lucide-react';

const FOOTER_LINKS = {
  Product: [
    { label: 'Browse Jobs',    href: '/jobs' },
    { label: 'Companies',      href: '/companies' },
    { label: 'Search',         href: '/search' },
    { label: 'Recommendations',href: '/recommendations' },
  ],
  Account: [
    { label: 'Sign Up',        href: '/auth/signup' },
    { label: 'Sign In',        href: '/auth/login' },
    { label: 'Dashboard',      href: '/dashboard' },
    { label: 'Settings',       href: '/profile/settings' },
  ],
  Company: [
    { label: 'About',          href: '/about' },
    { label: 'Blog',           href: '/blog' },
    { label: 'Careers',        href: '/careers' },
    { label: 'Contact',        href: '/contact' },
  ],
  Legal: [
    { label: 'Privacy Policy', href: '/privacy' },
    { label: 'Terms of Service',href: '/terms' },
    { label: 'Cookie Policy',  href: '/cookies' },
  ],
};

const SOCIAL_LINKS = [
  { icon: Github,   href: 'https://github.com/remotescope', label: 'GitHub' },
  { icon: Twitter,  href: 'https://twitter.com/remotescope', label: 'Twitter' },
  { icon: Linkedin, href: 'https://linkedin.com/company/remotescope', label: 'LinkedIn' },
];

export function Footer() {
  return (
    <footer className="border-t border-white/[0.06] bg-surface-900 mt-20">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Main footer */}
        <div className="py-16 grid grid-cols-2 md:grid-cols-6 gap-8">
          {/* Brand */}
          <div className="col-span-2">
            <Link href="/" className="flex items-center gap-2.5 mb-4 group w-fit">
              <div className="w-8 h-8 rounded-lg bg-gradient-brand flex items-center justify-center shadow-glow-sm">
                <Zap className="w-4 h-4 text-white" strokeWidth={2.5} />
              </div>
              <span className="font-bold text-lg">
                Remote<span className="text-brand-400">Scope</span>
              </span>
            </Link>
            <p className="text-sm text-surface-300 leading-relaxed mb-6 max-w-xs">
              Your mission control for finding the best remote jobs. Curated from top sources,
              powered by AI.
            </p>
            {/* Social links */}
            <div className="flex items-center gap-2">
              {SOCIAL_LINKS.map(({ icon: Icon, href, label }) => (
                <a
                  key={label}
                  href={href}
                  aria-label={label}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="p-2 rounded-lg text-surface-400 hover:text-white hover:bg-white/[0.05] transition-all"
                >
                  <Icon className="w-4 h-4" />
                </a>
              ))}
            </div>
          </div>

          {/* Links */}
          {Object.entries(FOOTER_LINKS).map(([category, links]) => (
            <div key={category}>
              <h3 className="section-label mb-4">{category}</h3>
              <ul className="space-y-2.5">
                {links.map((link) => (
                  <li key={link.href}>
                    <Link
                      href={link.href}
                      className="text-sm text-surface-300 hover:text-white transition-colors"
                    >
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Bottom bar */}
        <div className="py-6 border-t border-white/[0.06] flex flex-col sm:flex-row items-center justify-between gap-4">
          <p className="text-sm text-surface-400">
            © {new Date().getFullYear()} RemoteScope. All rights reserved.
          </p>
          <p className="text-sm text-surface-400 flex items-center gap-1.5">
            Made with <Heart className="w-3.5 h-3.5 text-accent-rose" fill="currentColor" /> for remote workers everywhere
          </p>
        </div>
      </div>
    </footer>
  );
}
