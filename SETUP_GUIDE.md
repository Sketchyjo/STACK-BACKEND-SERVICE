# Stack Service Setup Guide

This guide will help you set up and run the Stack Service application locally for development and testing.

## Prerequisites

1. **Go 1.21+**: Download from [https://golang.org/dl/](https://golang.org/dl/)
2. **PostgreSQL 14+**: Download from [https://postgresql.org/download/](https://postgresql.org/download/)
3. **Redis 6.0+**: Download from [https://redis.io/download](https://redis.io/download)
4. **Git**: For version control

## Installation

### 1. Clone the Repository

```bash
git clone <repository-url>
cd stack_service
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Database Setup

#### Create PostgreSQL Database

```sql
-- Connect to PostgreSQL as superuser
psql -U postgres

-- Create database and user
CREATE DATABASE stack_service;
CREATE USER stack_service_user WITH ENCRYPTED PASSWORD 'your_password_here';
GRANT ALL PRIVILEGES ON DATABASE stack_service TO stack_service_user;

-- Connect to the new database
\c stack_service

-- Grant schema permissions
GRANT ALL ON SCHEMA public TO stack_service_user;
```

### 4. Environment Configuration

Create a `.env` file in the project root:

```bash
# Application Environment
ENVIRONMENT=development
LOG_LEVEL=info
PORT=8080

# Database Configuration
DATABASE_URL=postgres://stack_service_user:your_password_here@localhost:5432/stack_service?sslmode=disable

# Security
JWT_SECRET=your-super-secret-jwt-key-change-in-production
ENCRYPTION_KEY=your-32-character-encryption-key-here

# Redis Configuration (Optional)
REDIS_HOST=localhost
REDIS_PORT=6379

# External Services (Development - uses mocks)
CIRCLE_API_KEY=
CIRCLE_ENVIRONMENT=sandbox

KYC_PROVIDER=mock
KYC_API_KEY=
KYC_API_SECRET=
KYC_CALLBACK_URL=http://localhost:8080/api/v1/webhooks/kyc

EMAIL_PROVIDER=mock
EMAIL_API_KEY=
BASE_URL=http://localhost:3000
```

### 5. Production Environment Variables

For production environments, set these environment variables:

```bash
# Required for production
export ENVIRONMENT=production
export DATABASE_URL=your-production-database-url
export JWT_SECRET=your-production-jwt-secret
export ENCRYPTION_KEY=your-production-encryption-key

# Circle API (Production)
export CIRCLE_API_KEY=your-circle-production-api-key
export CIRCLE_ENVIRONMENT=production

# KYC Provider (Jumio Example)
export KYC_PROVIDER=jumio
export KYC_API_KEY=your-jumio-api-key
export KYC_API_SECRET=your-jumio-api-secret
export KYC_CALLBACK_URL=https://yourapi.com/api/v1/webhooks/kyc

# Email Service (SendGrid Example)
export EMAIL_PROVIDER=sendgrid
export EMAIL_API_KEY=your-sendgrid-api-key
export BASE_URL=https://yourapp.com
```

## Running the Application

### 1. Start Dependencies

#### PostgreSQL
```bash
# macOS with Homebrew
brew services start postgresql

# Ubuntu/Debian
sudo systemctl start postgresql

# Docker
docker run --name postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres:14
```

#### Redis (Optional)
```bash
# macOS with Homebrew
brew services start redis

# Ubuntu/Debian
sudo systemctl start redis

# Docker
docker run --name redis -p 6379:6379 -d redis:6-alpine
```

### 2. Run Database Migrations

The application will automatically run migrations on startup, or you can run them manually:

```bash
# Using the application (recommended)
go run cmd/main.go

# Or using golang-migrate CLI
migrate -path migrations -database "postgres://stack_service_user:password@localhost/stack_service?sslmode=disable" up
```

### 3. Start the Application

#### Development Mode
```bash
go run cmd/main.go
```

#### Production Mode
```bash
go build -o stack_service cmd/main.go
./stack_service
```

### 4. Verify Installation

Check if the application is running:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
    "status": "ok",
    "timestamp": "2025-01-03T17:04:21Z",
    "version": "1.0.0",
    "environment": "development"
}
```

## Development Tools

### 1. Hot Reload (Optional)

Install Air for hot reloading during development:

```bash
go install github.com/cosmtrek/air@latest
air
```

### 2. Database GUI (Optional)

Recommended PostgreSQL GUI tools:
- **pgAdmin**: [https://www.pgadmin.org/](https://www.pgadmin.org/)
- **DBeaver**: [https://dbeaver.io/](https://dbeaver.io/)
- **TablePlus**: [https://tableplus.com/](https://tableplus.com/)

### 3. API Documentation

The API documentation will be available at:
- Swagger UI: `http://localhost:8080/swagger/index.html`

## Testing

### 1. Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/domain/services/onboarding/...
```

### 2. Integration Tests

```bash
# Run integration tests (requires database)
go test -tags=integration ./...
```

### 3. API Testing

Use the provided Postman collection and follow the [POSTMAN_TESTING_GUIDE.md](POSTMAN_TESTING_GUIDE.md) for comprehensive API testing.

## Configuration

### Application Configuration

The application uses a hierarchical configuration system:

1. **Default values** (hardcoded)
2. **Configuration file** (`config.yaml`)
3. **Environment variables** (highest priority)

### Example config.yaml

Create a `config.yaml` file in the project root:

```yaml
environment: development
log_level: info

server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30
  write_timeout: 30
  rate_limit_per_min: 100

database:
  host: localhost
  port: 5432
  name: stack_service
  user: stack_service_user
  password: your_password_here
  ssl_mode: disable
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 300

jwt:
  secret: your-jwt-secret-here
  access_token_ttl: 3600
  refresh_token_ttl: 2592000
  issuer: stack_service

security:
  encryption_key: your-32-character-encryption-key
  max_login_attempts: 5
  lockout_duration: 900
  require_mfa: false
  password_min_length: 8

circle:
  api_key: ""
  environment: sandbox

kyc:
  provider: mock
  api_key: ""
  api_secret: ""
  environment: development
  callback_url: http://localhost:8080/api/v1/webhooks/kyc

email:
  provider: mock
  api_key: ""
  from_email: no-reply@stackservice.com
  from_name: "Stack Service"
  base_url: http://localhost:3000
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Error
```
Error: failed to connect to database: dial tcp [::1]:5432: connect: connection refused
```

**Solution**: Ensure PostgreSQL is running and the connection parameters are correct.

#### 2. Migration Errors
```
Error: failed to run migrations: no change
```

**Solution**: This is normal if migrations are already applied. The application will continue to start.

#### 3. Port Already in Use
```
Error: bind: address already in use
```

**Solution**: Either stop the process using port 8080 or change the port in configuration.

#### 4. JWT Secret Missing
```
Error: JWT secret is required
```

**Solution**: Set the `JWT_SECRET` environment variable or add it to your config file.

### Logs and Debugging

The application uses structured logging. Set `LOG_LEVEL=debug` for detailed logs:

```bash
export LOG_LEVEL=debug
go run cmd/main.go
```

### Database Reset

To reset the database during development:

```bash
# Drop and recreate database
psql -U postgres -c "DROP DATABASE IF EXISTS stack_service;"
psql -U postgres -c "CREATE DATABASE stack_service;"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE stack_service TO stack_service_user;"

# Restart application (will run migrations)
go run cmd/main.go
```

## Production Deployment

### 1. Build for Production

```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o stack_service cmd/main.go
```

### 2. Docker Deployment

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o stack_service cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/stack_service .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./stack_service"]
```

Build and run:

```bash
docker build -t stack_service .
docker run -p 8080:8080 -e DATABASE_URL=your-db-url stack_service
```

### 3. Environment-Specific Configurations

Use different configuration files or environment variables for different environments:

- **Development**: Use mock services, detailed logging
- **Staging**: Use sandbox APIs, moderate logging  
- **Production**: Use production APIs, minimal logging

## Security Considerations

1. **JWT Secrets**: Use strong, unique secrets for each environment
2. **Database**: Use strong passwords and enable SSL in production
3. **Encryption Keys**: Use 32-character random keys for encryption
4. **API Keys**: Store securely and rotate regularly
5. **HTTPS**: Always use HTTPS in production
6. **Rate Limiting**: Configure appropriate rate limits
7. **CORS**: Configure allowed origins properly

This setup guide should help you get the Stack Service running locally and provide guidance for production deployment. Follow the Postman testing guide to verify everything is working correctly.