# 0G Integration Documentation

This document provides comprehensive documentation for the 0G Storage and Compute integration within the STACK platform, focusing on the AI-CFO service implementation.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Configuration](#configuration)
4. [API Endpoints](#api-endpoints)
5. [Usage Examples](#usage-examples)
6. [Development Guide](#development-guide)
7. [Testing](#testing)
8. [Monitoring](#monitoring)
9. [Troubleshooting](#troubleshooting)

## Overview

The 0G integration provides:
- **Secure object storage** for AI artifacts, summaries, and analysis results
- **AI inference capabilities** for portfolio analysis and summary generation
- **Scheduled weekly summaries** for active users
- **On-demand analysis** for specific portfolio aspects
- **Content-addressable storage** with metadata management
- **Comprehensive observability** with metrics, tracing, and structured logging

## Architecture

### Core Components

```
┌─────────────────┐
│   AI-CFO API    │
│   (Public)      │
└─────────┬───────┘
          │
┌─────────┴───────┐    ┌─────────────────┐
│  AI-CFO Service │────│ Weekly Summary  │
│                 │    │   Scheduler     │
└─────────┬───────┘    └─────────────────┘
          │
┌─────────┴───────┐    ┌─────────────────┐
│ Inference       │────│ Storage Client  │
│ Gateway         │    │                 │
└─────────────────┘    └─────────┬───────┘
                                 │
                       ┌─────────┴───────┐
                       │ Namespace       │
                       │ Manager         │
                       └─────────────────┘
```

### Service Layers

1. **API Layer**: HTTP endpoints for public and internal access
2. **Service Layer**: Business logic and orchestration
3. **Infrastructure Layer**: Storage and compute clients
4. **Domain Layer**: Entities and business rules

## Configuration

### Environment Variables

```bash
# 0G Storage Configuration
ZEROG_STORAGE_ENDPOINT=https://storage.0g.ai
ZEROG_STORAGE_ACCESS_KEY=your_access_key
ZEROG_STORAGE_SECRET_KEY=your_secret_key
ZEROG_STORAGE_BUCKET=stack-platform
ZEROG_STORAGE_REGION=us-west-2

# 0G Compute Configuration  
ZEROG_COMPUTE_ENDPOINT=https://compute.0g.ai
ZEROG_COMPUTE_API_KEY=your_api_key
ZEROG_COMPUTE_MODEL=gpt-4

# Scheduler Configuration
ZEROG_SCHEDULER_ENABLED=true
ZEROG_SCHEDULER_CRON="0 0 8 * * MON"  # Monday 8AM
ZEROG_SCHEDULER_BATCH_SIZE=50
ZEROG_SCHEDULER_CONCURRENCY_LIMIT=5

# Health Check Configuration
ZEROG_HEALTH_CHECK_INTERVAL=300s
ZEROG_HEALTH_CHECK_TIMEOUT=30s
```

### Configuration Structure

```yaml
zerog:
  storage:
    endpoint: "https://storage.0g.ai"
    access_key: "${ZEROG_STORAGE_ACCESS_KEY}"
    secret_key: "${ZEROG_STORAGE_SECRET_KEY}"
    bucket: "stack-platform"
    region: "us-west-2"
    max_retries: 3
    retry_delay: "1s"
    timeout: "30s"
    
  compute:
    endpoint: "https://compute.0g.ai"  
    api_key: "${ZEROG_COMPUTE_API_KEY}"
    model: "gpt-4"
    max_tokens: 2000
    temperature: 0.3
    max_retries: 3
    retry_delay: "2s"
    timeout: "60s"
    
  scheduler:
    enabled: true
    cron_expression: "0 0 8 * * MON"  # Every Monday at 8 AM
    batch_size: 50
    concurrency_limit: 5
    
  health_check:
    interval: "5m"
    timeout: "30s"
```

## API Endpoints

### Public AI-CFO Endpoints

All public endpoints require JWT authentication via `Authorization: Bearer <token>` header.

#### Get Latest Weekly Summary

```http
GET /api/v1/ai/summary/latest
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "987fcdeb-51a2-43d7-9f8c-123456789abc",
  "week_start": "2024-01-15",
  "title": "Weekly Summary: Jan 15 - Jan 21, 2024",
  "content": "# Weekly Portfolio Summary\n\n...",
  "created_at": "2024-01-22T08:00:00Z",
  "artifact_uri": "0g://ai-summaries/abc123...",
  "metadata": {
    "week_start": "2024-01-15",
    "week_end": "2024-01-21"
  }
}
```

#### Perform On-Demand Analysis

```http
POST /api/v1/ai/analyze
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "analysis_type": "risk",
  "parameters": {
    "timeframe": "30d",
    "include_benchmarks": true
  }
}
```

**Response:**
```json
{
  "request_id": "req_789xyz...",
  "analysis_type": "risk",
  "content": "## Risk Analysis\n\nYour portfolio shows...",
  "content_type": "text/markdown", 
  "insights": [
    {
      "type": "risk",
      "title": "Portfolio Risk Level",
      "description": "Your portfolio maintains a moderate risk profile",
      "impact": "medium",
      "confidence": 0.85
    }
  ],
  "recommendations": [
    "Monitor technology sector exposure...",
    "Consider adding defensive positions..."
  ],
  "metadata": {
    "analysis_type": "risk",
    "timeframe": "30d"
  },
  "tokens_used": 150,
  "processing_time": "2.5s",
  "created_at": "2024-01-22T10:30:00Z",
  "artifact_uri": "0g://ai-artifacts/analysis-789..."
}
```

#### Analysis Types

- `risk`: Portfolio risk assessment
- `performance`: Performance analysis vs benchmarks  
- `diversification`: Diversification analysis
- `allocation`: Asset allocation review
- `rebalancing`: Rebalancing recommendations

#### Health Check

```http
GET /api/v1/ai/health
Authorization: Bearer <jwt_token>
```

### Internal 0G Endpoints

Internal endpoints require API key authentication via `X-Internal-API-Key` header.

#### Storage Health Check

```http  
GET /_internal/0g/health/storage
X-Internal-API-Key: <internal_api_key>
```

#### Store Data

```http
POST /_internal/0g/storage/store  
X-Internal-API-Key: <internal_api_key>
Content-Type: application/json

{
  "data": "<base64_encoded_data>",
  "namespace": "ai-summaries",
  "content_type": "text/markdown",
  "metadata": {
    "user_id": "123...",
    "type": "weekly_summary"
  }
}
```

#### Generate Inference

```http
POST /_internal/0g/inference/generate
X-Internal-API-Key: <internal_api_key>  
Content-Type: application/json

{
  "type": "weekly_summary",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "parameters": {
    "timeframe": "7d"
  },
  "mock_portfolio_data": {
    "total_value": 100000,
    "positions": [...]
  }
}
```

## Usage Examples

### Integration Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/stack-service/stack_service/internal/config"
    "github.com/stack-service/stack_service/internal/zerog"
)

func main() {
    // Load configuration
    cfg := config.LoadZeroGConfig()
    
    // Initialize integration
    integration, err := zerog.NewZeroGIntegration(
        cfg,
        aiSummaryRepo,
        portfolioRepo, 
        logger,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Start services
    ctx := context.Background()
    if err := integration.Start(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Setup routes
    router := gin.New()
    integration.SetupRoutes(router)
    
    // Start server
    router.Run(":8080")
}
```

### Manual Summary Generation

```go
// Trigger weekly summary generation manually
ctx := context.Background()
err := integration.TriggerWeeklySummary(ctx)
if err != nil {
    log.Printf("Failed to generate summary: %v", err)
}
```

### Client Usage Examples

#### JavaScript/Node.js

```javascript
// Get latest summary
const response = await fetch('/api/v1/ai/summary/latest', {
  headers: {
    'Authorization': `Bearer ${jwt_token}`
  }
});

const summary = await response.json();
console.log('Latest summary:', summary.title);

// Perform risk analysis
const analysisResponse = await fetch('/api/v1/ai/analyze', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${jwt_token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    analysis_type: 'risk',
    parameters: { timeframe: '30d' }
  })
});

const analysis = await analysisResponse.json();
console.log('Risk insights:', analysis.insights);
```

#### Python

```python
import requests

# Get latest summary
headers = {'Authorization': f'Bearer {jwt_token}'}
response = requests.get('/api/v1/ai/summary/latest', headers=headers)
summary = response.json()

# Perform performance analysis  
analysis_data = {
    'analysis_type': 'performance',
    'parameters': {'timeframe': '90d'}
}
response = requests.post('/api/v1/ai/analyze', 
                        json=analysis_data, 
                        headers=headers)
analysis = response.json()
```

## Development Guide

### Adding New Analysis Types

1. **Define Constants**:
   ```go
   // In internal/domain/entities/zerog.go
   const (
       AnalysisTypeNewType = "new_type"
   )
   ```

2. **Update Validation**:
   ```go  
   // In handlers/aicfo_handlers.go
   func (h *AICfoHandler) isValidAnalysisType(analysisType string) bool {
       validTypes := []string{
           entities.AnalysisTypeNewType,
           // ... existing types
       }
       // ...
   }
   ```

3. **Add Business Logic**:
   ```go
   // In services/aicfo_service.go
   func (s *AICfoService) PerformOnDemandAnalysis(...) {
       switch analysisType {
       case entities.AnalysisTypeNewType:
           return s.performNewTypeAnalysis(ctx, userID, parameters)
       }
   }
   ```

### Custom Storage Namespaces

```go
// Define namespace
const CustomNamespace = "custom-data"

// Register namespace  
err := namespaceManager.EnsureNamespace(ctx, CustomNamespace, storage.NamespaceConfig{
    RetentionDays:    90,
    CompressionType: "gzip",
    EncryptionKey:   "custom-key",
})
```

### Error Handling

```go
// Service level error handling
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}

// Handler level error handling  
if err != nil {
    span.RecordError(err)
    h.logger.Error("Operation failed",
        zap.Error(err),
        zap.String("user_id", userID.String()),
    )
    
    c.JSON(http.StatusInternalServerError, ErrorResponse{
        Error: "Operation failed",
        Code: "OPERATION_FAILED", 
        Details: err.Error(),
    })
}
```

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/zerog/services/... -v

# Run benchmarks
go test -bench=. ./test/handlers/...
```

### Integration Tests

```bash
# Run integration tests (requires 0G services)
go test -tags=integration ./test/integration/...

# Test with real 0G endpoints (staging)
ZEROG_INTEGRATION_TEST=true go test ./test/integration/...
```

### Manual Testing

```bash
# Test storage health
curl -H "X-Internal-API-Key: test-key" \
     http://localhost:8080/_internal/0g/health/storage

# Test summary retrieval
curl -H "Authorization: Bearer $JWT_TOKEN" \
     http://localhost:8080/api/v1/ai/summary/latest

# Test analysis request
curl -X POST \
     -H "Authorization: Bearer $JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"analysis_type":"risk","parameters":{"timeframe":"30d"}}' \
     http://localhost:8080/api/v1/ai/analyze
```

## Monitoring

### Metrics

Key metrics exposed via Prometheus:

- `zerog_storage_requests_total`: Storage operation counters
- `zerog_storage_request_duration_seconds`: Storage operation latencies
- `zerog_inference_requests_total`: Inference request counters  
- `zerog_inference_tokens_used_total`: Token usage counters
- `zerog_scheduler_runs_total`: Scheduler execution counters
- `zerog_health_check_status`: Health check status gauge

### Tracing

OpenTelemetry spans are created for:

- HTTP requests
- Storage operations
- Inference calls
- Database queries
- Scheduled jobs

### Logging

Structured logging includes:

- Request IDs for correlation
- User IDs (when available)
- Operation types and parameters
- Performance metrics
- Error details with context

### Health Checks

```bash
# Overall health
curl http://localhost:8080/api/v1/ai/health

# Individual service health  
curl -H "X-Internal-API-Key: key" \
     http://localhost:8080/_internal/0g/health/all
```

### Grafana Dashboard Queries

```promql
# Request rate
rate(zerog_storage_requests_total[5m])

# Error rate
rate(zerog_storage_requests_total{status=~"5.."}[5m]) / 
rate(zerog_storage_requests_total[5m])

# 95th percentile latency
histogram_quantile(0.95, rate(zerog_storage_request_duration_seconds_bucket[5m]))

# Token usage trend
increase(zerog_inference_tokens_used_total[1h])
```

## Troubleshooting

### Common Issues

#### 1. Authentication Errors

**Symptoms**: 401 Unauthorized responses

**Solutions**:
- Verify JWT token validity
- Check internal API key configuration
- Ensure middleware is properly configured

#### 2. Storage Connection Issues

**Symptoms**: Storage health checks failing

**Solutions**:
- Verify 0G storage endpoint accessibility
- Check access key and secret key configuration
- Review network connectivity and firewall rules
- Check storage service status

#### 3. Inference Timeouts

**Symptoms**: Analysis requests timing out

**Solutions**:
- Increase inference timeout configuration
- Check 0G compute service availability
- Review model configuration and parameters
- Monitor token usage and limits

#### 4. Scheduler Not Running

**Symptoms**: Weekly summaries not generated

**Solutions**:
- Verify scheduler is enabled in configuration
- Check cron expression syntax
- Review scheduler logs for errors
- Ensure database connectivity for user queries

#### 5. High Memory Usage

**Symptoms**: Out of memory errors

**Solutions**:
- Reduce scheduler batch size
- Lower concurrency limits
- Implement request size limits
- Review data retention policies

### Debug Commands

```bash
# Check service status
curl http://localhost:8080/_internal/0g/health/all

# View scheduler status  
curl http://localhost:8080/api/v1/ai/health | jq '.services.scheduler'

# Check metrics
curl http://localhost:8080/metrics | grep zerog

# View recent logs
docker logs stack-service | grep -i zerog | tail -50
```

### Configuration Validation

```bash
# Test configuration loading
go run cmd/config-test/main.go

# Validate 0G credentials
go run cmd/0g-test/main.go
```

### Performance Tuning

- **Storage Operations**: Adjust `max_retries`, `retry_delay`, and `timeout`
- **Inference Calls**: Tune `max_tokens`, `temperature`, and `timeout`  
- **Scheduler**: Optimize `batch_size` and `concurrency_limit`
- **Health Checks**: Balance `interval` vs resource usage

For additional support, consult the 0G platform documentation or contact the development team.