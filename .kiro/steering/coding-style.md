# Coding Style Guide

## TypeScript/React Frontend

### Code Quality
- All TypeScript code must follow ESLint best practices with Airbnb configuration + Prettier
- Avoid using `any`; always prefer explicit and strict types
- Use `async/await` instead of `.then()` for promises
- Define interfaces or types for all complex objects and API responses
- Each file should be no longer than 200 lines; refactor into smaller modules if needed
- All functions must have explicit return types

### Naming Conventions
- **Components**: `PascalCase.tsx` (e.g., `NotificationProvider.tsx`)
- **Hooks**: `camelCase.ts` starting with `use` (e.g., `useDocumentOperations.ts`)
- **Utilities**: `camelCase.ts` (e.g., `apiUtils.ts`)
- **Types**: `camelCase.ts` (e.g., `document.ts`)
- **Tests**: `*.test.tsx` or `*.test.ts`
- **Variables/Functions**: `camelCase`
- **Interfaces/Types**: `PascalCase`
- **Constants**: `UPPER_SNAKE_CASE`

### Import Organization
Organize imports in this order with blank lines between groups:
1. React and Next.js imports
2. External packages (libraries)
3. Internal modules (relative imports)

```typescript
import React, { useState, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import { useApiError } from './useApiError';
import { DocumentUploadRequest } from '@/types/document';
```

### Component Structure
- Use functional components with hooks
- Define interfaces for all props
- Use `useCallback` for event handlers to prevent unnecessary re-renders
- Use `useMemo` for expensive calculations
- Implement proper error boundaries
- Always handle loading and error states

### Error Handling
- Use custom error handling hooks (e.g., `useApiError`)
- Implement graceful degradation for network failures
- Show user-friendly error messages
- Log errors appropriately (dev vs production)

### API Integration
- Use centralized API client with interceptors
- Implement retry logic with exponential backoff
- Handle authentication token refresh automatically
- Use React Query for caching and state management
- Implement offline support where applicable

## Go Backend

### Code Quality
- Follow Go standard formatting with `gofmt`
- Use `golint` and `go vet` for code quality
- All exported functions must have comments
- Use meaningful variable and function names
- Keep functions small and focused (max 50 lines)
- Handle all errors explicitly

### Naming Conventions
- **Files**: `snake_case.go`
- **Packages**: lowercase, single word when possible
- **Interfaces**: `PascalCase` with descriptive names (e.g., `UserRepository`)
- **Structs**: `PascalCase`
- **Functions/Methods**: `PascalCase` for exported, `camelCase` for private
- **Variables**: `camelCase`
- **Constants**: `PascalCase` or `UPPER_SNAKE_CASE` for package-level

### Import Organization
Organize imports in this order with blank lines between groups:
1. **Standard library imports** (alphabetically sorted)
2. **External package imports** (third-party libraries, alphabetically sorted)
3. **Internal package imports** (project modules, alphabetically sorted)

```go
import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"

    "digital-signature-system/internal/config"
    "digital-signature-system/internal/domain/entities"
    "digital-signature-system/internal/domain/repositories"
)
```

#### Import Rules:
- Use blank lines to separate import groups
- Sort imports alphabetically within each group
- Use import aliases for conflicting package names or long paths:
  ```go
  import (
      infraServices "digital-signature-system/internal/infrastructure/services"
  )
  ```
- Avoid dot imports (`.`) except for testing packages
- Use underscore imports (`_`) only for side effects (drivers, init functions)
- Group related imports together within the same category

### Package Structure
- Follow Clean Architecture layers:
  - `domain/`: Business logic (entities, repositories, services, usecases)
  - `infrastructure/`: External concerns (database, HTTP, services)
  - `cmd/`: Application entry points
- Use dependency injection container pattern
- Keep domain layer independent of infrastructure

### Error Handling
- Use custom error types for different error categories
- Wrap errors with context using `fmt.Errorf`
- Log errors at appropriate levels
- Return meaningful error messages to clients
- Use middleware for centralized error handling

### HTTP Handlers
- Use request/response structs with validation tags
- Implement proper HTTP status codes
- Use middleware for cross-cutting concerns
- Validate all input data
- Handle authentication and authorization consistently

### Database
- Use GORM for ORM operations
- Define clear entity models with proper tags
- Use migrations for schema changes
- Implement repository pattern for data access
- Use transactions for multi-step operations

## Testing Standards

### Frontend Testing
- Use Jest + React Testing Library
- Test user interactions, not implementation details
- Mock external dependencies
- Achieve minimum 80% code coverage
- Test error scenarios and edge cases

### Backend Testing
- Use testify for assertions
- Write unit tests for all business logic
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test error conditions thoroughly

### Test Structure
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Security Guidelines

### Input Validation
- Validate all user inputs at API boundaries
- Use struct tags for validation rules
- Sanitize data to prevent XSS attacks
- Implement rate limiting
- Use parameterized queries to prevent SQL injection

### Authentication & Authorization
- Use JWT tokens with proper expiration
- Implement refresh token rotation
- Store sensitive data securely (environment variables)
- Use HTTPS in production
- Implement proper session management

### Error Messages
- Don't expose internal system details in error messages
- Use generic error messages for authentication failures
- Log detailed errors server-side for debugging
- Implement audit logging for security events

## Performance Guidelines

### Frontend
- Use React.memo for expensive components
- Implement virtual scrolling for large lists
- Lazy load components and routes
- Optimize bundle size with code splitting
- Use service workers for caching

### Backend
- Use database indexes appropriately
- Implement caching for frequently accessed data
- Use connection pooling
- Optimize database queries
- Implement proper pagination

## Documentation
- Document all public APIs
- Use JSDoc for TypeScript functions
- Write clear commit messages
- Maintain README files for each module
- Document deployment procedures