# Alpaca Broker API Integration

## Overview

This document describes the Alpaca Broker API integration for the STACK service. The integration provides full brokerage capabilities including account management, trading, asset data, and market news.

## Architecture

The Alpaca integration follows STACK's adapter pattern with clean separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│  (Investing Service, Handlers, GraphQL Resolvers)           │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              BrokerageAdapter Interface                     │
│         (internal/infrastructure/adapters/)                 │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                  Alpaca Client                              │
│           (internal/adapters/alpaca/)                       │
│  - Circuit Breaker (gobreaker)                              │
│  - Exponential Backoff Retry                                │
│  - Structured Logging (Zap)                                 │
│  - TLS 1.2+ Security                                        │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Alpaca Broker API                              │
│  - Base URL: https://broker-api.sandbox.alpaca.markets      │
│  - Data URL: https://data.alpaca.markets                    │
│  - Auth: Basic Auth (API Key + Secret)                      │
└─────────────────────────────────────────────────────────────┘
```

## Configuration

### Environment Variables

Add the following to your `.env` file:

```bash
# Alpaca Broker API Configuration
ALPACA_API_KEY="YOUR_ALPACA_API_KEY_HERE"
ALPACA_SECRET_KEY="YOUR_ALPACA_SECRET_KEY_HERE"
ALPACA_BASE_URL=https://broker-api.sandbox.alpaca.markets
ALPACA_DATA_BASE_URL=https://data.alpaca.markets
ALPACA_ENVIRONMENT=sandbox # sandbox or production
ALPACA_TIMEOUT=30 # Request timeout in seconds
```

### Getting API Credentials

1. Sign up for an Alpaca Broker account at https://alpaca.markets
2. Navigate to your dashboard
3. Generate API key and secret for sandbox/production environment
4. Copy credentials to your `.env` file

**⚠️ IMPORTANT**: Never commit your API keys to version control!

## Features

### 1. Account Management

#### Create Account
```go
req := &entities.AlpacaCreateAccountRequest{
    Contact: entities.AlpacaContact{
        EmailAddress:  "user@example.com",
        PhoneNumber:   "+12125551234",
        StreetAddress: []string{"123 Main St", "Apt 4B"},
        City:          "New York",
        State:         "NY",
        PostalCode:    "10001",
    },
    Identity: entities.AlpacaIdentity{
        GivenName:   "John",
        FamilyName:  "Doe",
        DateOfBirth: "1990-01-01",
        TaxID:       "123-45-6789",
        TaxIDType:   "USA_SSN",
        CountryOfCitizenship: "USA",
        FundingSource: []string{"employment_income"},
    },
    Disclosures: entities.AlpacaDisclosures{
        IsControlPerson:             false,
        IsAffiliatedExchangeOrFINRA: false,
        IsPoliticallyExposed:        false,
        EmploymentStatus:            "employed",
    },
    Agreements: []entities.AlpacaAgreement{
        {
            Agreement: "account",
            SignedAt:  time.Now().Format(time.RFC3339),
            IPAddress: "192.168.1.1",
        },
    },
}

account, err := alpacaClient.CreateAccount(ctx, req)
```

#### Get Account
```go
account, err := alpacaClient.GetAccount(ctx, "account-id")
```

#### List Accounts
```go
query := map[string]string{
    "status": "ACTIVE",
    "limit":  "100",
}
accounts, err := alpacaClient.ListAccounts(ctx, query)
```

### 2. Trading

#### Place Market Order
```go
qty := decimal.NewFromFloat(10.5)
req := &entities.AlpacaCreateOrderRequest{
    Symbol:      "AAPL",
    Qty:         &qty,
    Side:        entities.AlpacaOrderSideBuy,
    Type:        entities.AlpacaOrderTypeMarket,
    TimeInForce: entities.AlpacaTimeInForceDay,
}

order, err := alpacaClient.CreateOrder(ctx, "account-id", req)
```

#### Place Limit Order
```go
qty := decimal.NewFromFloat(5)
limitPrice := decimal.NewFromFloat(150.50)

req := &entities.AlpacaCreateOrderRequest{
    Symbol:      "AAPL",
    Qty:         &qty,
    Side:        entities.AlpacaOrderSideBuy,
    Type:        entities.AlpacaOrderTypeLimit,
    TimeInForce: entities.AlpacaTimeInForceGTC,
    LimitPrice:  &limitPrice,
}

order, err := alpacaClient.CreateOrder(ctx, "account-id", req)
```

#### Fractional Shares (Dollar-Based Order)
```go
notional := decimal.NewFromFloat(100) // $100 worth

req := &entities.AlpacaCreateOrderRequest{
    Symbol:      "TSLA",
    Notional:    &notional,
    Side:        entities.AlpacaOrderSideBuy,
    Type:        entities.AlpacaOrderTypeMarket,
    TimeInForce: entities.AlpacaTimeInForceDay,
}

order, err := alpacaClient.CreateOrder(ctx, "account-id", req)
```

#### Get Order Status
```go
order, err := alpacaClient.GetOrder(ctx, "account-id", "order-id")
```

#### List Orders
```go
query := map[string]string{
    "status": "open",
    "limit":  "50",
}
orders, err := alpacaClient.ListOrders(ctx, "account-id", query)
```

#### Cancel Order
```go
err := alpacaClient.CancelOrder(ctx, "account-id", "order-id")
```

### 3. Assets

#### Get Asset by Symbol
```go
asset, err := alpacaClient.GetAsset(ctx, "AAPL")
```

#### List All Tradable Assets
```go
query := map[string]string{
    "status":   "active",
    "tradable": "true",
}
assets, err := alpacaClient.ListAssets(ctx, query)
```

#### Filter Assets
```go
query := map[string]string{
    "status":      "active",
    "asset_class": "us_equity",
    "exchange":    "NASDAQ",
}
assets, err := alpacaClient.ListAssets(ctx, query)
```

### 4. Positions

#### Get Position
```go
position, err := alpacaClient.GetPosition(ctx, "account-id", "AAPL")
```

#### List All Positions
```go
positions, err := alpacaClient.ListPositions(ctx, "account-id")
```

### 5. Market Data (News)

#### Get Latest News
```go
req := &entities.AlpacaNewsRequest{
    Limit: 20,
    Sort:  "DESC",
}
news, err := alpacaClient.GetNews(ctx, req)
```

#### Get News for Specific Symbols
```go
req := &entities.AlpacaNewsRequest{
    Symbols: []string{"AAPL", "TSLA", "MSFT"},
    Limit:   50,
    Sort:    "DESC",
    IncludeContent: true,
}
news, err := alpacaClient.GetNews(ctx, req)
```

#### Get News for Date Range
```go
start := time.Now().AddDate(0, 0, -7) // 7 days ago
end := time.Now()

req := &entities.AlpacaNewsRequest{
    Start:  &start,
    End:    &end,
    Limit:  100,
}
news, err := alpacaClient.GetNews(ctx, req)
```

## Resilience Features

### Circuit Breaker

The Alpaca client implements a circuit breaker pattern using `gobreaker`:

- **Max Requests**: 5 concurrent requests when half-open
- **Interval**: 10 seconds
- **Timeout**: 30 seconds before auto-retry
- **Threshold**: Opens after 5 consecutive failures

### Retry Logic

Automatic exponential backoff retry with:

- **Max Retries**: 3 attempts
- **Base Backoff**: 1 second
- **Max Backoff**: 16 seconds
- **Jitter**: ±10% randomization

Retries are triggered on:
- Rate limit errors (429)
- Server errors (5xx)
- Network timeouts
- Connection errors

### Error Handling

All errors are properly wrapped with context:

```go
account, err := alpacaClient.GetAccount(ctx, accountID)
if err != nil {
    if apiErr, ok := err.(*entities.AlpacaErrorResponse); ok {
        // Handle Alpaca API error
        log.Error("Alpaca API error",
            zap.Int("code", apiErr.Code),
            zap.String("message", apiErr.Message))
    } else {
        // Handle other errors (network, timeout, etc.)
        log.Error("Request failed", zap.Error(err))
    }
}
```

## Security

### Authentication

Alpaca uses Basic Authentication:
- Username: API Key
- Password: API Secret

Credentials are sent via HTTP Basic Auth header on every request.

### TLS

- Minimum TLS version: 1.2
- All requests use HTTPS
- Certificate validation enforced

### Secrets Management

- **NEVER** hardcode credentials
- Use environment variables or AWS Secrets Manager
- **NEVER** log API keys or secrets
- Rotate keys regularly

## Observability

### Logging

All operations are logged with structured logging (Zap):

```go
logger.Info("Creating Alpaca order",
    zap.String("account_id", accountID),
    zap.String("symbol", "AAPL"),
    zap.String("side", "buy"),
    zap.String("type", "market"))
```

### Metrics

Circuit breaker state changes are logged:

```
Circuit breaker state changed: closed -> open
Circuit breaker state changed: open -> half-open
Circuit breaker state changed: half-open -> closed
```

### Tracing

All requests include context propagation for distributed tracing with OpenTelemetry.

## Testing

### Unit Tests

Run unit tests:
```bash
go test ./internal/adapters/alpaca/... -v
```

### Integration Tests

Run integration tests against Alpaca sandbox:
```bash
go test ./test/integration/alpaca/... -v
```

**Note**: Integration tests require valid sandbox credentials.

## Error Handling Best Practices

### Check All Errors

```go
account, err := alpacaClient.GetAccount(ctx, accountID)
if err != nil {
    return fmt.Errorf("failed to get account: %w", err)
}
```

### Handle Specific Error Cases

```go
order, err := alpacaClient.CreateOrder(ctx, accountID, req)
if err != nil {
    if apiErr, ok := err.(*entities.AlpacaErrorResponse); ok {
        switch apiErr.Code {
        case 403:
            return errors.New("insufficient buying power")
        case 422:
            return errors.New("invalid order parameters")
        case 429:
            return errors.New("rate limit exceeded, retry later")
        default:
            return fmt.Errorf("order failed: %s", apiErr.Message)
        }
    }
    return fmt.Errorf("order request failed: %w", err)
}
```

### Use Context Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

account, err := alpacaClient.GetAccount(ctx, accountID)
```

## Rate Limits

Alpaca enforces rate limits on their API:

- **Trading API**: 200 requests/minute
- **Market Data API**: Varies by subscription tier

The client automatically handles rate limits with exponential backoff retry.

## Limitations & TODOs

### Current Implementation

✅ Account management (create, get, list)
✅ Order management (create, get, list, cancel)
✅ Asset data (get, list with filtering)
✅ Position tracking (get, list)
✅ Market news (get with filtering)
✅ Circuit breaker pattern
✅ Retry with exponential backoff
✅ Structured logging
✅ TLS 1.2+ security

### Future Enhancements

🔲 Basket order implementation in BrokerageAdapter
🔲 Real-time order status updates via webhooks
🔲 Advanced order types (bracket, OCO, OTO)
🔲 Portfolio analytics
🔲 Options trading support
🔲 Streaming market data (WebSocket)
🔲 Account funding/withdrawal flows
🔲 Tax reporting integration
🔲 Performance metrics dashboard

## References

- [Alpaca Broker API Documentation](https://docs.alpaca.markets/reference/broker-api)
- [Alpaca Market Data API](https://docs.alpaca.markets/reference/marketdata-api)
- [Project Architecture](./architecture.md)
- [WARP Project Rules](../WARP.md)

## Support

For issues or questions:

1. Check Alpaca documentation
2. Review this integration guide
3. Check application logs
4. Contact the development team

---

**Last Updated**: 2025-10-29
**Version**: 1.0
**Author**: AI Agent
