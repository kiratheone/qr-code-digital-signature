# Implementation Plan - Simple Clean Architecture

- [x] 1. Setup project structure and development environment
  - Create simple clean architecture: domain/services, infrastructure/handlers, infrastructure/database
  - Initialize Go modules (1.23+) and Next.js 15 project with TypeScript
  - Setup Docker and docker-compose with PostgreSQL 16
  - Configure environment variables and simple development scripts
  - _Requirements: 8.1, 8.2, 8.3, 8.4_

- [x] 2. Implement database schema and models - Simple approach
- [x] 2.1 Create PostgreSQL database schema with GORM auto-migrate
  - Create User, Session, Document entities with GORM tags
  - Use GORM auto-migrate instead of separate migration files
  - Create basic database indexes for user_id and document_hash
  - Setup simple database connection
  - _Requirements: 3.2, 5.4, 8.1_

- [x] 2.2 Implement simple repository layer
  - Create repository interfaces in domain/repositories
  - Implement concrete repositories in infrastructure/database
  - Use simple constructor injection (no complex DI container)
  - Write focused unit tests for business-critical operations only
  - _Requirements: 3.1, 3.2, 3.4, 8.1_

- [x] 3. Implement core cryptographic services
- [x] 3.1 Create digital signature service
  - Implement RSA key pair generation and management
  - Create document hash calculation using SHA-256
  - Implement digital signature creation and verification functions
  - Write comprehensive unit tests for cryptographic operations
  - _Requirements: 1.1, 1.2, 1.3, 6.1, 6.3_

- [x] 3.2 Implement secure key management
  - Create secure key storage using environment variables
  - Implement key loading and validation mechanisms
  - Add key rotation capability for future security updates
  - Write tests for key management security
  - _Requirements: 6.1, 6.4_

- [x] 4. Implement simple authentication system
- [x] 4.1 Create authentication service (combined business logic)
  - Implement AuthService with login, logout, and session validation
  - Create simple password hashing and validation utilities
  - Implement basic JWT authentication (no refresh tokens for simplicity)
  - Add basic user registration functionality
  - Write unit tests for authentication business logic only
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6_

- [x] 4.2 Implement authentication handlers and middleware
  - Create simple authentication middleware for protected routes
  - Implement login and logout API handlers
  - Add basic session validation
  - Create simple user management endpoints
  - Write integration tests for critical authentication flows only
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.6_

- [x] 5. Develop PDF processing capabilities
- [x] 5.1 Implement PDF hash calculation service
  - Create PDF file validation and processing functions
  - Implement SHA-256 hash calculation for PDF documents
  - Add support for various PDF formats and sizes up to 50MB
  - Write unit tests for PDF processing with different file types
  - _Requirements: 1.1, 7.1_

- [x] 5.2 Implement QR code generation and injection
  - Create QR code generation service with document metadata
  - Implement QR code injection into PDF at specified positions
  - Add support for default (last page) and custom positioning
  - Ensure PDF quality and format preservation after QR injection
  - Write tests for QR code generation and PDF injection
  - _Requirements: 1.4, 2.1, 2.2, 2.4_

- [x] 6. Build backend API endpoints - Simple service layer
- [x] 6.1 Implement document service and handlers
  - Create DocumentService with all document business logic (no separate use cases)
  - Implement DocumentHandler with POST /api/documents/sign endpoint
  - Add simple separation: handlers → services → repositories
  - Use simple constructor injection for dependencies
  - Write unit tests for service business logic only
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.3, 7.1_

- [x] 6.2 Implement document management service and handlers
  - Create document management logic in DocumentService
  - Implement handlers for GET /api/documents (simple pagination, no search)
  - Add GET /api/documents/:id and DELETE /api/documents/:id handlers
  - Keep simple separation of concerns between layers
  - Write unit tests for business-critical service methods only
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 6.3 Implement verification service and handlers
  - Create VerificationService with document verification business logic
  - Implement handlers for GET /api/verify/:docId and POST /api/verify/:docId/upload
  - Keep verification logic in service layer
  - Add simple verification result responses with clear status messages
  - Write unit tests for verification business logic only
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 7. Develop frontend with simple clean architecture
- [x] 7.1 Setup frontend clean architecture structure
  - Create folder structure: lib/services/, lib/api/, lib/types/, components/, hooks/
  - Setup TypeScript path aliases for clean imports
  - Create base ApiClient class with simple error handling
  - Setup React Query provider and basic configuration
  - _Requirements: Clean architecture foundation_

- [x] 7.2 Create frontend service layer
  - Create DocumentService in lib/services/ with all document business logic
  - Create AuthService in lib/services/ with authentication logic
  - Create VerificationService in lib/services/ with verification logic
  - Define TypeScript interfaces in lib/types/
  - Write unit tests for service layer business logic
  - _Requirements: 1.5, 3.1, 4.1, 5.1_

- [x] 7.3 Build custom hooks layer
  - Create useDocumentOperations hook wrapping React Query
  - Create useAuthOperations hook for authentication state
  - Create useVerificationOperations hook for verification flow
  - Handle loading states, error states, and data fetching in hooks
  - Write tests for critical hook behaviors
  - _Requirements: 7.2, 7.3, 8.1_

- [x] 7.4 Implement presentation components
  - Create DocumentUploadForm component (presentation only)
  - Create DocumentList component with simple UI
  - Create VerificationForm and VerificationResult components
  - Create simple UI components in components/ui/
  - Write tests for critical user interactions only
  - _Requirements: 1.5, 3.2, 4.2, 5.2, 7.4_

- [x] 7.5 Build Next.js pages with clean separation
  - Create app/documents/page.tsx using hooks and components
  - Create app/verify/[docId]/page.tsx for verification flow
  - Create app/(auth)/login/page.tsx for authentication
  - Ensure pages only orchestrate hooks and components
  - Write integration tests for critical user flows
  - _Requirements: 7.3, 8.2_

- [x] 8. Implement simple API integration and state management
- [x] 8.1 Setup simple API client and React Query
  - Create simple ApiClient class in lib/api/ with fetch
  - Implement React Query for basic caching and state management
  - Add simple error handling (no complex retry logic)
  - Create custom hooks that wrap React Query
  - Write tests for service layer (skip API client tests)
  - _Requirements: 7.2, 7.5_

- [x] 8.2 Implement simple frontend routing
  - Setup Next.js App Router for all pages
  - Create simple navigation components
  - Implement basic protected routes with authentication
  - Add simple loading states and error boundaries
  - Write tests for critical user flows only
  - _Requirements: 4.1, 7.3_

- [x] 9. Add simple error handling and logging
- [x] 9.1 Implement basic backend error handling
  - Create simple standardized error response format
  - Add basic file-based logging (no correlation IDs)
  - Implement basic request validation and sanitization
  - Add simple security middleware
  - Write tests for critical error scenarios only
  - _Requirements: 2.5, 5.5, 6.5, 7.5_

- [x] 9.2 Implement simple frontend error handling
  - Create basic error boundary components
  - Add simple user-friendly error messages
  - Implement basic loading states in custom hooks
  - Add simple form validation
  - Write tests for critical error scenarios only
  - _Requirements: 7.3, 7.5_

- [ ] 10. Implement basic security measures
- [x] 10.1 Add basic input validation and sanitization
  - Implement basic input validation for critical API endpoints
  - Add file type and size validation for uploads
  - Create simple sanitization for user inputs
  - Add basic security headers
  - Write tests for critical validation scenarios only
  - _Requirements: 6.4, 6.5_

- [x] 10.2 Implement basic audit logging
  - Create simple audit logging for document operations
  - Implement basic verification attempt logging
  - Add simple file-based logging with basic rotation
  - Write tests for critical audit scenarios only
  - _Requirements: 6.5_

- [ ] 11. Basic performance and focused testing
- [ ] 11.1 Implement basic performance optimizations
  - Add simple database connection pooling (max 10 connections)
  - Implement basic file streaming for PDF processing
  - Add simple frontend lazy loading for verification page
  - Skip complex caching and performance testing
  - _Requirements: 7.1, 7.4_

- [ ] 11.2 Create focused test suite (70% coverage target)
  - Write unit tests for business logic in services only
  - Create integration tests for critical API endpoints only
  - Focus on cryptographic operations, signing, and verification
  - Skip comprehensive end-to-end testing
  - Setup simple testing pipeline
  - _Requirements: Critical business logic validation only_

- [ ] 12. Setup simple deployment
- [ ] 12.1 Configure simple Docker setup
  - Create basic Docker images for frontend and backend
  - Setup simple multi-stage builds
  - Configure basic environment variable management
  - Write simple deployment documentation
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 12.2 Setup basic maintenance
  - Implement simple file-based logging
  - Create basic database backup procedures
  - Write simple operational documentation
  - Skip complex monitoring and alerting
  - _Requirements: 8.5_