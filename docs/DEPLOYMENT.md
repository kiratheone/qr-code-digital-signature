# Production Deployment Guide

This comprehensive guide covers the production deployment of the Digital Signature System using Docker and Docker Compose with enterprise-grade security, monitoring, and optimization features.

## Prerequisites

### System Requirements
- **Operating System**: Linux (Ubuntu 20.04+ recommended), macOS, or Windows with WSL2
- **Docker**: 20.10+ with BuildKit enabled
- **Docker Compose**: 2.0+
- **Memory**: Minimum 4GB RAM (8GB+ recommended for production)
- **Storage**: Minimum 20GB free disk space (SSD recommended)
- **Network**: Stable internet connection for image pulls and updates

### Required Tools
- **OpenSSL**: For secret and certificate generation
- **curl**: For health checks and API testing
- **git**: For version control and deployment tracking
- **jq**: For JSON processing (optional but recommended)

### Optional Tools for Enhanced Features
- **Trivy**: For container security scanning
- **Docker Buildx**: For multi-platform builds
- **Prometheus**: For advanced monitoring (included in compose)
- **Grafana**: For visualization (included in compose)

## Quick Start

### 1. Repository Setup
```bash
# Clone the repository
git clone <repository-url>
cd digital-signature-system

# Ensure scripts are executable
chmod +x scripts/*.sh
```

### 2. Environment Configuration
```bash
# Create production environment file
cp .env.prod.example .env.prod

# Edit with your actual values (see Environment Configuration section)
nano .env.prod
```

### 3. Security Setup
```bash
# Generate all required secrets and certificates
./scripts/setup-secrets.sh

# Validate generated secrets
./scripts/setup-secrets.sh validate
```

### 4. Build Optimized Images
```bash
# Build production-optimized Docker images
./scripts/build-optimized.sh

# Or build specific service
./scripts/build-optimized.sh -s backend
```

### 5. Deploy Application
```bash
# Deploy with comprehensive health checks
./scripts/deploy.sh

# Monitor deployment status
./scripts/health-check.sh monitor
```

### 6. Verify Deployment
```bash
# Run comprehensive health checks
./scripts/health-check.sh

# Check application status
./scripts/deploy.sh status
```

## Detailed Setup

### 1. Environment Configuration

Create your production environment file:

```bash
cp .env.prod.example .env.prod
```

Edit `.env.prod` with your actual values:

- `DB_PASSWORD`: Strong database password
- `JWT_SECRET`: Secure JWT signing key
- `NEXT_PUBLIC_API_URL`: Your domain URL
- `REDIS_PASSWORD`: Redis authentication password
- Other configuration values as needed

### 2. Secret Management

The application uses Docker secrets for sensitive data. Run the setup script:

```bash
./scripts/setup-secrets.sh
```

This creates the following secret files in `./secrets/`:
- `db_password.txt`: Database password
- `jwt_secret.txt`: JWT signing secret
- `private_key.pem`: RSA private key for digital signatures
- `public_key.pem`: RSA public key for verification

**Important**: Ensure these files are properly secured and backed up!

### 3. SSL Certificates

For HTTPS support, place your SSL certificates in `./nginx/ssl/`:
- `cert.pem`: SSL certificate
- `key.pem`: SSL private key

For development/testing, you can generate self-signed certificates:

```bash
mkdir -p nginx/ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
```

### 4. Deployment

Deploy the application using the deployment script:

```bash
./scripts/deploy.sh
```

The script will:
- Check requirements
- Setup secrets
- Build Docker images
- Deploy services
- Run health checks
- Clean up unused resources

## Architecture Overview

The production deployment includes:

- **Frontend**: Next.js application (port 3000)
- **Backend**: Go API server (port 8000)
- **Database**: PostgreSQL 16 with persistent storage
- **Cache**: Redis for session and data caching
- **Reverse Proxy**: Nginx for SSL termination and load balancing
- **Monitoring**: Health checks and logging

## Service Configuration

### Resource Limits

Each service has configured resource limits:

- **Frontend**: 512MB RAM, 0.5 CPU
- **Backend**: 1GB RAM, 1.0 CPU
- **Database**: 1GB RAM, 1.0 CPU
- **Redis**: 256MB RAM, 0.25 CPU
- **Nginx**: 128MB RAM, 0.25 CPU

### Health Checks

All services include health checks:
- **Frontend**: HTTP check on `/api/health`
- **Backend**: HTTP check on `/api/health`
- **Database**: PostgreSQL connection check
- **Redis**: Redis ping command

### Logging

Centralized logging configuration:
- Log rotation: 10MB max size, 3 files
- JSON format for structured logging
- Separate access and error logs for Nginx

## Security Features

### Network Security
- Isolated Docker network
- Internal service communication
- Nginx reverse proxy with security headers

### Data Security
- Docker secrets for sensitive data
- Encrypted database connections (SSL)
- HTTPS-only communication
- Rate limiting on API endpoints

### Application Security
- Non-root containers
- Read-only file systems where possible
- Security headers (CSP, HSTS, etc.)
- Input validation and sanitization

## Monitoring and Maintenance

### Health Monitoring

Check application health:
```bash
./scripts/deploy.sh health
```

### Service Status

View service status:
```bash
./scripts/deploy.sh status
```

### Logs

View application logs:
```bash
# All services
./scripts/deploy.sh logs

# Specific service
./scripts/deploy.sh logs backend
```

### Backup

Database backup is handled automatically. Manual backup:
```bash
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U postgres digital_signature > backup.sql
```

## Troubleshooting

### Common Issues

1. **Services not starting**
   - Check Docker daemon is running
   - Verify environment variables in `.env.prod`
   - Check secret files exist and have correct permissions

2. **Health checks failing**
   - Wait for services to fully initialize (30-60 seconds)
   - Check service logs for errors
   - Verify network connectivity between services

3. **SSL/HTTPS issues**
   - Ensure SSL certificates are valid and properly placed
   - Check certificate permissions (readable by nginx)
   - Verify domain name matches certificate

4. **Database connection issues**
   - Check database credentials in secrets
   - Verify PostgreSQL is healthy
   - Check network connectivity

### Debug Commands

```bash
# Check service logs
docker-compose -f docker-compose.prod.yml logs [service-name]

# Execute commands in containers
docker-compose -f docker-compose.prod.yml exec [service-name] [command]

# Check container resource usage
docker stats

# Inspect service configuration
docker-compose -f docker-compose.prod.yml config
```

## Scaling

### Horizontal Scaling

To scale services:
```bash
docker-compose -f docker-compose.prod.yml up -d --scale backend=3
```

### Load Balancing

Nginx is configured to load balance between multiple backend instances automatically.

## Updates and Maintenance

### Application Updates

1. Pull latest code
2. Rebuild images: `docker-compose -f docker-compose.prod.yml build`
3. Deploy: `./scripts/deploy.sh`

### Database Migrations

Migrations run automatically on backend startup. For manual migration:
```bash
docker-compose -f docker-compose.prod.yml exec backend ./main -migrate
```

### Key Rotation

For security, rotate keys periodically:
1. Generate new keys: `./scripts/setup-secrets.sh`
2. Update secrets
3. Restart services: `./scripts/deploy.sh restart`

## Performance Optimization

### Database Optimization
- Connection pooling configured
- Indexes on frequently queried columns
- Regular VACUUM and ANALYZE operations

### Caching Strategy
- Redis for session storage
- Application-level caching for frequently accessed data
- Nginx caching for static assets

### Resource Monitoring
- Monitor CPU and memory usage
- Database query performance
- Response times and error rates

## Security Checklist

- [ ] Strong passwords for all services
- [ ] SSL certificates properly configured
- [ ] Security headers enabled
- [ ] Rate limiting configured
- [ ] Regular security updates
- [ ] Backup and recovery tested
- [ ] Access logs monitored
- [ ] Secrets properly secured

## Support

For issues and questions:
1. Check this documentation
2. Review service logs
3. Check GitHub issues
4. Contact system administrator

## License

This deployment configuration is part of the Digital Signature System project.