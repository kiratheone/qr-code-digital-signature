---
applyTo: "**"
---
# Project Structure & Architecture

## Clean Architecture Pattern

The backend follows Clean Architecture with clear separation of concerns:

### Backend Structure (`backend/`) - Simple Clean Architecture
```
internal/
├── config/           # Configuration management
├── domain/           # Business logic layer
│   ├── entities/     # Domain models (User, Document, Session)
│   ├── repositories/ # Repository interfaces
│   └── services/     # Business logic services (all business logic here)
├── infrastructure/   # External concerns
│   ├── database/     # DB connection, GORM auto-migrate, repository implementations
│   ├── handlers/     # HTTP handlers, middleware, routes
│   ├── crypto/       # Cryptographic services
│   └── pdf/          # PDF processing services
cmd/
├── main.go          # Application entry point
└── keygen/          # Key generation utility
```

### Frontend Structure (`frontend/src/`) - Simple Clean Architecture
```
app/                 # Next.js 15+ app router
├── (auth)/         # Authentication pages (login, register)
├── documents/      # Document management pages
├── verify/         # Document verification pages
├── layout.tsx      # Root layout
└── globals.css     # Global styles

lib/                # Core business logic layer
├── services/       # Business logic services (document, auth, verification)
├── api/           # API client and HTTP calls
├── types/         # TypeScript type definitions
└── utils/         # Utility functions

components/         # Presentation layer
├── forms/         # Form components (upload, login)
├── ui/            # Generic UI components (button, modal, etc)
└── features/      # Feature-specific components
   ├── document/   # Document-related components
   ├── auth/       # Authentication components
   └── verify/     # Verification components

hooks/             # Custom React hooks (data fetching, state management)
```

## Key Architecture Patterns

### Backend Patterns - Simplified
- **Simple DI**: Constructor injection without complex DI container
- **Repository Pattern**: Abstract data access in `domain/repositories/`
- **Service Layer**: Combined business logic in `domain/services/` (no separate usecases)
- **Middleware Chain**: Authentication, validation, logging, rate limiting

### Frontend Patterns - Simple Clean Architecture
- **Service Layer**: Business logic in `lib/services/` (document, auth, verification services)
- **API Layer**: HTTP calls abstracted in `lib/api/` with simple error handling
- **Component Layer**: Presentation components in `components/` with clear separation
- **Custom Hooks**: Data fetching and state management in `hooks/`
- **Simple State**: React Query for server state, React Context for global client state

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
## Simple Application Design

### Design Principles
- **Keep It Simple**: Single service, auto-migrate database, minimal dependencies
- **Easy Maintenance**: Clear code structure, focused testing, simple deployment
- **Sufficient Performance**: Handle 50-100 concurrent users, 500 documents/day

### Deployment Strategy
- **Single Binary**: One Go binary with embedded frontend assets
- **Docker Compose**: Simple 2-container setup (app + database)
- **Local Storage**: Store PDFs on local filesystem (no external storage needed)
- **Basic Monitoring**: File-based logging with log rotation