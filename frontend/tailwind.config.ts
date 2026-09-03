import type { Config } from 'tailwindcss';

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['var(--font-inter)', 'system-ui', 'sans-serif'],
        mono: ['var(--font-mono)', 'monospace'],
      },
      colors: {
        brand: {
          50:  '#f0f4ff',
          100: '#e0e9ff',
          200: '#c4d4ff',
          300: '#98b0ff',
          400: '#6680ff',
          500: '#4355f5',
          600: '#3134eb',
          700: '#2728d4',
          800: '#2424ac',
          900: '#222488',
          950: '#151558',
        },
        surface: {
          900: '#0a0a0f',
          800: '#111118',
          700: '#18181f',
          600: '#1f1f2a',
          500: '#272735',
          400: '#32323f',
          300: '#3d3d4d',
        },
        accent: {
          cyan:    '#00d4ff',
          violet:  '#a855f7',
          emerald: '#10b981',
          amber:   '#f59e0b',
          rose:    '#f43f5e',
        },
      },
      backgroundImage: {
        'gradient-radial':   'radial-gradient(var(--tw-gradient-stops))',
        'gradient-conic':    'conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))',
        'gradient-brand':    'linear-gradient(135deg, #4355f5 0%, #a855f7 100%)',
        'gradient-glow':     'radial-gradient(circle at 50% 0%, rgba(67,85,245,0.15) 0%, transparent 60%)',
      },
      animation: {
        'fade-in':      'fadeIn 0.3s ease-out',
        'slide-up':     'slideUp 0.4s ease-out',
        'slide-down':   'slideDown 0.3s ease-out',
        'pulse-glow':   'pulseGlow 2s ease-in-out infinite',
        'shimmer':      'shimmer 1.5s infinite',
        'float':        'float 3s ease-in-out infinite',
      },
      keyframes: {
        fadeIn:    { from: { opacity: '0' }, to: { opacity: '1' } },
        slideUp:   { from: { opacity: '0', transform: 'translateY(16px)' }, to: { opacity: '1', transform: 'translateY(0)' } },
        slideDown: { from: { opacity: '0', transform: 'translateY(-8px)' }, to: { opacity: '1', transform: 'translateY(0)' } },
        pulseGlow: {
          '0%, 100%': { boxShadow: '0 0 20px rgba(67,85,245,0.3)' },
          '50%':      { boxShadow: '0 0 40px rgba(67,85,245,0.6)' },
        },
        shimmer: {
          '0%':   { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%':      { transform: 'translateY(-8px)' },
        },
      },
      boxShadow: {
        'glow-sm':   '0 0 15px rgba(67,85,245,0.2)',
        'glow-md':   '0 0 30px rgba(67,85,245,0.3)',
        'glow-lg':   '0 0 60px rgba(67,85,245,0.4)',
        'glass':     '0 8px 32px rgba(0,0,0,0.4), inset 0 1px 0 rgba(255,255,255,0.05)',
        'card':      '0 4px 24px rgba(0,0,0,0.3)',
        'card-hover':'0 8px 40px rgba(0,0,0,0.5)',
      },
      backdropBlur: {
        xs: '2px',
      },
    },
  },
  plugins: [],
};

export default config;
