# Security Documentation

## Overview

This document outlines the security measures implemented in the Task Tracker application and provides guidelines for maintaining security.

## Authentication & Authorization

### Session-Based Authentication

#### Implementation Details
- **Session Storage**: Server-side in Redis (not client-side)
- **Session ID**: Cryptographically random, generated using Go's crypto/rand
- **Cookie Settings**:
  - `HttpOnly`: Prevents JavaScript access
  - `Secure`: HTTPS-only in production
  - `SameSite`: CSRF protection
  - `Path=/`: Applies to entire application

#### Session Security Measures
```go
// Session creation with secure defaults
cookie := &http.Cookie{
    Name:     "session",
    Value:    sessionID,
    Path:     "/",
    MaxAge:   604800,  // 7 days
    HttpOnly: true,     // Prevent XSS
    Secure:   true,     // HTTPS only (production)
    SameSite: http.SameSiteLaxMode, // CSRF protection
}
```

#### Session Lifecycle
1. Created on successful login
2. Validated on every protected request
3. Extended on activity (sliding expiration)
4. Destroyed on logout or after 7 days of inactivity

### Password Security

#### Requirements Enforced
- Minimum 6 characters
- At least 1 special character (!@#$%^&*(),.?":{}|<>)
- At least 1 number
- Maximum 72 characters (bcrypt limitation)

#### Password Storage
- **Algorithm**: bcrypt with cost factor 10
- **Salting**: Automatic per-password salt
- **Storage**: Only hash stored, never plaintext

```go
// Password hashing implementation
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)

// Password verification
err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
```

#### Password Security Rationale
- **bcrypt**: Resistant to brute-force attacks
- **Cost factor 10**: Balance between security and performance
- **Per-password salt**: Prevents rainbow table attacks

### User Isolation

#### Implementation
All data access is scoped to the authenticated user:

```go
// Repository ensures user isolation
func (r *taskRepository) GetTaskByID(taskID string) (*domain.Task, error) {
    task := // ... fetch task
    
    // Verify ownership
    if task.UserID != currentUserID {
        return nil, ErrPermissionDenied
    }
    
    return task, nil
}
```

#### Isolation Boundaries
- Tasks: Only accessible by owner
- Categories: User-specific, no sharing
- Sessions: Isolated per user
- Deleted items: Only restorable by owner

## Input Validation

### Validation Layers

#### 1. Handler Layer (Request Validation)
```go
type CreateTaskRequest struct {
    Description string `json:"description" binding:"required,max=10000"`
    Category    string `json:"category" binding:"max=100"`
}
```

#### 2. Service Layer (Business Rules)
```go
func (s *taskService) CreateTask(userID, description, category string) (*domain.Task, error) {
    // Validate user ID
    if userID == "" {
        return nil, ErrUserIDRequired
    }
    
    // Validate description length
    if len(description) > 10000 {
        return nil, ErrDescriptionTooLong
    }
    
    // Additional business rules...
}
```

#### 3. Domain Layer (Entity Validation)
```go
func (t *Task) Validate() error {
    if t.UserID == "" {
        return ErrInvalidUserID
    }
    if t.Description == "" || len(t.Description) > 10000 {
        return ErrInvalidDescription
    }
    return nil
}
```

### Validation Rules by Field

| Field | Max Length | Format | Special Rules |
|-------|------------|--------|---------------|
| Email | 255 | RFC 5322 | Unique, lowercase |
| Password | 72 | - | Min 6 chars, 1 special, 1 number |
| DisplayName | 100 | Alphanumeric + spaces | Trimmed |
| Description | 10000 | Any UTF-8 | Required |
| Category | 100 | Any UTF-8 | Optional, auto-created |
| Session ID | 64 | Hex | Cryptographically random |

### SQL Injection Prevention

**Not Applicable**: The application uses Redis (NoSQL) exclusively. However, we still implement:

- **Parameterized Operations**: All Redis operations use parameterized commands
- **Input Sanitization**: Special characters are escaped where necessary
- **Type Safety**: Go's type system prevents many injection attempts

```go
// Safe Redis operation example
key := fmt.Sprintf("user:%s:tasks", userID)
// userID is validated and typed, preventing injection
```

### XSS Prevention Strategies

#### Current Measures
1. **API-Only Backend**: No server-side HTML rendering
2. **JSON Responses**: All responses are JSON, not HTML
3. **Content-Type Headers**: Properly set to `application/json`
4. **Input Validation**: HTML tags stripped/escaped in inputs

#### Frontend Responsibilities (Future)
```javascript
// Frontend should escape user content
function escapeHtml(unsafe) {
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}
```

## Data Protection

### Soft Delete for Data Recovery

#### Implementation
```go
func (t *Task) SoftDelete() {
    now := time.Now()
    t.DeletedAt = &now
    t.UpdatedAt = now
}
```

#### Security Benefits
- **Accidental Deletion**: 7-day recovery window
- **Audit Trail**: Deletion timestamp retained
- **Compliance**: Supports data retention requirements

#### Automatic Cleanup
```go
// Cleanup job for permanent deletion after 7 days
func CleanupExpiredTasks() {
    cutoff := time.Now().AddDate(0, 0, -7)
    // Delete tasks where DeletedAt < cutoff
}
```

### Session Expiration Strategy

#### Timeout Configuration
- **Idle Timeout**: 7 days (configurable)
- **Absolute Timeout**: None (sliding expiration)
- **Remember Me**: Default enabled (7-day sessions)

#### Security Trade-offs
- **Longer Sessions**: Better UX, higher risk if compromised
- **Shorter Sessions**: More secure, requires frequent login
- **Current Choice**: 7 days balances security and usability

### Redis Security Settings

#### Recommended Production Configuration
```conf
# redis.conf security settings

# Authentication
requirepass your-strong-redis-password

# Network security
bind 127.0.0.1 ::1  # Only localhost
protected-mode yes

# Command restrictions
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command KEYS ""
rename-command CONFIG ""

# SSL/TLS (Redis 6+)
tls-port 6379
port 0
tls-cert-file /path/to/redis.crt
tls-key-file /path/to/redis.key
tls-ca-cert-file /path/to/ca.crt
```

#### Data Persistence Security
```conf
# Persistence settings for data protection
save 900 1
save 300 10
save 60 10000

appendonly yes
appendfsync everysec

# Backup encryption (external)
# Use encrypted filesystem or backup tools
```

## Security Headers

### Recommended HTTP Headers

Add these headers in production (via reverse proxy or middleware):

```nginx
# Nginx configuration
add_header X-Content-Type-Options "nosniff" always;
add_header X-Frame-Options "DENY" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline';" always;

# HSTS (after SSL is configured)
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

### CORS Configuration

```go
// Current CORS settings
AllowedOrigins: []string{"http://localhost:3000"}  // Development
AllowedOrigins: []string{"https://example.com"}    // Production

// Headers allowed
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Origin, Content-Type, Accept
Access-Control-Allow-Credentials: true  // Required for cookies
```

## Rate Limiting

### Current Implementation
- **Limit**: 1000 requests per minute per IP
- **Scope**: Global (all endpoints)
- **Storage**: In-memory or Redis

### Future Enhancements
```go
// Per-endpoint rate limiting
rateLimits := map[string]int{
    "/api/v1/auth/login":    10,   // 10 per minute (brute-force protection)
    "/api/v1/auth/register": 10,   // 10 per minute (spam protection)
    "/api/v1/tasks":         100,  // 100 per minute (normal usage)
}
```

## Vulnerability Disclosure

### Reporting Security Issues

**DO NOT** create public GitHub issues for security vulnerabilities.

Instead, please email security reports to: [security@example.com]

Include:
1. Description of the vulnerability
2. Steps to reproduce
3. Potential impact
4. Suggested fix (if any)

### Security Response Process

1. **Acknowledgment**: Within 48 hours
2. **Investigation**: Within 1 week
3. **Fix Development**: Based on severity
4. **Disclosure**: Coordinated with reporter

### Severity Levels

| Level | Description | Response Time |
|-------|------------|--------------|
| Critical | Remote code execution, data breach | 24 hours |
| High | Authentication bypass, data exposure | 3 days |
| Medium | XSS, CSRF, session fixation | 1 week |
| Low | Information disclosure, DoS | 2 weeks |

## Security Checklist

### Development
- [ ] All user input is validated
- [ ] Passwords are hashed with bcrypt
- [ ] Sessions use secure settings
- [ ] Error messages don't leak sensitive info
- [ ] Dependencies are up to date
- [ ] Security headers are configured
- [ ] HTTPS is enforced in production
- [ ] Rate limiting is enabled
- [ ] Logs don't contain sensitive data

### Deployment
- [ ] Redis password is set
- [ ] Session secret is generated
- [ ] CORS origins are restricted
- [ ] Firewall rules are configured
- [ ] SSL certificates are valid
- [ ] Backups are encrypted
- [ ] Monitoring is enabled
- [ ] Incident response plan exists

### Regular Audits
- [ ] Weekly: Check for dependency updates
- [ ] Monthly: Review access logs
- [ ] Quarterly: Security assessment
- [ ] Yearly: Penetration testing

## Common Attack Vectors and Mitigations

### Brute Force Attacks
**Mitigation**:
- Rate limiting on login endpoint
- Account lockout after failures (future)
- CAPTCHA after failures (future)
- Strong password requirements

### Session Hijacking
**Mitigation**:
- HttpOnly cookies
- Secure flag (HTTPS)
- Session regeneration on login
- IP validation (optional)

### CSRF Attacks
**Mitigation**:
- SameSite cookie attribute
- Custom headers validation
- Origin checking

### Timing Attacks
**Mitigation**:
- Constant-time password comparison (bcrypt)
- Consistent error messages
- Random delays on auth failures

## Security Tools and Testing

### Recommended Security Tools

```bash
# Dependency scanning
go list -json -m all | nancy sleuth

# Static analysis
gosec ./...

# Docker scanning
docker scan task-tracker-backend:latest

# SSL testing
nmap --script ssl-cert,ssl-enum-ciphers -p 443 example.com

# API security testing
OWASP ZAP or Burp Suite
```

### Security Testing Checklist

```bash
# Test authentication
curl -X POST http://localhost:8080/api/v1/auth/login \
  -d '{"email":"test","password":"short"}'  # Should fail

# Test injection
curl -X POST http://localhost:8080/api/v1/tasks \
  -d '{"description":"<script>alert(1)</script>"}'  # Should be escaped

# Test authorization
# Try accessing another user's tasks (should fail)

# Test rate limiting
for i in {1..1001}; do
  curl http://localhost:8080/api/v1/tasks &
done
# Should get rate limited after 1000
```

## Compliance Considerations

### GDPR (If applicable)
- Right to deletion: Soft delete supports this
- Data portability: JSON export capability
- Consent: Explicit registration required
- Data minimization: Only essential data collected

### Security Standards
- **OWASP Top 10**: Addressed in implementation
- **CWE/SANS Top 25**: Mitigations in place
- **PCI DSS**: N/A (no payment processing)

## Incident Response

### Incident Response Plan

1. **Detection**: Monitor logs, alerts
2. **Containment**: Isolate affected systems
3. **Investigation**: Determine scope and cause
4. **Remediation**: Fix vulnerability, patch systems
5. **Recovery**: Restore normal operations
6. **Lessons Learned**: Document and improve

### Security Contacts

- Security Team: security@example.com
- On-call: [Phone number]
- Escalation: [Management contact]

## Security Updates

### Update Schedule
- **Critical**: Immediately
- **High**: Within 3 days
- **Medium**: Within 1 week
- **Low**: Next release cycle

### Dependency Management
```bash
# Check for vulnerabilities
go list -json -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy

# Verify updates
go test ./...
```