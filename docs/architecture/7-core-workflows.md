# 7\. Core Workflows

These sequence diagrams illustrate the key interactions between components and external partners for critical user journeys in the Go-based architecture.

### 7.1 Onboarding + Wallet Creation + Passcode Setup

```mermaid
sequenceDiagram
    participant App as React Native App
    participant GW as API Gateway (GraphQL)
    participant ONB as Onboarding Module (Go)
    participant WAL as Wallet Module (Go)
    participant IDP as Auth0/Cognito
    participant KYC as KYC Provider
    participant CIR as Circle API
    participant PG as PostgreSQL

    App->>IDP: User Initiates Sign Up/Login
    IDP-->>App: OIDC Flow / Receives ID Token
    App->>GW: SignUp/Login Mutation (with ID Token)
    GW->>ONB: Verify Token / Create User Request
    ONB->>IDP: Validate Token (Optional, depending on flow)
    ONB->>PG: Create/Fetch User Record
    ONB->>KYC: Initiate KYC Check
    ONB->>WAL: Trigger Wallet Creation (async)
    WAL->>CIR: Create Developer-Controlled Wallet
    CIR-->>WAL: Wallet ID/Details
    WAL->>PG: Store Wallet Info (address, circle_id)
    KYC-->>ONB: KYC Status Webhook/Callback (pass/fail)
    ONB->>PG: Update User KYC Status
    ONB-->>GW: User Profile & Status
    GW-->>App: Return User Profile

    %% Passcode Setup (Separate Flow/Mutation) %%
    App->>GW: Set Passcode Mutation (passcode)
    GW->>ONB: Hash and Store Passcode Request
    ONB->>PG: Store passcode_hash for user
    ONB-->>GW: Success Confirmation
    GW-->>App: Passcode Set
```

### 7.2 Funding Flow (USDC Deposit -\> Circle Off-Ramp -\> Alpaca Funding)

```mermaid
sequenceDiagram
    participant User
    participant App as React Native App
    participant GW as API Gateway (GraphQL)
    participant WAL as Wallet Module (Go)
    participant FND as Funding Module (Go)
    participant CIR as Circle API
    participant DW as Alpaca API
    participant PG as PostgreSQL
    participant Q as SQS Queue

    %% Step 1: Get Deposit Address %%
    App->>GW: Query Deposit Address (chain=ETH/SOL)
    GW->>WAL: GetDepositAddress(userId, chain)
    WAL->>PG: Fetch Wallet Info
    WAL-->>GW: Return Address
    GW-->>App: Display Address

    %% Step 2: User Deposits USDC (On-Chain) %%
    User->>Blockchain: Sends USDC to Address

    %% Step 3: Circle Detects Deposit & Notifies Backend %%
    Blockchain->>CIR: Transaction Confirmed
    CIR->>FND: Deposit Notification Webhook
    FND->>PG: Log Deposit (status: confirmed_on_chain)
    FND->>Q: Enqueue Off-Ramp Task (depositId)

    %% Step 4: Async Off-Ramp Processing %%
    Q->>FND: Process Off-Ramp Task (depositId)
    FND->>PG: Update Deposit Status (status: off_ramp_initiated)
    FND->>CIR: Initiate Payout (USDC -> USD to linked Bank/Broker)
    CIR-->>FND: Payout Accepted/Pending

    %% Step 5: Circle Confirms Off-Ramp %%
    CIR->>FND: Payout Completed Webhook
    FND->>PG: Update Deposit Status (status: off_ramp_complete)
    FND->>Q: Enqueue Broker Funding Task (depositId, usdAmount)

    %% Step 6: Async Broker Funding %%
    Q->>FND: Process Broker Funding Task
    FND->>DW: Initiate Account Funding (USD Transfer)
    DW-->>FND: Funding Accepted/Pending

    %% Step 7: Alpaca Confirms Funding %%
    DW->>FND: Funding Completed Webhook/Callback
    FND->>PG: Update Deposit Status (status: broker_funded)
    FND->>PG: Update User Balance (buying_power_usd)
    FND-->>GW: Notify User (e.g., via WebSocket/Push)
    GW-->>App: Balance Updated Notification
```

### 7.3 Investment Flow (Place Order for Stock/Option via Alpaca)

```mermaid
sequenceDiagram
    participant App as React Native App
    participant GW as API Gateway (GraphQL)
    participant INV as Investing Module (Go)
    participant FND as Funding Module (Go)
    participant DW as Alpaca API
    participant PG as PostgreSQL

    App->>GW: Place Order Mutation (basketId/optionDetails, amountUSD, side)
    GW->>INV: PlaceOrder Request
    INV->>FND: Check Balance (buying_power_usd)
    FND->>PG: Read balance record
    FND-->>INV: Balance Check Result (OK/Insufficient)

    alt Sufficient Balance
        INV->>PG: Create Order Record (status: pending)
        INV->>DW: Submit Order (Stock/ETF/Option)
        DW-->>INV: Order Accepted (brokerRef)
        INV->>PG: Update Order Record (status: accepted_by_broker, brokerRef)
        INV-->>GW: Order Accepted Confirmation
        GW-->>App: Order Accepted
    else Insufficient Balance
        INV-->>GW: Error: Insufficient Funds
        GW-->>App: Show Error Message
    end

    %% Later - Alpaca sends fill confirmation %%
    DW->>INV: Order Fill Webhook (partial/full fill details)
    INV->>PG: Update Order Record (status: partially_filled / filled)
    INV->>PG: Update/Cache Position Data (or trigger portfolio refresh)
    INV-->>GW: Notify User (e.g., via WebSocket/Push)
    GW-->>App: Order Fill Notification
```

### 7.4 Withdrawal Flow (Alpaca USD -\> Circle On-Ramp -\> USDC Transfer)

```mermaid
sequenceDiagram
    participant App as React Native App
    participant GW as API Gateway (GraphQL)
    participant FND as Funding Module (Go)
    participant DW as Alpaca API
    participant CIR as Circle API
    participant WAL as Wallet Module (Go)
    participant PG as PostgreSQL
    participant Q as SQS Queue

    App->>GW: Initiate Withdrawal Mutation (amountUSD, targetChain, targetAddress)
    GW->>FND: InitiateWithdrawal Request
    FND->>PG: Check Balance (buying_power_usd)
    alt Sufficient Balance
        FND->>PG: Log Withdrawal Request (status: pending)
        FND->>Q: Enqueue Broker Withdrawal Task
        FND-->>GW: Withdrawal Initiated Confirmation
        GW-->>App: Withdrawal Initiated
    else Insufficient Balance
        FND-->>GW: Error: Insufficient Funds
        GW-->>App: Show Error Message
    end

    %% Async Broker Withdrawal %%
    Q->>FND: Process Broker Withdrawal Task
    FND->>DW: Request USD Withdrawal
    DW-->>FND: Withdrawal Accepted/Pending

    %% Alpaca Confirms USD Sent %%
    DW->>FND: Withdrawal Completed Webhook/Callback (e.g., ACH settled)
    FND->>PG: Update Withdrawal Status (status: broker_withdrawal_complete)
    FND->>Q: Enqueue Circle On-Ramp Task

    %% Async Circle On-Ramp %%
    Q->>FND: Process Circle On-Ramp Task
    FND->>CIR: Initiate Payment/Transfer (USD -> USDC)
    CIR-->>FND: Payment Accepted/Pending

    %% Circle Confirms USDC Ready %%
    CIR->>FND: Payment Completed Webhook
    FND->>PG: Update Withdrawal Status (status: on_ramp_complete)
    FND->>Q: Enqueue On-Chain Transfer Task

    %% Async On-Chain Transfer %%
    Q->>FND: Process On-Chain Transfer Task
    FND->>WAL: Get User Wallet Details (Circle Wallet ID)
    FND->>CIR: Initiate Transfer (Send USDC from Circle Wallet to targetAddress)
    CIR-->>FND: Transfer Accepted/Pending (txHash)
    FND->>PG: Update Withdrawal Status (status: transfer_initiated, txHash)

    %% Circle Confirms Transfer %%
    CIR->>FND: Transfer Completed Webhook
    FND->>PG: Update Withdrawal Status (status: complete)
    FND-->>GW: Notify User
    GW-->>App: Withdrawal Complete Notification
```

