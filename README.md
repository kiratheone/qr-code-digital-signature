# Digital Signature System

A simple and secure QR Code-based digital signature system for PDF documents. Built with Go and Next.js, designed for small applications with minimal complexity and easy maintenance.

## 🚀 Features

### Core Functionality
- **Digital Document Signing**: Upload PDFs and generate RSA-2048 digital signatures with SHA-256 hashing
- **QR Code Integration**: Automatic QR code generation and PDF injection for easy verification
- **Document Verification**: Verify document authenticity by scanning QR codes or uploading documents
- **Document Management**: Upload, list, view, and delete signed documents
- **User Authentication**: Secure JWT-based authentication with refresh tokens
- **Audit Logging**: Complete audit trail for compliance and security monitoring

### Security Features
- RSA-2048 digital signatures with SHA-256 hashing
- JWT authentication with secure token management
- Input validation and sanitization at all layers
- Rate limiting and abuse prevention
- Secure key management through environment variables

## 🛠 Technology Stack

### Backend
- **Language**: Go 1.23+
- **Framework**: Gin web framework
- **Database**: PostgreSQL 16 with GORM ORM
- **Authentication**: JWT with golang-jwt/jwt/v5
- **PDF Processing**: UniPDF (unidoc/unipdf/v3)
- **QR Codes**: go-qrcode library
- **Cryptography**: RSA-2048 with SHA-256, golang.org/x/crypto
- **Testing**: Go standard testing + testify

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
- **SSL**: Self-signed certificates for development
- **Logging**: File-based logging with rotation

## 📋 Prerequisites

- Docker 20.10+ and Docker Compose 2.0+
- Node.js 18+ (for local development)
- Go 1.23+ (for local development)
- At least 2GB RAM and 5GB disk space

## 🚀 Getting Started

### 1. Clone the Repository

```bash
git clone <repository-url>
cd digital-signature-system
```

### 2. Environment Setup

Copy the environment template and configure your settings:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=digital_signature
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_SSL_MODE=disable

# JWT Configuration
JWT_SECRET=your-very-secure-jwt-secret-key-at-least-32-characters

# Server Configuration
PORT=8000

# CORS Configuration
CORS_ORIGINS=http://localhost:3000,http://localhost:8065

# Frontend Configuration
NEXT_PUBLIC_API_URL=http://localhost:8000
```

### 3. Generate RSA Keys

Generate RSA key pair for digital signatures:

```bash
make keygen
```

This creates `private_key.pem` and `public_key.pem` files. Convert them to base64 and add to your `.env`:

```bash
# Convert keys to base64 for environment variables
base64 -w 0 private_key.pem  # Add to PRIVATE_KEY in .env
base64 -w 0 public_key.pem   # Add to PUBLIC_KEY in .env
```

### 4. Start the Application

Using Docker (recommended):

```bash
make up
```

This starts:
- Frontend: http://localhost:8065
- Backend API: http://localhost:8112
- PostgreSQL: localhost:5434

## 💻 Development

### Docker Development (Recommended)

```bash
# Start all services
make up

# View logs
make logs

# Stop services
make down

# Clean up
make clean
```

### Local Development

#### Backend Development

```bash
cd backend
go mod tidy
go run cmd/main.go

# Run tests
go test ./...
go test -race ./...
go test -bench=. ./...
```

#### Frontend Development

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

### Available Make Commands

```bash
# Development
make up          # Start all services with Docker Compose
make down        # Stop all services
make build       # Build all Docker images
make logs        # View logs from all services
make clean       # Clean up containers and volumes
make dev-backend # Run backend in development mode
make dev-frontend# Run frontend in development mode
make test        # Run all tests
make keygen      # Generate RSA key pair

# Production
make prod-up     # Start services in production mode
make prod-down   # Stop production services
make prod-build  # Build production images
make prod-logs   # View production logs

# Maintenance
make backup      # Create database backup
make health      # Run system health checks
```

## 🚀 Deployment

### Production Deployment

1. **Environment Setup**:
   ```bash
   cp .env.example .env
   # Configure production values in .env
   ```

2. **Generate Production Keys**:
   ```bash
   make keygen
   # Add base64-encoded keys to .env
   ```

3. **Deploy**:
   ```bash
   make prod-build
   make prod-up
   ```

### Environment Variables

#### Required Variables
| Variable | Description | Example |
|----------|-------------|---------|
| `DB_NAME` | Database name | `digital_signature` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `secure_password` |
| `JWT_SECRET` | JWT signing secret | `your-32-char-secret` |
| `PRIVATE_KEY` | Base64 RSA private key | `LS0tLS1CRUdJTi...` |
| `PUBLIC_KEY` | Base64 RSA public key | `LS0tLS1CRUdJTi...` |

#### Optional Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Backend port | `8000` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_SSL_MODE` | SSL mode | `disable` |
| `CORS_ORIGINS` | Allowed origins | See .env.example |

### Health Checks

- **Backend**: `GET /health` - Returns `{"status": "ok"}`
- **Frontend**: `GET /api/health` - Returns service status

### Backup and Restore

```bash
# Create backup
make backup

# List backups
ls -la backups/

# Restore backup
docker-compose exec -T qds-postgres psql -U $DB_USER $DB_NAME < backups/backup_file.sql
```

## 🏗 Project Structure

```
├── backend/                 # Go backend application
│   ├── cmd/                # Application entry points
│   ├── internal/           # Internal packages
│   │   ├── config/         # Configuration management
│   │   ├── domain/         # Business logic layer
│   │   │   ├── entities/   # Domain models
│   │   │   ├── repositories/ # Repository interfaces
│   │   │   └── services/   # Business logic services
│   │   └── infrastructure/ # External concerns
│   │       ├── database/   # Database connection & repositories
│   │       ├── handlers/   # HTTP handlers & middleware
│   │       ├── crypto/     # Cryptographic services
│   │       └── pdf/        # PDF processing services
│   └── Dockerfile         # Backend container
├── frontend/               # Next.js frontend application
│   ├── src/
│   │   ├── app/           # Next.js 15+ app router
│   │   ├── lib/           # Core business logic
│   │   │   ├── services/  # Business logic services
│   │   │   ├── api/       # API client
│   │   │   ├── types/     # TypeScript definitions
│   │   │   └── utils/     # Utility functions
│   │   ├── components/    # React components
│   │   └── hooks/         # Custom React hooks
│   └── Dockerfile         # Frontend container
├── docker-compose.yml     # Development environment
├── docker-compose.prod.yml # Production environment
├── Makefile              # Development scripts
└── scripts/              # Utility scripts
```

## 🧪 Testing

### Backend Testing

```bash
cd backend
go test ./...                    # Run all tests
go test -race ./...             # Run with race detection
go test -bench=. ./...          # Run benchmarks
go test -coverage ./...         # Run with coverage
```

### Frontend Testing

```bash
cd frontend
npm test                        # Run tests
npm run test:coverage          # Run with coverage
npm run test:watch             # Watch mode
```

## 🔒 Security

### Security Features
- RSA-2048 digital signatures with SHA-256 hashing
- JWT authentication with secure token management
- Input validation and sanitization
- Rate limiting and abuse prevention
- Secure key management
- HTTPS support (configure reverse proxy)
- Non-root container users

### Security Checklist
- [ ] Use strong, unique database passwords
- [ ] Generate secure JWT secrets (32+ characters)
- [ ] Enable HTTPS in production
- [ ] Use SSL for database connections in production
- [ ] Regularly rotate RSA keys
- [ ] Monitor audit logs
- [ ] Keep dependencies updated

## 📊 Performance

### Specifications
- Handles 50-100 concurrent users
- Processes 500+ documents per day
- Database connection pooling
- Optimized Docker images
- Efficient PDF processing

### Monitoring

```bash
# View container stats
docker stats

# Check health endpoints
curl http://localhost:8000/health

# View application logs
make logs
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Follow the coding style guide in `.kiro/steering/coding-style.md`
- Write tests for new features
- Update documentation as needed
- Use conventional commit messages

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Troubleshooting

1. **Database Connection Issues**:
   ```bash
   make logs-postgres
   docker-compose ps qds-postgres
   ```

2. **Frontend Can't Connect to Backend**:
   ```bash
   curl http://localhost:8000/health
   make logs-backend
   ```

3. **RSA Key Issues**:
   ```bash
   make keygen
   openssl rsa -in private_key.pem -check
   ```

### Getting Help

- Check the logs: `make logs`
- Verify configuration: `docker-compose config`
- Test health endpoints: `curl http://localhost:8000/health`
- Review documentation in `DEPLOYMENT.md` and `OPERATIONS.md`

## 📈 Roadmap

- [ ] Multi-language support
- [ ] Advanced audit reporting
- [ ] Bulk document processing
- [ ] API rate limiting dashboard
- [ ] Document templates
- [ ] Integration with external storage providers

---

**Built with ❤️ using Go and Next.js**