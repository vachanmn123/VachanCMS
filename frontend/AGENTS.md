# VachanCMS - Agent Coding Guidelines

## Overview

This document provides coding guidelines and commands for agentic coding assistants working on the VachanCMS project. VachanCMS consists of a Go backend API and a Vue.js/TypeScript frontend admin interface.

## Build/Lint/Test Commands

### Frontend (Vue.js/TypeScript)

Located in `/frontend` directory.

**Development:**

```bash
npm run dev          # Start development server with hot reload
```

**Build:**

```bash
npm run build        # Production build (includes type-check)
npm run build-only   # Build without type checking
npm run preview      # Preview production build locally
```

**Type Checking:**

```bash
npm run type-check   # Run TypeScript compiler for type checking
```

**Linting & Formatting:**

```bash
npm run lint         # Run ESLint with auto-fix and cache
npm run format       # Run Prettier to format code
```

**Testing:**

- No automated unit tests configured
- Manual API testing via `tests.js` in root directory
- For single test runs: No specific command (no unit tests exist)

### Backend (Go)

Located in root directory.

**Development:**

```bash
go run main.go       # Start development server
```

**Dependencies:**

```bash
go mod tidy          # Download and organize dependencies
```

**Testing:**

- No automated tests configured
- Manual API testing via `tests.js`

## Code Style Guidelines

### Vue.js Components

- Use Vue 3 Composition API with `<script setup lang="ts">`
- Define props interfaces with TypeScript:

```typescript
interface Props {
  label: string
  required?: boolean
  modelValue: string
}

const props = withDefaults(defineProps<Props>(), {
  required: false,
})
```

- Use typed emits:

```typescript
const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
```

- Component naming: PascalCase (e.g., `TextField.vue`, `ContentTypesView.vue`)
- Template structure: Single root element, clear hierarchy
- Use `ref()` for reactive data, `computed()` for derived state

### TypeScript

- Strict TypeScript configuration enabled
- Use explicit types for function parameters and return values
- Prefer interfaces over types for object definitions
- Use union types for variant props: `type Variant = 'primary' | 'secondary'`

### Imports and Dependencies

- Import aliases: `@/` maps to `./src/`
- Group imports: external libraries first, then internal components
- Use absolute imports for internal modules:

```typescript
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
```

### State Management (Pinia)

- Use Pinia stores for global state
- Store naming: camelCase with descriptive names (e.g., `useAuthStore`)
- Actions for async operations, getters for computed state

### Styling (Tailwind CSS + Shadcn-vue)

- Use Shadcn-vue components from `@/components/ui/`
- Apply Tailwind classes directly in templates
- Use `cn()` utility for conditional classes:

```typescript
import { cn } from '@/lib/utils'

class={cn(
  "base-classes",
  variant === 'primary' && "primary-classes"
)}
```

### File Organization

```
frontend/src/
├── components/
│   ├── fields/        # Form field components
│   └── ui/           # Shadcn-vue UI components
├── views/            # Page components
├── stores/           # Pinia stores
├── router/           # Vue Router configuration
├── lib/              # Utilities and helpers
└── styles.css        # Global styles
```

### Naming Conventions

- **Files**: PascalCase for Vue components, camelCase for utilities
- **Variables**: camelCase
- **Constants**: UPPER_SNAKE_CASE
- **Functions**: camelCase, descriptive names
- **Components**: PascalCase, descriptive and specific

### Error Handling

- Use try/catch for async operations
- Validate API responses
- Show user-friendly error messages via toasts/dialogs
- Handle loading states appropriately

### API Integration

- Use Axios for HTTP requests
- Base URL configured via Vite proxy (`/api` → `http://localhost:8080`)
- Handle authentication tokens automatically via interceptors
- Type API responses when possible

### Vue Router

- Use named routes for programmatic navigation
- Lazy load route components:

```typescript
component: () => import('@/views/ContentTypesView.vue')
```

- Protect routes with meta fields and navigation guards

### Form Validation (Zod)

- Define schemas for form validation:

```typescript
import { z } from 'zod'

const schema = z.object({
  title: z.string().min(1, 'Title is required'),
  content: z.string().optional(),
})
```

- Validate forms before submission
- Provide user feedback for validation errors

### Performance Considerations

- Use `computed()` for expensive calculations
- Implement proper key attributes in v-for loops
- Lazy load components and routes
- Optimize images and assets

## Development Workflow

1. **Start Development:**

   ```bash
   cd frontend
   npm run dev
   ```

2. **Make Changes:**
   - Follow code style guidelines
   - Use TypeScript for type safety
   - Run linting: `npm run lint`
   - Format code: `npm run format`

3. **Type Check:**

   ```bash
   npm run type-check
   ```

4. **Build for Production:**

   ```bash
   npm run build
   ```

5. **Test Integration:**
   - Start backend: `go run main.go`
   - Test API endpoints via browser or `tests.js`

## Component Patterns

### Field Components

- Accept `modelValue` prop and emit `update:modelValue`
- Support validation states
- Include proper labeling and accessibility

### View Components

- Use reactive data from Pinia stores
- Handle loading and error states
- Implement responsive design with Tailwind

### Store Patterns

- Separate concerns: auth, repository, content management
- Use actions for API calls
- Provide reactive getters for components

## Security Considerations

- Never log sensitive data (tokens, passwords)
- Validate all user inputs
- Use HTTPS in production
- Implement proper CORS policies

## Commit Guidelines

- Use descriptive commit messages
- Group related changes
- Reference issue numbers when applicable
- Keep commits focused and atomic

## Tooling Configuration

- **ESLint**: Vue 3 + TypeScript recommended rules
- **Prettier**: Default configuration with skip formatting for ESLint
- **TypeScript**: Strict mode enabled
- **Vite**: Vue 3 + Tailwind CSS plugins
- **Shadcn-vue**: New York style, TypeScript enabled

## No External Rules Found

- No Cursor rules (.cursor/rules/ or .cursorrules)
- No Copilot instructions (.github/copilot-instructions.md)

## Additional Resources

- Vue 3 Composition API documentation
- TypeScript handbook
- Tailwind CSS documentation
- Shadcn-vue component library
- Pinia state management guide
