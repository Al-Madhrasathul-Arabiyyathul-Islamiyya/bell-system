export default {
    content: ['./index.html', './src/**/*.{vue,js,ts}'],
    theme: {
      extend: {
        colors: {
          school: {
            green: '#189543',
            surface: '#E8F5E9',
            text: '#F8FAFC',
          },
          primary: {
            green: '#189543',
            red: '#DA231C',
            accent: '#F5A623',
          },
          secondary: {
            green: '#E8F5E9',
            red: '#FFEBEE',
            gold: '#FFF3E0',
          },
          neutral: {
            background: '#F8FAFC',
            surface: '#F1F5F9',
            text: '#0F172A',
          },
          dark: {
            primary: {
              green: '#1FB954',
              red: '#E53935',
              accent: '#FFB74D',
            },
            secondary: {
              green: '#1B5E20',
              red: '#B71C1C',
              gold: '#FF8F00',
            },
            neutral: {
              background: '#121212',
              surface: '#1E1E1E',
              text: '#FFFFFF',
            },
          },
        },
      },
      fontFamily: {
        sans: [
          'Sofia Sans, sans-serif',
          {
            fontFeatureSettings: '"cv11", "ss01"',
            fontVariationSettings: '"opsz" 32',
          },
        ],
        serif: [
          'Playfair Display, serif',
          {
            fontFeatureSettings: '"cv11", "ss01"',
            fontVariationSettings: '"opsz" 32',
          },
        ],
        mono: ['Geist Mono, monospace'],
      },
    },
    plugins: [],
  }
