# Frontend (React + Vite + MUI)

This app is a new frontend for the e-commerce system. It uses React + TypeScript + Vite, Material UI for consistent UI components, React Router for client-side routing, and TanStack Query for server state.

## Getting started

1. Install dependencies:

```bash
cd frontend
npm install
```

2. Copy environment file and adjust if needed:

```bash
cp .env.example .env
```

3. Start development server:

```bash
npm run dev
```

4. Build for production:

```bash
npm run build
```

## Environment variables

- `VITE_API_BASE_URL` - Base URL for the gateway/BFF API.

Example value:

```env
VITE_API_BASE_URL=http://localhost:8080
```

## High-level structure

```text
src/
  app/          # app entry, providers, router
  components/   # shared common, layout, and thin UI wrappers
  features/     # feature-first modules: catalog, cart, checkout, orders
  lib/          # shared api client, env, query keys, formatters
  theme/        # centralized MUI theme setup
  types/        # shared TS types
  utils/        # generic helpers and error utilities
```

## Adding future features

Create a new folder under `src/features/<feature-name>/` and keep each feature split into:
- `api/` for request functions
- `hooks/` for TanStack Query hooks
- `components/` for feature-scoped UI
- `pages/` for route pages
- `types/` for DTOs/types

This keeps concerns separated and avoids scattering API logic into page components.
