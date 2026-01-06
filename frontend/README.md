# Finance AI Frontend

This is the premium dashboard for Finance AI, built with Next.js.

## Technologies
- **Next.js 14** (Pages/App Router)
- **Tailwind CSS** for premium styling.
- **Recharts** for interactive financial visualizations.
- **Lucide React** for iconography.

## Design System
The project uses a custom premium design system defined in `src/app/globals.css`, featuring:
- Glassmorphism effects (`.glass`)
- Premium Input/Button styles (`.input-premium`, `.btn-premium`)
- Dark mode optimized with custom radial gradients.

## Running the Frontend
```bash
npm install
npm run dev
```
The application will be available at `http://localhost:3000` (or the next available port like 3001, 3002).

## Authentication
Tokens are stored in `localStorage` and sent via the `Authorization: Bearer <token>` header to the backend.
