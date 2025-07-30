# Project Structure & Architecture

## Clean Architecture Pattern

The backend follows Clean Architecture with clear separation of concerns:

### Backend Structure (`backend/`)
```
internal/
├── config/           # Configuration management
├── domain/           # Business logic layer (entities, repositories, services, usecases)
│   ├── entities/     # Domain models (User, Document, Session, VerificationLog)
│   ├── repositories/ # Repository interfaces
│   ├── services/     # Domain service interfaces
│   └── usecases/     # Business logic orchestration
├── infrastructure/   # External concerns
│   ├── database/     # DB connection, migrations
│   ├── di/           # Dependency injection container
│   ├── repositories/impl/ # Repository implementations
│   ├── server/       # HTTP handlers, middleware, routes
│   ├── services/     # Service implementations
│   └── validation/   # Input validation
cmd/
├── main.go          # Application entry point
└── keygen/          # Key generation utility
```

### Frontend Structure (`frontend/src/`)
```
app/                 # Next.js 13+ app router
├── api/            # API route handlers
├── documents/      # Document management pages
├── verify/         # Document verification pages
└── globals.css     # Global styles

components/         # Reusable UI components
├── DocumentManagement/  # Document CRUD operations
├── DocumentUpload/      # File upload functionality
├── ErrorBoundary/       # Error handling components
├── Navigation/          # Navigation components
├── UI/                  # Generic UI components
└── Verification/        # Document verification UI

api/               # API client layer
hooks/             # Custom React hooks
providers/         # React context providers
types/             # TypeScript type definitions
utils/             # Utility functions
```

## Key Architecture Patterns

### Backend Patterns
- **Dependency Injection**: Container pattern in `infrastructure/di/`
- **Repository Pattern**: Abstract data access in `domain/repositories/`
- **Use Case Pattern**: Business logic orchestration in `domain/usecases/`
- **Service Layer**: Domain services for complex operations
- **Middleware Chain**: Authentication, validation, logging, rate limiting

### Frontend Patterns
- **Component Composition**: Reusable UI components with clear responsibilities
- **Custom Hooks**: Business logic abstraction (`useDocumentOperations`, `useVerificationOperations`)
- **Error Boundaries**: Graceful error handling at component level
- **API Client**: Centralized HTTP client with interceptors for auth/retry logic

## File Naming Conventions

### Backend (Go)
- Files: `snake_case.go`
- Packages: lowercase, single word when possible
- Interfaces: `PascalCase` with descriptive names
- Implementations: `PascalCase` often ending with `Impl`

### Frontend (TypeScript/React)
- Components: `PascalCase.tsx`
- Hooks: `camelCase.ts` starting with `use`
- Utilities: `camelCase.ts`
- Types: `camelCase.ts`
- Tests: `*.test.tsx` or `*.test.ts`

## Testing Structure
- Backend: Tests alongside source files (`*_test.go`)
- Frontend: Tests in `__tests__` directories or alongside components
- Integration tests: Separate files for cross-layer testing
- Test data: Centralized in `test_data.go` files

## Configuration Management
- Environment variables via `.env` files
- Backend config centralized in `internal/config/`
- Frontend config via Next.js environment variables
- Secrets managed through Docker secrets or environment injection