# Codebase Structure

**Analysis Date:** 2026-03-11

## Directory Layout

```
open-jarvis/
├── cmd/                    # Go application entrypoints
│   └── main.go            # Backend server entry point (empty, to be populated)
├── go.mod                 # Go module definition
├── tsconfig.json          # TypeScript compiler configuration
├── .vscode/               # VS Code configuration
├── .claude/               # Claude Code context (internal)
├── .serena/               # Project memory and context
├── .planning/             # GSD planning and analysis
│   └── codebase/          # Generated codebase analysis documents
└── README.md              # Project overview
```

## Directory Purposes

**cmd/:**
- Purpose: Go application entry points and command-line interfaces
- Contains: `main.go` with server initialization
- Key files: `cmd/main.go`
- Future: May expand with additional commands (CLI tools, migrations, etc.)

**go.mod:**
- Purpose: Go module metadata and dependency management
- Contains: Module name (open-jarvis), Go version (1.22)
- Configuration file; not code

**tsconfig.json:**
- Purpose: TypeScript compiler options and project configuration
- Contains: Compiler target (ES2020), module system (commonjs), strict mode enabled
- Configuration file; applies to future TypeScript/Next.js source

**.vscode/:**
- Purpose: Editor configuration for VS Code (local development)
- Contains: Custom settings, extensions, launch configs
- Not committed code; development convenience

**.claude/, .serena/:**
- Purpose: Internal context and memory management for AI agents
- Not part of application logic

**.planning/:**
- Purpose: GSD (Getting Stuff Done) planning and analysis artifacts
- Contains: Generated codebase analysis (ARCHITECTURE.md, STRUCTURE.md, etc.)
- Not application code

## Key File Locations

**Entry Points:**
- `cmd/main.go`: Backend server initialization (currently empty)

**Configuration:**
- `go.mod`: Go dependencies and module configuration
- `tsconfig.json`: TypeScript/Next.js compilation settings
- `.vscode/`: Development environment setup

**Core Logic:**
- Future: `internal/service/` - Business logic implementations
- Future: `internal/handler/` - API request handlers
- Future: Frontend app in separate location (app/pages/ or app/router/ in Next.js)

**Testing:**
- Future: `*_test.go` files alongside Go source files
- Future: `*.test.ts`, `*.spec.ts` files alongside TypeScript source

## Naming Conventions

**Files:**
- Go files: `lowercase_with_underscores.go` (standard Go convention)
- TypeScript files: `camelCase.ts` or `PascalCase.tsx` (components use PascalCase)
- Directories: `lowercase_with_underscores` for Go packages, `lowercase` for TypeScript

**Directories:**
- Go packages: `lowercase_with_underscores` (e.g., `internal/`, `api/`, `service/`)
- Next.js app: `app/`, `components/`, `lib/`, `public/` (standard Next.js App Router)

**Functions/Variables:**
- Go: `camelCase` for unexported, `PascalCase` for exported
- TypeScript: `camelCase` for functions/variables, `PascalCase` for types/components

## Where to Add New Code

**New Backend Feature:**
- API definition: `api/*.api` (go-zero API files)
- Handler: `internal/handler/*.go`
- Service logic: `internal/service/*.go`
- Tests: Alongside implementation files as `*_test.go`

**New Frontend Feature:**
- Components: Create new `.tsx` files under frontend app directory
- Pages: Add to frontend app router (App Router: `app/(routes)/`)
- API client: Add functions to `lib/api/` or similar client library
- Tests: Alongside component files as `*.test.tsx` or `*.spec.tsx`

**New Utilities:**
- Go utilities: `internal/util/` or feature-specific package
- TypeScript utilities: `lib/` directory (shared helpers)

**Database/Models:**
- Go models: `internal/model/` or `internal/domain/`
- Future: Consider ORM integration (GORM, sqlc, etc.)

## Special Directories

**cmd/:**
- Purpose: Executable entry points (main packages)
- Generated: No
- Committed: Yes
- Future expansion: May contain multiple commands (cmd/server, cmd/cli, etc.)

**internal/:**
- Purpose: Package-private code (not importable by external projects)
- Generated: No
- Committed: Yes
- Status: Not yet created; will be standard for Go service code

**Frontend root:**
- Purpose: Next.js application (location to be determined)
- Generated: Partially (.next/ build output)
- Committed: Source files yes, build output no
- Status: Not yet created; expected to be separate or in dedicated frontend/ directory

**.vscode/, .claude/, .serena/:**
- Purpose: Development and agent configuration
- Generated: Yes (cache/ subdirectories)
- Committed: Partial (configuration yes, cache no)

## Missing Directories (to be created)

**For Go Backend:**
- `internal/` - Service implementations, models, utilities
- `api/` - Go-zero API definitions (.api files)
- `test/` or `testdata/` - Test fixtures and data

**For Frontend:**
- `app/` or `src/` - Next.js application root (following Next.js 13+ App Router)
- `components/` - React components
- `lib/` - Utility functions and API clients
- `public/` - Static assets

---

*Structure analysis: 2026-03-11*
