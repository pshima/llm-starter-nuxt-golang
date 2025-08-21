# Deployment Guide

## Production Environment Configuration

### Required Environment Variables

```bash
# Server Configuration
SERVER_HOST=0.0.0.0              # Always use 0.0.0.0 in containers
SERVER_PORT=8080                 # Port to bind
SERVER_READ_TIMEOUT=30           # Read timeout in seconds
SERVER_WRITE_TIMEOUT=30          # Write timeout in seconds
ENVIRONMENT=production           # Set to production

# Redis Configuration
REDIS_HOST=redis                # Redis hostname/IP
REDIS_PORT=6379                  # Redis port
REDIS_PASSWORD=strong_password   # Set strong password in production
REDIS_DB=0                       # Database number
REDIS_POOL_SIZE=20              # Connection pool size

# Session Configuration
SESSION_DURATION=7               # Days
SESSION_SECRET=<generate-64-char-random-string>  # MUST change in production
SESSION_SECURE=true              # Use secure cookies (HTTPS only)
SESSION_HTTP_ONLY=true           # Prevent JS access

# Email Configuration
SMTP_HOST=smtp.example.com      # Your SMTP server
SMTP_PORT=587                   # SMTP port (587 for TLS, 465 for SSL)
SMTP_USERNAME=notifications@example.com
SMTP_PASSWORD=smtp_password
EMAIL_FROM_ADDRESS=noreply@example.com
EMAIL_FROM_NAME=Task Tracker

# Security Configuration
RATE_LIMIT=1000                 # Requests per minute per IP
ENABLE_CORS=true                # Enable CORS
FRONTEND_URL=https://example.com # Your frontend URL
```

### Generating Secure Secrets

```bash
# Generate session secret
openssl rand -base64 48

# Generate Redis password
openssl rand -base64 32
```

## Redis Configuration

### Persistence Configuration

Create `redis.conf`:

```conf
# Enable both RDB and AOF for maximum durability
save 900 1      # Save after 900 sec if at least 1 key changed
save 300 10     # Save after 300 sec if at least 10 keys changed
save 60 10000   # Save after 60 sec if at least 10000 keys changed

appendonly yes
appendfsync everysec  # Sync to disk every second

# Memory management
maxmemory 2gb
maxmemory-policy allkeys-lru  # Evict least recently used keys

# Security
requirepass your_redis_password
protected-mode yes
bind 127.0.0.1 ::1  # Only bind to localhost if using proxy

# Performance
tcp-backlog 511
tcp-keepalive 300
timeout 0

# Logging
loglevel notice
logfile /var/log/redis/redis-server.log
```

### Backup Strategy

```bash
#!/bin/bash
# backup-redis.sh - Run daily via cron

BACKUP_DIR="/backup/redis"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Save Redis data
redis-cli --rdb $BACKUP_DIR/dump_$DATE.rdb

# Keep only last 7 days of backups
find $BACKUP_DIR -name "dump_*.rdb" -mtime +7 -delete

# Optional: Upload to S3
aws s3 cp $BACKUP_DIR/dump_$DATE.rdb s3://your-backup-bucket/redis/
```

## Docker Production Setup

### Production Dockerfile

```dockerfile
# Production-optimized Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Final stage - minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Create non-root user
RUN addgroup -g 1000 -S appuser && \
    adduser -u 1000 -S appuser -G appuser

USER appuser

EXPOSE 8080

CMD ["./main"]
```

### Docker Compose Production

```yaml
# docker-compose.prod.yml
services:
  backend:
    image: your-registry/task-tracker-backend:latest
    container_name: task-tracker-backend
    restart: always
    ports:
      - "127.0.0.1:8080:8080"  # Only expose to localhost
    environment:
      - ENVIRONMENT=production
      # Load other env vars from .env file
    env_file:
      - .env.production
    depends_on:
      - redis
    networks:
      - task-tracker-network
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  redis:
    image: redis:7-alpine
    container_name: task-tracker-redis
    restart: always
    command: redis-server /usr/local/etc/redis/redis.conf
    volumes:
      - ./redis.conf:/usr/local/etc/redis/redis.conf:ro
      - redis-data:/data
    networks:
      - task-tracker-network
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  redis-data:
    driver: local

networks:
  task-tracker-network:
    driver: bridge
```

## Monitoring & Maintenance

### Health Check Endpoints

The application provides health checks at:

```bash
# Basic health check
GET /health

# Response
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

### Monitoring Setup

#### Prometheus Metrics (Future Enhancement)

```go
// Add to main.go for metrics
import "github.com/prometheus/client_golang/prometheus/promhttp"

router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

#### Logging Configuration

```bash
# Log rotation with logrotate
# /etc/logrotate.d/task-tracker

/var/log/task-tracker/*.log {
    daily
    rotate 14
    missingok
    notifempty
    compress
    delaycompress
    create 0640 appuser appuser
    sharedscripts
    postrotate
        docker exec task-tracker-backend kill -USR1 1
    endscript
}
```

### Redis Memory Management

Monitor Redis memory usage:

```bash
# Check memory usage
redis-cli INFO memory

# Monitor in real-time
redis-cli --stat

# Get memory stats for specific pattern
redis-cli --scan --pattern 'user:*' | xargs -L 1 redis-cli DEBUG OBJECT | grep serializedlength
```

### Soft-Deleted Task Cleanup

Although tasks are set with TTL, you can run manual cleanup:

```bash
#!/bin/bash
# cleanup-deleted-tasks.sh - Run daily via cron

# Connect to Redis and cleanup tasks deleted > 7 days ago
redis-cli EVAL "
local deleted_keys = redis.call('KEYS', 'user:*:tasks:deleted')
local cutoff = tonumber(ARGV[1])
local removed = 0

for _, key in ipairs(deleted_keys) do
    removed = removed + redis.call('ZREMRANGEBYSCORE', key, 0, cutoff)
end

return removed
" 0 $(date -d '7 days ago' +%s)
```

## Scaling Considerations

### Horizontal Scaling

#### Load Balancing with Nginx

```nginx
# /etc/nginx/sites-available/task-tracker
upstream backend {
    least_conn;
    server backend1:8080 weight=1;
    server backend2:8080 weight=1;
    server backend3:8080 weight=1;
}

server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Session affinity (sticky sessions)
        proxy_cookie_path / "/; HTTPOnly; Secure; SameSite=Strict";
    }

    location /health {
        access_log off;
        proxy_pass http://backend/health;
    }
}
```

### Redis Clustering

For high availability, use Redis Sentinel:

```yaml
# docker-compose.redis-ha.yml
services:
  redis-master:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    
  redis-replica1:
    image: redis:7-alpine
    command: redis-server --replicaof redis-master 6379 --requirepass ${REDIS_PASSWORD} --masterauth ${REDIS_PASSWORD}
    
  redis-replica2:
    image: redis:7-alpine
    command: redis-server --replicaof redis-master 6379 --requirepass ${REDIS_PASSWORD} --masterauth ${REDIS_PASSWORD}
    
  sentinel1:
    image: redis:7-alpine
    command: redis-sentinel /etc/redis-sentinel/sentinel.conf
    volumes:
      - ./sentinel.conf:/etc/redis-sentinel/sentinel.conf
```

### Session Storage Scaling

For distributed sessions across multiple backend instances:

1. **Sticky Sessions**: Use load balancer session affinity
2. **Shared Redis**: All instances connect to same Redis cluster
3. **Session Replication**: Use Redis replication for HA

## Security Hardening

### System Security

```bash
# Firewall rules (ufw)
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable

# Fail2ban for brute force protection
apt-get install fail2ban
```

### Docker Security

```bash
# Run security scan
docker scan task-tracker-backend:latest

# Use Docker secrets for sensitive data
docker secret create redis_password ./redis_password.txt
docker secret create session_secret ./session_secret.txt
```

### SSL/TLS Configuration

Use Let's Encrypt with Certbot:

```bash
# Install certbot
apt-get install certbot python3-certbot-nginx

# Get certificate
certbot --nginx -d api.example.com

# Auto-renewal cron job
echo "0 0,12 * * * root python3 -c 'import random; import time; time.sleep(random.random() * 3600)' && certbot renew" > /etc/cron.d/certbot
```

## Deployment Checklist

### Pre-Deployment

- [ ] Generate secure session secret
- [ ] Set strong Redis password
- [ ] Configure SMTP credentials
- [ ] Update CORS allowed origins
- [ ] Review and update rate limits
- [ ] Set up SSL certificates
- [ ] Configure firewall rules
- [ ] Set up monitoring and alerting
- [ ] Create backup strategy
- [ ] Test health check endpoints

### Deployment Steps

1. **Build and tag Docker image**:
   ```bash
   docker build -t task-tracker-backend:v1.0.0 -f Dockerfile .
   docker tag task-tracker-backend:v1.0.0 your-registry/task-tracker-backend:latest
   docker push your-registry/task-tracker-backend:latest
   ```

2. **Deploy with Docker Compose**:
   ```bash
   docker-compose -f docker-compose.prod.yml pull
   docker-compose -f docker-compose.prod.yml up -d
   ```

3. **Verify deployment**:
   ```bash
   curl https://api.example.com/health
   docker-compose -f docker-compose.prod.yml logs -f
   ```

### Post-Deployment

- [ ] Verify health checks are passing
- [ ] Test authentication flow
- [ ] Check Redis persistence
- [ ] Monitor error logs
- [ ] Verify backup scripts are running
- [ ] Load test the application
- [ ] Set up uptime monitoring
- [ ] Document any issues and resolutions

## Rollback Procedure

```bash
# Quick rollback to previous version
docker-compose -f docker-compose.prod.yml down
docker tag your-registry/task-tracker-backend:previous your-registry/task-tracker-backend:latest
docker-compose -f docker-compose.prod.yml up -d

# Restore Redis backup if needed
docker-compose -f docker-compose.prod.yml stop redis
docker cp backup/dump_20240101.rdb task-tracker-redis:/data/dump.rdb
docker-compose -f docker-compose.prod.yml start redis
```

## Troubleshooting Production Issues

### Common Issues and Solutions

1. **High Memory Usage**:
   ```bash
   # Check Redis memory
   redis-cli INFO memory
   # Clear expired keys
   redis-cli --scan --pattern '*' | xargs -L 100 redis-cli TTL | grep -c '^-2$'
   ```

2. **Connection Pool Exhaustion**:
   - Increase `REDIS_POOL_SIZE`
   - Check for connection leaks in code
   - Monitor with `redis-cli CLIENT LIST`

3. **Session Issues**:
   - Verify `SESSION_SECRET` is consistent across deployments
   - Check cookie settings match domain
   - Verify Redis persistence is working

4. **Performance Degradation**:
   - Check Redis slow log: `redis-cli SLOWLOG GET 10`
   - Monitor backend response times
   - Review indexes and query patterns