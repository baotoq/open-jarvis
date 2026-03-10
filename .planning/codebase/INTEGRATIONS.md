# External Integrations

**Analysis Date:** 2026-03-11

## APIs & External Services

**AI/LLM Integration (Planned):**
- OpenAI - Mentioned in project inspiration (OpenClaw is AI-powered personal assistant)
  - SDK/Client: Not yet integrated
  - Auth: To be configured via environment variables (standard pattern)

**Code Assistance:**
- shadcn UI - Component library for frontend
  - SDK/Client: `shadcn@latest` (via npm MCP server)
  - Use case: UI component management and integration

## Data Storage

**Databases:**
- No database currently configured
- No ORM/database client selected
- Planned: Storage layer to be added as project develops

**File Storage:**
- Local filesystem only (no external storage integration)

**Caching:**
- None currently configured
- Planned: Cache layer to be added as project scales

## Authentication & Identity

**Auth Provider:**
- Custom (planned)
- Implementation: Not yet started
- Planned approach: go-zero patterns for API authentication (likely JWT or similar)

**Frontend Auth:**
- TypeScript/Next.js implementation planned
- Pattern: To follow Next.js auth conventions (middleware, session management)

## Monitoring & Observability

**Error Tracking:**
- None configured
- Standard Go error handling returns (explicit error propagation pattern used)

**Logs:**
- Console logging via Go's standard library
- No structured logging framework currently implemented
- Planned: Add structured logging as project grows

## CI/CD & Deployment

**Hosting:**
- Not yet determined
- Backend: Typical Go deployment patterns (Docker, binary distribution, cloud platform)
- Frontend: Typical Next.js patterns (Vercel, self-hosted, Docker)

**CI Pipeline:**
- None currently configured
- Planned: GitHub Actions or similar (git repository infrastructure present)

**Build Commands:**
- Go: `go build ./...`, `go run ./cmd/...`
- TypeScript: `npm run build`, `npm run dev`

## Environment Configuration

**Required env vars:**
- None currently configured
- Planned: API keys, database connection strings, AI service credentials (to be added)

**Secrets location:**
- Not yet established
- Development: Planned to use `.env` or `.env.local` files
- Production: Planned to use platform-specific secret management

## Webhooks & Callbacks

**Incoming:**
- Not yet implemented
- Planned: API endpoints via go-zero framework

**Outgoing:**
- Not yet implemented
- Planned: Webhooks to external AI services or integrations

## External Libraries & Services

**GitHub:**
- Repository hosting: Yes (git remote configured)
- CI/CD integration: Planned

**go-zero Framework:**
- Provides RPC/API service generation
- Code generation via `.proto` or `.api` files (pattern in place, no files yet)

## Code Quality Tools

**Development:**
- Serena MCP server - AI-assisted code navigation and analysis
- VSCode MCP integration for IDE support

---

*Integration audit: 2026-03-11*
