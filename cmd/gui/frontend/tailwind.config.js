/** @type {import('tailwindcss').Config} */
export default {
    content: [
        './src/**/*.{html,js,svelte,ts}',
        './node_modules/flowbite-svelte/**/*.{html,js,svelte,ts}',
    ],
    plugins: [require('flowbite/plugin')],
    darkMode: 'class',
    theme: {
        extend: {
            colors: {
                // flowbite-svelte
                primary: {
                    50: '#f0f9ff',
                    100: '#e0f2fe',
                    200: '#bae6fd',
                    300: '#7dd3fc',
                    400: '#38bdf8',
                    500: '#0ea5e9',
                    600: '#0284c7',
                    700: '#0369a1',
                    800: '#075985',
                    900: '#0c4a6e',
                    950: '#082f49'
                },
                // Your custom colors
                'bg-color': '#1e1e1e',
                'container-bg': '#2a2a2a',
                'text-color': '#d4d4d4',
                'accent-blue': '#7cafc2',
                'accent-green': '#99c794'
            }
        }
    }
};