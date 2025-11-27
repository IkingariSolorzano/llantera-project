/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./src/**/*.{html,ts}",
    ],
    darkMode: 'selector',
    theme: {
        extend: {
            boxShadow: {
                '3xl': 'rgba(149, 157, 165, 0.2) 0px 8px 24px'
            }
        },
        colors: {
            primary: {
                50: '#ECF0FF',
                100: '#DDE4FF',
                200: '#C2CDFF',
                300: '#9CAAFF',
                400: '#757DFF',
                500: '#605DFF',
                600: '#4936F5',
                700: '#3E2AD8',
                800: '#3225AE',
                900: '#2D2689'
            },
            secondary: {
                50: '#EEF6FF',
                100: '#DAEBFF',
                200: '#BDDCFF',
                300: '#90C7FF',
                400: '#5DA8FF',
                500: '#3584FC',
                600: '#1F64F1',
                700: '#174EDE',
                800: '#1940B4',
                900: '#1A3A8E'
            },
            purple: {
                50: '#FAF5FF',
                100: '#F3E8FF',
                200: '#E9D5FF',
                300: '#D7B5FD',
                400: '#BF85FB',
                500: '#AD63F6',
                600: '#9135E8',
                700: '#7C24CC',
                800: '#6A22A7',
                900: '#571D86'
            },
            success: {
                50: '#eeffe5',
                100: '#d8ffc8',
                200: '#b2ff97',
                300: '#82fc5a',
                400: '#58f229',
                500: '#37d80a',
                600: '#25b003',
                700: '#1e8308',
                800: '#1d670d',
                900: '#1a5710'
            },
            orange: {
                50: '#fff5ed',
                100: '#ffe8d4',
                200: '#ffcea9',
                300: '#ffaa72',
                400: '#fe7a36',
                500: '#fd5812',
                600: '#ee3e08',
                700: '#c52b09',
                800: '#9c2410',
                900: '#7e2010'
            },
            danger: {
                50: '#fff2f0',
                100: '#ffe1dd',
                200: '#ffc8c0',
                300: '#ffa294',
                400: '#ff6d57',
                500: '#FF4023',
                600: '#ec1f00',
                700: '#d71c00',
                800: '#b11a03',
                900: '#921c0a'
            },
            gray: {
                50: '#f6f7f9',
                100: '#eceef2',
                200: '#d5d9e2',
                300: '#b1bbc8',
                400: '#8695aa',
                500: '#64748b',
                600: '#526077',
                700: '#434e61',
                800: '#3a4252',
                900: '#23272e'
            },
            warning: {
                50: '#fff8e1',
                100: '#ffecb3',
                200: '#ffe082',
                300: '#ffd54f',
                400: '#ffca28',
                500: '#ffc107',
                600: '#ffb300',
                700: '#ffa000',
                800: '#ff8f00',
                900: '#ff6f00'
            },
            info: {
                50: '#e1f7fe',
                100: '#b3ecfc',
                200: '#80e0fa',
                300: '#48d3f5',
                400: '#0dcaf0',
                500: '#00c1eb',
                600: '#00b1d7',
                700: '#009cbb',
                800: '#0088a2',
                900: '#006574'
            },
            white: '#ffffff',
            black: '#3A4252',
            dark: '#000000'
        },
        fontSize: {
            xs: '12px',
            sm: '13px',
            base: '14px',
            md: '16px',
            lg: '18px',
            xl: '24px',
            '2xl': '28px',
            '3xl': '32px',
            '4xl': '36px',
            '5xl': '40px',
        },
        fontFamily: {
            body: ['Inter', 'system-ui', 'sans-serif']
        }
    },
    plugins: [],
}