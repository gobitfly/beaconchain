/** @type {import('tailwindcss').Config} */

export default {
  content: [],
  corePlugins: {
    // we have to get rid of all the other .container classes from scss first
    container: false,
  },
  plugins: [],
  theme: {
    colors: ({ colors }) => ({
      ...colors,
      orange: {
        50: '#fff8eb',
        100: '#ffecc6',
        200: '#ffd788',
        300: '#ffbd4a',
        400: '#ffaa31',
        500: '#f97f07',
        600: '#dd5a02',
        700: '#b73b06',
        800: '#942d0c',
        900: '#7a250d',
        950: '#461102',
      },
    }),
    fontFamily: {
      montserrat: [
        'Montserrat',
        'sans-serif',
      ],
    },
  },
}
