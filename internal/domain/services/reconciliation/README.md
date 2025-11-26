# Reconciliation Service

## Overview

The Reconciliation Service provides automated daily verification of balance integrity across the STACK financial system. It ensures that the ledger, external partners (Circle, Alpaca), and database tables remain in sync.

## Architecture

### Components

1. **Domain Entities** (`entities/reconciliation_entities.go`)
   - `ReconciliationReport`: Overall reconciliation run metadata
   - `ReconciliationCheck`: Individual check execution details
   - `ReconciliationException`: Discrepancies detected during checks
   - Severity levels: Low, Medium, High, Critical

2. **Repository Layer** (`repositories/reconciliation_repository.go`)
   - PostgreSQL persistence for reports, checks, and exceptions
   - Batch operations for efficiency
   - Queries for unresolved exceptions

3. **Service Layer** (`service.go`)
   - Orchestrates reconciliation runs
   - Executes all checks
   - Handles exception detection and auto-correction
   - Sends alerts for critical issues

4. **Check Implementations** (`checks.go`)
   - **Ledger Consistency**: Verifies double-entry bookkeeping (debits = credits)
   - **Circle Balance**: USDC buffer matches Circle wallet balances
   - **Alpaca Balance**: User fiat exposure matches Alpaca buying power
   - **Deposits**: Deposit table totals match ledger entries
   - **Conversion Jobs**: All completed conversions have ledger entries
   - **Withdrawals**: Withdrawal table totals match ledger entries

5. **Scheduler** (`scheduler.go`)
   - Automated hourly and daily reconciliation runs
   - Configurable intervals
   - Graceful start/stop

6. **Metrics** (`metrics/reconciliation_metrics.go`)
   - Prometheus metrics for monitoring
   - Run duration, check success rates, exception counts
   - Discrepancy amounts by check type

## Database Schema

### `reconciliation_reports`
- Stores metadata for each reconciliation run
- Tracks total checks, passed/failed counts, exceptions
- Status: pending → in_progress → completed/failed

### `reconciliation_checks`
- Individual check execution records
- Expected vs actual values, differences
- Execution time and pass/fail status

### `reconciliation_exceptions`
- Discrepancies detected during checks
- Severity classification (low/medium/high/critical)
- Auto-correction tracking
- Resolution workflow (unresolved → resolved)

## Usage

### Starting the Scheduler

```go
config := &reconciliation.SchedulerConfig{
    HourlyInterval: 1 * time.Hour,
    DailyInterval:  24 * time.Hour,
}

scheduler := reconciliation.NewScheduler(service, logger, config)
err := scheduler.Start(ctx)
defer scheduler.Stop()
```

### Manual Reconciliation Run

```go
report, err := reconciliationService.RunReconciliation(ctx, "manual")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Reconciliation completed: %d checks, %d exceptions\n", 
    report.TotalChecks, report.ExceptionsCount)
```

### Querying Unresolved Exceptions

```go
exceptions, err := reconciliationService.GetUnresolvedExceptions(
    ctx, 
    entities.ExceptionSeverityHigh,
)

for _, exc := range exceptions {
    fmt.Printf("Exception: %s - $%.2f\n", exc.Description, exc.Difference)
}
```

## Configuration

### Service Config (`Config`)

```go
config := &reconciliation.Config{
    AutoCorrectLowSeverity: true,                        // Auto-correct <$1 discrepancies
    ToleranceCircle:        decimal.NewFromFloat(10.0),  // $10 tolerance for Circle
    ToleranceAlpaca:        decimal.NewFromFloat(100.0), // $100 tolerance for Alpaca
    EnableAlerting:         true,                        // Send alerts for high/critical
    AlertWebhookURL:        "https://alerts.example.com/webhook",
}
```

### Severity Thresholds

- **Low**: ≤ $1 (auto-correctable)
- **Medium**: $1 - $100
- **High**: $100 - $1,000
- **Critical**: > $1,000

## Checks Detail

### 1. Ledger Consistency Check

**Purpose**: Verify double-entry bookkeeping integrity

**Validations**:
- Sum of all debits equals sum of all credits
- No orphaned ledger entries (entries without transactions)
- All transactions have exactly 2 entries (1 debit + 1 credit)

**Failure Impact**: Critical - indicates data corruption or logic bug

---

### 2. Circle Balance Check

**Purpose**: Ensure on-chain USDC buffer matches Circle wallets

**Validation**:
```
ledger.system_buffer_usdc == Circle.getTotalUSDCBalance()
```

**Tolerance**: $10 (for pending transactions)

**Failure Impact**: High - affects withdrawal liquidity

---

### 3. Alpaca Balance Check

**Purpose**: Verify total user fiat exposure matches Alpaca buying power

**Validation**:
```
SUM(ledger.fiat_exposure) == Alpaca.getTotalBuyingPower()
```

**Tolerance**: $100 (for pending orders)

**Failure Impact**: High - affects trading integrity

---

### 4. Deposit Check

**Purpose**: Ensure deposit table totals match ledger deposit entries

**Validation**:
```
SUM(deposits.amount WHERE status='completed') == 
SUM(ledger_entries WHERE transaction_type='deposit')
```

**Failure Impact**: Medium - indicates deposit processing issue

---

### 5. Conversion Jobs Check

**Purpose**: Verify all completed conversion jobs have ledger entries

**Validation**: No completed `conversion_jobs` without corresponding ledger entries

**Failure Impact**: High - indicates conversion not recorded

---

### 6. Withdrawal Check

**Purpose**: Ensure withdrawal table totals match ledger withdrawal entries

**Validation**:
```
SUM(withdrawals.amount WHERE status='completed') == 
SUM(ledger_entries WHERE transaction_type='withdrawal')
```

**Failure Impact**: Medium - indicates withdrawal processing issue

## Auto-Correction

Low severity exceptions (≤ $1) can be automatically corrected:

1. Exception detected with severity = Low
2. System logs the discrepancy
3. Exception marked as `auto_corrected = true`
4. Resolution action recorded
5. Metric incremented

**Note**: Actual correction logic is intentionally conservative. Current implementation logs discrepancies for manual review.

## Alerting

High and critical exceptions trigger alerts:

1. **Webhook Notification**: POST to configured webhook URL
2. **Metrics Alert**: Prometheus alert rules fire
3. **Log Alert**: Structured error logs for SIEM

### Alert Payload Example

```json
{
  "severity": "critical",
  "check_type": "circle_balance",
  "description": "Circle wallet balance does not match ledger",
  "expected": "10000.00",
  "actual": "8500.00",
  "difference": "-1500.00",
  "currency": "USDC",
  "report_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Metrics

### Prometheus Metrics

- `reconciliation_runs_total{run_type, status}`
- `reconciliation_run_duration_seconds{run_type}`
- `reconciliation_checks_total{check_type}`
- `reconciliation_checks_passed_total{check_type}`
- `reconciliation_checks_failed_total{check_type}`
- `reconciliation_exceptions_total{check_type, severity}`
- `reconciliation_exceptions_unresolved{severity}`
- `reconciliation_discrepancy_amount{check_type, currency}`
- `reconciliation_alerts_total{check_type, severity}`

### Grafana Dashboard Queries

**Success Rate**:
```promql
sum(rate(reconciliation_checks_passed_total[5m])) /
sum(rate(reconciliation_checks_total[5m]))
```

**Unresolved Critical Exceptions**:
```promql
reconciliation_exceptions_unresolved{severity="critical"}
```

**Average Discrepancy by Check**:
```promql
avg(reconciliation_discrepancy_amount) by (check_type)
```

## Monitoring & Alerting Rules

### Recommended Alerts

```yaml
- alert: ReconciliationCriticalException
  expr: reconciliation_exceptions_unresolved{severity="critical"} > 0
  for: 5m
  annotations:
    summary: Critical reconciliation exception detected

- alert: ReconciliationCheckFailureRate
  expr: |
    sum(rate(reconciliation_checks_failed_total[1h])) /
    sum(rate(reconciliation_checks_total[1h])) > 0.1
  for: 15m
  annotations:
    summary: High reconciliation check failure rate (>10%)

- alert: ReconciliationRunFailed
  expr: increase(reconciliation_runs_total{status="failed"}[10m]) > 0
  annotations:
    summary: Reconciliation run failed
```

## Operational Procedures

### Daily Operations

1. **Review Daily Report**: Check Grafana dashboard each morning
2. **Investigate Exceptions**: Review any high/critical exceptions
3. **Resolve Manually**: Use admin tools to mark exceptions as resolved
4. **Adjust Thresholds**: Tune tolerance values if needed

### Exception Resolution Workflow

```go
// Query unresolved exceptions
exceptions, _ := service.GetUnresolvedExceptions(ctx, 
    entities.ExceptionSeverityHigh)

// Investigate and resolve
for _, exc := range exceptions {
    // Manual investigation...
    
    err := service.ResolveException(ctx, 
        exc.ID, 
        "admin@example.com",
        "Investigated - due to pending conversion",
    )
}
```

### Troubleshooting

**High Failure Rate**:
- Check external API availability (Circle, Alpaca)
- Review recent code deployments
- Verify database connectivity

**Persistent Discrepancies**:
- Run manual SQL queries to identify root cause
- Check for stuck transactions or webhooks
- Review conversion job statuses

**Performance Issues**:
- Check database query performance
- Review OpenTelemetry traces
- Consider read replicas for reconciliation queries

## Testing

### Unit Tests

```bash
go test ./internal/domain/services/reconciliation/... -v
```

### Integration Tests

```bash
go test ./internal/domain/services/reconciliation/... -tags=integration -v
```

### Test Coverage

```bash
go test ./internal/domain/services/reconciliation/... -cover
```

Target: >80% coverage for core reconciliation logic

## Future Enhancements

1. **Predictive Alerting**: ML-based anomaly detection
2. **Automated Remediation**: Auto-correct more exception types
3. **Multi-Provider Support**: Reconcile across multiple conversion providers
4. **Real-time Reconciliation**: Continuous reconciliation instead of batched
5. **User-Level Reconciliation**: Per-user balance verification
6. **Historical Trending**: Track discrepancy patterns over time
7. **Integration Testing**: Automated E2E tests with testnet APIs

## References

- [Double-Entry Bookkeeping](https://en.wikipedia.org/wiki/Double-entry_bookkeeping)
- [Financial Reconciliation Best Practices](https://stripe.com/docs/account/reconciliation)
- [OpenTelemetry Tracing](https://opentelemetry.io/docs/concepts/observability-primer/)
- Circle Developer Wallets Documentation
- Alpaca Journal API Documentation
