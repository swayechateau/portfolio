/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./templates/**/*.html",
    "./static/**/*.js",
    "./main.go"
  ],
  theme: {
    extend: {
      backgroundImage: {
        'fade-bottom':
          'linear-gradient(180deg, transparent, rgba(37, 37, 37, 0.61), #111)',
        'fade-bottom-2':
          'linear-gradient(180deg, transparent, rgba(37, 37, 37, 0.61), #222)',
        'fade-top':
          'linear-gradient(360deg, transparent, rgba(37, 37, 37, 0.21), #111)',
        'main': 'linear-gradient(104deg, rgba(35, 32, 36, 0.5) 0%, rgba(24, 67, 23, 0.5) 50%, rgba(43, 105, 18, 0.5) 100%)',
      },
      animation: {
        glow: 'glow 4s infinite',
        blink: 'blink 1s infinite',
        gradient: 'gradient 15s ease infinite',
        para: 'para 50s linear infinite',
      },
      keyframes: {
        blink: {
          'from, to': {
            color: 'transparent',
          },
          '50%': {
            color: 'inherit',
          },
        },
        glow: {
          '0%': {
            'text-shadow': '0 0 10px inherit',
          },
          '15%': {
            'text-shadow':
              '2px 2px 10px rgba(0, 255, 0, 1), -2px -2px 10px rgba(0, 255, 0, 1)',
          },
          '30%': {
            'text-shadow':
              '2px 2px 4px rgba(0, 255, 0, .7), -2px -2px 4px rgba(0, 255, 0, .7)',
          },
          '50%': {
            'text-shadow':
              '20px 20px 50px rgba(255, 255, 255, .5), -20px -20px 50px rgba(255, 255, 255, .5)',
          },
        },
        gradient: {
          '0%': {
            'background-position': '0% 50%',
          },
          '50%': {
            'background-position': '100% 50%',
          },
          '100%': {
            'background-position': '0% 50%',
          },
        },
        para: {
          '100%': {
            'background-position':
              '-5000px 20%, -800px 95%, 500px 50%, 1000px 100%, 400px 0',
          },
        },
      },
    },
  },
  plugins: [],
}

