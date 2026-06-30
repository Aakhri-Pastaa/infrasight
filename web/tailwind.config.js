/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
    "./src/**/*.css"
  ],
  theme: {
    extend: {
      colors: {
        background: 'var(--bg)',
        surface: 'var(--surface)',
        'surface-container-lowest': 'var(--surface-container-lowest)',
        'surface-container-low': 'var(--surface-container-low)',
        'surface-container': 'var(--surface-container)',
        'surface-container-high': 'var(--surface-container-high)',
        'surface-container-highest': 'var(--surface-container-highest)',
        primary: 'var(--primary)',
        'primary-container': 'var(--primary-container)',
        'primary-fixed-dim': 'var(--primary-fixed-dim)',
        secondary: 'var(--secondary)',
        'secondary-container': 'var(--secondary-container)',
        tertiary: 'var(--tertiary)',
        'tertiary-container': 'var(--tertiary-container)',
        error: 'var(--error)',
        'error-container': 'var(--error-container)',
        outline: 'var(--outline)',
        'outline-variant': 'var(--outline-variant)',
        'on-background': 'var(--on-background)',
        'on-surface': 'var(--on-surface)',
        'on-surface-variant': 'var(--on-surface-variant)',
        'on-primary': 'var(--on-primary)',
        'on-primary-container': 'var(--on-primary-container)',
      },
      fontFamily: {
        headline: ['Geist', 'sans-serif'],
        display: ['Geist', 'sans-serif'],
        body: ['Geist', 'sans-serif'],
        label: ['Geist', 'sans-serif'],
        mono: ['Geist Mono', 'monospace'],
      }
    },
  },
  plugins: [],
}
