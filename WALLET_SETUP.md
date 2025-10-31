# Developer-Controlled Wallets Implementation Guide

This guide covers the complete setup and implementation of developer-controlled wallets using Circle's Web3 Services API.

## Table of Contents

1. [Overview](#overview)
2. [Architecture Flow](#architecture-flow)
3. [Prerequisites & Setup](#prerequisites--setup)
4. [Implementation Steps](#implementation-steps)
5. [API Endpoints](#api-endpoints)
6. [Testing](#testing)
7. [Troubleshooting](#troubleshooting)
8. [Security Considerations](#security-considerations)

---

## Overview

The developer-controlled wallet implementation allows users to:
- Create wallets after email verification and passcode setup
- Generate wallets across supported testnet blockchains
- Manage wallet lifecycle and retrieve wallet addresses
- Use wallets for transactions

### Supported Chains (Testnet Only)

| Chain | ID | Network |
|-------|-----|---------|
| **Solana** | `SOL-DEVNET` | Solana Devnet |
| **Aptos** | `APTOS-TESTNET` | Aptos Testnet |
| **Polygon** | `MATIC-AMOY` | Polygon Amoy (Mumbai Testnet) |
| **Base** | `BASE-SEPOLIA` | Base Sepolia Testnet |

⚠️ **Important**: Mainnet chains are NOT supported at this time.

---

## Architecture Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     User Signup Flow                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  1. User Registration (Email/Password)                          │
│     POST /api/v1/auth/register                                  │
│     └─> Sends verification code to email                        │
│                                                                   │
│  2. Email Verification                                          │
│     POST /api/v1/auth/verify-code                               │
│     └─> User marked as verified, JWT tokens issued              │
│                                                                   │
│  3. Passcode Setup (4-digit PIN)                                │
│     POST /api/v1/security/passcode                              │
│     └─> Stores hashed passcode                                  │
│                                                                   │
│  4. Passcode Verification ⭐ Triggers Wallet Creation            │
│     POST /api/v1/security/passcode/verify                       │
│     └─> User enters passcode, receives session token            │
│                                                                   │
│  5. Wallet Initiation ⭐ NEW ENDPOINT                            │
│     POST /api/v1/wallets/initiate                               │
│     └─> Creates wallet set, initiates wallet generation         │
│                                                                   │
│  6. Wallet Provisioning (Background Worker)                     │
│     ├─> Creates Circle WalletSet if needed                      │
│     ├─> Creates wallets for each chain                          │
│     └─> Stores wallet addresses & IDs in database               │
│                                                                   │
│  7. Wallet Ready                                                 │
│     └─> User can retrieve addresses and use wallets             │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Prerequisites & Setup

### Step 1: Circle Account Setup

1. **Create Circle Developer Account**
   - Visit: https://dashboard.circle.com
   - Sign up or log in
   - Create a new API key

2. **Generate Entity Secret**
   ```bash
   cd cmd/entitysecret
   go run main.go
   # Output: Hex encoded entity secret: cb82a7c71abcc0b262a061e3897deb4d...
   ```

3. **Register Entity Secret with Circle**
   - Go to: https://dashboard.circle.com/developers/wallets/entity-secrets
   - Click "Register Entity Secret"
   - Paste your hex-encoded Entity Secret
   - Circle will generate and return the **Ciphertext**
   - Save this Ciphertext (you'll need it for configuration)
   - Download and securely store the recovery file

### Step 2: Environment Configuration

Update your `.env` file with Circle credentials:

```bash
# Copy the example
cp .env.example .env

# Edit .env with your values
CIRCLE_API_KEY=your_api_key_from_circle_dashboard
CIRCLE_ENTITY_SECRET_CIPHERTEXT=the_ciphertext_from_circle_dashboard
CIRCLE_ENVIRONMENT=sandbox  # Use 'sandbox' for development
```

### Step 3: Database Setup

Ensure the following tables exist:

```sql
-- Wallet Sets table
CREATE TABLE wallet_sets (
    id UUID PRIMARY KEY,
    circle_wallet_set_id VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    entity_secret_ciphertext TEXT,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Managed Wallets table
CREATE TABLE managed_wallets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    wallet_set_id UUID NOT NULL REFERENCES wallet_sets(id),
    circle_wallet_id VARCHAR(255) NOT NULL UNIQUE,
    chain VARCHAR(50) NOT NULL,
    address VARCHAR(255) NOT NULL,
    account_type VARCHAR(50) DEFAULT 'SCA',
    status VARCHAR(50) DEFAULT 'live',
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    UNIQUE(user_id, chain)
);

-- Provisioning Jobs table
CREATE TABLE wallet_provisioning_jobs (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    chains TEXT[] NOT NULL,
    status VARCHAR(50) DEFAULT 'queued',
    attempt_count INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    circle_requests JSONB,
    error_message TEXT,
    next_retry_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

---

## Implementation Steps

### Step 1: Update Circle Client Configuration

The Circle client in `internal/infrastructure/circle/client.go` has been updated to use the pre-registered Entity Secret Ciphertext:

```go
// Config now includes EntitySecretCiphertext
type Config struct {
    APIKey                     string
    BaseURL                    string
    EntitySecretCiphertext     string  // Pre-registered ciphertext
}

// CreateWalletSet and CreateWallet now use the configured ciphertext
func (c *Client) CreateWalletSet(ctx context.Context, name string, _ string) error {
    // Uses c.config.EntitySecretCiphertext instead of generating
    // ...
}
```

### Step 2: Wallet Initiation Endpoint

The new `/api/v1/wallets/initiate` endpoint is called after passcode verification:

**Request:**
```json
POST /api/v1/wallets/initiate
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "chains": ["SOL-DEVNET", "MATIC-AMOY"]  // Optional, defaults to all supported testnet chains
}
```

**Response (202 Accepted):**
```json
{
  "message": "Wallet creation initiated successfully",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "chains": ["SOL-DEVNET", "MATIC-AMOY", "APTOS-TESTNET", "BASE-SEPOLIA"],
  "job": {
    "id": "job-550e8400-e29b-41d4",
    "status": "in_progress",
    "progress": "0%",
    "attemptCount": 1,
    "maxAttempts": 3,
    "errorMessage": null,
    "nextRetryAt": null,
    "createdAt": "2025-10-20T02:54:00Z"
  }
}
```

### Step 3: Wallet Storage

After successful creation, wallets are stored with:
- **circle_wallet_id**: Unique identifier from Circle API
- **address**: Blockchain address for the wallet
- **chain**: Blockchain network (e.g., SOL-DEVNET)
- **status**: Current status (live, creating, failed)
- **user_id**: Associated user
- **wallet_set_id**: Associated wallet set

---

## API Endpoints

### Wallet Management Endpoints

#### 1. Initiate Wallet Creation (NEW)
```
POST /api/v1/wallets/initiate
Authorization: Bearer <JWT_TOKEN>
```

Initiates wallet creation for the authenticated user across specified chains.

**Parameters:**
```json
{
  "chains": ["SOL-DEVNET", "MATIC-AMOY"]  // Optional
}
```

**Returns:** 202 Accepted with job details

**Errors:**
- `INVALID_REQUEST` (400): Invalid request format
- `UNAUTHORIZED` (401): Missing or invalid JWT
- `INVALID_CHAIN` (400): Unsupported chain specified
- `MAINNET_NOT_SUPPORTED` (400): Mainnet chain attempted
- `WALLET_INITIATION_FAILED` (500): Server error

---

#### 2. Get Wallet Addresses
```
GET /api/v1/wallet/addresses?chain=SOL-DEVNET
Authorization: Bearer <JWT_TOKEN>
```

Retrieves all wallet addresses for the user, optionally filtered by chain.

**Query Parameters:**
- `chain` (optional): Filter by specific chain

**Returns:**
```json
{
  "wallets": [
    {
      "chain": "SOL-DEVNET",
      "address": "4sGoLvUV5VwvmsAWSVgsk76ajGs9pG5YNFPRtn5mCZE1",
      "status": "live"
    }
  ]
}
```

---

#### 3. Get Wallet Status
```
GET /api/v1/wallet/status
Authorization: Bearer <JWT_TOKEN>
```

Retrieves comprehensive wallet status including provisioning progress.

**Returns:**
```json
{
  "userId": "550e8400-e29b-41d4-a716-446655440000",
  "totalWallets": 4,
  "readyWallets": 2,
  "pendingWallets": 2,
  "failedWallets": 0,
  "walletsByChain": {
    "SOL-DEVNET": {
      "chain": "SOL-DEVNET",
      "address": "4sGoLvUV5VwvmsAWSVgsk76ajGs9pG5YNFPRtn5mCZE1",
      "status": "live"
    }
  },
  "provisioningJob": {
    "id": "job-550e8400-e29b-41d4",
    "status": "in_progress",
    "progress": "50% complete",
    "attemptCount": 1,
    "maxAttempts": 3,
    "createdAt": "2025-10-20T02:54:00Z"
  }
}
```

---

#### 4. Get Wallet Address by Chain
```
GET /api/v1/wallets/:chain/address
Authorization: Bearer <JWT_TOKEN>
```

Retrieves wallet address for a specific chain.

**Path Parameters:**
- `chain`: Blockchain network (e.g., SOL-DEVNET)

**Returns:**
```json
{
  "chain": "SOL-DEVNET",
  "address": "4sGoLvUV5VwvmsAWSVgsk76ajGs9pG5YNFPRtn5mCZE1",
  "status": "live"
}
```

---

## Testing

### Quick Start

1. **Set environment variables:**
```bash
export CIRCLE_API_KEY=your_api_key
export CIRCLE_ENTITY_SECRET_CIPHERTEXT=your_ciphertext
export API_BASE_URL=http://localhost:8080
```

2. **Run the test script:**
```bash
chmod +x test/integration/wallet_integration_test.sh
./test/integration/wallet_integration_test.sh
```

3. **Run individual tests:**
```bash
source test/integration/wallet_integration_test.sh

# Test user signup
test_user_signup

# Create passcode
export ACCESS_TOKEN=your_jwt_token
test_create_passcode

# Verify passcode
test_verify_passcode

# Initiate wallet creation
test_initiate_wallet_creation

# Check wallet status
test_get_wallet_status
```

### Manual Testing with cURL

#### 1. Create Passcode
```bash
curl -X POST http://localhost:8080/api/v1/security/passcode \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "passcode": "1234",
    "confirm_passcode": "1234"
  }'
```

#### 2. Verify Passcode
```bash
curl -X POST http://localhost:8080/api/v1/security/passcode/verify \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "passcode": "1234"
  }'
```

#### 3. Initiate Wallets (Default Chains)
```bash
curl -X POST http://localhost:8080/api/v1/wallets/initiate \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "chains": []
  }'
```

#### 4. Initiate Wallets (Specific Chains)
```bash
curl -X POST http://localhost:8080/api/v1/wallets/initiate \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "chains": ["SOL-DEVNET", "MATIC-AMOY", "APTOS-TESTNET"]
  }'
```

#### 5. Get Wallet Addresses
```bash
curl -X GET http://localhost:8080/api/v1/wallet/addresses \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

#### 6. Get Wallet Status
```bash
curl -X GET http://localhost:8080/api/v1/wallet/status \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

## Troubleshooting

### Issue: "entity secret ciphertext not configured"

**Cause:** `CIRCLE_ENTITY_SECRET_CIPHERTEXT` environment variable not set.

**Solution:**
1. Register Entity Secret with Circle Dashboard
2. Copy the Ciphertext from the response
3. Set environment variable: `export CIRCLE_ENTITY_SECRET_CIPHERTEXT=your_ciphertext`
4. Restart the service

### Issue: "WALLET_INITIATION_FAILED"

**Cause:** Missing Circle API key or invalid ciphertext.

**Solution:**
1. Verify `CIRCLE_API_KEY` is set correctly
2. Verify `CIRCLE_ENTITY_SECRET_CIPHERTEXT` length is ~684 characters
3. Check that ciphertext is base64 encoded
4. Verify Circle account has wallet API access
5. Check service logs for detailed error: `docker logs stack-service`

### Issue: "MAINNET_NOT_SUPPORTED"

**Cause:** Attempting to create wallet on mainnet chains.

**Solution:**
- Use only testnet chains: SOL-DEVNET, APTOS-TESTNET, MATIC-AMOY, BASE-SEPOLIA
- Mainnet support will be added in future releases

### Issue: Wallet creation stuck in "creating" status

**Cause:** Background provisioning worker not processing jobs.

**Solution:**
1. Verify worker service is running
2. Check worker logs: `docker logs stack-service-worker`
3. Verify database connection from worker
4. Check Circle API limits and rate limiting

---

## Security Considerations

### 1. Entity Secret Ciphertext

- ✅ **Good:** Ciphertext stored as environment variable (not in code)
- ✅ **Good:** Registered once in Circle Dashboard, never transmitted
- ⚠️ **Important:** Never expose or log the ciphertext
- ⚠️ **Important:** Store recovery file in secure location

### 2. Wallet IDs Storage

- ✅ **Good:** `circle_wallet_id` stored securely in database
- ✅ **Good:** Associated with user_id for access control
- ✅ **Good:** Never expose IDs to untrusted clients
- ⚠️ **Important:** Implement row-level security if sharing database

### 3. Authentication

- ✅ **Good:** All endpoints require JWT authentication
- ✅ **Good:** Passcode verification adds additional layer
- ⚠️ **Important:** Use HTTPS in production
- ⚠️ **Important:** Implement rate limiting on auth endpoints

### 4. Rate Limiting

- ✅ **Good:** Global rate limiting configured
- ✅ **Good:** Specific limits on sensitive endpoints
- Recommendation: Implement per-user rate limiting for wallet creation

### 5. Audit Logging

- ✅ **Good:** All wallet operations logged with user_id and timestamp
- Recommendation: Implement immutable audit logs for compliance

---

## Related Documentation

- [Circle Developer Docs](https://developers.circle.com/w3s)
- [Entity Secret Setup Guide](./cmd/entitysecret/README.md)
- [API Reference](./WALLET_API.md)
- [Database Schema](./internal/infrastructure/migrations/)

---

## Support & Next Steps

### Immediate Next Steps

1. ✅ Configure Circle credentials (DONE)
2. ✅ Implement wallet endpoints (DONE)
3. ⬜ Set up background worker for wallet provisioning
4. ⬜ Implement wallet balance retrieval
5. ⬜ Implement wallet transactions (transfers)

### Future Features

- [ ] Mainnet wallet support
- [ ] Multiple wallet accounts per user
- [ ] Wallet import/export
- [ ] Hardware wallet integration
- [ ] Multi-sig support
- [ ] Advanced transaction features

---

**Last Updated:** 2025-10-20
**Implementation Status:** ✅ Wallet Endpoints Complete | ⏳ Background Worker Required
