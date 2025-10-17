Authentication & Onboarding API Testing Guide

Overview
- This guide documents how to test authentication and onboarding endpoints implemented in the service.
- It covers endpoint-by-endpoint testing steps, expected responses, required authentication, common error scenarios, and end-to-end onboarding flow.

Prerequisites
- Running server on `http://localhost:8080` (e.g., `go run cmd/main.go`).
- A test database configured and reachable per `configs` settings.
- JWT configured (`config.JWT.secret`) and middleware enabled for protected routes.
- Optional: Redis configured if using the verification flow.

Endpoint Summary
- Authentication (public):
  - `POST /api/v1/auth/register` — register with email/password; sends verification code when applicable.
  - `POST /api/v1/auth/login` — login and obtain JWT tokens.
  - `POST /api/v1/auth/verify-code` — verify email/SMS code (for signup flows).
  - `POST /api/v1/auth/resend-code` — resend verification code (rate limited).
  - `POST /api/v1/auth/refresh` — refresh tokens.
  - `POST /api/v1/auth/logout` — invalidate session/refresh token.
  - `POST /api/v1/auth/forgot-password` — initiate password reset.
  - `POST /api/v1/auth/reset-password` — complete password reset.
  - `POST /api/v1/auth/verify-email` — verify email via token/link.
- Onboarding (protected):
  - `POST /api/v1/onboarding/start` — start onboarding for a user.
  - `GET /api/v1/onboarding/status` — get onboarding + KYC + wallet status.
  - `POST /api/v1/onboarding/kyc/submit` — submit KYC documents.
  - `POST /api/v1/kyc/callback/:provider_ref` — provider webhook for KYC status updates.

Testing Guidelines

General Notes
- Use `Content-Type: application/json` for JSON requests.
- Protected endpoints require `Authorization: Bearer <jwt>`.
- Follow rate limits where applicable (verification/resend flows).
- Validate both positive and negative paths (success, invalid input, unauthorized).

1) POST /api/v1/auth/register
- Purpose: Register a new user.
- Request example:
  {
    "email": "user@example.com",
    "password": "StrongP@ssw0rd!"
  }
- Expected responses:
  - 201 Created: user registered (and verification initiated if applicable).
  - 400 Bad Request: missing/invalid fields.
  - 409 Conflict: user already exists.
- Test cases:
  - Valid email/password returns 201.
  - Duplicate registration returns 409.
  - Invalid email format returns 400.
  - Weak/missing password returns 400.

2) POST /api/v1/auth/login
- Purpose: Obtain JWT token with valid credentials.
- Request example:
  {
    "email": "user@example.com",
    "password": "StrongP@ssw0rd!"
  }
- Expected responses:
  - 200 OK: JSON with `token` or `accessToken` and optional `refreshToken`.
  - 401 Unauthorized: invalid credentials.
- Test cases:
  - Valid credentials return 200 and token fields present.
  - Invalid password returns 401.
  - Nonexistent user returns 401.

3) POST /api/v1/auth/verify-code
- Purpose: Verify a 6‑digit code sent to email/phone.
- Request example:
  {
    "email": "user@example.com",
    "code": "123456"
  }
- Expected responses:
  - 200 OK: returns user profile + `accessToken` and `refreshToken`.
  - 400 Bad Request: missing or malformed payload.
  - 401/403: invalid or expired code; too many attempts.
- Test cases:
  - Correct code → 200 with tokens.
  - Wrong/expired code → 401/403.
  - Rate limit exceeded → appropriate error response.

4) POST /api/v1/onboarding/start
- Purpose: Start onboarding: create user record if missing, initialize steps, send verification.
- Request example:
  {
    "email": "user@example.com",
    "phone": "+1234567890" // optional
  }
- Expected responses:
  - 201 Created: `{ "userId", "onboardingStatus", "nextStep", "sessionToken"? }`.
  - 400 Bad Request: invalid payload (email missing/invalid, phone format invalid).
  - 409 Conflict: user already exists and onboarding previously started.
  - 500 Internal Server Error: unexpected failure.
- Test cases:
  - New email → 201 with nextStep `email_verification`.
  - Duplicate email → 409.
  - Invalid email/phone → 400.

5) GET /api/v1/onboarding/status
- Purpose: Retrieve current onboarding status, KYC state, wallet provisioning summary.
- Auth: Bearer JWT required.
- Request: `GET /api/v1/onboarding/status?user_id=<uuid>` (user_id optional; otherwise derived from auth context).
- Expected responses:
  - 200 OK: `{ userId, onboardingStatus, kycStatus, currentStep?, completedSteps[], walletStatus?, canProceed, requiredActions[] }`.
  - 400 Bad Request: invalid `user_id` format.
  - 404 Not Found: valid UUID but user does not exist or inaccessible.
  - 401 Unauthorized: missing/invalid token.
- Test cases:
  - Valid user id + auth → 200; fields have correct types.
  - Invalid UUID string → 400.
  - Random valid UUID → 404.
  - No token → 401.

6) POST /api/v1/onboarding/kyc/submit
- Purpose: Submit KYC documents to provider.
- Auth: Bearer JWT required.
- Request example:
  {
    "documentType": "passport",
    "documents": [
      { "type": "id_front", "fileUrl": "https://example.com/docs/id_front.jpg", "contentType": "image/jpeg" },
      { "type": "selfie", "fileUrl": "https://example.com/docs/selfie.jpg", "contentType": "image/jpeg" }
    ],
    "personalInfo": {
      "firstName": "Test",
      "lastName": "User",
      "dateOfBirth": "1990-01-01T00:00:00Z",
      "country": "US",
      "address": { "street": "123 Test St", "city": "Testville", "postalCode": "12345", "country": "US" }
    },
    "metadata": { "purpose": "standard_kyc" }
  }
- Expected responses:
  - 202 Accepted: submission accepted and under review.
  - 400 Bad Request: validation errors in payload.
  - 403 Forbidden: user not eligible for KYC (e.g., failed preconditions).
  - 401 Unauthorized: missing/invalid token.
  - 500 Internal Server Error: provider submission failure.
- Test cases:
  - Complete valid payload → 202.
  - Missing required fields → 400.
  - Submit when not eligible → 403.

7) POST /api/v1/kyc/callback/:provider_ref
- Purpose: Process KYC provider callback (webhook) by provider reference.
- Auth: Usually no auth (external webhook), but should include signature verification in production.
- Request example:
  {
    "status": "approved",
    "reason": null,
    "reviewedAt": "2025-01-01T12:00:00Z"
  }
- Expected responses:
  - 200 OK/204 No Content: processed successfully and state updated.
  - 400 Bad Request: malformed callback.
  - 404 Not Found: unknown provider_ref.
- Test cases:
  - Valid callback advances user KYC → status updated; wallet provisioning job enqueued.
  - Invalid payload → 400.

Onboarding Flow Documentation

User Journey Steps
- Registration: create account using `POST /api/v1/auth/register`.
- Verification: complete email/SMS verification with `POST /api/v1/auth/verify-code`.
- Start Onboarding: call `POST /api/v1/onboarding/start` to initialize onboarding steps.
- Submit KYC: `POST /api/v1/onboarding/kyc/submit` with required documents and personal info.
- Provider Review: KYC provider calls `/api/v1/kyc/callback/:provider_ref` with decision.
- Wallet Provisioning: upon KYC approval, managed wallets are created; status visible via `GET /api/v1/onboarding/status`.
- Completion: onboarding status reaches `completed`; user can fund and invest.

Sequence of API Calls
1. `POST /api/v1/auth/register` → 201
2. `POST /api/v1/auth/verify-code` → 200 with `accessToken`
3. `POST /api/v1/onboarding/start` (authorized if required) → 201 (`nextStep`, `userId`)
4. `POST /api/v1/onboarding/kyc/submit` (Bearer JWT) → 202 (`processing`)
5. `POST /api/v1/kyc/callback/:provider_ref` → 200/204 (`approved`/`rejected`)
6. `GET /api/v1/onboarding/status` (Bearer JWT) → 200 (includes `walletStatus`, `completedSteps`, `requiredActions`)

Required Inputs per Stage
- Register: `email`, `password`.
- Verify-code: `email` or `phone`, `code`.
- Onboarding-start: `email` (required), `phone` (optional, E.164).
- KYC-submit: `documentType`, `documents[]`, `personalInfo`, optional `metadata`.
- Callback: `status` (`approved`, `rejected`, `expired`, etc.), optional `reason`.

Success/Failure States
- KYC Status: `pending` → `processing` → `approved`/`rejected`/`expired`.
- Onboarding Steps: `pending` | `in_progress` | `completed` | `failed` | `skipped`.
- Wallet Status: Per-chain readiness and failures summarized in `walletStatus`.
- Handling procedures:
  - On `rejected`: surface reasons, allow resubmission (`kyc_rejected` state) and restart KYC.
  - On transient failure: retry submission; log audit events; show `requiredActions`.
  - On wallet failures: retry via provisioning job; expose status in `walletStatus`.

Code Examples

JavaScript (fetch)
// Register
await fetch('/api/v1/auth/register', {
  method: 'POST', headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email: 'user@example.com', password: 'StrongP@ssw0rd!' })
});

// Login
const loginRes = await fetch('/api/v1/auth/login', {
  method: 'POST', headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email: 'user@example.com', password: 'StrongP@ssw0rd!' })
});
const { token } = await loginRes.json();

// Start onboarding
const startRes = await fetch('/api/v1/onboarding/start', {
  method: 'POST', headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email: 'user@example.com' })
});
const startData = await startRes.json();

// Submit KYC
await fetch('/api/v1/onboarding/kyc/submit', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
  body: JSON.stringify({
    documentType: 'passport',
    documents: [
      { type: 'id_front', fileUrl: 'https://example.com/docs/id_front.jpg', contentType: 'image/jpeg' },
      { type: 'selfie', fileUrl: 'https://example.com/docs/selfie.jpg', contentType: 'image/jpeg' }
    ],
    personalInfo: {
      firstName: 'Test', lastName: 'User', dateOfBirth: '1990-01-01T00:00:00Z', country: 'US',
      address: { street: '123 Test St', city: 'Testville', postalCode: '12345', country: 'US' }
    }
  })
});

// Get status
const statusRes = await fetch('/api/v1/onboarding/status', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const status = await statusRes.json();

Python (requests)
import requests

base = 'http://localhost:8080'
headers = { 'Content-Type': 'application/json' }

# Register
requests.post(f'{base}/api/v1/auth/register', json={
  'email': 'user@example.com', 'password': 'StrongP@ssw0rd!'
}, headers=headers)

# Login
r = requests.post(f'{base}/api/v1/auth/login', json={
  'email': 'user@example.com', 'password': 'StrongP@ssw0rd!'
}, headers=headers)
token = r.json().get('token')

# Start onboarding
start = requests.post(f'{base}/api/v1/onboarding/start', json={ 'email': 'user@example.com' }, headers=headers)
user_id = start.json().get('userId')

# Submit KYC
auth = { 'Authorization': f'Bearer {token}', 'Content-Type': 'application/json' }
requests.post(f'{base}/api/v1/onboarding/kyc/submit', json={
  'documentType': 'passport',
  'documents': [
    { 'type': 'id_front', 'fileUrl': 'https://example.com/docs/id_front.jpg', 'contentType': 'image/jpeg' },
    { 'type': 'selfie', 'fileUrl': 'https://example.com/docs/selfie.jpg', 'contentType': 'image/jpeg' }
  ],
  'personalInfo': {
    'firstName': 'Test', 'lastName': 'User', 'dateOfBirth': '1990-01-01T00:00:00Z', 'country': 'US',
    'address': { 'street': '123 Test St', 'city': 'Testville', 'postalCode': '12345', 'country': 'US' }
  }
}, headers=auth)

# Status
status = requests.get(f'{base}/api/v1/onboarding/status', headers=auth).json()

Expected Response Examples

Register (201):
{
  "message": "User registered successfully",
  "userId": "<uuid>"
}

Login (200):
{
  "token": "<jwt>",
  "refreshToken": "<refresh_jwt>",
  "expiresAt": "2025-01-01T00:00:00Z"
}

Onboarding Start (201):
{
  "userId": "<uuid>",
  "onboardingStatus": "started",
  "nextStep": "email_verification",
  "sessionToken": null
}

Onboarding Status (200):
{
  "userId": "<uuid>",
  "onboardingStatus": "kyc_pending",
  "kycStatus": "pending",
  "currentStep": "kyc_submission",
  "completedSteps": ["registration", "email_verification"],
  "walletStatus": null,
  "canProceed": true,
  "requiredActions": ["Submit KYC documents"]
}

KYC Submit (202):
{
  "message": "KYC documents submitted successfully",
  "status": "processing",
  "user_id": "<uuid>",
  "next_steps": ["Wait for KYC review", "Check onboarding status for updates"]
}

Error Handling Scenarios
- Authentication:
  - Missing token on protected endpoints → `401 Unauthorized`.
  - Invalid/expired token → `401 Unauthorized`.
- Validation:
  - Invalid email/phone format → `400 Bad Request`.
  - Missing required fields → `400 Bad Request`.
- KYC:
  - Not eligible for KYC → `403 Forbidden` with `KYC_NOT_ELIGIBLE`.
  - Malformed callback → `400 Bad Request`.
  - Unknown provider_ref → `404 Not Found`.
- Conflict:
  - Duplicate registration/onboarding → `409 Conflict`.

Troubleshooting
- Ensure route registration matches server logs (see `internal/api/routes/routes.go`).
- Check DI container wiring for handlers and services.
- Validate database migrations are applied (users, onboarding_flows, kyc_submissions, wallets).
- Inspect logs for `request_id` on failures; verify payload structure.
- Ensure requests to protected onboarding endpoints include a valid JWT obtained from the authentication flow.

References
- `internal/api/routes/routes.go` — route definitions.
- `internal/api/handlers/onboarding_handlers.go` — onboarding handler logic and error responses.
- `internal/api/handlers/auth_signup_handlers.go` — signup/verify-code/resend-code handlers.
- `internal/domain/services/onboarding/service.go` — onboarding business logic.
- `internal/domain/entities/onboarding_entities.go` — types and response models.
