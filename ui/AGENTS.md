# Scope: Frontend Dashboard (`ui/`)

## Stack

- Vite 7 + React 19 + TypeScript + **Tailwind CSS v4** (via `@tailwindcss/vite` — config lives in CSS, there is no `tailwind.config.js`).
- Data fetching / server state: **`@tanstack/react-query`** over `axios`. Tables: `@tanstack/react-table`. Routing: `react-router-dom` v7.
- Package manager: **yarn** (classic). No new dependencies without explicit approval.

## Rules

- TypeScript strict mode. Explicit or implicit `any` is forbidden.
- Types mirroring backend resources (e.g. a `TerraformLayer`) must match the CRD shapes in `api/v1alpha1/*_types.go`.
- Server/remote state goes through `react-query`; component state stays local. Do not add a new state library (Redux, Zustand…) without approval.
- Build with atomic components from `src/components`. Style with Tailwind utility classes; don't add raw `.css` beyond the existing Tailwind theme entrypoint.
- Terraform plan logs can be huge — keep log rendering virtualized/lazy so the browser doesn't freeze.

## Validate

Before declaring a task done: `yarn lint` and `yarn build` (runs `tsc`). `yarn format-check` for formatting.
