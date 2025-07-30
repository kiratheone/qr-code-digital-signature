# Technology Stack & Build System

## Tech Stack

### Backend
- **Language**: Go 1.23+ 
- **Framework**: Gin web framework
- **Database**: PostgreSQL 16 with GORM ORM
- **Authentication**: JWT with golang-jwt/jwt/v5
- **PDF Processing**: UniPDF (unidoc/unipdf/v3)
- **QR Codes**: go-qrcode library
- **Cryptography**: RSA-2048 with SHA-256, golang.org/x/crypto
- **Migrations**: golang-migrate/migrate/v4
- **Testing**: testify for unit tests

### Frontend
- **Framework**: Next.js 15 with TypeScript
- **Styling**: Tailwind CSS
- **State Management**: TanStack React Query v5
- **HTTP Client**: Axios with custom interceptors
- **Forms**: React Hook Form
- **File Upload**: React Dropzone
- **Testing**: Jest + React Testing Library
- **Date Handling**: date-fns

### Infrastructure
- **Containerization**: Docker & Docker Compose
- **Database**: PostgreSQL 16 Alpine
- **Reverse Proxy**: Nginx (production)
- **Monitoring**: Prometheus, Grafana, Loki
- **SSL**: Self-signed certificates for development

## Build Commands

### Docker (Recommended)
```bash
# Start full stack
make up
# or
docker-compose up -d

# Build images
make build

# View logs
make logs

# Stop services
make down

# Clean up
make clean
```

### Backend Development
```bash
cd backend
go mod tidy
go run cmd/main.go

# Run tests
go test ./...
go test -race ./...
go test -bench=. ./...
```

### Frontend Development
```bash
cd frontend
npm install
npm run dev

# Run tests
npm test
npm run test:coverage
npm run lint
npm run type-check
npm run build
```

### Key Generation (Required Setup)
```bash
# Generate RSA key pair for signatures
openssl genrsa -out private_key.pem 2048
openssl rsa -in private_key.pem -pubout -out public_key.pem
```

## Environment Configuration
- Copy `.env.example` to `.env` and configure
- Required: Database credentials, RSA keys, JWT secret
- Frontend API URL: `NEXT_PUBLIC_API_URL`
- Database connection via `DB_*` variables