package circle

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stack-service/stack_service/internal/domain/entities"
	// "github.com/stack-service/stack_service/pkg/logger"
	"go.uber.org/zap"
)

const (
	// Circle API URLs
	ProductionBaseURL = "https://api.circle.com"
	SandboxBaseURL    = "https://api-sandbox.circle.com"

	// API Endpoints
	walletSetsEndpoint = "/v1/w3s/developer/walletSets"
	walletsEndpoint    = "/v1/w3s/developer/wallets"

	// Timeouts and limits
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
	baseBackoff    = 1 * time.Second
	maxBackoff     = 16 * time.Second
)

// Config represents Circle API configuration
type Config struct {
	APIKey      string        `json:"api_key"`
	BaseURL     string        `json:"base_url"`
	Environment string        `json:"environment"` // "sandbox" or "production"
	Timeout     time.Duration `json:"timeout"`
}

// CircuitBreakerState represents circuit breaker states
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// CircuitBreaker implements a simple circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	timeout         time.Duration
	maxFailures     int
	logger          *zap.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration, logger *zap.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		logger:      logger,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(ctx context.Context, fn func() error) error {
	cb.mu.RLock()
	state := cb.state
	cb.mu.RUnlock()

	switch state {
	case CircuitBreakerOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.setState(CircuitBreakerHalfOpen)
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	err := fn()
	cb.recordResult(err)
	return err
}

func (cb *CircuitBreaker) setState(state CircuitBreakerState) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = state
	cb.logger.Info("Circuit breaker state changed", zap.String("state", cb.stateString(state)))
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.state == CircuitBreakerHalfOpen {
			cb.state = CircuitBreakerOpen
		} else if cb.failureCount >= cb.maxFailures {
			cb.state = CircuitBreakerOpen
		}
	} else {
		cb.failureCount = 0
		cb.successCount++

		if cb.state == CircuitBreakerHalfOpen {
			cb.state = CircuitBreakerClosed
		}
	}
}

func (cb *CircuitBreaker) stateString(state CircuitBreakerState) string {
	switch state {
	case CircuitBreakerClosed:
		return "closed"
	case CircuitBreakerOpen:
		return "open"
	case CircuitBreakerHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Client represents a Circle API client
type Client struct {
	config         Config
	httpClient     *http.Client
	circuitBreaker *CircuitBreaker
	logger         *zap.Logger
}

// NewClient creates a new Circle API client
func NewClient(config Config, logger *zap.Logger) *Client {
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	if config.BaseURL == "" {
		if config.Environment == "production" {
			config.BaseURL = ProductionBaseURL
		} else {
			config.BaseURL = SandboxBaseURL
		}
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	circuitBreaker := NewCircuitBreaker(5, 30*time.Second, logger)

	return &Client{
		config:         config,
		httpClient:     httpClient,
		circuitBreaker: circuitBreaker,
		logger:         logger,
	}
}

// CreateWalletSet creates a new wallet set
func (c *Client) CreateWalletSet(ctx context.Context, name string) (*entities.CircleWalletSetResponse, error) {
	request := entities.CircleWalletSetRequest{
		Name: name,
	}

	var response entities.CircleWalletSetResponse
	err := c.circuitBreaker.Call(ctx, func() error {
		return c.doRequestWithRetry(ctx, "POST", walletSetsEndpoint, request, &response)
	})

	if err != nil {
		c.logger.Error("Failed to create wallet set",
			zap.String("name", name),
			zap.Error(err))
		return nil, fmt.Errorf("create wallet set failed: %w", err)
	}

	c.logger.Info("Created wallet set successfully",
		zap.String("name", name),
		zap.String("walletSetId", response.WalletSet.ID))

	return &response, nil
}

// GetWalletSet retrieves a wallet set by ID
func (c *Client) GetWalletSet(ctx context.Context, walletSetID string) (*entities.CircleWalletSetResponse, error) {
	endpoint := fmt.Sprintf("%s/%s", walletSetsEndpoint, walletSetID)

	var response entities.CircleWalletSetResponse
	err := c.circuitBreaker.Call(ctx, func() error {
		return c.doRequestWithRetry(ctx, "GET", endpoint, nil, &response)
	})

	if err != nil {
		c.logger.Error("Failed to get wallet set",
			zap.String("walletSetId", walletSetID),
			zap.Error(err))
		return nil, fmt.Errorf("get wallet set failed: %w", err)
	}

	return &response, nil
}

// CreateWallet creates a new wallet
func (c *Client) CreateWallet(ctx context.Context, req entities.CircleWalletCreateRequest) (*entities.CircleWalletCreateResponse, error) {
	var response entities.CircleWalletCreateResponse
	err := c.circuitBreaker.Call(ctx, func() error {
		return c.doRequestWithRetry(ctx, "POST", walletsEndpoint, req, &response)
	})

	if err != nil {
		c.logger.Error("Failed to create wallet",
			zap.String("walletSetId", req.WalletSetID),
			zap.Strings("blockchains", req.Blockchains),
			zap.String("accountType", req.AccountType),
			zap.Error(err))
		return nil, fmt.Errorf("create wallet failed: %w", err)
	}

	c.logger.Info("Created wallet successfully",
		zap.String("walletSetId", req.WalletSetID),
		zap.String("walletId", response.Wallet.ID),
		zap.Strings("blockchains", req.Blockchains))

	return &response, nil
}

// GetWallet retrieves a wallet by ID
func (c *Client) GetWallet(ctx context.Context, walletID string) (*entities.CircleWalletCreateResponse, error) {
	endpoint := fmt.Sprintf("%s/%s", walletsEndpoint, walletID)

	var response entities.CircleWalletCreateResponse
	err := c.circuitBreaker.Call(ctx, func() error {
		return c.doRequestWithRetry(ctx, "GET", endpoint, nil, &response)
	})

	if err != nil {
		c.logger.Error("Failed to get wallet",
			zap.String("walletId", walletID),
			zap.Error(err))
		return nil, fmt.Errorf("get wallet failed: %w", err)
	}

	return &response, nil
}

// doRequestWithRetry performs HTTP request with exponential backoff retry
func (c *Client) doRequestWithRetry(ctx context.Context, method, endpoint string, requestBody, responseBody interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			backoff := time.Duration(1<<uint(attempt-1)) * baseBackoff
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			c.logger.Info("Retrying Circle API request",
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
				zap.String("method", method),
				zap.String("endpoint", endpoint))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := c.doRequest(ctx, method, endpoint, requestBody, responseBody)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on certain types of errors
		if !c.shouldRetry(err) {
			c.logger.Warn("Not retrying Circle API request due to error type",
				zap.Error(err),
				zap.String("method", method),
				zap.String("endpoint", endpoint))
			break
		}

		c.logger.Warn("Circle API request failed, will retry",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Int("maxRetries", maxRetries),
			zap.String("method", method),
			zap.String("endpoint", endpoint))
	}

	return fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, lastErr)
}

// doRequest performs a single HTTP request
func (c *Client) doRequest(ctx context.Context, method, endpoint string, requestBody, responseBody interface{}) error {
	url := c.config.BaseURL + endpoint

	var reqBody io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Stack-Service/1.0")

	// Add request ID for tracing
	if requestID := ctx.Value("request_id"); requestID != nil {
		req.Header.Set("X-Request-ID", requestID.(string))
	}

	c.logger.Debug("Making Circle API request",
		zap.String("method", method),
		zap.String("url", url),
		zap.Any("headers", req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debug("Received Circle API response",
		zap.String("method", method),
		zap.String("url", url),
		zap.Int("statusCode", resp.StatusCode),
		zap.String("body", string(body)))

	// Handle error responses
	if resp.StatusCode >= 400 {
		var circleErr entities.CircleErrorResponse
		if err := json.Unmarshal(body, &circleErr); err != nil {
			return fmt.Errorf("circle API error %d: %s", resp.StatusCode, string(body))
		}
		circleErr.Code = resp.StatusCode
		return circleErr
	}

	// Unmarshal successful response
	if responseBody != nil && len(body) > 0 {
		if err := json.Unmarshal(body, responseBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// shouldRetry determines if a request should be retried based on the error
func (c *Client) shouldRetry(err error) bool {
	// Don't retry on context cancellation
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Check if it's a Circle API error
	if circleErr, ok := err.(entities.CircleErrorResponse); ok {
		// Don't retry on client errors (4xx), except for rate limiting and timeouts
		if circleErr.Code >= 400 && circleErr.Code < 500 {
			return circleErr.Code == 429 || circleErr.Code == 408
		}
		// Retry on server errors (5xx)
		return circleErr.Code >= 500
	}

	// Retry on network errors
	return true
}

// HealthCheck performs a health check against Circle API
func (c *Client) HealthCheck(ctx context.Context) error {
	// Use a simple GET request to wallet sets to check connectivity
	endpoint := walletSetsEndpoint

	req, err := http.NewRequestWithContext(ctx, "GET", c.config.BaseURL+endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("circle API health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("circle API health check failed with status %d", resp.StatusCode)
	}

	c.logger.Info("Circle API health check successful", zap.Int("statusCode", resp.StatusCode))
	return nil
}

// GenerateDepositAddress generates a deposit address for the specified chain and user
func (c *Client) GenerateDepositAddress(ctx context.Context, chain entities.Chain, userID uuid.UUID) (string, error) {
	// For MVP, we'll simulate address generation based on chain type
	// In production, this would call Circle's actual deposit address generation API

	c.logger.Info("Generating deposit address",
		zap.String("chain", string(chain)),
		zap.String("user_id", userID.String()))

	// Simulate address generation with deterministic but unique addresses
	var address string
	switch chain {
	case entities.ChainSolana:
		// Generate Solana-style address (Base58)
		address = fmt.Sprintf("So1%s%d", userID.String()[:8], time.Now().Unix()%10000)
	case entities.ChainPolygon:
		// Generate EVM-style address
		address = fmt.Sprintf("0x%s%x", userID.String()[:8], time.Now().Unix()%10000)
	case entities.ChainAptos:
		// Generate Aptos-style address
		address = fmt.Sprintf("0x%s%x", userID.String()[:8], time.Now().Unix()%10000)
	case entities.ChainStarknet:
		// Generate StarkNet-style address
		address = fmt.Sprintf("0x%s%x", userID.String()[:8], time.Now().Unix()%10000)
	default:
		return "", fmt.Errorf("unsupported chain: %s", chain)
	}

	c.logger.Info("Generated deposit address",
		zap.String("chain", string(chain)),
		zap.String("user_id", userID.String()),
		zap.String("address", address))

	return address, nil
}

// ValidateDeposit validates a deposit transaction using Circle's validation service
func (c *Client) ValidateDeposit(ctx context.Context, txHash string, amount decimal.Decimal) (bool, error) {
	c.logger.Info("Validating deposit",
		zap.String("tx_hash", txHash),
		zap.String("amount", amount.String()))

	// For MVP, we'll simulate validation
	// In production, this would call Circle's transaction validation API

	// Simple validation: check if amount is positive and txHash is not empty
	if amount.IsZero() || amount.IsNegative() {
		c.logger.Warn("Invalid deposit amount",
			zap.String("tx_hash", txHash),
			zap.String("amount", amount.String()))
		return false, nil
	}

	if txHash == "" {
		c.logger.Warn("Empty transaction hash", zap.String("tx_hash", txHash))
		return false, nil
	}

	// For demo purposes, reject transactions with "invalid" in the hash
	if len(txHash) > 7 && txHash[:7] == "invalid" {
		c.logger.Warn("Invalid transaction detected", zap.String("tx_hash", txHash))
		return false, nil
	}

	c.logger.Info("Deposit validation successful",
		zap.String("tx_hash", txHash),
		zap.String("amount", amount.String()))

	return true, nil
}

// ConvertToUSD converts stablecoin amount to USD buying power
func (c *Client) ConvertToUSD(ctx context.Context, amount decimal.Decimal, token entities.Stablecoin) (decimal.Decimal, error) {
	c.logger.Info("Converting to USD",
		zap.String("amount", amount.String()),
		zap.String("token", string(token)))

	// For MVP, we'll use fixed conversion rates
	// In production, this would call Circle's price oracle or conversion API

	var conversionRate decimal.Decimal
	switch token {
	case entities.StablecoinUSDC:
		// USDC is pegged 1:1 to USD
		conversionRate = decimal.NewFromInt(1)
	default:
		return decimal.Zero, fmt.Errorf("unsupported token: %s", token)
	}

	usdAmount := amount.Mul(conversionRate)

	c.logger.Info("Conversion to USD completed",
		zap.String("original_amount", amount.String()),
		zap.String("token", string(token)),
		zap.String("usd_amount", usdAmount.String()),
		zap.String("conversion_rate", conversionRate.String()))

	return usdAmount, nil
}

// GetMetrics returns circuit breaker metrics for monitoring
func (c *Client) GetMetrics() map[string]interface{} {
	c.circuitBreaker.mu.RLock()
	defer c.circuitBreaker.mu.RUnlock()

	return map[string]interface{}{
		"circuit_breaker_state":     c.circuitBreaker.stateString(c.circuitBreaker.state),
		"circuit_breaker_failures":  c.circuitBreaker.failureCount,
		"circuit_breaker_successes": c.circuitBreaker.successCount,
		"last_failure_time":         c.circuitBreaker.lastFailureTime,
	}
}
