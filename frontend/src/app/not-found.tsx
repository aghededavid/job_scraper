import Link from 'next/link';
import { Home, Search, ArrowLeft } from 'lucide-react';

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center px-4 bg-glow-top">
      <div className="text-center max-w-md">
        <div className="text-8xl font-black text-transparent bg-clip-text bg-gradient-to-b from-white/20 to-white/5 mb-6 select-none">
          404
        </div>
        <h1 className="text-2xl font-bold text-white mb-3">Page not found</h1>
        <p className="text-surface-300 mb-8">
          The page you&apos;re looking for doesn&apos;t exist or has been moved.
        </p>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
          <Link href="/" className="btn-primary">
            <Home className="w-4 h-4" />
            Go Home
          </Link>
          <Link href="/jobs" className="btn-ghost">
            <Search className="w-4 h-4" />
            Browse Jobs
          </Link>
        </div>
      </div>
    </div>
  );
}
