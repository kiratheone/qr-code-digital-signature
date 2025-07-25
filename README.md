# Digital Signature System

A QR Code-based digital signature system for PDF documents that enables automatic validation of document authenticity.

## Features

- 🔐 Digital document signing with SHA-256 hash
- 📄 QR Code generation and PDF injection
- 📦 Document management and verification
- 🔍 Document verification through QR codes
- ✅ Comprehensive validation system
- 🛡️ Secure authentication and key management
- 🐳 Docker-ready deployment

## Technology Stack

- **Frontend**: Next.js 15 with TypeScript, Tailwind CSS
- **Backend**: Golang 1.22+ with Gin framework
- **Database**: PostgreSQL 16
- **Containerization**: Docker & Docker Compose

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Node.js 20+ (for local development)
- Go 1.22+ (for local development)

### Using Docker Compose

1. Clone the repository
2. Copy environment file:
   ```bash
   cp .env.example .env
   ```
3. Generate RSA key pair for digital signatures:
   ```bash
   # Generate private key
   openssl genrsa -out private_key.pem 2048
   
   # Generate public key
   openssl rsa -in private_key.pem -pubout -out public_key.pem
   
   # Add keys to .env file (base64 encoded or as single line)
   ```
4. Start the application:
   ```bash
   docker-compose up -d
   ```

The application will be available at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8000
- Database: localhost:5432

### Local Development

#### Backend

```bash
cd backend
go mod tidy
go run cmd/main.go
```

#### Frontend

```bash
cd frontend
npm install
npm run dev
```

## API Endpoints

### Authentication
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `POST /api/auth/register` - User registration

### Documents
- `POST /api/documents/sign` - Sign a document
- `GET /api/documents` - List user documents
- `GET /api/documents/:id` - Get document details
- `DELETE /api/documents/:id` - Delete document

### Verification
- `GET /api/verify/:docId` - Get verification info
- `POST /api/verify/:docId/upload` - Verify document

## Architecture

The system follows Clean Architecture principles with clear separation of concerns:

- **Presentation Layer**: HTTP handlers and frontend components
- **Business Logic Layer**: Use cases and domain services
- **Data Layer**: Repository interfaces and database models
- **Infrastructure Layer**: External services and configurations

## Security

- RSA-2048 digital signatures with SHA-256 hashing
- Secure key management through environment variables
- JWT-based authentication
- Input validation and sanitization
- Audit logging for all operations

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.