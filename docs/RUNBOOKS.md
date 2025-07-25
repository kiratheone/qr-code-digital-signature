# Digital Signature System - Operational Runbooks

This document contains step-by-step procedures for common operational tasks and incident response for the Digital Signature System.

## Table of Contents

1. [Emergency Procedures](#emergency-procedures)
2. [Service Management](#service-management)
3. [Database Operations](#database-operations)
4. [Monitoring and Alerting](#monitoring-and-alerting)
5. [Backup and Recovery](#backup-and-recovery)
6. [Security Incidents](#security-incidents)
7. [Performance Issues](#performance-issues)
8. [Maintenance Procedures](#maintenance-procedures)

## Emergency Procedures

### System Down - Complete Outage

**Symptoms**: All services are unresponsive, health checks failing

**Immediate Actions**:
1. Check system status:
   ```bash
   ./scripts/monitor.sh check
   ```

2. Check Docker services:
   ```bash
   docker-compose -f docker-compose.prod.yml ps
   ```

3. Restart all services:
   ```bash
   docker-compose -f docker-compose.prod.yml restart
   ```

4. If restart fails, perform full redeployment:
   ```bash
   ./scripts/deploy.sh stop
   ./scripts/deploy.sh
   ```

5. Verify recovery:
   ```bash
   ./scripts/health-check.sh
   ```

**Escalation**: If services don't recover within 10 minutes, escalate to senior engineer.

### Database Connection Issues

**Symptoms**: Backend returns database connection errors

**Immediate Actions**:
1. Check PostgreSQL status:
   ```bash
   docker-compose -f docker-compose.prod.yml exec postgres pg_isready -U postgres
   ```

2. Check database connections:
   ```bash
   docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -d digital_signature -c "SELECT count(*) FROM pg_stat_activity;"
   ```

3. If too many connections, restart backend:
   ```bash
   docker-compose -f docker-compose.prod.yml restart backend
   ```

4. If database is down, restart PostgreSQL:
   ```bash
   docker-compose -f docker-compose.prod.yml restart postgres
   ```

5. Check logs for root cause:
   ```bash
   docker-compose -f docker-compose.prod.yml logs postgres
   ```

### High Memory Usage Alert

**Symptoms**: Memory usage > 85%, system becoming slow

**Immediate Actions**:
1. Check memory usage:
   ```bash
   free -h
   docker stats --no-stream
   ```

2. Identify memory-consuming containers:
   ```bash
   docker stats --format "table {{.Container}}\t{{.MemUsage}}\t{{.MemPerc}}" --no-stream | sort -k3 -nr
   ```

3. Restart high-memory containers:
   ```bash
   docker-compose -f docker-compose.prod.yml restart [service-name]
   ```

4. If system memory is low, clean up Docker resources:
   ```bash
   ./scripts/maintenance.sh docker
   ```

### SSL Certificate Expired

**Symptoms**: HTTPS connections failing, certificate warnings

**Immediate Actions**:
1. Check certificate status:
   ```bash
   openssl x509 -in nginx/ssl/cert.pem -noout -enddate
   ```

2. If expired, generate new self-signed certificate (temporary):
   ```bash
   ./scripts/setup-secrets.sh
   ```

3. Restart nginx:
   ```bash
   docker-compose -f docker-compose.prod.yml restart nginx
   ```

4. For production, obtain new certificate from CA and replace files.

## Service Management

### Starting Services

```bash
# Start all services
docker-compose -f docker-compose.prod.yml up -d

# Start specific service
docker-compose -f docker-compose.prod.yml up -d [service-name]

# Start with build
docker-compose -f docker-compose.prod.yml up -d --build
```

### Stopping Services

```bash
# Stop all services
docker-compose -f docker-compose.prod.yml down

# Stop specific service
docker-compose -f docker-compose.prod.yml stop [service-name]

# Stop and remove volumes (DANGEROUS)
docker-compose -f docker-compose.prod.yml down -v
```

### Restarting Services

```bash
# Restart all services
docker-compose -f docker-compose.prod.yml restart

# Restart specific service
docker-compose -f docker-compose.prod.yml restart [service-name]

# Rolling restart (zero downtime)
for service in backend frontend; do
  docker-compose -f docker-compose.prod.yml up -d --no-deps $service
  sleep 30
done
```

### Viewing Logs

```bash
# View all logs
docker-compose -f docker-compose.prod.yml logs -f

# View specific service logs
docker-compose -f docker-compose.prod.yml logs -f [service-name]

# View last 100 lines
docker-compose -f docker-compose.prod.yml logs --tail=100 [service-name]

# View logs since specific time
docker-compose -f docker-compose.prod.yml logs --since=1h [service-name]
```

### Scaling Services

```bash
# Scale backend service to 3 instances
docker-compose -f docker-compose.prod.yml up -d --scale backend=3

# Scale down to 1 instance
docker-compose -f docker-compose.prod.yml up -d --scale backend=1
```

## Database Operations

### Database Backup

```bash
# Create backup
./scripts/postgres-backup.sh backup

# List available backups
./scripts/postgres-backup.sh list

# Verify backup integrity
./scripts/postgres-backup.sh verify [backup-file]
```

### Database Restore

```bash
# Restore from specific backup
./scripts/postgres-backup.sh restore [backup-file]

# Restore latest backup
LATEST_BACKUP=$(ls -t ./backups/postgres_backup_*.sql.gz | head -1)
./scripts/postgres-backup.sh restore "$LATEST_BACKUP"
```

### Database Maintenance

```bash
# Run database optimization
./scripts/maintenance.sh database

# Manual database maintenance
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -d digital_signature -c "VACUUM ANALYZE;"
```

### Database Monitoring

```bash
# Check database size
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -d digital_signature -c "SELECT pg_size_pretty(pg_database_size('digital_signature'));"

# Check active connections
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -d digital_signature -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';"

# Check long-running queries
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -d digital_signature -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes';"
```

## Monitoring and Alerting

### Health Checks

```bash
# Comprehensive health check
./scripts/monitor.sh check

# Check specific components
./scripts/monitor.sh services
./scripts/monitor.sh health
./scripts/monitor.sh resources
./scripts/monitor.sh database
```

### Continuous Monitoring

```bash
# Start continuous monitoring
./scripts/monitor.sh continuous

# Monitor with custom interval (30 seconds)
./scripts/monitor.sh -i 30 continuous

# Monitor with webhook alerts
./scripts/monitor.sh -w https://hooks.slack.com/... continuous
```

### Accessing Monitoring Dashboards

- **Grafana**: http://localhost:3001 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Application Metrics**: http://localhost:8000/metrics

### Alert Response

When receiving alerts:

1. **Acknowledge the alert** in your monitoring system
2. **Check system status** using monitoring scripts
3. **Follow specific runbook** for the alert type
4. **Document actions taken** in incident log
5. **Update stakeholders** on status and resolution

## Backup and Recovery

### Automated Backups

Backups run automatically via cron:
```bash
# Check cron jobs
crontab -l

# Add backup job (runs daily at 2 AM)
echo "0 2 * * * /path/to/scripts/postgres-backup.sh backup" | crontab -
```

### Manual Backup

```bash
# Full system backup
./scripts/maintenance.sh backup

# Database only
./scripts/postgres-backup.sh backup

# Configuration backup
tar -czf config-backup-$(date +%Y%m%d).tar.gz .env.prod docker-compose.prod.yml nginx/ secrets/
```

### Disaster Recovery

**Complete System Recovery**:

1. **Prepare new environment**:
   ```bash
   git clone [repository-url]
   cd digital-signature-system
   ```

2. **Restore configuration**:
   ```bash
   # Extract configuration backup
   tar -xzf config-backup-[date].tar.gz
   ```

3. **Setup secrets**:
   ```bash
   ./scripts/setup-secrets.sh
   ```

4. **Deploy services**:
   ```bash
   ./scripts/deploy.sh
   ```

5. **Restore database**:
   ```bash
   ./scripts/postgres-backup.sh restore [backup-file]
   ```

6. **Verify system**:
   ```bash
   ./scripts/health-check.sh
   ```

## Security Incidents

### Suspected Breach

**Immediate Actions**:
1. **Isolate the system**:
   ```bash
   # Block external access
   docker-compose -f docker-compose.prod.yml stop nginx
   ```

2. **Preserve evidence**:
   ```bash
   # Capture logs
   docker-compose -f docker-compose.prod.yml logs > incident-logs-$(date +%Y%m%d-%H%M%S).txt
   
   # Capture system state
   ./scripts/monitor.sh report
   ```

3. **Check for unauthorized access**:
   ```bash
   # Check authentication logs
   docker-compose -f docker-compose.prod.yml logs backend | grep -i "login\|auth"
   
   # Check for suspicious activity
   docker-compose -f docker-compose.prod.yml logs | grep -E "(403|404|500)" | tail -100
   ```

4. **Notify security team** and follow incident response procedures

### Failed Login Attempts

**Investigation**:
```bash
# Check failed login attempts
docker-compose -f docker-compose.prod.yml logs backend | grep -i "failed\|unauthorized" | tail -50

# Check source IPs
docker-compose -f docker-compose.prod.yml logs nginx | grep -E "(403|401)" | awk '{print $1}' | sort | uniq -c | sort -nr
```

**Mitigation**:
```bash
# Block suspicious IPs (if using fail2ban)
fail2ban-client set nginx-limit-req banip [IP_ADDRESS]

# Restart services to clear sessions
docker-compose -f docker-compose.prod.yml restart backend
```

### SSL/TLS Issues

**Check certificate**:
```bash
# Verify certificate
openssl x509 -in nginx/ssl/cert.pem -text -noout

# Test SSL connection
openssl s_client -connect localhost:443 -servername [domain]
```

**Renew certificate**:
```bash
# For Let's Encrypt
certbot renew --dry-run

# Update nginx configuration
docker-compose -f docker-compose.prod.yml restart nginx
```

## Performance Issues

### High CPU Usage

**Investigation**:
```bash
# Check system CPU
top -bn1 | head -20

# Check container CPU usage
docker stats --no-stream | sort -k3 -nr

# Check application metrics
curl -s http://localhost:8000/metrics | grep cpu
```

**Mitigation**:
```bash
# Scale backend services
docker-compose -f docker-compose.prod.yml up -d --scale backend=3

# Restart high-CPU containers
docker-compose -f docker-compose.prod.yml restart [service-name]
```

### Slow Response Times

**Investigation**:
```bash
# Check response times
curl -w "@curl-format.txt" -o /dev/null -s http://localhost/api/health

# Check database performance
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -d digital_signature -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"
```

**Optimization**:
```bash
# Optimize database
./scripts/maintenance.sh database

# Clear Redis cache
docker-compose -f docker-compose.prod.yml exec redis redis-cli FLUSHALL

# Restart services
docker-compose -f docker-compose.prod.yml restart
```

### Memory Leaks

**Investigation**:
```bash
# Monitor memory usage over time
while true; do
  docker stats --no-stream --format "table {{.Container}}\t{{.MemUsage}}\t{{.MemPerc}}"
  sleep 60
done
```

**Mitigation**:
```bash
# Restart affected services
docker-compose -f docker-compose.prod.yml restart [service-name]

# Implement memory limits
# Edit docker-compose.prod.yml to add memory limits
```

## Maintenance Procedures

### Scheduled Maintenance

**Pre-maintenance**:
1. **Notify users** of scheduled maintenance
2. **Enable maintenance mode**:
   ```bash
   ./scripts/maintenance.sh -m full
   ```
3. **Create backup**:
   ```bash
   ./scripts/postgres-backup.sh backup
   ./scripts/maintenance.sh backup
   ```

**During maintenance**:
1. **Update system**:
   ```bash
   ./scripts/maintenance.sh system
   ```
2. **Update Docker images**:
   ```bash
   ./scripts/maintenance.sh images
   ```
3. **Optimize database**:
   ```bash
   ./scripts/maintenance.sh database
   ```
4. **Clean up resources**:
   ```bash
   ./scripts/maintenance.sh docker
   ```

**Post-maintenance**:
1. **Run health checks**:
   ```bash
   ./scripts/maintenance.sh health
   ```
2. **Disable maintenance mode**:
   ```bash
   # Maintenance mode is automatically disabled
   ```
3. **Generate report**:
   ```bash
   ./scripts/maintenance.sh report
   ```

### Emergency Maintenance

For urgent security updates or critical fixes:

1. **Assess impact** and determine if maintenance mode is needed
2. **Apply fixes** with minimal downtime:
   ```bash
   # Rolling update
   docker-compose -f docker-compose.prod.yml build [service]
   docker-compose -f docker-compose.prod.yml up -d --no-deps [service]
   ```
3. **Verify fix** is working correctly
4. **Document changes** and notify stakeholders

### Regular Maintenance Tasks

**Daily**:
- Check system health
- Review error logs
- Monitor resource usage

**Weekly**:
- Review backup integrity
- Check SSL certificate expiry
- Update security patches

**Monthly**:
- Full system maintenance
- Performance optimization
- Security audit

**Quarterly**:
- Disaster recovery testing
- Security penetration testing
- Capacity planning review

## Contact Information

### Escalation Contacts

- **Level 1 Support**: [support@company.com]
- **Level 2 Engineering**: [engineering@company.com]
- **Security Team**: [security@company.com]
- **Management**: [management@company.com]

### Emergency Contacts

- **On-call Engineer**: [phone-number]
- **Security Incident**: [security-phone]
- **Management Escalation**: [management-phone]

### External Vendors

- **Cloud Provider Support**: [provider-support]
- **SSL Certificate Authority**: [ca-support]
- **Monitoring Service**: [monitoring-support]

## Appendix

### Useful Commands Reference

```bash
# Quick system status
docker-compose -f docker-compose.prod.yml ps && docker stats --no-stream

# Emergency restart
docker-compose -f docker-compose.prod.yml restart

# View all logs
docker-compose -f docker-compose.prod.yml logs -f --tail=100

# Database connection test
docker-compose -f docker-compose.prod.yml exec postgres pg_isready -U postgres

# Health check
curl -f http://localhost/api/health && echo "OK" || echo "FAILED"

# Disk space check
df -h && docker system df

# Memory usage
free -h && docker stats --no-stream
```

### Log Locations

- **Application logs**: Docker container logs
- **System logs**: `/var/log/`
- **Maintenance logs**: `/var/log/digital-signature-maintenance.log`
- **Monitor logs**: `/var/log/digital-signature-monitor.log`
- **Backup logs**: `./backups/`

### Configuration Files

- **Environment**: `.env.prod`
- **Docker Compose**: `docker-compose.prod.yml`
- **Nginx**: `nginx/nginx.conf`
- **Prometheus**: `monitoring/prometheus.yml`
- **Grafana**: `monitoring/grafana/`
- **Secrets**: `secrets/` (encrypted)

---

**Document Version**: 1.0  
**Last Updated**: $(date)  
**Next Review**: $(date -d "+3 months")