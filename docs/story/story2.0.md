✅ User Story: Stablecoin Deposit → Buying Power (EVM & Solana)

Epic: 2 — Stablecoin Deposits
Goal: Allow users to fund their STACK account instantly with stablecoins (EVM + Solana) and convert deposits into buying power.

⸻

🎯 Functional Summary

As a STACK user
I want to generate a stablecoin deposit address on my preferred chain and see my balance update automatically once funds are confirmed
So that I can start investing without waiting for bank transfers.

⸻

✅ Acceptance Criteria

Area	Criteria
Deposit Address	- User can request deposit address for EVM or SOL.  - Endpoint: POST /funding/deposit/address.  - Returns QR + chain/network metadata.  - Address generation is idempotent per user/chain.
On-Chain Detection	- Inbound webhook (webhooks.chainDeposit) processes transactions.  - Confirms stablecoin transfer via chain RPCs.  - Applies chain-specific confirmations (e.g. 12 for EVM, finality for Solana).
Validation	- Only whitelisted tokens (e.g. USDC) accepted.  - Wrong token or chain triggers warning entry in funding history.
Credit to Buying Power	- After confirmed, create record in deposits and credit balances.buying_power.  - /balances endpoint reflects update immediately.
Funding History	- /funding/confirmations lists recent funding events (status: Pending / Confirmed / Credited).
UX	- User sees deposit address + QR.  - On success: success toast + updated balance.  - On error: clear message + retry option.
Reliability	- Transient errors retry with exponential backoff.  - Duplicate events are ignored safely (idempotency by tx hash).


⸻

🧩 API Endpoints (per OpenAPI Spec)

Method	Path	Purpose
POST	/funding/deposit/address	Create or return deposit address for chain.
GET	/funding/confirmations	List funding confirmations (paginated).
GET	/balances	Get current user balances.
POST	webhooks.chainDeposit	Receive deposit events from EVM/SOL watchers.

Source: open_api.yaml Funding & Investing section.

⸻

🧱 Data Model

Table	Key Columns	Description
deposits	id, user_id, chain, tx_hash, token, amount, status, confirmed_at	Tracks incoming stablecoin transactions.
balances	user_id, buying_power, updated_at	Reflects user’s investing power.

Source: architecture.md MVP schema.

⸻

⚙️ Technical Implementation Notes
	•	Chain Watchers: Use lightweight webhooks (Alchemy/Moralis/Helius or internal) to push deposit events to /webhooks/chainDeposit.
	•	Circle Wallet API: Deposits land in managed wallets (EVM + Solana addresses already provisioned).
	•	Ledger Logic: Credit ledger → trigger balance recalculation → persist to balances.
	•	Retries: Implement retry queue for failed webhooks and RPC lookups.
	•	Security: Validate JWT bearer tokens on all user endpoints; webhook secret verification required.
	•	Audit Logging: All deposit events recorded in audit_logs with before/after balances.

⸻

🧪 Definition of Done (Tests)

Type	Scenarios
Unit	Address creation idempotent; correct token validation; retries logic triggered on 5xx.
Integration	Simulated EVM + Solana deposits processed to credited status; buying_power updates correctly.
E2E	User requests address → sends testnet USDC → webhook fires → balance increases → investable immediately.
Negative	Unsupported token triggers warning entry; duplicate webhook ignored safely.


⸻

📋 Tasks Breakdown
	•	Implement POST /funding/deposit/address.
	•	Implement webhooks.chainDeposit handler (EVM/Solana).
	•	Implement deposit persistence + ledger credit.
	•	Implement GET /funding/confirmations (paginated).
	•	Implement GET /balances.
	•	Add retry/backoff + idempotency for webhook handling.
	•	Unit, Integration, and E2E tests (EVM + Solana).

⸻

📌 Dependencies & Risks
	•	Dependencies: Circle Wallet service (for managed addresses), Blockchain RPC providers, DB schema migrations.
	•	Risks:
	•	RPC delays or webhook duplicates could delay credits.
	•	Need to ensure Solana finality handling is consistent across networks.

⸻

✅ Status: Ready for Dev
📦 Owner: Funding Service (Backend Team)
📅 Priority: High (launch-critical for MVP)
