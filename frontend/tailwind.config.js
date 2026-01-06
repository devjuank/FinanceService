/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./src/**/*.{js,ts,jsx,tsx,mdx}",
    ],
    theme: {
        extend: {
            colors: {
                background: "var(--background)",
                foreground: "var(--foreground)",
                surface: "var(--surface)",
                border: "var(--border)",
                'p-blue': "var(--p-blue)",
                'p-green': "var(--p-green)",
                'p-red': "var(--p-red)",
                'p-purple': "var(--p-purple)",
            },
        },
    },
    plugins: [],
}
