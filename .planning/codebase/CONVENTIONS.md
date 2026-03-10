# Coding Conventions

**Analysis Date:** 2026-03-11

## Project Overview

Open-Jarvis is a dual-language project with separate Go (backend) and TypeScript (frontend) services. Conventions differ between the two languages but follow standard industry practices for each ecosystem.

## Go (Backend) Conventions

### Naming Patterns

**Files:**
- Snake_case for filenames (e.g., `api_handler.go`, `user_service.go`)
- `*_test.go` suffix for test files in the same package
- Package directories use lowercase with underscores if multi-word

**Functions:**
- PascalCase for exported functions (visible outside package): `GetUser()`, `CreateAccount()`
- camelCase for unexported functions (package-private): `getUserByID()`, `validateInput()`

**Variables:**
- camelCase for all variables: `userID`, `firstName`, `isActive`
- ALL_CAPS only for constants: `MAX_RETRIES`, `DEFAULT_TIMEOUT`

**Types:**
- PascalCase for all type names (structs, interfaces): `User`, `UserService`, `ApiRequest`
- Interface names typically end with `-er` suffix: `Reader`, `Writer`, `Handler`

**Packages:**
- Lowercase single-word package names when possible: `api`, `service`, `model`
- Multi-word packages use underscores: `user_service`, `api_handler`
- No dots or hyphens in package names

### Code Style

**Formatting:**
- Use `gofmt` for all code formatting
- Indentation: tabs (gofmt standard)
- Line length: conventional Go follows ~80 char limit but allows flexibility for readability

**Linting:**
- Run `go vet ./...` for static analysis
- Follow standard Go idioms and conventions
- `go vet` checks are mandatory before commits

**Error Handling:**
- Explicit error returns: functions return `(value, error)` tuples
- Check errors immediately: `if err != nil { return err }`
- Never silently ignore errors
- Wrap errors with context when appropriate: use `fmt.Errorf("operation failed: %w", err)`
- Pattern in `cmd/main.go`:
  ```go
  package main

  func main() {}
  ```

### Import Organization

**Order (Go convention):**
1. Standard library imports: `fmt`, `io`, `net/http`
2. Third-party imports: `github.com/zeromicro/go-zero`, other external packages
3. Local package imports: `open-jarvis/internal`, `open-jarvis/pkg`

**Example:**
```go
import (
    "fmt"
    "log"

    "github.com/zeromicro/go-zero/core/logx"

    "open-jarvis/internal/user"
    "open-jarvis/pkg/config"
)
```

### go-zero Framework Patterns

The project uses [go-zero](https://go-zero.dev) framework for API and RPC definitions:
- API definitions in `.api` files or protobuf
- Service layer separation: API handlers, service logic, data access
- Dependency injection through constructor functions
- Middleware for cross-cutting concerns

### Module Organization

- `cmd/main.go`: Entry point for the application
- `internal/`: Private packages not exposed to external consumers
- `pkg/`: Public packages that may be consumed by other modules
- `api/`: go-zero API definitions or protobuf files (when created)

## TypeScript (Frontend) Conventions

### Naming Patterns

**Files:**
- PascalCase for React components: `UserProfile.tsx`, `LoginForm.tsx`
- camelCase for utility/service files: `userService.ts`, `apiClient.ts`
- kebab-case for pages in Next.js App Router: `user-settings.tsx`, `login-page.tsx`

**Functions:**
- camelCase for utility functions: `formatDate()`, `validateEmail()`
- PascalCase for React component functions: `UserCard()`, `NavigationBar()`
- camelCase for event handlers with `handle` prefix: `handleClick()`, `handleSubmit()`

**Variables:**
- camelCase for all variables: `userName`, `isLoading`, `userId`
- UPPER_SNAKE_CASE for constants: `MAX_RETRIES`, `API_BASE_URL`

**Types & Interfaces:**
- PascalCase for all type/interface names: `User`, `ApiResponse`, `ButtonProps`
- Prefix interfaces with `I` only if necessary for clarity (generally avoid)
- Use `type` for unions and primitives, `interface` for object shapes

**React-specific:**
- Component files match component name: `UserProfile.tsx` exports `UserProfile`
- Custom hooks use `use` prefix: `useAuth()`, `useFetchUser()`

### Code Style

**Formatting:**
- Configuration in `tsconfig.json`: ES2020 target, CommonJS module, strict mode
- Use Prettier or equivalent for consistent formatting
- 2-space indentation standard
- Semicolons required

**Strict Mode:**
- TypeScript strict mode enabled in `tsconfig.json`
- All variables must have explicit types or inferred correctly
- No `any` type except in very specific documented cases
- Null/undefined must be explicitly handled

**Linting:**
- Run `npm run lint` before commits
- Follow ESLint rules for React/Next.js projects
- Rules should prevent common mistakes and enforce consistency

### Next.js App Router Conventions

- Use Next.js App Router (not Pages Router)
- Place pages in `app/` directory with route segments
- Server components by default, use `'use client'` for interactive components
- API routes in `app/api/` directory

### Import Organization

**Order (TypeScript convention):**
1. External libraries: `react`, `next`
2. Third-party libraries: `axios`, `lodash`, etc.
3. Relative imports: `../services`, `./components`
4. Type imports: `import type { User } from '...'` separated at end

**Path Aliases:**
- If configured: use `@/` for `src/` or project root imports
- Example: `import { Button } from '@/components/Button'`

### Error Handling

**Patterns:**
- Use try-catch for async operations
- Throw descriptive errors with context
- Type error handling with custom error classes when needed
- Always handle promise rejections in async code

**Example pattern:**
```typescript
try {
    const user = await fetchUser(id);
    return user;
} catch (error) {
    console.error('Failed to fetch user:', error);
    throw new Error(`User fetch failed: ${error.message}`);
}
```

### Module Design

**Component Structure:**
```
components/
├── Button.tsx           # Single responsibility
├── UserCard.tsx
└── Form/
    ├── LoginForm.tsx
    └── useLoginForm.ts  # Hook for form logic
```

**Service Files:**
- `services/userService.ts` - API calls
- `services/authService.ts` - Authentication logic
- `utils/helpers.ts` - Utility functions

**Exports:**
- One main export per file when possible
- Re-export related items from index files: `export { Button, ButtonProps }`

## Shared Conventions (Both Languages)

### Comments

**When to Comment:**
- Explain *why*, not *what* - code should be self-documenting
- Document public APIs, complex algorithms, non-obvious decisions
- Mark temporary workarounds with `TODO` or `FIXME`

**Documentation:**
- Go: Use godoc comments for exported items
- TypeScript: Use JSDoc comments for exported functions/types
- Example (Go):
  ```go
  // GetUser retrieves a user by ID from the database.
  func GetUser(id string) (*User, error) {
  ```
- Example (TypeScript):
  ```typescript
  /**
   * Fetches a user by ID
   * @param id - The user ID
   * @returns The user object
   */
  async function getUser(id: string): Promise<User> {
  ```

### Error Messages

- Be specific about what failed
- Include relevant context values
- Use consistent format: "Operation name: what went wrong: details"
- Example: "User creation failed: invalid email format: user@invalid"

### Logging

**When to Log:**
- Significant operations: authentication, database calls, external service calls
- Errors and warnings
- Debug info when troubleshooting
- Never log passwords, API keys, or PII

**Format:**
- Go: use `logx` from go-zero framework when available
- TypeScript: use `console.log`, `console.error`, or appropriate logging library
- Include context: operation name, relevant IDs, timestamp

## Testing Conventions

See `TESTING.md` for detailed testing patterns and frameworks.

## Deviations & Exceptions

**When to deviate:**
- Document deviation in code comment
- Include rationale for deviation
- Get team agreement if possible
- Plan to address later if it's technical debt

---

*Convention analysis: 2026-03-11*
