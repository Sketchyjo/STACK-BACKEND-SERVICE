# Stack Service API Testing Guide with Postman

This guide provides comprehensive instructions for testing the Stack Service APIs using Postman.

## Prerequisites

1. **Install Postman**: Download from [https://www.postman.com/downloads/](https://www.postman.com/downloads/)
2. **Run the application**: Follow the setup instructions to start the Stack Service
3. **Database setup**: Ensure PostgreSQL is running and migrations have been applied

## Environment Setup

### 1. Create a Postman Environment

Create a new environment in Postman with the following variables:

```
base_url: http://localhost:8080
api_version: /api/v1
auth_token: (will be set during authentication)
user_id: (will be set after registration)
```

### 2. Environment Variables

- **base_url**: Your API base URL (default: `http://localhost:8080`)
- **api_version**: API version path (`/api/v1`)
- **auth_token**: JWT token (automatically set after successful authentication)
- **user_id**: User UUID (automatically set after registration)

## API Endpoints Collection

### 1. Health Check

**GET** `{{base_url}}/health`

**Description**: Check if the API is running and healthy.

**Expected Response**:
```json
{
    "status": "ok",
    "timestamp": "2025-01-03T17:04:21Z",
    "version": "1.0.0",
    "environment": "development"
}
```

### 2. Onboarding Endpoints

#### 2.1 Start Onboarding

**POST** `{{base_url}}{{api_version}}/onboarding/start`

**Headers**:
```
Content-Type: application/json
```

**Body** (JSON):
```json
{
    "email": "testuser@example.com",
    "phone": "+12345678901"
}
```

**Tests Script** (Add to Tests tab):
```javascript
pm.test("Status code is 201", function () {
    pm.response.to.have.status(201);
});

pm.test("Response has required fields", function () {
    const responseJson = pm.response.json();
    pm.expect(responseJson).to.have.property('userId');
    pm.expect(responseJson).to.have.property('onboardingStatus');
    pm.expect(responseJson).to.have.property('nextStep');
    
    // Set environment variables
    pm.environment.set("user_id", responseJson.userId);
    if (responseJson.sessionToken) {
        pm.environment.set("auth_token", responseJson.sessionToken);
    }
});
```

#### 2.2 Get Onboarding Status

**GET** `{{base_url}}{{api_version}}/onboarding/status/{{user_id}}`

**Headers**:
```
Authorization: Bearer {{auth_token}}
```

**Tests Script**:
```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has user status", function () {
    const responseJson = pm.response.json();
    pm.expect(responseJson).to.have.property('userId');
    pm.expect(responseJson).to.have.property('onboardingStatus');
    pm.expect(responseJson).to.have.property('canProceed');
});
```

#### 2.3 Submit KYC Documents

**POST** `{{base_url}}{{api_version}}/onboarding/kyc/submit`

**Headers**:
```
Authorization: Bearer {{auth_token}}
Content-Type: application/json
```

**Body** (JSON):
```json
{
    "documentType": "passport",
    "documents": [
        {
            "type": "passport",
            "fileUrl": "https://example.com/document.jpg",
            "contentType": "image/jpeg"
        }
    ],
    "personalInfo": {
        "firstName": "John",
        "lastName": "Doe",
        "dateOfBirth": "1995-01-15T00:00:00Z",
        "country": "US",
        "address": {
            "street": "123 Main St",
            "city": "New York",
            "state": "NY",
            "postalCode": "10001",
            "country": "US"
        }
    }
}
```

**Tests Script**:
```javascript
pm.test("Status code is 201", function () {
    pm.response.to.have.status(201);
});

pm.test("KYC submission created", function () {
    const responseJson = pm.response.json();
    pm.expect(responseJson).to.have.property('submissionId');
    pm.expect(responseJson).to.have.property('status');
});
```

### 3. Wallet Management Endpoints

#### 3.1 Get Wallet Status

**GET** `{{base_url}}{{api_version}}/wallets/status/{{user_id}}`

**Headers**:
```
Authorization: Bearer {{auth_token}}
```

**Tests Script**:
```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Wallet status response structure", function () {
    const responseJson = pm.response.json();
    pm.expect(responseJson).to.have.property('userId');
    pm.expect(responseJson).to.have.property('totalWallets');
    pm.expect(responseJson).to.have.property('walletsByChain');
});
```

#### 3.2 Get Wallet Addresses

**GET** `{{base_url}}{{api_version}}/wallets/addresses/{{user_id}}`

**Headers**:
```
Authorization: Bearer {{auth_token}}
```

**Query Parameters** (Optional):
```
chain: ETH
```

**Tests Script**:
```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Wallets array exists", function () {
    const responseJson = pm.response.json();
    pm.expect(responseJson).to.have.property('wallets');
    pm.expect(responseJson.wallets).to.be.an('array');
});
```

#### 3.3 Create Wallets

**POST** `{{base_url}}{{api_version}}/wallets/create/{{user_id}}`

**Headers**:
```
Authorization: Bearer {{auth_token}}
Content-Type: application/json
```

**Body** (JSON):
```json
{
    "chains": ["ETH", "MATIC", "AVAX"]
}
```

**Tests Script**:
```javascript
pm.test("Status code is 202", function () {
    pm.response.to.have.status(202);
});

pm.test("Job created successfully", function () {
    const responseJson = pm.response.json();
    pm.expect(responseJson).to.have.property('jobId');
    pm.expect(responseJson).to.have.property('status');
    pm.expect(responseJson.status).to.equal('queued');
});
```

## Pre-request Scripts

### Authentication Helper

Add this to the Collection's Pre-request Script for automatic token handling:

```javascript
// Check if we have a valid token
const token = pm.environment.get("auth_token");
if (token) {
    pm.request.headers.add({
        key: 'Authorization',
        value: 'Bearer ' + token
    });
}

// Add request ID for tracking
const requestId = pm.variables.replaceIn('{{$guid}}');
pm.request.headers.add({
    key: 'X-Request-ID',
    value: requestId
});
```

## Testing Scenarios

### Scenario 1: Complete Onboarding Flow

1. **Start Onboarding** - Register a new user
2. **Check Status** - Verify onboarding has started
3. **Submit KYC** - Upload identity documents
4. **Check Status** - Verify KYC is processing
5. **Create Wallets** - Provision crypto wallets
6. **Check Wallet Status** - Verify wallets are created

### Scenario 2: Wallet Management

1. **Get Wallet Status** - Check current wallet state
2. **Get Wallet Addresses** - Retrieve wallet addresses by chain
3. **Create Additional Wallets** - Add new blockchain support

### Scenario 3: Error Handling

Test various error scenarios:
- Invalid user IDs
- Missing authentication tokens
- Invalid request payloads
- Duplicate email registration

## Environment Configuration

### Development Environment
```json
{
    "base_url": "http://localhost:8080",
    "api_version": "/api/v1",
    "auth_token": "",
    "user_id": ""
}
```

### Staging Environment
```json
{
    "base_url": "https://api-staging.stackservice.com",
    "api_version": "/api/v1",
    "auth_token": "",
    "user_id": ""
}
```

## Expected Response Codes

| Endpoint | Method | Success Code | Description |
|----------|--------|--------------|-------------|
| `/health` | GET | 200 | Service healthy |
| `/onboarding/start` | POST | 201 | Onboarding started |
| `/onboarding/status/:id` | GET | 200 | Status retrieved |
| `/onboarding/kyc/submit` | POST | 201 | KYC submitted |
| `/wallets/status/:id` | GET | 200 | Wallet status |
| `/wallets/addresses/:id` | GET | 200 | Addresses retrieved |
| `/wallets/create/:id` | POST | 202 | Wallet creation queued |

## Common Error Responses

### 400 Bad Request
```json
{
    "error": "validation_error",
    "message": "Invalid request payload",
    "details": {
        "field": "email",
        "code": "invalid_format"
    },
    "timestamp": "2025-01-03T17:04:21Z",
    "requestId": "req-123456"
}
```

### 401 Unauthorized
```json
{
    "error": "unauthorized",
    "message": "Authentication token required",
    "timestamp": "2025-01-03T17:04:21Z",
    "requestId": "req-123456"
}
```

### 404 Not Found
```json
{
    "error": "not_found",
    "message": "User not found",
    "timestamp": "2025-01-03T17:04:21Z",
    "requestId": "req-123456"
}
```

### 500 Internal Server Error
```json
{
    "error": "internal_server_error",
    "message": "An unexpected error occurred",
    "timestamp": "2025-01-03T17:04:21Z",
    "requestId": "req-123456"
}
```

## Additional Testing Tips

### 1. Collection Variables
Set up collection-level variables for commonly used values:
- Test user emails
- Common request headers
- Default timeout values

### 2. Data-Driven Testing
Use Postman's data files to test multiple scenarios:
- Create CSV files with different user data
- Test edge cases with various input combinations

### 3. Monitoring
Set up Postman monitors to run tests automatically:
- Schedule regular health checks
- Monitor critical user flows
- Alert on failures

### 4. Documentation
Use Postman's documentation feature to:
- Add descriptions for each endpoint
- Include example requests and responses
- Share with team members

## Environment Variables Reference

| Variable | Description | Example |
|----------|-------------|---------|
| `base_url` | API base URL | `http://localhost:8080` |
| `api_version` | API version path | `/api/v1` |
| `auth_token` | JWT authentication token | `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...` |
| `user_id` | User UUID | `550e8400-e29b-41d4-a716-446655440000` |

This guide should help you thoroughly test all the Stack Service APIs using Postman. Start with the health check endpoint to ensure everything is working, then proceed through the onboarding flow systematically.