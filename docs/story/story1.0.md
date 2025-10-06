Awesome—adding API mappings + Circle Wallets integration (Aptos, EVM, Solana) and tightening the story.

⸻

User Story: Simple Onboarding & Managed Wallet Creation (Expanded)

As a new Gen Z user
I want to sign up quickly and have secure wallets created for me on supported chains (EVM, Solana, Aptos)
So that I can start using the app without seed phrases or technical setup.

⸻

Acceptance Criteria ✅

A. Sign-Up & KYC
	1.	User can sign up with email or phone; verification (OTP/email link) required before continuing.
	2.	KYC flow requests ID + selfie; status clearly shown: Pending / Approved / Failed; failed state offers next steps.

B. Managed Wallet Creation (Automated)
	1.	On KYC = Approved, backend provisions developer-controlled wallets via Circle for:
	•	EVM (unified address across EVM chains where possible)
	•	Solana
	•	Aptos (EOA only)
Creation uses Circle Wallet Set + Wallet APIs; Solana/Aptos are EOA only; SCAs are not supported on Solana/Aptos.  ￼
	2.	Persist wallet metadata in wallets table; associate chain, address, provider ref, status.
	3.	User never sees private keys; custody abstracted.
	4.	Show success state and land on dashboard; handle friendly errors & retries; target ≥90% completion in tests.

C. App/Backend API Contracts (STACK)
	•	POST /onboarding/start → create user + kick off KYC.
	•	GET /onboarding/status → returns KYC + wallet provisioning status.
	•	GET /wallet/addresses?chain=eth|sol|aptos → returns deposit/receive address for requested chain.
	•	GET /wallet/status → returns per-chain wallet readiness.
(Endpoints align with the architecture API outline and system context.)

D. Observability & Audit
	•	Log provisioning attempts and results in audit_logs with provider refs; emit metrics for success/failure and time-to-ready.

⸻

Backend Design Notes (Circle Wallets integration)

Goal: Create wallets immediately after KYC approval using Circle Programmable Wallets (developer-controlled) so users can transact without seed phrases.

1) Wallet Set (one per tenant/app)
	•	POST https://api.circle.com/v1/w3s/developer/walletSets → create/set if missing.
	•	Stores walletSetId used for all user wallets.
	•	Enables Unified Wallet Addressing across EVM chains (same address on L1/L2 where supported).  ￼

2) Create Wallets per User (after KYC pass)
	•	POST https://api.circle.com/v1/w3s/developer/wallets with body:
	•	walletSetId, blockchains: ["ETH-SEPOLIA" (staging) | "ETH", ... EVMs, "SOL" | "SOL-DEVNET", "APTOS" | "APTOS-TESTNET"]
	•	accountType: "EOA" for Solana and Aptos (SCA not supported on these).
	•	entitySecretCiphertext (required by Circle).  ￼

3) Chain-Specific Notes
	•	EVM: Use Wallet Set for unified EVM addressing; supports SCA/EOA (choose EOA for MVP).  ￼
	•	Solana: Follow Circle’s Solana guide; accountType: "EOA", blockchains: "SOL"; no SCA.  ￼
	•	Aptos: Circle added Aptos support (2025); use blockchains: "APTOS"; EOA only.  ￼

4) Read/Verify
	•	GET wallets to confirm addresses and status; store id, address, blockchain in wallets.  ￼

5) Security & Secrets
	•	Keep Circle API key in server secrets manager; register Entity Secret as required by Circle before wallet calls.  ￼

6) SDKs (Optional)
	•	Use Circle SDKs for Node/Python in the Wallet Service worker to simplify calls.  ￼

⸻

Updated Sequence (KYC → Wallets)
	1.	App → API: POST /onboarding/start
	2.	API → KYC provider: start + poll; on Approved →
	3.	API → Circle:
	•	Ensure/lookup walletSetId (create if missing).  ￼
	•	Create wallets for ["ETH"(+other EVMs), "SOL", "APTOS"].  ￼
	4.	Persist wallet records; API → App: status Ready with per-chain addresses.

⸻

Validation Tests (Done = True)
	•	API contract: /wallet/addresses returns valid addresses for EVM, Solana, Aptos after KYC pass.
	•	Circle calls: Creating wallets sends correct params per chain; Solana/Aptos created as EOA; responses persisted.  ￼
	•	DX/UX: On failure from Circle, user sees retry with safe copy; no secrets or keys exposed.

⸻

Dev Notes & Risks
	•	KYC dependency: Provision wallets only after Approved to avoid orphaned wallets.
	•	Aptos is new (2025): confirm GA/region availability and quotas in Circle console before launch.  ￼
	•	Costs: Creating many EOA wallets on mainnet has per-wallet costs; consider pre-creating in batches monthly.  ￼

⸻

If you want, I can immediately spin off two follow-up stories:
	1.	Wallet Provisioning Worker (idempotent retries, backoff, audit logs).
	2.	/wallet/addresses endpoint (contract tests + Circle read-through).

Or say the word and I’ll mark this story Ready for Dev with task breakdowns.