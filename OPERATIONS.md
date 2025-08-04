# Operations Guide

This guide covers day-to-day operations, maintenance, and troubleshooting for the Digital Signature System.

## Table of Contents

1. [System Monitoring](#system-monitoring)
2. [Log Management](#log-management)
3. [Database Maintenance](#database-maintenance)
4. [Security Operations](#security-operations)
5. [Backup and Recovery](#backup-and-recovery)
6. [Troubleshooting](#troubleshooting)
7. [Performance Monitoring](#performance-monitoring)
8. [Maintenance Schedule](#maintenance-schedule)

## System Monitoring

### Health Checks

The system provides health check endpoints for monitoring:

```bash
# Backend health check
curl http://localhost:8000/health

# Frontend health check
curl http://localhost:3000/api/health

# Database health check (via Docker)
docker-compose exec qds-postgres pg_isready -U $DB_USER
```

### Service Status

Check service status using Docker Compose:

```bash
# View all services
docker-compose ps

# View service logs
make logs

# View specific service logs
make logs-backend
make logs-frontend
make logs-postgres
```

### Resource Monitoring

Monitor system resources:

```bash
# Container resource usage
docker stats

# Disk usage
df -h

# Log directory size
du -sh backend/internal/infrastructure/handlers/logs/

# Database size
docker-compose exec qds-postgres psql -U $DB_USER -d $DB_NAME -c "SELECT pg_size_pretty(pg_database_size('$DB_NAME'));"
```

## Log Management

### Log Files

The system maintains three types of logs:

1. **Application Logs** (`logs/app.log`)
   - General application events
   - Info, debug, and warning messages
   - Performance metrics

2. **Error Logs** (`logs/error.log`)
   - Error messages and stack traces
   - System failures
   - Critical issues

3. **Audit Logs** (`logs/audit.log`)
   - User authentication events
   - Document operations
   - Security events
   - Verification attempts

### Log Rotation

Automatic log rotation is implemented:

- **Trigger**: When log files exceed 10MB
- **Retention**: Keep last 5 rotated files
- **Format**: `logfile.log.YYYY-MM-DD-HH-MM-SS`

Manual log rotation:

```bash
# Rotate all logs
./scripts/rotate-logs.sh

# Check log sizes
ls -lh backend/internal/infrastructure/handlers/logs/
```

### Log Analysis

Common log analysis commands:

```bash
# View recent errors
tail -f backend/internal/infrastructure/handlers/logs/error.log

# Search for specific events
grep "DOCUMENT_SIGN" backend/internal/infrastructure/handlers/logs/audit.log

# Count authentication failures
grep "AUTH_FAILURE" backend/internal/infrastructure/handlers/logs/audit.log | wc -l

# View verification attempts
grep "VERIFICATION_ATTEMPT" backend/internal/infrastructure/handlers/logs/audit.log | tail -10
```

### Log Monitoring Alerts

Set up monitoring for critical events:

```bash
# Monitor for authentication failures
tail -f logs/audit.log | grep "AUTH_FAILURE"

# Monitor for system errors
tail -f logs/error.log | grep "ERROR"

# Monitor for suspicious activity
tail -f logs/audit.log | grep "SUSPICIOUS_ACTIVITY"
```

## Database Maintenance

### Database Backups

Create regular backups:

```bash
# Create backup
make backup

# List available backups
ls -la backups/

# Backup with custom name
docker-compose exec qds-postgres pg_dump -U $DB_USER $DB_NAME > backups/manual_backup_$(date +%Y%m%d).sql
```

### Database Restore

Restore from backup:

```bash
# Stop services
make down

# Start only database
docker-compose up -d qds-postgres

# Wait for database to be ready
sleep 10

# Restore backup
docker-compose exec -T qds-postgres psql -U $DB_USER $DB_NAME < backups/backup_20240101_120000.sql

# Start all services
make up
```

### Database Maintenance

Regular maintenance tasks:

```bash
# Connect to database
docker-compose exec qds-postgres psql -U $DB_USER $DB_NAME

# Check database size
SELECT pg_size_pretty(pg_database_size(current_database()));

# Check table sizes
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

# Vacuum and analyze
VACUUM ANALYZE;

# Check index usage
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation
FROM pg_stats
WHERE schemaname = 'public';
```

## Security Operations

### Key Management

Monitor and rotate RSA keys:

```bash
# Check current key status
cd backend && go run cmd/keygen/main.go -check

# Generate new key pair
make keygen

# Rotate keys (requires service restart)
cd backend && go run cmd/keyrotate/main.go
```

### Security Monitoring

Monitor security events:

```bash
# Authentication failures
grep "AUTH_FAILURE" logs/audit.log | tail -20

# Rate limit violations
grep "RATE_LIMIT_EXCEEDED" logs/audit.log | tail -10

# Suspicious activities
grep "SUSPICIOUS_ACTIVITY" logs/audit.log

# Failed verification attempts
grep "VERIFICATION_FAILURE" logs/audit.log | tail -10
```

### Access Control

Review user access:

```bash
# Connect to database
docker-compose exec qds-postgres psql -U $DB_USER $DB_NAME

# List all users
SELECT id, username, email, role, created_at, is_active FROM users;

# List active sessions
SELECT s.id, u.username, s.created_at, s.last_accessed 
FROM sessions s 
JOIN users u ON s.user_id = u.id 
WHERE s.expires_at > NOW();

# Deactivate user
UPDATE users SET is_active = false WHERE username = 'suspicious_user';
```

## Backup and Recovery

### Automated Backup Strategy

1. **Daily Backups**: Automated via cron job
2. **Weekly Full Backups**: Complete system backup
3. **Monthly Archive**: Long-term storage

### Backup Script Setup

Create cron job for automated backups:

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /path/to/qr-code-digital-signature/scripts/backup.sh

# Add weekly log rotation at 3 AM on Sundays
0 3 * * 0 /path/to/qr-code-digital-signature/scripts/rotate-logs.sh
```

### Disaster Recovery

Complete system recovery procedure:

```bash
# 1. Stop all services
make down

# 2. Restore database
docker-compose up -d qds-postgres
sleep 10
docker-compose exec -T qds-postgres psql -U $DB_USER $DB_NAME < backups/latest_backup.sql

# 3. Restore configuration
cp .env.backup .env

# 4. Restore RSA keys
cp private_key.pem.backup private_key.pem
cp public_key.pem.backup public_key.pem

# 5. Start all services
make up

# 6. Verify system health
curl http://localhost:8000/health
curl http://localhost:3000/api/health
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Failed

```bash
# Check database status
docker-compose ps qds-postgres

# Check database logs
make logs-postgres

# Restart database
docker-compose restart qds-postgres

# Test connection
docker-compose exec qds-postgres pg_isready -U $DB_USER
```

#### 2. Frontend Can't Connect to Backend

```bash
# Check backend status
curl http://localhost:8000/health

# Check network connectivity
docker-compose exec qds-frontend ping qds-backend

# Check environment variables
docker-compose exec qds-frontend env | grep API_URL

# Restart services
docker-compose restart qds-frontend qds-backend
```

#### 3. RSA Key Issues

```bash
# Verify private key
openssl rsa -in private_key.pem -check

# Verify public key
openssl rsa -in private_key.pem -pubout | diff - public_key.pem

# Regenerate keys if corrupted
make keygen
```

#### 4. High Memory Usage

```bash
# Check container memory usage
docker stats

# Check database connections
docker-compose exec qds-postgres psql -U $DB_USER $DB_NAME -c "SELECT count(*) FROM pg_stat_activity;"

# Restart services to clear memory
make down && make up
```

### Log Analysis for Troubleshooting

```bash
# Find errors in last hour
find logs/ -name "*.log" -exec grep "$(date -d '1 hour ago' '+%Y-%m-%d %H')" {} \; | grep ERROR

# Check for database connection errors
grep "database" logs/error.log | tail -10

# Check for authentication issues
grep "auth" logs/error.log | tail -10

# Monitor real-time errors
tail -f logs/error.log | grep ERROR
```

## Performance Monitoring

### Key Metrics

Monitor these performance indicators:

1. **Response Times**
   - Document signing: < 30 seconds
   - Document verification: < 10 seconds
   - Authentication: < 2 seconds

2. **Resource Usage**
   - CPU: < 70% average
   - Memory: < 80% of available
   - Disk: < 80% of available

3. **Database Performance**
   - Connection count: < 50
   - Query response time: < 1 second
   - Lock waits: minimal

### Performance Commands

```bash
# Monitor API response times
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8000/health

# Database performance
docker-compose exec qds-postgres psql -U $DB_USER $DB_NAME -c "
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements 
ORDER BY total_time DESC 
LIMIT 10;"

# Container resource usage
docker stats --no-stream
```

### Performance Optimization

```bash
# Database optimization
docker-compose exec qds-postgres psql -U $DB_USER $DB_NAME -c "
VACUUM ANALYZE;
REINDEX DATABASE $DB_NAME;"

# Clear Docker cache
docker system prune -f

# Restart services for fresh start
make down && make up
```

## Maintenance Schedule

### Daily Tasks

- [ ] Check service health endpoints
- [ ] Review error logs for critical issues
- [ ] Monitor disk space usage
- [ ] Verify backup completion

### Weekly Tasks

- [ ] Rotate log files
- [ ] Review audit logs for security events
- [ ] Check database performance metrics
- [ ] Update system packages (if needed)

### Monthly Tasks

- [ ] Full system backup
- [ ] Review and archive old logs
- [ ] Security audit review
- [ ] Performance analysis
- [ ] Update documentation

### Quarterly Tasks

- [ ] RSA key rotation
- [ ] Security penetration testing
- [ ] Disaster recovery testing
- [ ] Capacity planning review

## Maintenance Commands

Quick reference for common maintenance tasks:

```bash
# System health check
make logs | grep ERROR
curl http://localhost:8000/health
curl http://localhost:3000/api/health

# Backup and cleanup
make backup
./scripts/rotate-logs.sh

# Service management
make down
make up
make clean

# Database maintenance
docker-compose exec qds-postgres psql -U $DB_USER $DB_NAME -c "VACUUM ANALYZE;"

# Security check
grep "AUTH_FAILURE\|SUSPICIOUS_ACTIVITY" logs/audit.log | tail -20
```

## Emergency Contacts

In case of critical issues:

1. **System Administrator**: [Contact Information]
2. **Database Administrator**: [Contact Information]
3. **Security Team**: [Contact Information]
4. **Development Team**: [Contact Information]

## Support Resources

- **Documentation**: This file and DEPLOYMENT.md
- **Logs**: `backend/internal/infrastructure/handlers/logs/`
- **Configuration**: `.env` file
- **Backups**: `backups/` directory
- **Scripts**: `scripts/` directory

Remember to keep this documentation updated as the system evolves and new operational procedures are established.