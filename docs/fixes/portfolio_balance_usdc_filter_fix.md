# Portfolio Balance USDC Token Address Filter Fix

## Issue
The portfolio overview endpoint was returning zero balance despite the Circle API showing a balance of 20 USDC for wallet `6de94cea-8895-5e29-ac79-cb61413d5c60` with token address `4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU` on SOL-DEVNET.

## Root Cause
There were two issues:

1. **Missing token address filter**: The `GetWalletBalances` method was not filtering by token address when calling the Circle API. The balances endpoint accepts an optional `tokenAddress` query parameter:
   ```
   GET /v1/w3s/wallets/{walletId}/balances?tokenAddress={tokenAddress}
   ```

2. **Incorrect response parsing**: The Circle API wraps the response in a `data` object, but our entity struct expected `tokenBalances` at the root level:
   ```json
   {
     "data": {
       "tokenBalances": [...]
     }
   }
   ```
   Our struct was trying to parse:
   ```json
   {
     "tokenBalances": [...]
   }
   ```

## Solution
Added token address filtering to the Circle API balance fetching flow:

### 1. Added USDC Token Address Constants
**File**: `internal/domain/entities/wallet_entities.go`

```go
const (
    // Solana - Only SOL-DEVNET is currently supported
    ChainSOLDevnet WalletChain = "SOL-DEVNET"
    
    // USDC Token Addresses by Chain
    // SOL-DEVNET USDC token address
    USDCTokenAddressSOLDevnet = "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU"
)

// GetUSDCTokenAddress returns the USDC token address for the chain
func (c WalletChain) GetUSDCTokenAddress() string {
    switch c {
    case ChainSOLDevnet:
        return USDCTokenAddressSOLDevnet
    default:
        return ""
    }
}
```

### 2. Updated Circle Client Method
**File**: `internal/infrastructure/circle/client.go`

Modified `GetWalletBalances` to accept an optional `tokenAddress` parameter:

```go
// GetWalletBalances retrieves token balances for a specific wallet
// tokenAddress is optional - if provided, filters results to only that token
func (c *Client) GetWalletBalances(ctx context.Context, walletID string, tokenAddress ...string) (*entities.CircleWalletBalancesResponse, error) {
    endpoint := fmt.Sprintf("%s/%s/balances", c.config.BalancesEndpoint, walletID)
    
    // Add tokenAddress query parameter if provided
    if len(tokenAddress) > 0 && tokenAddress[0] != "" {
        endpoint = fmt.Sprintf("%s?tokenAddress=%s", endpoint, tokenAddress[0])
        c.logger.Info("Getting wallet balances", 
            zap.String("walletId", walletID),
            zap.String("tokenAddress", tokenAddress[0]))
    } else {
        c.logger.Info("Getting wallet balances", zap.String("walletId", walletID))
    }
    
    // ... rest of implementation
}
```

### 3. Updated Service Interfaces
**Files**: 
- `internal/domain/services/investing/service.go`
- `internal/domain/services/funding/service.go`
- `internal/infrastructure/di/container.go`

Updated all interface definitions and implementations to support the optional tokenAddress parameter:

```go
// CircleClient interface for fetching wallet balances from Circle
type CircleClient interface {
    GetWalletBalances(ctx context.Context, walletID string, tokenAddress ...string) (*entities.CircleWalletBalancesResponse, error)
}
```

### 4. Fixed Response Parsing with Custom UnmarshalJSON
**File**: `internal/domain/entities/wallet_entities.go`

Added custom `UnmarshalJSON` method to handle Circle API's `data` wrapper:

```go
// UnmarshalJSON normalizes Circle balance responses that wrap data
func (r *CircleWalletBalancesResponse) UnmarshalJSON(data []byte) error {
    aux := struct {
        Data struct {
            TokenBalances []CircleTokenBalance `json:"tokenBalances"`
        } `json:"data"`
        TokenBalances []CircleTokenBalance `json:"tokenBalances"`
    }{}

    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }

    // Check if wrapped in data.tokenBalances first
    if len(aux.Data.TokenBalances) > 0 {
        r.TokenBalances = aux.Data.TokenBalances
        return nil
    }

    // Fallback to direct tokenBalances field
    if len(aux.TokenBalances) > 0 {
        r.TokenBalances = aux.TokenBalances
        return nil
    }

    // Default to empty array (not nil)
    r.TokenBalances = []CircleTokenBalance{}
    return nil
}
```

### 5. Updated Portfolio Overview Logic
**File**: `internal/domain/services/investing/service.go`

Modified `GetPortfolioOverview` to fetch the USDC token address for each wallet's chain and pass it when fetching balances:

```go
for _, wallet := range wallets {
    // Get USDC token address for this wallet's chain
    usdcTokenAddress := wallet.Chain.GetUSDCTokenAddress()
    if usdcTokenAddress == "" {
        s.logger.Warn("No USDC token address configured for chain",
            "wallet_id", wallet.CircleWalletID,
            "chain", wallet.Chain)
        continue
    }
    
    // Fetch balance from Circle API for each wallet, filtering by USDC token address
    balancesResp, err := s.circleClient.GetWalletBalances(ctx, wallet.CircleWalletID, usdcTokenAddress)
    if err != nil {
        s.logger.Warn("Failed to fetch Circle balance",
            "wallet_id", wallet.CircleWalletID,
            "chain", wallet.Chain,
            "token_address", usdcTokenAddress,
            "error", err)
        continue
    }
    
    // Extract USDC balance
    usdcBalanceStr := balancesResp.GetUSDCBalance()
    if usdcBalance, err := decimal.NewFromString(usdcBalanceStr); err == nil {
        totalUSDCBalance = totalUSDCBalance.Add(usdcBalance)
        s.logger.Debug("Fetched Circle wallet balance",
            "wallet_id", wallet.CircleWalletID,
            "chain", wallet.Chain,
            "token_address", usdcTokenAddress,
            "usdc_balance", usdcBalanceStr)
    }
}
```

## Expected Circle API Response
When calling the Circle API with the token address filter:

```bash
GET https://api.circle.com/v1/w3s/wallets/6de94cea-8895-5e29-ac79-cb61413d5c60/balances?tokenAddress=4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU
```

Expected response:
```json
{
  "tokenBalances": [
    {
      "token": {
        "id": "8fb3cadb-0ef4-573d-8fcd-e194f961c728",
        "blockchain": "SOL-DEVNET",
        "tokenAddress": "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU",
        "standard": "Fungible",
        "name": "USD Coin",
        "symbol": "USDC",
        "decimals": 6,
        "isNative": false,
        "updateDate": "2024-02-28T13:56:21Z",
        "createDate": "2024-02-28T13:56:21Z"
      },
      "amount": "20",
      "updateDate": "2025-10-27T14:53:52Z"
    }
  ]
}
```

## Benefits
1. **Accurate Balance Reporting**: Portfolio overview now correctly displays USDC balances from Circle wallets
2. **Efficient API Calls**: Filtering by token address reduces response payload and processing time
3. **Extensible Design**: Easy to add support for other chains/tokens in the future
4. **Backward Compatible**: Optional parameter means existing calls without token address still work

## Testing
1. Build verification: `go build -o /dev/null ./cmd` ✅
2. Portfolio overview endpoint should now return the correct USDC balance
3. Verify logs show token address being passed to Circle API

## Additional Notes
- The token address is specific to each blockchain network (testnet vs mainnet)
- Currently only SOL-DEVNET is supported with token address `4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU`
- When adding support for other chains (e.g., SOL-MAINNET, ETH-SEPOLIA), add corresponding constants and update the `GetUSDCTokenAddress()` method

## Related Files
- `internal/domain/entities/wallet_entities.go`
- `internal/infrastructure/circle/client.go`
- `internal/domain/services/investing/service.go`
- `internal/domain/services/funding/service.go`
- `internal/infrastructure/di/container.go`
