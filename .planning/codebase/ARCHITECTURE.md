# Architecture

**Analysis Date:** 2026-03-11

## Pattern Overview

**Overall:** Microservices with decoupled frontend and backend

**Key Characteristics:**
- Two independent services communicating over HTTP/RPC
- Backend-driven logic and AI integrations in Go
- Frontend UI as separate Next.js application
- Services can evolve independently unless API contract changes
- Backend uses go-zero framework for API/RPC structure

## Layers

**Backend (Go - go-zero):**
- Purpose: API server, business logic, AI integrations, data management
- Location: `cmd/` and future `internal/`, `api/`, `service/` directories
- Contains: API/RPC handlers, service logic, data models, external integrations
- Depends on: go-zero framework, standard Go libraries
- Used by: Frontend application (via HTTP/RPC calls)

**Frontend (TypeScript/Next.js):**
- Purpose: User-facing UI, client state management, request coordination
- Location: Future Next.js app structure (app router, pages, components)
- Contains: React components, pages, API client calls, UI state
- Depends on: Backend APIs (HTTP/RPC)
- Used by: End users via web browser

**Communication:**
- Protocol: HTTP/RPC between backend and frontend
- Decoupling: API contract defines the interface
- Independence: Backend and frontend can be deployed separately

## Data Flow

**User Request Flow:**

1. User interacts with frontend (Next.js/React component)
2. Frontend sends HTTP/RPC request to backend service
3. Backend service processes request (validates, applies logic, calls external APIs/AI)
4. Backend returns response to frontend
5. Frontend updates UI based on response

**State Management:**
- Backend: Stateless request/response model (each request processed independently)
- Frontend: Client-side state management (to be determined as project grows)
- Persistence: Handled by backend (database, external services)

## Key Abstractions

**API/RPC Services (go-zero):**
- Purpose: Define service contracts and handlers
- Examples: Future `api/user.api`, `rpc/assistant.proto`
- Pattern: go-zero API definition files (.api) for HTTP endpoints and .proto for RPC

**Service Layer (Go):**
- Purpose: Encapsulate business logic, separate from HTTP transport
- Examples: Future `internal/service/` implementations
- Pattern: Service interfaces with concrete implementations

**API Client (TypeScript):**
- Purpose: Abstract backend API calls from UI components
- Examples: Future `lib/api/` client functions
- Pattern: Typed fetch wrappers or generated clients

## Entry Points

**Backend:**
- Location: `cmd/main.go`
- Triggers: Go runtime invocation
- Responsibilities: Initialize go-zero server, register handlers, start HTTP/RPC listeners

**Frontend:**
- Location: Future `app/page.tsx` or `pages/index.tsx`
- Triggers: Browser navigation to application URL
- Responsibilities: Render initial UI, hydrate application state

## Error Handling

**Strategy:** Explicit error propagation and user-facing error responses

**Patterns:**
- Backend (Go): Explicit error returns in function signatures, no panics in library code
- Frontend: Error boundaries and graceful degradation on failed API calls
- API Responses: HTTP status codes and error messages in response bodies

## Cross-Cutting Concerns

**Logging:** Not yet standardized (to be implemented as project grows)

**Validation:**
- Backend: Input validation at API boundaries (go-zero middleware/handlers)
- Frontend: Form validation before submission

**Authentication:** Not yet implemented (placeholder for future AI authentication/authorization)

---

*Architecture analysis: 2026-03-11*
