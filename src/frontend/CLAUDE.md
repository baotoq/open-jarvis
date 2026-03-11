# CLAUDE.md — Frontend

Next.js 16 frontend for open-jarvis. TypeScript strict mode, Tailwind v4, shadcn/ui.

## Commands

> Run all commands from `src/frontend/`.

```bash
npm run dev      # dev server → http://localhost:3000
npm run build    # production build (must pass before committing)
npm run lint     # ESLint
npx tsc --noEmit # type-check without emitting
```

## Stack

- **Next.js 16** — App Router, React 19
- **Tailwind v4** — CSS-first (`@import "tailwindcss"` in globals.css, no `tailwind.config.ts`)
- **shadcn/ui** — component primitives in `components/ui/`
- **date-fns v4** — time formatting
- **lucide-react** — icons

## Structure

```
app/
  layout.tsx        # root layout — fonts, body classes
  page.tsx          # main page — sidebar + chat area composition
  globals.css       # Tailwind v4 @import, CSS variables
components/
  Sidebar.tsx       # conversation list with delete
  ChatArea.tsx      # SSE streaming chat + message history
  ui/               # shadcn/ui primitives (button, etc.)
hooks/
  useSession.ts     # localStorage session management
lib/
  api.ts            # typed fetch wrappers for backend API
  utils.ts          # cn() helper (clsx + tailwind-merge)
```

## Key Patterns

**`'use client'` boundaries:** Any component using hooks, browser APIs (localStorage, Date), or event handlers must have `'use client'` at the top. Server Components cannot use these.

**Hydration safety:** Never call `formatDistanceToNow` or `new Date()` in Server Components — use `'use client'` components for time-relative displays.

**localStorage access:** Only in `useEffect` or event handlers, never at module level. Avoids SSR hydration mismatch.

**Session management:** `useSession` hook in `hooks/useSession.ts` — reads/writes `jarvis-session-id` from localStorage, validates against backend on load, clears stale IDs silently.

**SSE parsing:** Read `ReadableStream` from `fetch` response; parse `data: ` prefix; handle final `{"done":true,"sessionId":"..."}` event to capture backend-assigned session ID.

**API calls:** All backend calls go through `lib/api.ts`. Backend base URL defaults to `http://localhost:8888`. Override with `NEXT_PUBLIC_API_URL` env var.

## Conventions

- PascalCase for components (`Sidebar.tsx`, `ChatArea.tsx`)
- camelCase for hooks (`useSession.ts`) with `use` prefix
- TypeScript strict mode — no `any`, no `// @ts-ignore`
- Tailwind classes only — no inline styles, no CSS modules
- `cn()` from `lib/utils.ts` for conditional class merging

## Tailwind v4 Notes

- Config is in `globals.css` via `@theme { }` block — **no `tailwind.config.ts`**
- CSS variables defined in `:root` and `.dark` via `@layer base`
- `tw-animate-css` for animation utilities

## Backend API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat/stream` | POST | SSE streaming chat |
| `/api/conversations` | GET | list all conversations |
| `/api/conversations/:id` | GET | get single conversation |
| `/api/conversations/:id` | DELETE | delete conversation |
| `/api/conversations/:id/messages` | GET | get messages for conversation |

Request body for chat: `{ sessionId: string, message: string }`

## Gotchas

- `formatDistanceToNow` from date-fns must only run client-side — `'use client'` required on any component using it
- Don't clear message history when starting a new SSE stream — wait for the stream to resolve before updating
- `useSearchParams` requires a Suspense boundary — wrap in `<Suspense>` if used
