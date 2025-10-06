# Stack Service

A professional GenZ Web3 multi-chain investment platform API that allows users to fund accounts with stablecoins from different chains, create curated investment baskets, use debit cards, and engage in copy trading.

## 🚀 Features

- **Multi-Chain Wallet Support**: Ethereum, Polygon, Binance Smart Chain, and more
- **Stablecoin Integration**: USDC, USDT, BUSD across multiple networks
- **Investment Baskets**: Curated and custom portfolios with automatic rebalancing
- **Copy Trading**: Follow top traders and mirror their strategies
- **Debit Cards**: Virtual and physical cards linked to crypto portfolios
- **Advanced Security**: JWT authentication, 2FA, encryption, audit trails
- **Real-time Analytics**: Portfolio tracking, performance metrics, risk analysis
- **RESTful API**: Comprehensive API with Swagger documentation
- **Scalable Architecture**: Clean architecture with repository pattern

## 🏗️ Architecture

```
stack_service/
├── cmd/                    # Application entry points
│   └── main.go
├── internal/               # Private application code
│   ├── api/               # API layer
│   │   ├── handlers/      # HTTP request handlers
│   │   ├── middleware/    # HTTP middleware
│   │   └── routes/        # Route definitions
│   ├── domain/            # Business domain
│   │   ├── entities/      # Domain entities/models
│   │   ├── repositories/  # Repository interfaces
│   │   └── services/      # Business logic services
│   └── infrastructure/    # External concerns
│       ├── blockchain/    # Blockchain integrations
│       ├── database/      # Database connections
│       ├── config/        # Configuration management
│       └── cache/         # Caching layer
├── pkg/                   # Public libraries
│   ├── auth/             # Authentication utilities
│   ├── crypto/           # Cryptographic functions
│   ├── logger/           # Logging utilities
│   └── utils/            # General utilities
├── migrations/           # Database migrations
├── configs/              # Configuration files
├── deployments/          # Deployment configurations
├── scripts/              # Build and deployment scripts
└── tests/                # Test files
    ├── unit/             # Unit tests
    ├── integration/      # Integration tests
    └── e2e/              # End-to-end tests
```

## 🛠️ Technology Stack

- **Language**: Go 1.21
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Authentication**: JWT tokens
- **Blockchain**: Ethereum, Polygon, BSC
- **Containerization**: Docker & Docker Compose
- **Documentation**: Swagger/OpenAPI
- **Testing**: Go testing, Testify
- **Monitoring**: Prometheus, Grafana

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15
- Redis 7
- Git

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/your-org/stack_service.git
cd stack_service
```

2. **Copy configuration**
```bash
cp configs/config.yaml.example configs/config.yaml
```

3. **Edit configuration**
Update the configuration file with your settings:
- Database credentials
- JWT secret
- Encryption key
- Blockchain RPC endpoints
- API keys for external services

4. **Start with Docker Compose**
```bash
# Basic services
docker-compose up -d

# With admin tools (pgAdmin, RedisInsight)
docker-compose --profile admin up -d

# With monitoring (Prometheus, Grafana)
docker-compose --profile monitoring up -d
```

5. **Run database migrations**
```bash
# Migrations run automatically on startup
# To run manually:
go run cmd/main.go migrate
```

### Development Setup

1. **Install dependencies**
```bash
go mod download
```

2. **Set environment variables**
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/stack_service_dev?sslmode=disable"
export JWT_SECRET="your-super-secret-jwt-key"
export ENCRYPTION_KEY="your-32-byte-encryption-key"
```

3. **Run the application**
```bash
go run cmd/main.go
```

4. **Access the API**
- API: http://localhost:8080
- Health: http://localhost:8080/health
- Swagger: http://localhost:8080/swagger/index.html
- Metrics: http://localhost:8080/metrics

## 📚 API Documentation

### Authentication

All protected endpoints require a JWT token in the Authorization header:

```bash
Authorization: Bearer <your-jwt-token>
```

### Key Endpoints

#### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout

#### Wallets
- `GET /api/v1/wallets` - Get user wallets
- `POST /api/v1/wallets` - Create new wallet
- `GET /api/v1/wallets/{id}/balance` - Get wallet balance

#### Investment Baskets
- `GET /api/v1/baskets` - Get user baskets
- `POST /api/v1/baskets` - Create custom basket
- `GET /api/v1/curated/baskets` - Get curated baskets
- `POST /api/v1/baskets/{id}/invest` - Invest in basket

#### Copy Trading
- `GET /api/v1/copy/traders` - Get top traders
- `POST /api/v1/copy/traders/{id}/follow` - Follow trader

#### Cards
- `GET /api/v1/cards` - Get user cards
- `POST /api/v1/cards` - Create virtual card
- `POST /api/v1/cards/{id}/freeze` - Freeze card

### Complete API documentation is available at `/swagger/index.html` when running the server.

## 🧪 Testing

### Unit Tests
```bash
go test ./...
```

### Integration Tests
```bash
go test ./tests/integration/...
```

### End-to-End Tests
```bash
go test ./tests/e2e/...
```

### Test Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🔒 Security

### Security Features
- JWT authentication with refresh tokens
- Password hashing with bcrypt
- AES-256-GCM encryption for sensitive data
- Rate limiting
- CORS protection
- Security headers
- Input validation and sanitization
- Audit logging
- Session management

### Security Best Practices
- All sensitive data is encrypted at rest
- Private keys are encrypted before storage
- API rate limiting prevents abuse
- Comprehensive audit trails
- Two-factor authentication support
- IP whitelisting for admin endpoints

## 🔧 Configuration

Key configuration options in `configs/config.yaml`:

```yaml
# Server configuration
server:
  port: 8080
  host: 0.0.0.0
  rate_limit_per_min: 100

# Database configuration
database:
  host: localhost
  port: 5432
  name: stack_service
  user: postgres
  password: postgres

# JWT configuration
jwt:
  secret: "your-secret-key"
  access_token_ttl: 3600
  refresh_token_ttl: 2592000

# Blockchain networks
blockchain:
  networks:
    ethereum:
      chain_id: 1
      rpc: "https://eth-mainnet.alchemyapi.io/v2/YOUR-API-KEY"
      # ... more networks
```

## 🚀 Deployment

### Docker Deployment

1. **Build production image**
```bash
docker build -t stack_service:latest .
```

2. **Run container**
```bash
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  -e JWT_SECRET="..." \
  stack_service:latest
```

### Kubernetes Deployment

Kubernetes manifests are available in the `deployments/` directory:

```bash
kubectl apply -f deployments/k8s/
```

### Cloud Deployment

The application is cloud-ready and can be deployed on:
- AWS ECS/EKS
- Google Cloud Run/GKE  
- Azure Container Instances/AKS
- DigitalOcean App Platform

## 📊 Monitoring & Observability

### Health Checks
- `GET /health` - Application health
- `GET /metrics` - Prometheus metrics

### Logging
- Structured logging with Zap
- Request/response logging
- Error tracking with stack traces
- Audit trail logging

### Metrics
- HTTP request metrics
- Database connection metrics
- Business metrics (transactions, users, etc.)
- Custom application metrics

### Monitoring Stack
- **Prometheus**: Metrics collection
- **Grafana**: Visualization dashboards
- **AlertManager**: Alert notifications

## 🤝 Contributing

See [CONTRIBUTING.md](./docs/CONTRIBUTING.md) for detailed contribution guidelines.

### Development Workflow
1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Follow coding standards (see below)
4. Write tests for new functionality
5. Ensure all tests pass
6. Create pull request

### Coding Standards
- Follow Go conventions and best practices
- Use meaningful variable and function names
- Write comprehensive tests
- Document public APIs
- Follow the established project structure
- Use dependency injection
- Handle errors appropriately

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: GitHub Issues
- **Documentation**: `/docs` directory
- **API Docs**: Swagger UI at `/swagger`
- **Community**: GitHub Discussions

## 🛣️ Roadmap

### Phase 1 (Current)
- [x] Basic authentication and user management
- [x] Multi-chain wallet integration
- [x] Investment baskets foundation
- [ ] Copy trading implementation
- [ ] Debit card integration

### Phase 2
- [ ] Advanced portfolio analytics
- [ ] Mobile app API
- [ ] DeFi protocol integrations
- [ ] Yield farming strategies
- [ ] NFT portfolio tracking

### Phase 3
- [ ] AI-powered investment recommendations
- [ ] Social trading features
- [ ] Institutional features
- [ ] Options and derivatives
- [ ] Cross-chain bridge integration

---

**Built with ❤️ for the GenZ Web3 community**