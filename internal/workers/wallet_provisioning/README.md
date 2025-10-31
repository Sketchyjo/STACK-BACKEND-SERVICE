# Wallet Provisioning Worker

Asynchronous wallet provisioning worker with idempotent processing, exponential backoff retries, comprehensive audit logging, and full OpenTelemetry instrumentation.

## Features

### Core Functionality
- ✅ **Idempotent Processing**: Safe to retry, checks for existing wallets
- ✅ **Multi-Chain Support**: ETH, SOL, APTOS with per-chain status
- ✅ **Exponential Backoff**: Configurable retries (1m → 30m max) with jitter
- ✅ **Smart Error Classification**: Distinguishes retryable vs non-retryable errors
- ✅ **Circuit Breaker**: Integrates with Circle API client protection
- ✅ **No Key Exposure**: All keys remain with Circle custody

### Observability
- ✅ **Audit Logging**: Every operation persisted to database with before/after snapshots
- ✅ **OpenTelemetry Tracing**: Distributed tracing across all operations
- ✅ **Metrics Collection**: Success rate, duration, error types, active jobs
- ✅ **Structured Logging**: JSON logs with trace correlation

## Architecture

```
┌─────────────────────┐
│  KYC Approved       │
│  (Webhook/Callback) │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Onboarding Service  │
│  - Enqueues Job     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐     ┌──────────────────┐
│ Provisioning Job    │────▶│  audit_logs      │
│ (status: queued)    │     │  (job created)   │
└──────────┬──────────┘     └──────────────────┘
           │
           ▼
┌─────────────────────┐
│   Scheduler         │
│   (polls every 30s) │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐     ┌──────────────────┐
│   Worker            │────▶│  Circle API      │
│   - Process Job     │     │  (create wallets)│
│   - Create Wallets  │     └──────────────────┘
│   - Retry on Fail   │
└──────────┬──────────┘
           │
           ├─Success──▶ managed_wallets (status: live)
           │            audit_logs (provision_complete)
           │
           └─Failure──▶ job (status: retry, next_retry_at)
                        audit_logs (provision_failed)
```

## Configuration

### Worker Config
```go
type Config struct {
    MaxAttempts         int           // Default: 5
    BaseBackoffDuration time.Duration // Default: 1 minute
    MaxBackoffDuration  time.Duration // Default: 30 minutes
    JitterFactor        float64       // Default: 0.1 (10%)
    ChainsToProvision   []WalletChain // Default: [ETH, SOL, APTOS]
}
```

### Scheduler Config
```go
type SchedulerConfig struct {
    PollInterval     time.Duration // Default: 30 seconds
    MaxConcurrency   int           // Default: 5 concurrent jobs
    JobBatchSize     int           // Default: 10 jobs per poll
    ShutdownTimeout  time.Duration // Default: 60 seconds
    EnableRetries    bool          // Default: true
}
```

## OpenTelemetry Instrumentation

### Metrics Emitted

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `wallet_provisioning_jobs_processed_total` | Counter | Total jobs processed | status |
| `wallet_provisioning_job_duration_seconds` | Histogram | Job processing duration | status |
| `wallet_provisioning_job_retries_total` | Counter | Total retry attempts | job.id, attempt.count |
| `wallet_provisioning_job_errors_total` | Counter | Total errors by type | error.type |
| `wallet_provisioning_wallets_created_total` | Counter | Wallets created | user.id, chain |
| `wallet_provisioning_active_jobs` | UpDownCounter | Currently active jobs | - |
| `wallet_provisioning_scheduler_running` | UpDownCounter | Scheduler status (1=running) | - |

### Traces

**Top-Level Spans:**
- `ProcessWalletProvisioningJob` - Full job processing
- `SchedulerStart` - Scheduler startup
- `SchedulerStop` - Scheduler shutdown

**Child Spans:**
- `CreateWalletForChain` - Per-chain wallet creation
- `CircleAPI:CreateWalletSet` - Circle API calls
- `CircleAPI:CreateWallet` - Circle wallet creation

**Span Attributes:**
- `job.id` - Provisioning job UUID
- `user.id` - User UUID
- `chain` - Blockchain (ETH, SOL, APTOS)
- `job.status` - success/failed/retry
- `job.duration_seconds` - Processing duration
- `error.type` - Error classification
- `attempt.count` - Retry attempt number

### Using Telemetry Wrappers

```go
// Create instrumented worker
worker := NewWorker(...)
telemetryWorker, err := NewTelemetryWorker(worker, logger)
if err != nil {
    log.Fatal(err)
}

// Create instrumented scheduler
scheduler := NewScheduler(...)
telemetryScheduler, err := NewTelemetryScheduler(scheduler, logger)
if err != nil {
    log.Fatal(err)
}

// Start with telemetry
telemetryScheduler.Start()
defer telemetryScheduler.Stop()
```

## API Endpoints

### Worker Management

#### Get Worker Status
```bash
GET /api/v1/admin/workers/status
```

**Response:**
```json
{
  "worker": {
    "type": "wallet_provisioning",
    "status": "operational"
  },
  "scheduler": {
    "is_running": true,
    "poll_interval": "30s",
    "max_concurrency": 5,
    "active_jobs": 2
  },
  "metrics": {
    "total_jobs_processed": 150,
    "successful_jobs": 142,
    "failed_jobs": 8,
    "average_duration_ms": 2341,
    "last_processed_at": "2025-10-04T04:30:00Z",
    "errors_by_type": {
      "circle_5xx": 3,
      "timeout": 2,
      "network": 3
    }
  }
}
```

#### Get Metrics
```bash
GET /api/v1/admin/workers/metrics
```

#### Health Check
```bash
GET /api/v1/admin/workers/health
```

#### Restart Scheduler
```bash
POST /api/v1/admin/workers/restart
```

## Error Handling

### Error Types

| Type | Retryable | Description |
|------|-----------|-------------|
| `circle_5xx` | ✅ Yes | Circle API server error |
| `circle_rate_limit` | ✅ Yes | Rate limit (429) |
| `circle_4xx` | ❌ No | Client error (bad request) |
| `timeout` | ✅ Yes | Request timeout |
| `network` | ✅ Yes | Network/connection issues |
| `validation` | ❌ No | Input validation error |

### Retry Strategy

1. **Attempt 1**: Immediate (0s delay)
2. **Attempt 2**: 1 minute + jitter
3. **Attempt 3**: 2 minutes + jitter
4. **Attempt 4**: 4 minutes + jitter
5. **Attempt 5**: 8 minutes + jitter
6. **Failed**: Marked as permanently failed, alert triggered

## Audit Trail

Every operation is logged to `audit_logs` table:

```sql
SELECT * FROM audit_logs 
WHERE action LIKE '%wallet-worker%' 
ORDER BY created_at DESC;
```

**Audit Events:**
- `wallet-worker:provision_start` - Job processing started
- `wallet-worker:wallet_created` - Individual wallet created
- `wallet-worker:provision_complete` - Job completed successfully
- `wallet-worker:provision_failed` - Job failed (with retry info)

## Monitoring

### Prometheus Queries

**Success Rate:**
```promql
rate(wallet_provisioning_jobs_processed_total{status="success"}[5m]) 
/ 
rate(wallet_provisioning_jobs_processed_total[5m])
```

**Average Duration:**
```promql
rate(wallet_provisioning_job_duration_seconds_sum[5m]) 
/ 
rate(wallet_provisioning_job_duration_seconds_count[5m])
```

**Error Rate by Type:**
```promql
sum by (error_type) (
  rate(wallet_provisioning_job_errors_total[5m])
)
```

**Active Jobs:**
```promql
wallet_provisioning_active_jobs
```

### Grafana Dashboard

Recommended panels:
1. **Success Rate** (gauge) - Target: >90%
2. **Processing Duration** (histogram) - P50, P95, P99
3. **Active Jobs** (graph) - Real-time concurrency
4. **Error Types** (pie chart) - Error distribution
5. **Jobs Processed** (counter) - Throughput over time

## Troubleshooting

### Job Stuck in `in_progress`
```bash
# Check worker logs
grep "job_id=<JOB_ID>" /var/log/stack_service.log

# Check if scheduler is running
curl http://localhost:8080/api/v1/admin/workers/health
```

### High Retry Rate
```sql
-- Check job retry patterns
SELECT 
    user_id,
    status,
    attempt_count,
    error_message,
    next_retry_at
FROM wallet_provisioning_jobs
WHERE status = 'retry'
ORDER BY next_retry_at ASC;
```

### Wallet Creation Failures
```sql
-- Check audit logs for specific user
SELECT 
    action,
    resource_type,
    changes,
    status,
    error_message,
    created_at
FROM audit_logs
WHERE user_id = '<USER_UUID>'
AND action LIKE '%wallet%'
ORDER BY created_at DESC;
```

## Testing

### Local Testing

```bash
# Start application with worker
go run cmd/main.go

# Check scheduler started
curl http://localhost:8080/api/v1/admin/workers/status

# Trigger KYC approval (creates job)
curl -X POST http://localhost:8080/api/v1/onboarding/kyc/callback/<provider_ref> \
  -H "Content-Type: application/json" \
  -d '{"status": "approved"}'

# Monitor job processing
watch -n 1 'curl -s http://localhost:8080/api/v1/admin/workers/metrics | jq'
```

### Integration Testing

See `internal/workers/wallet_provisioning/worker_test.go` for:
- Idempotency tests
- Retry logic tests
- Error classification tests
- Audit logging verification

## Production Deployment

### Environment Variables
```bash
# Worker configuration (optional, uses defaults if not set)
WORKER_MAX_ATTEMPTS=5
WORKER_BASE_BACKOFF=1m
WORKER_MAX_BACKOFF=30m
WORKER_POLL_INTERVAL=30s
WORKER_MAX_CONCURRENCY=5

# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
OTEL_SERVICE_NAME=wallet-provisioning-worker
```

### Alerts

Recommended alerts:
- Success rate < 90% for 5 minutes
- Average duration > 10 seconds
- Scheduler down
- Job stuck in `in_progress` > 30 minutes
- Error rate > 10%

## License

Internal use only - Stack Service
