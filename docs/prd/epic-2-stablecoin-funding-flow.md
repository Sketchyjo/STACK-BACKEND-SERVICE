# Epic 2: Stablecoin Funding Flow
**Summary:**
Enable users to fund their brokerage accounts instantly with stablecoins.

**In-Scope Features:**
- Support deposits from Solana (non-EVM) into the user's Circle wallet.
- **NEW:** Orchestrate an immediate **USDC-to-USD off-ramp** via Circle.
- **NEW:** Securely transfer the resulting USD into the user's **Alpaca** brokerage account.
- **NEW:** Handle the reverse flow: **USD (Alpaca) -> USDC (Circle)** for withdrawals.

**Success Criteria:**
- Users can fund their brokerage account within minutes of a confirmed stablecoin deposit.
- At least 2 supported deposit pathways at launch.
- End-to-end funding/withdrawal success rate of >99%.

---
x