import type { Metadata, Viewport } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { Navbar } from '@/components/layout/Navbar';
import { Footer } from '@/components/layout/Footer';
import { AuthProvider } from '@/contexts/AuthContext';

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap',
});

export const metadata: Metadata = {
  title: {
    default: 'RemoteScope — Find Your Remote Dream Job',
    template: '%s | RemoteScope',
  },
  description:
    'Discover thousands of remote jobs aggregated from top platforms. Filter by skills, salary, and timezone. Land your ideal remote position with RemoteScope.',
  keywords: ['remote jobs', 'work from home', 'remote work', 'job board', 'tech jobs'],
  authors: [{ name: 'RemoteScope' }],
  metadataBase: new URL(process.env.NEXT_PUBLIC_APP_URL || 'https://remotescope.app'),
  openGraph: {
    type: 'website',
    siteName: 'RemoteScope',
    title: 'RemoteScope — Find Your Remote Dream Job',
    description: 'Discover thousands of curated remote jobs from top companies worldwide.',
    images: ['/og-image.png'],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'RemoteScope — Find Your Remote Dream Job',
    description: 'Discover thousands of curated remote jobs from top companies worldwide.',
  },
};

export const viewport: Viewport = {
  themeColor: '#0a0a0f',
  colorScheme: 'dark',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${inter.variable} dark`}>
      <body className="bg-surface-900 text-white antialiased">
        <AuthProvider>
          <div className="min-h-screen flex flex-col">
            <Navbar />
            <main className="flex-1">{children}</main>
            <Footer />
          </div>
        </AuthProvider>
      </body>
    </html>
  );
}
