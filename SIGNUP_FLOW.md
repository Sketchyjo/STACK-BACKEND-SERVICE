# Sign-Up Flow Implementation

## Overview

The sign-up flow has been successfully implemented with email/SMS verification, Redis-backed code storage, and async onboarding/wallet provisioning.

## New Endpoints

### POST /api/v1/auth/signup
Register a new user with email OR phone verification.

**Request:**
```json
{
  "email": "user@example.com",  // OR phone
  "phone": "+1234567890",       // OR email
  "password": "password123"
}
```

**Response:**
```json
{
  "message": "Verification code sent",
  "identifier": "us***@***.com"
}
```

### POST /api/v1/auth/verify-code
Verify the 6-digit code sent during registration.

**Request:**
```json
{
  "email": "user@example.com",  // OR phone
  "phone": "+1234567890",       // OR email
  "code": "123456"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "emailVerified": true,
    "onboardingStatus": "kyc_pending"
  },
  "accessToken": "jwt_token",
  "refreshToken": "refresh_token",
  "expiresAt": "2024-01-01T00:00:00Z"
}
```

### POST /api/v1/auth/resend-code
Resend verification code (rate limited).

**Request:**
```json
{
  "email": "user@example.com"  // OR phone
}
```

## Configuration

Add to `configs/config.yaml`:

```yaml
sms:
  provider: "twilio" # or "mock"
  api_key: ""
  api_secret: ""
  from_number: "+1234567890"
  environment: "development"

verification:
  code_length: 6
  code_ttl_minutes: 10
  max_attempts: 3
  rate_limit_per_hour: 3
```

## Environment Variables

```bash
# SMS Configuration
TWILIO_API_KEY=your_twilio_api_key
TWILIO_API_SECRET=your_twilio_api_secret
TWILIO_FROM_NUMBER=+1234567890

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Features Implemented

✅ **Email & SMS Verification**: Support both email and phone number registration
✅ **6-Digit Codes**: Cryptographically secure random codes
✅ **Redis Storage**: Fast, secure code storage with automatic expiry
✅ **Rate Limiting**: 3 codes per hour per identifier
✅ **JWT Tokens**: Access and refresh tokens issued after verification
✅ **Async Onboarding**: Background KYC and wallet provisioning
✅ **Input Validation**: Email format, E.164 phone format, password strength
✅ **Error Handling**: Contextual error messages without user enumeration
✅ **Security**: Password hashing, code masking, attempt limits

## Architecture

```
Signup Request → Verification Service → Redis → Email/SMS Service
     ↓
User Created (unverified) → Onboarding Job Created → Worker Processes
     ↓
Verify Code → JWT Tokens → User Verified → Onboarding Started
```

## Testing

Run integration tests:
```bash
go test ./test/integration/signup_flow_test.go -v
```

## Database Migration

Run the new migration:
```bash
make migrate-up
```

This creates the `onboarding_jobs` table for async processing.

## Monitoring

The system includes comprehensive logging and metrics:
- Verification code generation/sending
- Rate limiting events
- Failed verification attempts
- Onboarding job processing
- Worker performance metrics

## Security Considerations

- Codes are cryptographically secure (crypto/rand)
- Passwords hashed with bcrypt (cost factor 12)
- Rate limiting prevents abuse
- Sensitive data masked in logs
- Codes expire after 10 minutes
- Maximum 3 verification attempts per code
