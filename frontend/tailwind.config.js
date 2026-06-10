/** @type {import('tailwindcss').Config} */
// 合并自 docs/prototypes 各原型的 tailwind.config，保留原始视觉风格。
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        ink: '#172033',
        muted: '#64748B',
        line: '#E2E8F0',
        canvas: '#F6F8FB',
        soft: '#F7F9FC',
        brand: {
          50: '#EEF6FF',
          100: '#D9ECFF',
          500: '#2563EB',
          600: '#1D4ED8',
          700: '#1E40AF',
        },
        violetx: {
          50: '#F5F3FF',
          100: '#EDE9FE',
          500: '#7C3AED',
          600: '#6D28D9',
        },
      },
      boxShadow: {
        soft: '0 1px 2px rgba(15, 23, 42, .05), 0 10px 30px rgba(15, 23, 42, .06)',
      },
      fontFamily: {
        sans: [
          'Inter',
          'ui-sans-serif',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          '"Segoe UI"',
          'sans-serif',
        ],
      },
    },
  },
  plugins: [],
}
