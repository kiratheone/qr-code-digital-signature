# Deployment Troubleshooting Guide

This guide covers common issues encountered during deployment and their solutions.

## Network Issues

### Error: "invalid pool request: Pool overlaps with other one on this address space"

**Cause**: Docker network subnet conflict with existing networks.

**Solution**:
1. Check existing networks: `docker network ls`
2. Check IP ranges: `bash scripts/check-network-conflicts.sh`
3. Update subnet in `docker-compose.prod.yml` if needed
4. Run deployment script which includes automatic cleanup

**Prevention**: The deploy script now includes automatic network cleanup.

## Volume Mount Issues

### Error: "failed to populate volume: no such file or directory"

**Cause**: Required data directories don't exist for bind mounts.

**Solution**:
1. Run: `bash scripts/setup-data-dirs.sh`
2. Or use the deploy script which creates directories automatically

**Directory Structure**:
```
data/
├── postgres/     # PostgreSQL data (permissions: 700)
├── redis/        # Redis data (permissions: 755)
├── nginx/        # Nginx logs (permissions: 755)
├── prometheus/   # Prometheus data (permissions: 755)
├── grafana/      # Grafana data (permissions: 755)
└── loki/         # Loki logs (permissions: 755)
```

## Environment Variable Issues

### Error: "export: not a valid identifier"

**Cause**: Invalid characters in environment file (comments, empty lines).

**Solution**: The deploy script now properly filters environment variables.

**Manual Fix**:
```bash
# Test environment loading
bash scripts/test-env-loading.sh

# Check for invalid lines
grep -n '^#\|^$' .env.prod
```

## Secret Files Issues

### Error: "secret file not found"

**Cause**: Required secret files are missing.

**Solution**:
1. Run: `bash scripts/setup-secrets.sh`
2. Ensure all required files exist in `./secrets/`:
   - `db_password.txt`
   - `jwt_secret.txt`
   - `private_key.pem`
   - `public_key.pem`
   - `redis_password.txt`

## Service Health Issues

### Error: "service is not healthy"

**Cause**: Service failed to start or health check failed.

**Solution**:
1. Check service logs: `docker-compose -f docker-compose.prod.yml logs <service>`
2. Check service status: `docker-compose -f docker-compose.prod.yml ps`
3. Restart specific service: `docker-compose -f docker-compose.prod.yml restart <service>`

**Common Service Issues**:

#### PostgreSQL
- Check data directory permissions: `ls -la data/postgres`
- Verify password file: `cat secrets/db_password.txt`
- Check logs: `docker-compose -f docker-compose.prod.yml logs postgres`

#### Redis
- Check password configuration
- Verify Redis data directory
- Check logs: `docker-compose -f docker-compose.prod.yml logs redis`

#### Backend
- Check environment variables
- Verify secret files exist
- Check database connection
- Check logs: `docker-compose -f docker-compose.prod.yml logs backend`

#### Frontend
- Check API URL configuration
- Verify backend is healthy
- Check logs: `docker-compose -f docker-compose.prod.yml logs frontend`

## Build Issues

### Error: "build failed"

**Cause**: Build process failed due to missing dependencies or configuration.

**Solution**:
1. Check build logs for specific errors
2. Ensure all required files exist
3. Try rebuilding: `docker-compose -f docker-compose.prod.yml build --no-cache`

## Port Conflicts

### Error: "port already in use"

**Cause**: Another service is using the same port.

**Solution**:
1. Check what's using the port: `sudo netstat -tulpn | grep :<port>`
2. Stop conflicting service or change port in docker-compose.prod.yml
3. Update environment variables accordingly

## SSL/TLS Issues

### Error: "certificate not found"

**Cause**: SSL certificates are missing.

**Solution**:
1. Generate self-signed certificates for development
2. Place certificates in `nginx/ssl/` directory
3. Update paths in `nginx/nginx.conf`

## Memory/Resource Issues

### Error: "container killed (OOMKilled)"

**Cause**: Container exceeded memory limits.

**Solution**:
1. Increase memory limits in docker-compose.prod.yml
2. Check system resources: `docker stats`
3. Optimize application memory usage

## Debugging Commands

### Check Service Status
```bash
# All services
docker-compose -f docker-compose.prod.yml ps

# Specific service
docker-compose -f docker-compose.prod.yml ps <service>
```

### View Logs
```bash
# All services
docker-compose -f docker-compose.prod.yml logs

# Specific service
docker-compose -f docker-compose.prod.yml logs <service>

# Follow logs
docker-compose -f docker-compose.prod.yml logs -f <service>
```

### Check Resource Usage
```bash
# Container stats
docker stats

# System resources
df -h
free -h
```

### Network Debugging
```bash
# List networks
docker network ls

# Inspect network
docker network inspect <network_name>

# Check connectivity
docker exec <container> ping <target>
```

## Recovery Procedures

### Complete Reset
```bash
# Stop all services
docker-compose -f docker-compose.prod.yml down

# Remove volumes (WARNING: This deletes all data)
docker-compose -f docker-compose.prod.yml down -v

# Clean up networks
docker network prune -f

# Rebuild and restart
bash scripts/deploy.sh deploy
```

### Partial Reset (Keep Data)
```bash
# Stop services
docker-compose -f docker-compose.prod.yml down

# Remove containers and networks only
docker system prune -f

# Restart
bash scripts/deploy.sh deploy
```

## Getting Help

1. Check service logs first
2. Verify all prerequisites are met
3. Run diagnostic scripts:
   - `bash scripts/check-network-conflicts.sh`
   - `bash scripts/test-env-loading.sh`
   - `bash scripts/setup-data-dirs.sh`
4. Check this troubleshooting guide
5. Review Docker and service documentation