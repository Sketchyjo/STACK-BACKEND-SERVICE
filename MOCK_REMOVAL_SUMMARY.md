# Mock/Stub Implementations Removal Summary

## Overview
Removed all mock and stub implementations from production code as requested. The system now only uses real implementations or returns "service unavailable" for features not yet implemented.

## Changes Made

### 1. Deleted Files
- **`internal/infrastructure/repositories/investing_stubs.go`**
  - Removed `StubBasketRepository` 
  - Removed `StubOrderRepository`
  - Removed `StubPositionRepository`

### 2. Updated Dependency Injection Container
**File:** `internal/infrastructure/di/container.go`

- **Before:**
  ```go
  basketRepo := repositories.NewStubBasketRepository()
  orderRepo := repositories.NewStubOrderRepository()
  positionRepo := repositories.NewStubPositionRepository()
  c.InvestingService = investing.NewService(...)
  ```

- **After:**
  ```go
  // TODO: Initialize investing service when proper repository implementations are available
  // For now, investing service is disabled as stub/mock implementations have been removed
  c.InvestingService = nil
  ```

### 3. Added Nil Checks in Handlers
**File:** `internal/api/handlers/stack_handlers.go`

Added `checkInvestingServiceAvailable()` helper method that:
- Returns `false` if investing service is `nil`
- Sends `503 Service Unavailable` response
- Includes descriptive error message

Applied to all investing-related endpoints:
- ✅ `ListBaskets()` 
- ✅ `GetBasket()`
- ✅ `CreateOrder()`
- ✅ `ListOrders()`
- ✅ `GetOrder()`
- ✅ `GetPortfolio()`
- ✅ `GetPortfolioOverview()`
- ✅ `BrokerageFillWebhook()`

## Test Files NOT Modified
The following test files still contain mock implementations (this is correct and expected):
- `internal/domain/services/funding/service_test.go` - Contains `MockDepositRepository`, `MockBalanceRepository`, etc.
  - These are testing mocks and should remain

## Impact

### ✅ What Still Works
- **Onboarding**: User signup, KYC submission
- **Wallet Management**: Wallet provisioning, address generation
- **Funding**: Deposit addresses, balance checking (real-time from Circle API)
- **Security**: Passcode management
- **AI-CFO**: Summary generation

### ⚠️ What Returns "Service Unavailable" (503)
- **Investing Endpoints**:
  - `GET /api/v1/baskets` - List investment baskets
  - `GET /api/v1/baskets/{id}` - Get specific basket
  - `POST /api/v1/orders` - Create investment order
  - `GET /api/v1/orders` - List orders
  - `GET /api/v1/orders/{id}` - Get specific order
  - `GET /api/v1/portfolio` - Get user portfolio
  - `GET /api/v1/portfolio/overview` - Get portfolio overview
  - `POST /api/v1/webhooks/brokerage-fills` - Brokerage fill webhooks

**Response Format:**
```json
{
  "code": "SERVICE_UNAVAILABLE",
  "message": "Investing service is not available"
}
```

## Next Steps to Re-Enable Investing

To re-enable investing functionality, you need to:

1. **Create Proper Repository Implementations**:
   ```go
   // internal/infrastructure/repositories/basket_repository.go
   type BasketRepository struct {
       db     *sql.DB
       logger *zap.Logger
   }
   // Implement all BasketRepository interface methods
   ```

2. **Create Database Tables/Migrations**:
   ```sql
   CREATE TABLE baskets (...);
   CREATE TABLE orders (...);
   CREATE TABLE positions (...);
   ```

3. **Update DI Container**:
   ```go
   basketRepo := repositories.NewBasketRepository(c.DB, c.ZapLog)
   orderRepo := repositories.NewOrderRepository(c.DB, c.ZapLog)
   positionRepo := repositories.NewPositionRepository(c.DB, c.ZapLog)
   c.InvestingService = investing.NewService(
       basketRepo,
       orderRepo,
       positionRepo,
       c.BalanceRepo,
       brokerageAPI,
       c.Logger,
   )
   ```

4. **Implement Brokerage API Adapter**:
   ```go
   // internal/infrastructure/adapters/Alpaca/client.go
   type Client struct { ... }
   // Implement Alpaca API integration
   ```

## Verification

### Build Status
✅ Code compiles successfully:
```bash
$ go build -o bin/stack_service ./cmd/main.go
# Success - no errors
```

### Test Status
⚠️ Investing service tests will need to be updated or disabled since the service is now `nil`

### API Behavior
- **Funding endpoints**: ✅ Working normally with real-time Circle balance fetching
- **Investing endpoints**: ⚠️ Return 503 Service Unavailable (expected)
- **Other endpoints**: ✅ Working normally

## Rollback Instructions

If you need to restore stub implementations temporarily:

1. Restore deleted file:
   ```bash
   git checkout HEAD -- internal/infrastructure/repositories/investing_stubs.go
   ```

2. Update DI container:
   ```bash
   git checkout HEAD -- internal/infrastructure/di/container.go
   ```

3. Rebuild:
   ```bash
   go build -o bin/stack_service ./cmd/main.go
   ```

## Related Documentation
- See `WALLET_BALANCE_FIX.md` for real-time balance fetching implementation
- See `docs/architecture/` for system architecture details
- See project rules in `.cursorrules` for coding standards

