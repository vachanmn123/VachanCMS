# VachanCMS Frontend Implementation Plan

## Overview
This plan outlines the implementation of a Vue-based admin frontend for VachanCMS, a headless CMS using GitHub repos for storage. The frontend will provide a user-friendly interface for managing repositories, content types, and content values, with a focus on dark mode, sidebar layout, and extensible form components.

## Tech Stack
- **Vue 3** with Composition API and `<script setup>`.
- **Vue Router** for navigation.
- **TailwindCSS v4.1**: Hardcoded dark mode (no toggle).
- **Shadcn-vue**: For UI components (Dialog, Button, Input, Table, etc.).
- **Zod**: For form validation (per-element and per-form, constructed dynamically).
- **Pinia**: For state management (auth, repo).
- **Axios**: For API calls.
- **Vite**: For build tool and dev server.
- Build output: `frontend/dist`.

## Backend Modifications
- **API Grouping**: In `main.go`, create `/api` group and pass to `SetupRoutes(router.Group("/api"))`.
- **/me Route**: Add protected GET `/api/me` in `routes.go` and `handlers/auth.go`. Decode JWT, fetch user from GitHub API, return: `{ "login": "string", "name": "string", "avatar_url": "string" }`.
- **Static Serving**: In `main.go`, serve `./frontend/dist` at root. Check if file exists; if yes, serve it; else, serve `index.html`.

## Project Structure
```
frontend/
├── src/
|   |-- styles.css # Tailwind imports and custom styles
│   ├── components/
│   │   ├── fields/  # Separate Vue files: TextField.vue, NumberField.vue, BooleanField.vue, SelectField.vue
│   │   └── ui/      # Shadcn components
│   ├── views/
│   │   ├── Login.vue
│   │   ├── RepoSelect.vue  # With search/filter
│   │   └── Dashboard.vue   # Main layout with sidebar
│   ├── stores/
│   │   ├── auth.js  # { user: null, isAuthenticated: false }
│   │   ├── repo.js  # { selected: { owner, repo }, config: null }
│   ├── router/index.js  # Routes: /, /repos, /dashboard/:owner/:repo, /dashboard/:owner/:repo/:ctSlug
│   ├── utils/
│   │   ├── fieldComponents.js  # Map: { text: TextField, ... }
│   │   ├── zodSchemas.js       # Helper to build schemas per field/form
│   ├── App.vue
│   └── main.js
├── vite.config.js  # Proxy /api to backend
└── components.json
```

## Key Features & Implementation Details

### 1. Authentication
- On app load: Check `auth_token` cookie. If present, fetch `/api/me` and set user in store. Else, redirect to `/`.
- Login.vue: Button redirects to `/api/auth/login`.

### 2. Repository Selection
- RepoSelect.vue: Fetch `/api/repos`, display list with search/filter input.
- On select: Check `/api/:owner/:repo/config`. If 404, show dialog: "Initialize repo? This will create/edit files." Prompt for site_name, call `/api/:owner/:repo/init` if confirmed. Then navigate to `/dashboard/:owner/:repo`.

### 3. Dashboard Layout
- Hardcoded dark theme.
- Sidebar: List content types from store (fetched on mount via `/api/:owner/:repo/config`). "New Type" button opens dialog.
- Main: Conditional render based on route (e.g., content types list or values table).

### 4. Content Type Management
- New Type Dialog: Fields for name, slug (auto-generate), dynamic array of fields (add/remove buttons). Use fieldComponents.js for rendering (each emits `update:value`). Zod schemas: Base per field type, combine into form schema.
- POST to `/api/:owner/:repo/content-types`.

### 5. Content Value Management
- Table: Columns from content type fields. Pagination via `?page=`.
- Add/Edit Dialog: Dynamic form using fieldComponents.js. Zod: Construct schema from fields array.
- POST/PUT to `/api/:owner/:repo/:ctSlug` or `/:id`.

### 6. Routing
- `/`: Login redirect.
- `/repos`: RepoSelect.
- `/dashboard/:owner/:repo`: Dashboard (sidebar + content types list).
- `/dashboard/:owner/:repo/:ctSlug`: Values table for that type.

## Assumptions & Notes
- Field components (TextField.vue, etc.): Simple inputs that emit `update:value`. Can be extended later.
- Zod schemas: Helper function in `zodSchemas.js` to generate per field (e.g., `z.string()` for text), then combine for form.
- Error handling: Toasts/dialogs for API errors.
- Mobile: Tailwind responsive classes.
- No media yet, as per scope.
- Component naming: Follow default ESLint and Prettier rules.

## Implementation Steps
1. Modify backend (main.go, routes.go, handlers/auth.go).
2. Create frontend project structure with Vite.
3. Install dependencies and set up Shadcn-vue, TailwindCSS v4.1.
4. Implement authentication and stores.
5. Build views and routing.
6. Add content type and value management with dynamic forms.
7. Test integration with backend.
8. Build and verify static serving.
