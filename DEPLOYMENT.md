# Deployment Guide

This guide covers deploying the Digital Signature System using Docker and Docker Compose.

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- At least 2GB RAM
- At least 5GB disk space

## Quick Start

### 1. Environment Setup

Copy the environment template and configure your settings:

```bash
cp .env.example .env
```

Edit `.env` with your production values:

```bash
# Database Configuration
DB_HOST=qds-postgres
DB_PORT=5432
DB_NAME=digital_signature
DB_USER=your_db_user
DB_PASSWORD=your_secure_password
DB_SSL_MODE=require

# JWT Configuration (use a strong secret in production)
JWT_SECRET=your-very-secure-jwt-secret-key-at-least-32-characters

# RSA Keys for Digital Signatures (generate using make keygen)
PRIVATE_KEY=your-base64-encoded-private-key
PUBLIC_KEY=your-base64-encoded-public-key

# Server Configuration
PORT=8000

# CORS Configuration (comma-separated list of allowed origins)
CORS_ORIGINS=https://your-frontend-domain.com,https://your-api-domain.com

# Frontend Configuration
NEXT_PUBLIC_API_URL=http://localhost:8000
```

### 2. Generate RSA Keys

Generate RSA key pair for digital signatures:

```bash
make keygen
```

This will generate `private_key.pem` and `public_key.pem` files. Convert them to base64 and add to your `.env` file:

```bash
# Convert keys to base64 for environment variables
base64 -w 0 private_key.pem
base64 -w 0 public_key.pem
```

### 3. Development Deployment

Start the development environment:

```bash
make up
```

This will start:
- Frontend on http://localhost:3000
- Backend API on http://localhost:8000
- PostgreSQL database on localhost:5432

### 4. Production Deployment

For production deployment with optimized settings:

```bash
make prod-build
make prod-up
```

## Docker Images

### Backend Image

The backend uses a multi-stage build:

1. **Builder stage**: Uses `golang:1.23-alpine` to compile the Go application
2. **Runtime stage**: Uses `scratch` for minimal image size (~10MB)

Features:
- Static binary compilation
- No runtime dependencies
- Optimized for size and security

### Frontend Image

The frontend uses Next.js standalone output:

1. **Dependencies stage**: Installs production dependencies
2. **Builder stage**: Builds the Next.js application
3. **Runtime stage**: Runs the standalone server

Features:
- Standalone Next.js server
- Optimized for production
- Non-root user for security

## Environment Variables

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_NAME` | Database name | `digital_signature` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `secure_password` |
| `JWT_SECRET` | JWT signing secret | `your-32-char-secret` |
| `PRIVATE_KEY` | Base64 RSA private key | `LS0tLS1CRUdJTi...` |
| `PUBLIC_KEY` | Base64 RSA public key | `LS0tLS1CRUdJTi...` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Backend port | `8000` |
| `DB_HOST` | Database host | `qds-postgres` |
| `DB_PORT` | Database port | `5432` |
| `DB_SSL_MODE` | SSL mode | `disable` (dev), `require` (prod) |
| `CORS_ORIGINS` | Allowed CORS origins (comma-separated) | See config |

## Health Checks

Both services include health check endpoints:

- **Backend**: `GET /health` - Returns `{"status": "ok"}`
- **Frontend**: `GET /api/health` - Returns service status and timestamp

Docker Compose includes health checks with automatic restart policies.

## Backup and Restore

### Database Backup

Create a database backup:

```bash
make backup
```

Backups are stored in the `backups/` directory with timestamp.

### Database Restore

To restore from a backup:

```bash
# List available backups
ls -la backups/

# Restore specific backup
docker-compose exec -T qds-postgres psql -U $DB_USER $DB_NAME < backups/backup_20240101_120000.sql
```

## Monitoring and Logs

### View Logs

```bash
# All services
make logs

# Specific service
make logs-frontend
make logs-backend
make logs-postgres

# Production logs
make prod-logs
```

### Log Files

Application logs are stored in:
- Backend: `backend/internal/infrastructure/handlers/logs/`
  - `app.log` - Application logs
  - `error.log` - Error logs
  - `audit.log` - Audit trail

## Security Considerations

### Production Security Checklist

- [ ] Use strong, unique passwords for database
- [ ] Generate secure JWT secret (32+ characters)
- [ ] Use HTTPS in production (configure reverse proxy)
- [ ] Enable SSL for database connections (`DB_SSL_MODE=require`)
- [ ] Regularly rotate RSA keys
- [ ] Monitor audit logs
- [ ] Keep Docker images updated
- [ ] Use non-root users in containers
- [ ] Implement proper firewall rules

### Key Management

- Store RSA keys securely (environment variables, not files)
- Rotate keys periodically using the key rotation utility
- Backup keys securely before rotation

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check database status
   docker-compose ps qds-postgres
   
   # Check database logs
   make logs-postgres
   ```

2. **Frontend Can't Connect to Backend**
   ```bash
   # Verify backend is running
   curl http://localhost:8000/health
   
   # Check network connectivity
   docker-compose exec qds-frontend ping qds-backend
   ```

3. **RSA Key Issues**
   ```bash
   # Regenerate keys
   make keygen
   
   # Verify key format
   openssl rsa -in private_key.pem -check
   ```

### Debug Mode

Enable debug logging by setting environment:

```bash
# In .env file
ENVIRONMENT=development
```

### Container Resource Usage

Monitor resource usage:

```bash
# View container stats
docker stats

# View container resource limits
docker-compose config
```

## Scaling Considerations

For higher loads, consider:

1. **Database**: Use connection pooling (already configured)
2. **Backend**: Scale horizontally with load balancer
3. **Frontend**: Use CDN for static assets
4. **Storage**: Use external storage for PDF files

## Maintenance

### Regular Maintenance Tasks

1. **Weekly**:
   - Check disk space
   - Review error logs
   - Verify backups

2. **Monthly**:
   - Update Docker images
   - Rotate log files
   - Review audit logs

3. **Quarterly**:
   - Rotate RSA keys
   - Security audit
   - Performance review

### Updates

To update the application:

```bash
# Pull latest code
git pull

# Rebuild images
make prod-build

# Restart services
make prod-down
make prod-up
```

## Support

For issues and questions:

1. Check logs first: `make logs`
2. Verify configuration: `docker-compose config`
3. Test health endpoints: `curl http://localhost:8000/health`
4. Review this documentation

## Performance Tuning

### Database Optimization

```sql
-- Add indexes for better performance (already included in migrations)
CREATE INDEX IF NOT EXISTS idx_documents_user_id ON documents(user_id);
CREATE INDEX IF NOT EXISTS idx_documents_hash ON documents(document_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(session_token);
```

### Docker Optimization

```bash
# Clean up unused resources
docker system prune -f

# Optimize image builds
docker-compose build --no-cache
```

This deployment guide provides a complete setup for both development and production environments with proper security, monitoring, and maintenance procedures.