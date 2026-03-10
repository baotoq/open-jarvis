# Technology Stack

**Analysis Date:** 2026-03-11

## Languages

**Primary:**
- Go 1.22 - Backend API service and core logic
- TypeScript - Frontend and UI layer (planned, configuration present)

**Secondary:**
- Bash - Build and test scripts
- YAML - Configuration files

## Runtime

**Environment:**
- Go 1.22 runtime for backend service
- Node.js (version unspecified) for TypeScript/Next.js frontend

**Package Manager:**
- `go mod` 1.22 - Go dependency management
  - Lockfile: `go.sum` (minimal, no external dependencies yet)
- `npm` - TypeScript/Next.js package manager
  - Lockfile: Not yet created (frontend not initialized)

## Frameworks

**Core:**
- [go-zero](https://go-zero.dev) - Go backend framework for API/RPC services, microservices pattern
- Next.js (planned, version unspecified) - React-based TypeScript frontend with App Router
- React (via Next.js) - Frontend UI library

**Testing:**
- Go built-in `testing` package - Unit testing for Go
- Testing framework for TypeScript (not yet selected)

**Build/Dev:**
- Go build tool - Native Go compilation
- TypeScript compiler (via Next.js toolchain) - TypeScript compilation
- Serena - Development assistant tool for code navigation and editing

## Key Dependencies

**Critical:**
- go-zero module (imported via git submodule) - Framework for building Go microservices
  - Location: `.claude/skills/zero-skills` (git submodule from https://github.com/zeromicro/zero-skills.git)

**Infrastructure:**
- Standard Go stdlib - Core language functionality
- shadcn/ui MCP server (referenced in VSCode config) - UI component library integration for frontend

## Configuration

**Environment:**
- No `.env` configuration currently in place
- Configuration follows go-zero patterns (to be implemented)
- Frontend will use Next.js environment variables (not yet configured)

**Build:**
- `tsconfig.json` - TypeScript configuration with strict mode enabled
  - Target: ES2020
  - Module: commonjs
  - Strict: true
- Go build: Standard `go build`, `go run` commands
- VSCode integration via MCP (Model Context Protocol)

## Platform Requirements

**Development:**
- macOS/Unix-like environment (Darwin platform)
- Go 1.22 installed and in PATH
- Node.js/npm for TypeScript frontend (when development begins)
- git (version control)
- Standard Unix tools (bash, grep, find, ls)

**Production:**
- Deployment target: Not yet determined
- Backend: Go binary deployment (typical patterns: Docker, cloud platform, standalone binary)
- Frontend: Next.js deployment (typical patterns: Vercel, Docker, self-hosted Node.js)
- No container configuration (Dockerfile) currently present

## Development Tools

**Editors/IDEs:**
- VSCode - Primary IDE
- MCP integration for code assistance

**Code Analysis:**
- `go vet` - Go static analysis
- `gofmt` - Go code formatting (standard)

---

*Stack analysis: 2026-03-11*
