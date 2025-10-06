# Webhook Reliability Worker Implementation

## Overview
Professional implementation of a robust, scalable webhook reliability system with automatic retries, dead letter queue (DLQ), and reconciliation for deposit processing.

## Architecture

### Components Implemented

1. **Funding Event Job Entities** (`internal/domain/entities/funding_job_entities.go`)
   - `FundingEventJob`: Core job entity with retry metadata
   - `ProcessingLogEntry`: Tracks individual processing attempts
   - `ReconciliationCandidate`: Represents pending deposits for reconciliation
   - `WebhookMetrics`: Tracks processing metrics
   - Status management: pending → processing → completed/failed/dlq
   - Error categorization: transient, permanent, rpc_failure, unknown
   - Exponential backoff with jitter calculation
   - Idempotency support via tx_hash + chain

2. **Database Schema** (`migrations/006_funding_event_jobs.up.sql`)
   - `funding_event_jobs` table with comprehensive indexing
   - Unique constraint on (tx_hash, chain) for idempotency
   - JSONB fields for webhook_payload and processing_logs
   - Optimized indexes for queue operations and reconciliation
   - Automatic updated_at trigger

3. **Repository Layer** (`internal/infrastructure/repositories/funding_event_job_repository.go`)
   - `FundingEventJobRepository`: Full CRUD operations
   - `Enqueue()`: Idempotent job creation with ON CONFLICT handling
   - `GetNextPendingJobs()`: Uses FOR UPDATE SKIP LOCKED for concurrency
   - `Update()`: Atomic status and metadata updates
   - `GetDLQJobs()`: Retrieves failed jobs for manual review
   - `GetPendingDepositsForReconciliation()`: Finds stuck deposits
   - `GetMetrics()`: Real-time performance metrics

## Next Steps

### 4. Worker Processor (Priority: High)
**File**: `internal/workers/funding_webhook/processor.go`

```go
// Key responsibilities:
- Poll GetNextPendingJobs() continuously
- Process jobs via FundingService.ProcessChainDeposit()
- Implement error categorization:
  * HTTP 5xx, timeout, network → ErrorTypeTransient (retry)
  * HTTP 4xx, validation → ErrorTypePermanent (DLQ)
  * RPC failures → ErrorTypeRPCFailure (retry with longer backoff)
- Calculate next retry delay using job.GetRetryDelay()
- Add processing logs with timing and metadata
- Emit OpenTelemetry metrics:
  * webhook_processed_total{status}
  * webhook_processing_duration_seconds
  * webhook_retry_count
  * webhook_dlq_depth
- Audit logging via audit_logs table
```

**Concurrency Strategy**:
- Run N worker goroutines (configurable, default 5)
- Each worker polls independently using SKIP LOCKED
- Graceful shutdown with context cancellation
- Circuit breaker for RPC providers

### 5. Reconciliation Worker (Priority: High)
**File**: `internal/workers/funding_webhook/reconciliation.go`

```go
// Key responsibilities:
- Cron job runs every 10 minutes
- Call GetPendingDepositsForReconciliation(15min threshold)
- For each candidate:
  * Query RPC/explorer API for tx status
  * If confirmed: call FundingService.ProcessChainDeposit()
  * If failed: mark deposit as failed
  * If still pending: log and skip
- Track reconciliation metrics:
  * reconciliation_runs_total
  * reconciliation_recovered_deposits_total
  * reconciliation_duration_seconds
```

**RPC Validators**:
- Implement chain-specific validators
- Solana: Use `getTransaction()` + `confirmTransaction()`
- EVM: Use `eth_getTransactionReceipt()`
- Cache results to avoid redundant RPC calls
- Rate limiting per RPC provider

### 6. Enhanced Webhook Handler (Priority: High)
**File**: `internal/api/handlers/funding_investing_handlers.go` (update)

```go
// Updates to existing ChainDepositWebhookHandler:
- Enqueue job immediately (don't process inline)
- Return 200 OK after successful enqueue
- Log audit entry: actor=webhook-receiver, action=enqueue_deposit
- Add request tracing with OpenTelemetry
```

### 7. Queue Interface (Priority: Medium)
**File**: `internal/infrastructure/queue/interface.go`

```go
// Optional: Abstract queue operations for future Redis/SQS migration
type JobQueue interface {
    Enqueue(ctx context.Context, job *entities.FundingEventJob) error
    Dequeue(ctx context.Context, count int) ([]*entities.FundingEventJob, error)
    UpdateStatus(ctx context.Context, jobID uuid.UUID, status entities.FundingEventJobStatus) error
}

// PostgresQueue implements JobQueue using repository
// Future: RedisQueue, SQSQueue implementations
```

### 8. Configuration
**File**: `internal/infrastructure/config/config.go` (add)

```go
type FundingWorkerConfig struct {
    Enabled               bool          `env:"FUNDING_WORKER_ENABLED" default:"true"`
    WorkerCount           int           `env:"FUNDING_WORKER_COUNT" default:"5"`
    PollInterval          time.Duration `env:"FUNDING_WORKER_POLL_INTERVAL" default:"5s"`
    MaxAttempts           int           `env:"FUNDING_WORKER_MAX_ATTEMPTS" default:"5"`
    
    ReconciliationEnabled  bool          `env:"FUNDING_RECONCILIATION_ENABLED" default:"true"`
    ReconciliationInterval time.Duration `env:"FUNDING_RECONCILIATION_INTERVAL" default:"10m"`
    ReconciliationThreshold time.Duration `env:"FUNDING_RECONCILIATION_THRESHOLD" default:"15m"`
    ReconciliationBatchSize int          `env:"FUNDING_RECONCILIATION_BATCH_SIZE" default:"50"`
    
    CircuitBreakerThreshold int          `env:"FUNDING_CIRCUIT_BREAKER_THRESHOLD" default:"5"`
    CircuitBreakerTimeout   time.Duration `env:"FUNDING_CIRCUIT_BREAKER_TIMEOUT" default:"60s"`
}
```

## Testing Strategy

### Unit Tests
1. **Job Entity Tests** (`internal/domain/entities/funding_job_entities_test.go`)
   - CanRetry() logic with different error types
   - ShouldMoveToDLQ() conditions
   - MarkFailed() with retry calculation
   - GetRetryDelay() exponential backoff
   - Validate() input validation

2. **Repository Tests** (`internal/infrastructure/repositories/funding_event_job_repository_test.go`)
   - Enqueue idempotency (duplicate tx_hash + chain)
   - GetNextPendingJobs() concurrency with SKIP LOCKED
   - Update atomic operations
   - GetDLQJobs() filtering
   - GetMetrics() aggregation

3. **Worker Tests** (`internal/workers/funding_webhook/processor_test.go`)
   - Error categorization (5xx vs 4xx vs timeout)
   - Retry backoff timing
   - DLQ routing on max attempts
   - Graceful shutdown
   - Circuit breaker triggering

### Integration Tests
1. **End-to-End Flow** (`test/integration/webhook_reliability_test.go`)
   - Webhook received → job enqueued → processed → completed
   - Transient error → retry → success
   - Permanent error → immediate DLQ
   - Max attempts exhausted → DLQ
   - Reconciliation recovers stuck deposit

2. **RPC Timeout Scenarios**
   - Simulate RPC timeouts
   - Verify retry with backoff
   - Test circuit breaker

3. **Concurrency Tests**
   - Multiple workers processing simultaneously
   - SKIP LOCKED prevents double-processing
   - No race conditions

## Monitoring & Alerts

### Metrics (Prometheus/OpenTelemetry)
```
webhook_received_total{chain,token}
webhook_processed_total{status,chain}
webhook_processing_duration_seconds{quantile}
webhook_retry_count_total{chain}
webhook_dlq_depth{chain}
webhook_reconciliation_runs_total
webhook_reconciliation_recovered_total
webhook_circuit_breaker_trips_total{provider}
```

### Alerts
```yaml
- name: WebhookDLQDepthHigh
  expr: webhook_dlq_depth > 10
  for: 5m
  severity: warning

- name: WebhookSuccessRateLow
  expr: rate(webhook_processed_total{status="completed"}[5m]) / rate(webhook_received_total[5m]) < 0.9
  for: 10m
  severity: critical

- name: WebhookProcessingP99High
  expr: histogram_quantile(0.99, webhook_processing_duration_seconds) > 30
  for: 5m
  severity: warning
```

### Dashboards
- Success rate (last 24h)
- Average retry count
- DLQ depth over time
- Processing latency p50/p95/p99
- Reconciliation recovery stats

## Security Considerations

1. **HMAC Signature Verification** (Already implemented)
   - Verify webhook signatures using `pkg/webhook/security.go`
   - Prevent spoofed webhooks

2. **Rate Limiting**
   - Per-user rate limits on webhook endpoint
   - Per-RPC-provider rate limits

3. **Secrets Management**
   - Store RPC API keys in environment variables
   - Never log sensitive data
   - Use read-only RPC endpoints

4. **Audit Trail**
   - All processing attempts logged to `audit_logs`
   - Actor: `funding-worker`
   - Actions: `enqueue_webhook`, `process_deposit`, `reconcile_deposit`
   - Include before/after state

## Performance Optimization

1. **Database**
   - Indexes optimized for queue operations
   - FOR UPDATE SKIP LOCKED for concurrency
   - Vacuum and analyze regularly
   - Partition funding_event_jobs by month if high volume

2. **Worker Scaling**
   - Start with 5 workers
   - Scale horizontally: run multiple instances
   - Each instance polls independently (SKIP LOCKED prevents conflicts)

3. **RPC Caching**
   - Cache transaction results for 5 minutes
   - Reduces redundant RPC calls during retries

4. **Batch Processing**
   - Process multiple jobs in parallel within each worker
   - Use errgroup for concurrent processing

## Migration Path

### Phase 1: Basic Reliability (Completed)
- ✅ Job entities with retry logic
- ✅ Database schema with DLQ
- ✅ Repository with queue operations

### Phase 2: Worker Implementation (Next)
- Worker processor with error categorization
- Exponential backoff with jitter
- Circuit breaker for RPC providers
- Basic metrics and logging

### Phase 3: Reconciliation (Next)
- Reconciliation worker cron job
- RPC validators (Solana, EVM)
- Recovery of stuck deposits
- Reconciliation metrics

### Phase 4: Observability (Next)
- Comprehensive OpenTelemetry metrics
- Prometheus dashboards
- Alert rules
- Audit logging

### Phase 5: Production Hardening
- Load testing
- Chaos engineering (inject RPC failures)
- Performance optimization
- Documentation

## Files Created

1. ✅ `internal/domain/entities/funding_job_entities.go`
2. ✅ `migrations/006_funding_event_jobs.up.sql`
3. ✅ `migrations/006_funding_event_jobs.down.sql`
4. ✅ `internal/infrastructure/repositories/funding_event_job_repository.go`
5. 🔄 `internal/infrastructure/repositories/simple_wallet_repository.go` (for funding service)

## Next Actions

1. **Immediate**: Implement worker processor
2. **Immediate**: Implement reconciliation worker
3. **Short-term**: Add RPC validators
4. **Short-term**: Enhanced metrics and monitoring
5. **Medium-term**: Write comprehensive tests
6. **Medium-term**: Performance testing and optimization

## References

- User Story: Funding Webhook Reliability Worker
- Clean Architecture principles
- Twelve-Factor App methodology
- OpenTelemetry best practices
- PostgreSQL FOR UPDATE SKIP LOCKED pattern
