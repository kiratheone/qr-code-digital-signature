# Maintenance Checklist

This document provides quick maintenance checklists for the Digital Signature System.

## Daily Maintenance Checklist

### System Health (5 minutes)

- [ ] Run health check: `make health`
- [ ] Check service status: `docker-compose ps`
- [ ] Verify disk space: `df -h`
- [ ] Check recent errors: `tail -20 backend/internal/infrastructure/handlers/logs/error.log`

### Quick Commands
```bash
# Daily health check
make health

# Check for recent errors
grep "$(date '+%Y-%m-%d')" backend/internal/infrastructure/handlers/logs/error.log | grep ERROR
```

## Weekly Maintenance Checklist

### Log Management (10 minutes)

- [ ] Rotate logs: `make rotate-logs`
- [ ] Review audit logs for security events
- [ ] Check log directory size: `du -sh backend/internal/infrastructure/handlers/logs/`
- [ ] Archive old logs if needed

### Database Health (5 minutes)

- [ ] Create backup: `make backup`
- [ ] Check database size
- [ ] Verify backup integrity

### Quick Commands
```bash
# Weekly maintenance
make rotate-logs
make backup

# Check database size
docker-compose exec qds-postgres psql -U $DB_USER $DB_NAME -c "SELECT pg_size_pretty(pg_database_size('$DB_NAME'));"

# Review security events
grep "AUTH_FAILURE\|SUSPICIOUS_ACTIVITY" backend/internal/infrastructure/handlers/logs/audit.log | tail -20
```

## Monthly Maintenance Checklist

### Security Review (15 minutes)

- [ ] Review authentication failures
- [ ] Check for suspicious activities
- [ ] Verify RSA key integrity
- [ ] Review user access logs

### Performance Review (10 minutes)

- [ ] Check container resource usage: `docker stats`
- [ ] Review database performance
- [ ] Check response times
- [ ] Clean up old data if needed

### System Updates (20 minutes)

- [ ] Update Docker images: `make prod-build`
- [ ] Review and update environment variables
- [ ] Test system after updates
- [ ] Update documentation if needed

### Quick Commands
```bash
# Monthly security review
grep "AUTH_FAILURE" backend/internal/infrastructure/handlers/logs/audit.log | wc -l
grep "SUSPICIOUS_ACTIVITY" backend/internal/infrastructure/handlers/logs/audit.log

# Performance check
docker stats --no-stream
docker-compose exec qds-postgres psql -U $DB_USER $DB_NAME -c "SELECT count(*) FROM pg_stat_activity;"

# System update
make prod-build
make prod-down
make prod-up
make health
```

## Quarterly Maintenance Checklist

### Security Audit (30 minutes)

- [ ] Rotate RSA keys: `cd backend && go run cmd/keyrotate/main.go`
- [ ] Review all user accounts
- [ ] Check for unused sessions
- [ ] Audit system access logs

### Disaster Recovery Test (45 minutes)

- [ ] Test backup restoration process
- [ ] Verify system recovery procedures
- [ ] Update disaster recovery documentation
- [ ] Test monitoring and alerting

### Capacity Planning (20 minutes)

- [ ] Review system growth metrics
- [ ] Check storage usage trends
- [ ] Plan for scaling if needed
- [ ] Update resource requirements

### Quick Commands
```bash
# Quarterly security audit
cd backend && go run cmd/keyrotate/main.go

# Check user accounts
docker-compose exec qds-postgres psql -U $DB_USER $DB_NAME -c "SELECT username, email, created_at, is_active FROM users ORDER BY created_at DESC;"

# Test backup restoration
make backup
# Follow restoration procedure in OPERATIONS.md

# Capacity planning
du -sh .
docker system df
```

## Emergency Procedures

### Service Down

1. Check service status: `docker-compose ps`
2. Check logs: `make logs`
3. Restart services: `make down && make up`
4. Verify health: `make health`

### Database Issues

1. Check database logs: `make logs-postgres`
2. Test connection: `docker-compose exec qds-postgres pg_isready -U $DB_USER`
3. Restart database: `docker-compose restart qds-postgres`
4. Restore from backup if needed

### High Resource Usage

1. Check resource usage: `docker stats`
2. Check disk space: `df -h`
3. Rotate logs: `make rotate-logs`
4. Clean up Docker: `docker system prune -f`
5. Restart services: `make down && make up`

### Security Incident

1. Check audit logs: `grep "SUSPICIOUS_ACTIVITY\|AUTH_FAILURE" backend/internal/infrastructure/handlers/logs/audit.log`
2. Review recent user activities
3. Disable suspicious accounts if needed
4. Change RSA keys if compromised
5. Document incident

## Maintenance Scripts

All maintenance scripts are located in the `scripts/` directory:

- `scripts/backup.sh` - Database backup
- `scripts/rotate-logs.sh` - Log rotation
- `scripts/health-check.sh` - System health check

## Monitoring Commands

Quick reference for monitoring:

```bash
# System health
make health

# Service status
docker-compose ps

# Resource usage
docker stats --no-stream

# Recent errors
tail -50 backend/internal/infrastructure/handlers/logs/error.log

# Security events
grep "$(date '+%Y-%m-%d')" backend/internal/infrastructure/handlers/logs/audit.log | grep -E "AUTH_FAILURE|SUSPICIOUS_ACTIVITY"

# Database status
docker-compose exec qds-postgres pg_isready -U $DB_USER

# Disk usage
df -h
du -sh backend/internal/infrastructure/handlers/logs/
```

## Maintenance Log

Keep a record of maintenance activities:

| Date | Activity | Performed By | Notes |
|------|----------|--------------|-------|
| YYYY-MM-DD | Daily health check | [Name] | All systems normal |
| YYYY-MM-DD | Log rotation | [Name] | Rotated 3 log files |
| YYYY-MM-DD | Database backup | [Name] | Backup size: 50MB |

## Contact Information

For maintenance issues:

- **System Administrator**: [Contact]
- **Database Administrator**: [Contact]
- **Security Team**: [Contact]
- **On-call Support**: [Contact]

## Additional Resources

- **DEPLOYMENT.md** - Deployment procedures
- **OPERATIONS.md** - Detailed operational guide
- **README.md** - General system information
- **scripts/** - Maintenance scripts