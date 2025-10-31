# Contributing to Stack Service

Thank you for your interest in contributing to Stack Service! This document provides guidelines for developers who want to contribute to the project.

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- PostgreSQL 15+
- Redis 7+
- Git
- Basic understanding of blockchain concepts
- Familiarity with Go, REST APIs, and database design

### Development Environment Setup

1. **Fork the repository** on GitHub

2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/stack_service.git
   cd stack_service
   ```

3. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/original-org/stack_service.git
   ```

4. **Set up local environment**
   ```bash
   cp configs/config.yaml.example configs/config.yaml
   # Edit config.yaml with your local settings
   ```

5. **Install dependencies**
   ```bash
   go mod download
   ```

6. **Start services**
   ```bash
   docker-compose up -d db redis
   ```

7. **Run migrations**
   ```bash
   go run cmd/main.go migrate
   ```

8. **Start the application**
   ```bash
   go run cmd/main.go
   ```

## 📋 Development Guidelines

### Code Style and Standards

#### Go Style
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` and `goimports` for formatting
- Use `golangci-lint` for linting
- Follow Go naming conventions
- Use meaningful package, variable, and function names

#### Project Structure
```
internal/
├── api/              # HTTP layer
│   ├── handlers/     # Request handlers
│   ├── middleware/   # HTTP middleware
│   └── routes/       # Route definitions
├── domain/           # Business logic
│   ├── entities/     # Data models
│   ├── repositories/ # Data access interfaces
│   └── services/     # Business services
└── infrastructure/   # External dependencies
    ├── database/     # Database implementations
    ├── blockchain/   # Blockchain clients
    └── config/       # Configuration
```

#### Clean Architecture Principles
- **Dependencies point inward**: Higher layers can depend on lower layers, not vice versa
- **Repository pattern**: Abstract data access behind interfaces
- **Dependency injection**: Use interfaces and inject dependencies
- **Separation of concerns**: Keep business logic separate from infrastructure

### Testing Standards

#### Test Coverage
- Aim for **80%+ test coverage**
- All business logic must have unit tests
- Critical paths must have integration tests
- New features must include tests

#### Test Structure
```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := &mocks.UserRepository{}
    service := NewUserService(mockRepo)
    
    // Act
    result, err := service.CreateUser(ctx, userInput)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockRepo.AssertExpectations(t)
}
```

#### Test Types
- **Unit tests**: Test individual functions/methods in isolation
- **Integration tests**: Test component interactions
- **E2E tests**: Test complete user workflows

### Security Guidelines

#### Sensitive Data
- **Never log sensitive data** (passwords, private keys, tokens)
- **Encrypt sensitive data** at rest using AES-256-GCM
- **Use secure random generators** for tokens and secrets
- **Validate all inputs** to prevent injection attacks

#### Authentication & Authorization
- **Always validate JWT tokens** in protected endpoints
- **Use role-based access control** (RBAC)
- **Implement rate limiting** to prevent abuse
- **Log all security-related events** for auditing

#### Blockchain Security
- **Encrypt private keys** before storing in database
- **Validate all blockchain addresses** before use
- **Use proper gas estimation** to prevent failed transactions
- **Implement transaction monitoring** for security events

## 🔄 Development Workflow

### Branch Strategy
- `main`: Production-ready code
- `develop`: Integration branch for features
- `feature/feature-name`: Individual features
- `fix/bug-name`: Bug fixes
- `hotfix/critical-fix`: Critical production fixes

### Workflow Steps

1. **Create a feature branch**
   ```bash
   git checkout develop
   git pull upstream develop
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Follow coding standards
   - Write tests
   - Update documentation

3. **Test your changes**
   ```bash
   # Run tests
   go test ./...
   
   # Check coverage
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   
   # Run linting
   golangci-lint run
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add user registration endpoint"
   ```

5. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   ```
   Then create a Pull Request on GitHub.

### Commit Message Convention

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(auth): add JWT token refresh endpoint
fix(wallet): resolve balance calculation error
docs(api): update swagger documentation
test(user): add unit tests for user service
```

## 🧪 Testing

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/domain/services

# With coverage
go test -coverprofile=coverage.out ./...

# Verbose output
go test -v ./...

# Integration tests
go test ./tests/integration/...

# E2E tests
go test ./tests/e2e/...
```

### Writing Tests

#### Unit Test Example
```go
func TestWalletService_CreateWallet(t *testing.T) {
    tests := []struct {
        name           string
        input          CreateWalletInput
        mockSetup      func(*mocks.WalletRepository)
        expectedError  string
        expectedResult *entities.Wallet
    }{
        {
            name: "successful wallet creation",
            input: CreateWalletInput{
                UserID:  testUserID,
                Network: "ethereum",
                Name:    "My Ethereum Wallet",
            },
            mockSetup: func(m *mocks.WalletRepository) {
                m.On("Create", mock.AnythingOfType("*entities.Wallet")).
                  Return(nil)
            },
            expectedResult: &entities.Wallet{
                UserID:  testUserID,
                Network: "ethereum",
                Name:    "My Ethereum Wallet",
            },
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### Integration Test Example
```go
func TestUserRegistration_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer teardownTestDB(t, db)
    
    // Create test server
    server := setupTestServer(t, db)
    defer server.Close()
    
    // Test the endpoint
    payload := `{"email":"test@example.com","password":"password123"}`
    resp := httptest.NewRequest("POST", "/api/v1/auth/register", 
        strings.NewReader(payload))
    
    w := httptest.NewRecorder()
    server.ServeHTTP(w, resp)
    
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

## 📖 Documentation

### Code Documentation
- **Document all public functions** with Go doc comments
- **Include examples** in documentation when helpful
- **Document complex algorithms** with inline comments
- **Keep documentation up to date** with code changes

### API Documentation
- **Use Swagger annotations** for API endpoints
- **Include request/response examples**
- **Document error responses**
- **Update Swagger docs** with new endpoints

Example:
```go
// CreateUser creates a new user account
// @Summary Create user account
// @Description Create a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User creation data"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse "User already exists"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) CreateUser(c *gin.Context) {
    // Implementation
}
```

## 🔍 Code Review Process

### Before Submitting
- [ ] All tests pass
- [ ] Code is properly formatted
- [ ] Documentation is updated
- [ ] No security vulnerabilities
- [ ] Performance implications considered

### Review Criteria
- **Correctness**: Does the code work as intended?
- **Security**: Are there any security vulnerabilities?
- **Performance**: Will this impact system performance?
- **Maintainability**: Is the code easy to understand and modify?
- **Testing**: Are there adequate tests?
- **Documentation**: Is the code properly documented?

### Reviewer Guidelines
- Be constructive and respectful
- Focus on the code, not the person
- Provide specific suggestions for improvements
- Approve only when confident in the changes

## 🚀 Deployment

### Environment Setup
- **Development**: Local development environment
- **Staging**: Pre-production testing environment
- **Production**: Live production environment

### Deployment Process
1. Code reviewed and approved
2. Merged to `develop` branch
3. Automated tests run
4. Deployed to staging for testing
5. Approved for production deployment
6. Tagged release created

## 🐛 Issue Reporting

### Bug Reports
When reporting bugs, include:
- **Environment details** (OS, Go version, etc.)
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Error messages and logs**
- **Minimal code example** if applicable

### Feature Requests
When requesting features, include:
- **Use case description**
- **Proposed solution**
- **Alternative considerations**
- **Impact assessment**

## 📞 Getting Help

### Communication Channels
- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Code Reviews**: PR comments for implementation feedback

### Resources
- [Go Documentation](https://golang.org/doc/)
- [Gin Framework Documentation](https://gin-gonic.com/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Blockchain Development Resources](https://ethereum.org/developers/)

## ✨ Recognition

Contributors will be recognized in:
- `CONTRIBUTORS.md` file
- Release notes for significant contributions
- Project documentation where applicable

Thank you for contributing to Stack Service! 🚀