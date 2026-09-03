/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Required for the production Docker stage (standalone server.js)
  output: process.env.NODE_ENV === 'production' ? 'standalone' : undefined,
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: '**' },
    ],
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.NEXT_PUBLIC_API_URL || 'http://backend:8080'}/api/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;
