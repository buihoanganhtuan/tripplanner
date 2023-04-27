/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/*.{js,ts,jsx,tsx}",
    "./src/components/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      gridTemplateRows: {
        'landing': '75px 0.2fr 1fr',
        'login-pane': '1fr 10fr 3fr',
        'login-input': '0.5fr 1fr 0.5fr 1fr 10px'
      },
    },
  },
  plugins: [],
}

