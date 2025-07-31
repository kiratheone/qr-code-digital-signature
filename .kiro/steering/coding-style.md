# Coding Style Guide

## TypeScript/React Frontend

### Code Quality
- All TypeScript code must follow ESLint best practices with Airbnb configuration + Prettier
- Avoid using `any`; always prefer explicit and strict types
- Use `async/await` instead of `.then()` for promises
- Define interfaces or types for all complex objects and API responses
- Each file should be no longer than 200 lines; refactor into smaller modules if needed
- All functions must have explicit return types

### Naming Conventions - Simple Clean Architecture
- **Services**: `PascalCase.ts` in `lib/services/` (e.g., `DocumentService.ts`)
- **API Client**: `camelCase.ts` in `lib/api/` (e.g., `apiClient.ts`)
- **Components**: `PascalCase.tsx` in `components/` (e.g., `DocumentUploadForm.tsx`)
- **Hooks**: `camelCase.ts` starting with `use` in `hooks/` (e.g., `useDocumentOperations.ts`)
- **Types**: `camelCase.ts` in `lib/types/` (e.g., `document.ts`)
- **Utils**: `camelCase.ts` in `lib/utils/` (e.g., `formatDate.ts`)
- **Pages**: `page.tsx` in `app/` directories
- **Tests**: `*.test.tsx` or `*.test.ts` alongside source files
- **Variables/Functions**: `camelCase`
- **Interfaces/Types**: `PascalCase`
- **Constants**: `UPPER_SNAKE_CASE`

### Import Organization - Clean Architecture
Organize imports in this order with blank lines between groups:
1. React and Next.js imports
2. External packages (libraries)
3. Internal modules (services, api, types, utils)
4. Components and hooks
5. Relative imports

```typescript
import React, { useState, useCallback } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';

import { DocumentService } from '@/lib/services/DocumentService';
import { ApiClient } from '@/lib/api/apiClient';
import { Document } from '@/lib/types/document';

import { Button } from '@/components/ui/Button';
import { useDocumentOperations } from '@/hooks/useDocumentOperations';

import './DocumentUpload.css';
```

### Component Structure - Simple Clean Architecture
- **Presentation Components**: Only handle UI rendering and user interactions
- **Business Logic**: Keep in services (`lib/services/`) and custom hooks (`hooks/`)
- **Data Fetching**: Use custom hooks that wrap React Query
- **Props Interface**: Define clear interfaces for all component props
- **Error Handling**: Use error boundaries and simple error states
- **Loading States**: Handle loading states in custom hooks, display in components

```typescript
// Good: Presentation component with clear separation
export function DocumentUploadForm({ onUpload, isLoading }: DocumentUploadFormProps) {
  const [file, setFile] = useState<File | null>(null);
  const [issuer, setIssuer] = useState('');
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (file && issuer) {
      onUpload(file, issuer); // Business logic handled by parent/hook
    }
  };
  
  return (
    <form onSubmit={handleSubmit}>
      {/* UI only */}
    </form>
  );
}

// Good: Page component using custom hook for business logic
export function DocumentsPage() {
  const { documents, signDocument, isLoading } = useDocumentOperations();
  
  return (
    <div>
      <DocumentUploadForm onUpload={signDocument} isLoading={isLoading} />
      <DocumentList documents={documents} />
    </div>
  );
}
```

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

## Testing Standards - Focused on Business Logic

### Backend Testing Priority
1. **Business Logic Tests (High Priority)**
   - Test all services in `domain/services/` thoroughly
   - Focus on cryptographic operations, document signing, verification logic
   - Use table-driven tests for multiple scenarios
   - Mock external dependencies (database, file system)

2. **Integration Tests (Medium Priority)**
   - Test critical API endpoints end-to-end
   - Focus on authentication flow and document operations
   - Use test containers for database integration

3. **Unit Tests (Lower Priority)**
   - Test complex utility functions
   - Test error handling scenarios

### Frontend Testing - Simple & Focused
1. **Service Layer Tests (High Priority)**
   - Test business logic in `lib/services/`
   - Mock API client, test service methods
   - Focus on error handling and data transformation

2. **Hook Tests (Medium Priority)**
   - Test custom hooks with React Query
   - Mock services, test hook behavior
   - Test loading states and error scenarios

3. **Component Tests (Lower Priority)**
   - Test critical user interactions only
   - Mock hooks and services
   - Focus on form submissions and error displays

```typescript
// Example: Service test
describe('DocumentService', () => {
  it('should sign document with correct data', async () => {
    const mockApiClient = {
      post: jest.fn().mockResolvedValue({ id: '123', status: 'signed' })
    };
    const service = new DocumentService(mockApiClient);
    
    const result = await service.signDocument(mockFile, 'John Doe');
    
    expect(mockApiClient.post).toHaveBeenCalledWith('/documents/sign', expect.any(FormData));
    expect(result.status).toBe('signed');
  });
});

// Example: Component test
describe('DocumentUploadForm', () => {
  it('should call onUpload when form is submitted', () => {
    const mockOnUpload = jest.fn();
    render(<DocumentUploadForm onUpload={mockOnUpload} isLoading={false} />);
    
    // User interactions
    fireEvent.change(screen.getByLabelText(/file/i), { target: { files: [mockFile] } });
    fireEvent.change(screen.getByLabelText(/issuer/i), { target: { value: 'John Doe' } });
    fireEvent.click(screen.getByRole('button', { name: /sign/i }));
    
    expect(mockOnUpload).toHaveBeenCalledWith(mockFile, 'John Doe');
  });
});
```

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

## What NOT to Do - Keep It Simple

### Avoid Over-Engineering
- **Don't add abstractions until we actually need them** - Start with simple, direct implementations
- **Don't build for imaginary future requirements** - Solve today's problems, not hypothetical ones
- **Don't add complex error handling for edge cases that probably won't happen** - Handle common errors well, ignore unlikely scenarios
- **Don't suggest design patterns unless the problem actually requires them** - Prefer simple functions over complex patterns
- **Don't optimize prematurely** - Write clear code first, optimize only when performance becomes an issue
- **Don't add configuration for things that rarely change** - Hard-code reasonable defaults instead of making everything configurable

### Practical Examples
```typescript
// Good: Simple and direct
function uploadDocument(file: File, issuer: string): Promise<Document> {
  return apiClient.post('/documents', { file, issuer });
}

// Bad: Over-engineered with unnecessary abstraction
interface DocumentUploadStrategy {
  upload(file: File, issuer: string): Promise<Document>;
}

class StandardDocumentUploadStrategy implements DocumentUploadStrategy {
  constructor(private client: ApiClient, private validator: Validator) {}
  // ... complex implementation
}
```

```go
// Good: Simple error handling
func SignDocument(doc *Document) error {
    if doc == nil {
        return errors.New("document is required")
    }
    
    signature, err := generateSignature(doc.Hash)
    if err != nil {
        return fmt.Errorf("failed to generate signature: %w", err)
    }
    
    doc.Signature = signature
    return nil
}

// Bad: Over-engineered error handling
type DocumentError struct {
    Code    string
    Message string
    Details map[string]interface{}
}

func (e *DocumentError) Error() string { /* complex implementation */ }

func SignDocument(doc *Document) *DocumentError {
    // Complex validation and error categorization for unlikely edge cases
}
```

### When to Add Complexity
Only add abstractions, patterns, or configuration when:
- You have the same code in 3+ places (Rule of Three)
- You have concrete evidence of a performance problem
- You have actual requirements that demand flexibility
- The simple approach is causing real maintenance pain