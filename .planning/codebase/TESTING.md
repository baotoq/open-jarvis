# Testing Patterns

**Analysis Date:** 2026-03-11

## Go (Backend) Testing

### Test Framework

**Built-in Testing:**
- Standard library `testing` package (no external framework needed)
- Run command: `go test ./...` (test all packages)
- Run specific package: `go test ./path/to/package/...`
- Verbose output: `go test -v ./...`
- Run specific test: `go test -run TestName ./path/...`
- Coverage: `go test -cover ./...`

**Test Configuration:**
- Go 1.22 as per `go.mod`
- Tests placed in `*_test.go` files alongside implementation
- Test functions named `TestXxx` where Xxx is descriptive
- Benchmark functions named `BenchmarkXxx`

### Test File Organization

**Location:**
- Co-located with implementation: `user_service.go` → `user_service_test.go`
- Same package as implementation, allows testing unexported functions
- Tests run with `go test` from project root

**Naming Convention:**
- Test files: `{module}_test.go`
- Test functions: `Test{FunctionName}` or `Test{FunctionName}_{Scenario}`
- Example: `TestGetUser`, `TestGetUser_NotFound`, `TestCreateUser_InvalidInput`

**File Structure:**
```go
package userservice

import (
    "testing"
)

func TestGetUser(t *testing.T) {
    // Arrange
    // Act
    // Assert
}

func TestGetUser_NotFound(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```

### Test Structure

**Standard Go Pattern (Arrange-Act-Assert):**
```go
func TestGetUser(t *testing.T) {
    // Arrange: Set up test data and dependencies
    userID := "123"
    expectedUser := &User{ID: "123", Name: "John"}

    // Act: Execute the code being tested
    user, err := GetUser(userID)

    // Assert: Verify the result
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.ID != expectedUser.ID {
        t.Errorf("got ID %s, want %s", user.ID, expectedUser.ID)
    }
}
```

**Error Assertion Pattern:**
```go
func TestGetUser_NotFound(t *testing.T) {
    user, err := GetUser("nonexistent")

    if err == nil {
        t.Error("expected error, got nil")
    }
    if user != nil {
        t.Errorf("expected nil user, got %v", user)
    }
}
```

**Table-Driven Tests:**
```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Mocking

**Approach:**
- Use interfaces for dependency injection
- Mock by creating test implementations of interfaces
- No external mocking library required for most cases
- `interface{}` for flexible test doubles

**Pattern:**
```go
// Production interface
type UserRepository interface {
    GetUser(id string) (*User, error)
}

// Mock for tests
type mockUserRepository struct {
    getUserFunc func(id string) (*User, error)
}

func (m *mockUserRepository) GetUser(id string) (*User, error) {
    return m.getUserFunc(id)
}

// Test usage
func TestUserService(t *testing.T) {
    mock := &mockUserRepository{
        getUserFunc: func(id string) (*User, error) {
            return &User{ID: id}, nil
        },
    }

    service := NewUserService(mock)
    // ... test service
}
```

**What to Mock:**
- External service calls (APIs, databases)
- Random number generators
- Time-dependent operations
- Hard-to-set-up dependencies

**What NOT to Mock:**
- Business logic
- Validation functions
- Pure functions
- Things you're testing

### Fixtures and Test Data

**Test Data Organization:**
- Create helper functions to build test objects
- Use constructors or factory functions
- Keep test data minimal but realistic

**Example:**
```go
func createTestUser(id, name string) *User {
    return &User{
        ID:        id,
        Name:      name,
        Email:     "test@example.com",
        CreatedAt: time.Now(),
    }
}

func TestGetUser(t *testing.T) {
    user := createTestUser("1", "Alice")
    // ... use user in test
}
```

### Coverage

**Viewing Coverage:**
```bash
go test -cover ./...          # Show coverage percentage
go test -coverprofile=cov.out ./...  # Generate coverage profile
go tool cover -html=cov.out   # View as HTML report
```

**Target:** No explicit requirement set; aim for >70% coverage of critical paths

**What to Prioritize:**
- Error paths and edge cases
- Public APIs
- Business logic
- Go-zero generated code: follow framework's testing patterns

## TypeScript (Frontend) Testing

### Test Framework & Configuration

**Framework:**
- Jest or Vitest (configuration not yet present in skeleton)
- When configured: watch mode for development
- Coverage reporting with threshold enforcement

**Run Commands (once configured):**
```bash
npm test                # Run all tests
npm test -- --watch    # Watch mode
npm test -- --coverage # Coverage report
npm run test:ci        # CI mode (no watch)
```

**Typical Configuration:**
- Test environment: jsdom (for React/DOM testing)
- Module paths configured to match tsconfig aliases
- Transform TypeScript with ts-jest or esbuild-jest

### Test File Organization

**Location Pattern:**
- Co-located with components: `UserCard.tsx` → `UserCard.test.tsx`
- Or separate `__tests__/` directory
- Recommended: co-located for easier maintenance

**Naming Convention:**
- Test files: `{Component}.test.tsx` or `{module}.test.ts`
- Test suites: `describe('ComponentName', () => {})`
- Test cases: `it('should do something', () => {})`

**Structure:**
```
components/
├── UserCard.tsx
├── UserCard.test.tsx
├── Button.tsx
└── Button.test.tsx
```

### Test Structure

**Testing Library Pattern (React):**
```typescript
import { render, screen } from '@testing-library/react';
import { UserCard } from './UserCard';

describe('UserCard', () => {
    it('should display user name', () => {
        // Arrange
        const user = { id: '1', name: 'Alice', email: 'alice@example.com' };

        // Act
        render(<UserCard user={user} />);

        // Assert
        expect(screen.getByText('Alice')).toBeInTheDocument();
    });

    it('should call onClick when button clicked', () => {
        // Arrange
        const handleClick = jest.fn();
        render(<UserCard user={user} onClick={handleClick} />);

        // Act
        const button = screen.getByRole('button');
        button.click();

        // Assert
        expect(handleClick).toHaveBeenCalledTimes(1);
    });
});
```

**Async Test Pattern:**
```typescript
it('should load user data', async () => {
    // Arrange
    const mockFetch = jest.fn().mockResolvedValue({
        json: async () => ({ id: '1', name: 'Alice' })
    });

    // Act
    render(<UserProfile userId="1" />);

    // Assert
    const userName = await screen.findByText('Alice');
    expect(userName).toBeInTheDocument();
});
```

### Mocking

**Framework:**
- Jest mocks: `jest.fn()`, `jest.mock()`, `jest.spyOn()`
- Mock modules: `jest.mock('@/services/apiClient')`
- Mock fetch/axios for API calls

**Mocking Patterns:**

1. **Module mocking:**
```typescript
jest.mock('@/services/userService', () => ({
    getUser: jest.fn().mockResolvedValue({ id: '1', name: 'Alice' })
}));

import { getUser } from '@/services/userService';

it('should fetch user', async () => {
    const user = await getUser('1');
    expect(user.name).toBe('Alice');
});
```

2. **Mock function assertions:**
```typescript
const mockCallback = jest.fn();
component.onClick = mockCallback;

// Test the callback
expect(mockCallback).toHaveBeenCalled();
expect(mockCallback).toHaveBeenCalledWith('arg1', 'arg2');
expect(mockCallback).toHaveBeenCalledTimes(1);
```

3. **Component mocking:**
```typescript
jest.mock('@/components/Button', () => ({
    Button: ({ children, onClick }: any) => (
        <button onClick={onClick}>{children}</button>
    )
}));
```

**What to Mock:**
- API calls (external services)
- Child components in isolation tests
- Date/time (if time-dependent)
- Random functions
- Browser APIs (localStorage, fetch, etc.)

**What NOT to Mock:**
- The component under test
- Simple utility functions
- Constants
- Basic React hooks in component tests (unless testing hook interaction)

### Fixtures and Test Data

**Test Data Builders:**
```typescript
function createUser(overrides = {}): User {
    return {
        id: '1',
        name: 'Test User',
        email: 'test@example.com',
        ...overrides
    };
}

it('should handle user with custom email', () => {
    const user = createUser({ email: 'custom@example.com' });
    // ... test with user
});
```

**Fixtures Directory:**
```
__fixtures__/
├── users.json
├── responses.json
└── mockData.ts
```

### Coverage

**Viewing Coverage:**
```bash
npm test -- --coverage
# View detailed HTML report
open coverage/lcov-report/index.html
```

**Configuration (example):**
```javascript
// jest.config.js
module.exports = {
    collectCoverageFrom: [
        'src/**/*.{ts,tsx}',
        '!src/**/*.d.ts',
        '!src/index.ts'
    ],
    coverageThreshold: {
        global: {
            branches: 60,
            functions: 60,
            lines: 60,
            statements: 60
        }
    }
};
```

**Target:** No explicit requirement set; aim for >70% coverage

## Test Types

### Unit Tests

**Scope:**
- Test single function, component, or module in isolation
- Mock all external dependencies
- Fast execution
- Primary focus for most development

**Go Example:**
```go
func TestParseEmail(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"john.doe@example.com", "john.doe"},
        {"user+tag@example.com", "user+tag"},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := ParseEmail(tt.input)
            if result != tt.expected {
                t.Errorf("got %s, want %s", result, tt.expected)
            }
        })
    }
}
```

### Integration Tests

**Scope:**
- Test multiple components working together
- Use real or in-memory databases when practical
- Slower than unit tests but catch integration issues
- Example: API handler + service + repository

**Approach (Go):**
```go
func TestGetUserFlow(t *testing.T) {
    // Set up test database
    db := setupTestDB()
    defer db.Close()

    repo := NewUserRepository(db)
    service := NewUserService(repo)

    // Test the flow
    user, err := service.GetUser("1")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // ... assertions
}
```

### E2E Tests

**Status:** Not yet configured in project skeleton

**When Needed:**
- Test critical user workflows
- API contracts
- Full request/response cycles

**Tools (when configured):**
- Go: httptest package or integration tests
- TypeScript/Next.js: Playwright, Cypress, or similar

## Common Patterns

### Async Testing (Go)

```go
func TestAsyncOperation(t *testing.T) {
    done := make(chan error, 1)

    go func() {
        result, err := LongRunningOperation()
        if err != nil {
            done <- err
        }
        done <- nil
    }()

    select {
    case err := <-done:
        if err != nil {
            t.Errorf("operation failed: %v", err)
        }
    case <-time.After(5 * time.Second):
        t.Error("operation timeout")
    }
}
```

### Async Testing (TypeScript)

```typescript
it('should complete async operation', async () => {
    const result = await asyncOperation();
    expect(result).toBeDefined();
});

it('should handle async error', async () => {
    await expect(failingAsyncOperation()).rejects.toThrow();
});
```

### Error Testing (Go)

```go
func TestErrorScenarios(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        wantError bool
        errorType string
    }{
        {"invalid input", "", true, "validation"},
        {"timeout", "slow", true, "timeout"},
        {"success", "valid", false, ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := Operation(tt.input)
            if (err != nil) != tt.wantError {
                t.Errorf("expected error: %v, got: %v", tt.wantError, err)
            }
        })
    }
}
```

### Error Testing (TypeScript)

```typescript
it('should throw validation error', () => {
    expect(() => validateEmail('invalid')).toThrow('Invalid email');
});

it('should catch promise rejection', async () => {
    await expect(fetchData('invalid')).rejects.toThrow('Not found');
});
```

## Best Practices

**General:**
- Test behavior, not implementation
- Write tests before or alongside code (TDD encourages)
- Keep tests focused and readable
- Use descriptive test names
- Avoid test interdependencies
- Clean up resources (databases, files, etc.)

**Naming:**
- Test names should describe what is being tested and expected outcome
- Good: `TestGetUser_WithValidID_ReturnsUser`
- Bad: `TestGetUser`, `Test1`

**Isolation:**
- Each test should be independent
- No setup dependencies between tests
- No shared mutable state

**Coverage:**
- Prioritize critical paths over coverage percentage
- Aim for meaningful assertions, not just line coverage
- Test error conditions and edge cases

---

*Testing analysis: 2026-03-11*
