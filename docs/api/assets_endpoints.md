# Alpaca Assets API Endpoints

## Overview

These endpoints provide access to tradable assets (stocks, ETFs) available through the Alpaca Broker API. All endpoints require authentication.

**Base Path**: `/api/v1/assets`

**Authentication**: Required (Bearer Token)

---

## Endpoints

### 1. List All Assets

Get a paginated list of all tradable assets with optional filtering.

**Endpoint**: `GET /api/v1/assets`

**Query Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `status` | string | No | `active` | Asset status filter (`active`, `inactive`) |
| `asset_class` | string | No | - | Asset class filter (`us_equity`, `crypto`) |
| `exchange` | string | No | - | Exchange filter (`NASDAQ`, `NYSE`, `ARCA`, `BATS`) |
| `tradable` | boolean | No | `true` | Filter by tradability |
| `fractionable` | boolean | No | - | Filter by fractional shares support |
| `shortable` | boolean | No | - | Filter by short selling support |
| `easy_to_borrow` | boolean | No | - | Filter by easy-to-borrow status |
| `search` | string | No | - | Search by symbol or name |
| `page` | integer | No | `1` | Page number |
| `page_size` | integer | No | `100` | Items per page (max 500) |

**Response**: `200 OK`

```json
{
  "assets": [
    {
      "id": "b0b6dd9d-8b9b-48a9-ba46-b9d54906e415",
      "class": "us_equity",
      "exchange": "NASDAQ",
      "symbol": "AAPL",
      "name": "Apple Inc.",
      "status": "active",
      "tradable": true,
      "marginable": true,
      "shortable": true,
      "easy_to_borrow": true,
      "fractionable": true,
      "min_order_size": "0.001",
      "min_trade_increment": "0.001",
      "price_increment": "0.01"
    }
  ],
  "total_count": 10500,
  "page": 1,
  "page_size": 100
}
```

**Example Requests**:

```bash
# Get all active, tradable assets
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets"

# Filter by exchange
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets?exchange=NASDAQ&page_size=50"

# Search for assets
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets?search=apple"

# Get fractionable assets only
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets?fractionable=true"
```

---

### 2. Get Asset Details

Retrieve detailed information about a specific asset by symbol or asset ID.

**Endpoint**: `GET /api/v1/assets/{symbol_or_id}`

**Path Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol_or_id` | string | Yes | Asset symbol (e.g., `AAPL`) or Asset ID (UUID) |

**Response**: `200 OK`

```json
{
  "id": "b0b6dd9d-8b9b-48a9-ba46-b9d54906e415",
  "class": "us_equity",
  "exchange": "NASDAQ",
  "symbol": "AAPL",
  "name": "Apple Inc.",
  "status": "active",
  "tradable": true,
  "marginable": true,
  "shortable": true,
  "easy_to_borrow": true,
  "fractionable": true,
  "min_order_size": "0.001",
  "min_trade_increment": "0.001",
  "price_increment": "0.01"
}
```

**Error Responses**:

- `400 Bad Request`: Invalid symbol or ID format
- `404 Not Found`: Asset not found
- `500 Internal Server Error`: Server error

**Example Requests**:

```bash
# Get Apple stock details
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets/AAPL"

# Get asset by ID
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets/b0b6dd9d-8b9b-48a9-ba46-b9d54906e415"
```

---

### 3. Search Assets

Search for assets by symbol or company name.

**Endpoint**: `GET /api/v1/assets/search`

**Query Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `q` | string | Yes | - | Search query (symbol or name) |
| `limit` | integer | No | `20` | Maximum results to return (max 100) |

**Response**: `200 OK`

```json
{
  "assets": [
    {
      "id": "b0b6dd9d-8b9b-48a9-ba46-b9d54906e415",
      "class": "us_equity",
      "exchange": "NASDAQ",
      "symbol": "AAPL",
      "name": "Apple Inc.",
      "status": "active",
      "tradable": true,
      "marginable": true,
      "shortable": true,
      "easy_to_borrow": true,
      "fractionable": true
    }
  ],
  "total_count": 1,
  "page": 1,
  "page_size": 1
}
```

**Example Requests**:

```bash
# Search for Apple
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets/search?q=apple"

# Search for Tesla with limit
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets/search?q=tesla&limit=5"

# Search by partial symbol
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets/search?q=AAP"
```

---

### 4. Get Popular Assets

Retrieve a curated list of popular/trending stocks and ETFs.

**Endpoint**: `GET /api/v1/assets/popular`

**Response**: `200 OK`

```json
{
  "assets": [
    {
      "id": "b0b6dd9d-8b9b-48a9-ba46-b9d54906e415",
      "class": "us_equity",
      "exchange": "NASDAQ",
      "symbol": "AAPL",
      "name": "Apple Inc.",
      "status": "active",
      "tradable": true,
      "fractionable": true
    },
    {
      "id": "f801f835-bfe6-4a9d-a6b1-ccbb84bfd75f",
      "class": "us_equity",
      "exchange": "NASDAQ",
      "symbol": "TSLA",
      "name": "Tesla, Inc.",
      "status": "active",
      "tradable": true,
      "fractionable": true
    }
  ],
  "total_count": 17,
  "page": 1,
  "page_size": 17
}
```

**Popular Assets Include**:
- Tech Giants: AAPL, MSFT, GOOGL, AMZN, META, NVDA, TSLA
- Popular ETFs: SPY, QQQ, VOO, VTI, IVV
- Other Popular Stocks: NFLX, AMD, INTC, DIS, BA

**Example Request**:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets/popular"
```

---

### 5. Get Assets by Exchange

Retrieve assets listed on a specific exchange.

**Endpoint**: `GET /api/v1/assets/exchange/{exchange}`

**Path Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `exchange` | string | Yes | Exchange code (`NASDAQ`, `NYSE`, `ARCA`, `BATS`, `AMEX`) |

**Query Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | `1` | Page number |
| `page_size` | integer | No | `50` | Items per page (max 200) |

**Response**: `200 OK`

```json
{
  "assets": [
    {
      "id": "b0b6dd9d-8b9b-48a9-ba46-b9d54906e415",
      "class": "us_equity",
      "exchange": "NASDAQ",
      "symbol": "AAPL",
      "name": "Apple Inc.",
      "status": "active",
      "tradable": true
    }
  ],
  "total_count": 3500,
  "page": 1,
  "page_size": 50
}
```

**Example Requests**:

```bash
# Get NASDAQ assets
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets/exchange/NASDAQ"

# Get NYSE assets with pagination
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.stackservice.com/api/v1/assets/exchange/NYSE?page=2&page_size=100"
```

---

## Asset Object Schema

### AlpacaAssetResponse

| Field | Type | Description |
|-------|------|-------------|
| `id` | string (UUID) | Unique asset identifier |
| `class` | string | Asset class (`us_equity`, `crypto`) |
| `exchange` | string | Exchange where asset is listed |
| `symbol` | string | Trading symbol |
| `name` | string | Company or asset name |
| `status` | string | Asset status (`active`, `inactive`) |
| `tradable` | boolean | Whether asset can be traded |
| `marginable` | boolean | Whether asset can be traded on margin |
| `shortable` | boolean | Whether asset can be sold short |
| `easy_to_borrow` | boolean | Whether asset is easy to borrow for short selling |
| `fractionable` | boolean | Whether fractional shares are supported |
| `min_order_size` | string (decimal) | Minimum order size (optional) |
| `min_trade_increment` | string (decimal) | Minimum trade increment (optional) |
| `price_increment` | string (decimal) | Minimum price increment (optional) |

---

## Error Responses

### Standard Error Format

```json
{
  "code": "ERROR_CODE",
  "error": "Human-readable error message",
  "details": "Additional error details (optional)"
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_PARAMETER` | 400 | Invalid or missing required parameter |
| `INVALID_EXCHANGE` | 400 | Invalid exchange code provided |
| `ASSET_NOT_FOUND` | 404 | Asset does not exist |
| `ASSETS_FETCH_ERROR` | 500 | Failed to retrieve assets from broker |
| `SEARCH_ERROR` | 500 | Failed to search assets |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication token |

---

## Usage Examples

### React Native / JavaScript

```javascript
// List all assets
const getAssets = async (filters = {}) => {
  const params = new URLSearchParams(filters);
  const response = await fetch(
    `https://api.stackservice.com/api/v1/assets?${params}`,
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      }
    }
  );
  return response.json();
};

// Get specific asset
const getAsset = async (symbol) => {
  const response = await fetch(
    `https://api.stackservice.com/api/v1/assets/${symbol}`,
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    }
  );
  return response.json();
};

// Search assets
const searchAssets = async (query) => {
  const response = await fetch(
    `https://api.stackservice.com/api/v1/assets/search?q=${encodeURIComponent(query)}`,
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    }
  );
  return response.json();
};

// Get popular assets
const getPopularAssets = async () => {
  const response = await fetch(
    'https://api.stackservice.com/api/v1/assets/popular',
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    }
  );
  return response.json();
};
```

### Python

```python
import requests

BASE_URL = "https://api.stackservice.com/api/v1"
ACCESS_TOKEN = "your_access_token"

headers = {
    "Authorization": f"Bearer {ACCESS_TOKEN}",
    "Content-Type": "application/json"
}

# List all assets
def get_assets(filters=None):
    response = requests.get(
        f"{BASE_URL}/assets",
        headers=headers,
        params=filters or {}
    )
    return response.json()

# Get specific asset
def get_asset(symbol):
    response = requests.get(
        f"{BASE_URL}/assets/{symbol}",
        headers=headers
    )
    return response.json()

# Search assets
def search_assets(query, limit=20):
    response = requests.get(
        f"{BASE_URL}/assets/search",
        headers=headers,
        params={"q": query, "limit": limit}
    )
    return response.json()
```

---

## Best Practices

### 1. Caching

Assets data changes infrequently. Consider caching responses:

```javascript
// Cache popular assets for 1 hour
const CACHE_TTL = 3600000; // 1 hour in ms

const getCachedPopularAssets = async () => {
  const cached = localStorage.getItem('popular_assets');
  if (cached) {
    const { data, timestamp } = JSON.parse(cached);
    if (Date.now() - timestamp < CACHE_TTL) {
      return data;
    }
  }
  
  const data = await getPopularAssets();
  localStorage.setItem('popular_assets', JSON.stringify({
    data,
    timestamp: Date.now()
  }));
  return data;
};
```

### 2. Pagination

For large lists, use pagination efficiently:

```javascript
const getAllAssets = async () => {
  let allAssets = [];
  let page = 1;
  let hasMore = true;
  
  while (hasMore) {
    const response = await getAssets({ page, page_size: 500 });
    allAssets = [...allAssets, ...response.assets];
    
    // Check if there are more pages
    hasMore = response.assets.length === response.page_size;
    page++;
  }
  
  return allAssets;
};
```

### 3. Search Debouncing

Implement debouncing for search to reduce API calls:

```javascript
const debounce = (func, wait) => {
  let timeout;
  return (...args) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func.apply(this, args), wait);
  };
};

const debouncedSearch = debounce(async (query) => {
  if (query.length < 2) return;
  const results = await searchAssets(query);
  // Update UI with results
}, 300);
```

### 4. Error Handling

Always handle errors gracefully:

```javascript
const getAssetSafely = async (symbol) => {
  try {
    return await getAsset(symbol);
  } catch (error) {
    if (error.response?.status === 404) {
      console.warn(`Asset ${symbol} not found`);
      return null;
    }
    throw error; // Re-throw other errors
  }
};
```

---

## Rate Limits

- **Alpaca Trading API**: 200 requests/minute
- The client automatically handles rate limits with exponential backoff retry

---

## Additional Resources

- [Alpaca Broker API Documentation](https://docs.alpaca.markets/reference/broker-api)
- [Alpaca Integration Guide](../alpaca_integration.md)
- [Project Architecture](../architecture.md)

---

**Last Updated**: 2025-10-29  
**Version**: 1.0  
**Author**: Stack Service Team
