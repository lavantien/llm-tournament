/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./templates/**/*.html",
    "./tools/screenshots/**/*.html",
  ],
  safelist: [
    // Dynamic classes used in JavaScript for score theming
    'score-0', 'score-20', 'score-40', 'score-60', 'score-80', 'score-100',
    // Profile grouping classes
    'profile-group-divider',
    'profile-group-start',
    'profile-group-end',
  ],
  theme: {
    extend: {
      colors: {
        // Background colors
        bg: {
          0: '#070a12',
          1: '#0b1020',
        },
        // Surface colors with transparency
        surface: {
          0: 'rgba(16, 24, 38, 0.72)',
          1: 'rgba(10, 15, 26, 0.7)',
          2: 'rgba(14, 20, 33, 0.62)',
        },
        // Border colors
        stroke: {
          DEFAULT: 'rgba(255, 255, 255, 0.14)',
          strong: 'rgba(255, 255, 255, 0.22)',
        },
        // Accent colors
        cyan: '#19f7ff',
        magenta: '#ff2bd6',
        violet: '#a77bff',
        green: '#7cff6b',
        amber: '#ffc857',
        red: '#ff4d4d',
        // Text colors
        text: {
          DEFAULT: 'rgba(244, 248, 255, 0.92)',
          muted: 'rgba(244, 248, 255, 0.64)',
          faint: 'rgba(244, 248, 255, 0.46)',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'Segoe UI', 'Roboto', 'Arial', 'sans-serif'],
        mono: ['JetBrains Mono', 'ui-monospace', 'Cascadia Mono', 'Consolas', 'Menlo', 'monospace'],
      },
      fontSize: {
        'xs': ['12px', { lineHeight: '16px', letterSpacing: '0.08em' }],
        'sm': ['13px', { lineHeight: '20px' }],
        'base': ['14px', { lineHeight: '22px', letterSpacing: '0.01em' }],
      },
      borderRadius: {
        'lg': '18px',
        'md': '14px',
      },
      boxShadow: {
        'glass': '0 18px 60px rgba(0, 0, 0, 0.55)',
        'glow-cyan': '0 0 22px rgba(25, 247, 255, 0.14)',
        'glow-magenta': '0 0 22px rgba(255, 43, 214, 0.14)',
        'btn': '0 0 0 1px rgba(255, 255, 255, 0.07) inset, 0 14px 44px rgba(0, 0, 0, 0.35)',
        'btn-hover': '0 0 0 1px rgba(255, 255, 255, 0.1) inset, 0 16px 58px rgba(0, 0, 0, 0.45), 0 0 22px rgba(25, 247, 255, 0.14)',
        'btn-focus': '0 0 0 1px rgba(255, 255, 255, 0.1) inset, 0 0 0 3px rgba(25, 247, 255, 0.22), 0 0 30px rgba(25, 247, 255, 0.18)',
        'score': '0 0 0 1px rgba(255, 255, 255, 0.06) inset, 0 12px 36px rgba(0, 0, 0, 0.35)',
        'input': '0 0 0 1px rgba(255, 255, 255, 0.06) inset',
        'input-focus': '0 0 0 3px rgba(25, 247, 255, 0.16)',
      },
      backgroundImage: {
        'gradient-radial-cyan': 'radial-gradient(1200px 800px at 15% 10%, rgba(25, 247, 255, 0.12), transparent 55%)',
        'gradient-radial-magenta': 'radial-gradient(900px 700px at 80% 20%, rgba(255, 43, 214, 0.1), transparent 50%)',
        'gradient-radial-violet': 'radial-gradient(700px 520px at 70% 85%, rgba(167, 123, 255, 0.12), transparent 55%)',
        'gradient-vertical': 'linear-gradient(180deg, #0b1020, #070a12)',
        'gradient-surface': 'linear-gradient(180deg, rgba(255, 255, 255, 0.08), transparent 22%), linear-gradient(180deg, rgba(16, 24, 38, 0.72), rgba(14, 20, 33, 0.62))',
        'gradient-glow': 'radial-gradient(900px 140px at 20% 0%, rgba(25, 247, 255, 0.12), transparent 55%), radial-gradient(800px 120px at 80% 0%, rgba(255, 43, 214, 0.11), transparent 55%)',
        'gradient-btn': 'linear-gradient(90deg, rgba(25, 247, 255, 0.95), rgba(255, 43, 214, 0.95), rgba(167, 123, 255, 0.95))',
        'gradient-input': 'linear-gradient(180deg, rgba(10, 15, 26, 0.78), rgba(10, 15, 26, 0.55))',
      },
      spacing: {
        'shell-pad': '12px',
        'shell-gap': '12px',
        'topbar-height': '66px',
        'sidebar-width': '140px',
        'content-max-width': '1320px',
      },
      backdropBlur: {
        'glass': '14px',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}
