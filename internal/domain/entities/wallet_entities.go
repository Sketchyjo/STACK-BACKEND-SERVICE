package entities

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// WalletChain represents supported blockchain networks for Circle integration
type WalletChain string

const (
	// EVM chains
	ChainETH         WalletChain = "ETH"
	ChainETHSepolia  WalletChain = "ETH-SEPOLIA"
	ChainMATIC       WalletChain = "MATIC"
	ChainMATICAmoy   WalletChain = "MATIC-AMOY"
	ChainAVAX        WalletChain = "AVAX"
	ChainBASE        WalletChain = "BASE"
	ChainBASESepolia WalletChain = "BASE-SEPOLIA"

	// Solana
	ChainSOL       WalletChain = "SOL"
	ChainSOLDevnet WalletChain = "SOL-DEVNET"

	// Aptos
	ChainAPTOS        WalletChain = "APTOS"
	ChainAPTOSTestnet WalletChain = "APTOS-TESTNET"
)

// GetMainnetChains returns production chains
func GetMainnetChains() []WalletChain {
	return []WalletChain{ChainETH, ChainMATIC, ChainAVAX, ChainSOL, ChainAPTOS, ChainBASE}
}

// GetTestnetChains returns testnet chains
func GetTestnetChains() []WalletChain {
	return []WalletChain{ChainETHSepolia, ChainSOLDevnet, ChainAPTOSTestnet, ChainMATICAmoy, ChainBASESepolia}
}

// IsValid checks if the chain is supported
func (c WalletChain) IsValid() bool {
	validChains := append(GetMainnetChains(), GetTestnetChains()...)
	for _, chain := range validChains {
		if chain == c {
			return true
		}
	}
	return false
}

// IsTestnet checks if the chain is a testnet
func (c WalletChain) IsTestnet() bool {
	testnets := GetTestnetChains()
	for _, testnet := range testnets {
		if testnet == c {
			return true
		}
	}
	return false
}

// GetChainFamily returns the chain family (EVM, Solana, Aptos)
func (c WalletChain) GetChainFamily() string {
	switch c {
	case ChainETH, ChainETHSepolia, ChainMATIC, ChainMATICAmoy, ChainAVAX, ChainBASE, ChainBASESepolia:
		return "EVM"
	case ChainSOL, ChainSOLDevnet:
		return "Solana"
	case ChainAPTOS, ChainAPTOSTestnet:
		return "Aptos"
	default:
		return "Unknown"
	}
}

// WalletAccountType represents the type of wallet account
type WalletAccountType string

const (
	AccountTypeEOA WalletAccountType = "EOA" // Externally Owned Account
	AccountTypeSCA WalletAccountType = "SCA" // Smart Contract Account
)

// IsValid checks if account type is valid
func (t WalletAccountType) IsValid() bool {
	return t == AccountTypeEOA || t == AccountTypeSCA
}

// WalletStatus represents the status of a wallet
type WalletStatus string

const (
	WalletStatusCreating WalletStatus = "creating"
	WalletStatusLive     WalletStatus = "live"
	WalletStatusFailed   WalletStatus = "failed"
)

// IsValid checks if wallet status is valid
func (s WalletStatus) IsValid() bool {
	return s == WalletStatusCreating || s == WalletStatusLive || s == WalletStatusFailed
}

// WalletSetStatus represents the status of a wallet set
type WalletSetStatus string

const (
	WalletSetStatusActive   WalletSetStatus = "active"
	WalletSetStatusInactive WalletSetStatus = "inactive"
)

// WalletSet represents a Circle wallet set
type WalletSet struct {
	ID                     uuid.UUID       `json:"id" db:"id"`
	Name                   string          `json:"name" db:"name" validate:"required"`
	CircleWalletSetID      string          `json:"circle_wallet_set_id" db:"circle_wallet_set_id"`
	EntitySecretCiphertext string          `json:"-" db:"entity_secret_ciphertext"` // Never expose in JSON
	Status                 WalletSetStatus `json:"status" db:"status"`
	CreatedAt              time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at" db:"updated_at"`
}

// Validate performs validation on wallet set
func (ws *WalletSet) Validate() error {
	if ws.Name == "" {
		return fmt.Errorf("wallet set name is required")
	}

	if len(ws.Name) > 100 {
		return fmt.Errorf("wallet set name cannot exceed 100 characters")
	}

	if ws.CircleWalletSetID == "" {
		return fmt.Errorf("circle wallet set ID is required")
	}

	if ws.EntitySecretCiphertext == "" {
		return fmt.Errorf("entity secret ciphertext is required")
	}

	if ws.Status != WalletSetStatusActive && ws.Status != WalletSetStatusInactive {
		return fmt.Errorf("invalid wallet set status: %s", ws.Status)
	}

	return nil
}

// ManagedWallet represents a Circle-managed wallet
type ManagedWallet struct {
	ID             uuid.UUID         `json:"id" db:"id"`
	UserID         uuid.UUID         `json:"user_id" db:"user_id"`
	Chain          WalletChain       `json:"chain" db:"chain"`
	Address        string            `json:"address" db:"address"`
	CircleWalletID string            `json:"circle_wallet_id" db:"circle_wallet_id"`
	WalletSetID    uuid.UUID         `json:"wallet_set_id" db:"wallet_set_id"`
	AccountType    WalletAccountType `json:"account_type" db:"account_type"`
	Status         WalletStatus      `json:"status" db:"status"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" db:"updated_at"`

	// Related entities (not stored in DB)
	WalletSet *WalletSet `json:"wallet_set,omitempty"`
}

// Validate performs validation on managed wallet
func (w *ManagedWallet) Validate() error {
	if w.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if !w.Chain.IsValid() {
		return fmt.Errorf("invalid chain: %s", w.Chain)
	}

	if w.Address == "" {
		return fmt.Errorf("wallet address is required")
	}

	if w.CircleWalletID == "" {
		return fmt.Errorf("circle wallet ID is required")
	}

	if w.WalletSetID == uuid.Nil {
		return fmt.Errorf("wallet set ID is required")
	}

	if !w.AccountType.IsValid() {
		return fmt.Errorf("invalid account type: %s", w.AccountType)
	}

	if !w.Status.IsValid() {
		return fmt.Errorf("invalid wallet status: %s", w.Status)
	}

	return nil
}

// IsReady checks if wallet is ready for use
func (w *ManagedWallet) IsReady() bool {
	return w.Status == WalletStatusLive && w.Address != ""
}

// CanReceive checks if wallet can receive funds
func (w *ManagedWallet) CanReceive() bool {
	return w.IsReady()
}

// GetDisplayAddress returns a user-friendly display of the address
func (w *ManagedWallet) GetDisplayAddress() string {
	if len(w.Address) <= 8 {
		return w.Address
	}
	return fmt.Sprintf("%s...%s", w.Address[:6], w.Address[len(w.Address)-4:])
}

// WalletProvisioningJobStatus represents the status of wallet provisioning
type WalletProvisioningJobStatus string

const (
	ProvisioningStatusQueued     WalletProvisioningJobStatus = "queued"
	ProvisioningStatusInProgress WalletProvisioningJobStatus = "in_progress"
	ProvisioningStatusCompleted  WalletProvisioningJobStatus = "completed"
	ProvisioningStatusFailed     WalletProvisioningJobStatus = "failed"
	ProvisioningStatusRetry      WalletProvisioningJobStatus = "retry"
)

// WalletProvisioningJob represents an async wallet provisioning job
type WalletProvisioningJob struct {
	ID             uuid.UUID                   `json:"id" db:"id"`
	UserID         uuid.UUID                   `json:"user_id" db:"user_id"`
	Chains         []string                    `json:"chains" db:"chains"`
	Status         WalletProvisioningJobStatus `json:"status" db:"status"`
	AttemptCount   int                         `json:"attempt_count" db:"attempt_count"`
	MaxAttempts    int                         `json:"max_attempts" db:"max_attempts"`
	CircleRequests map[string]any              `json:"circle_requests" db:"circle_requests"`
	ErrorMessage   *string                     `json:"error_message" db:"error_message"`
	NextRetryAt    *time.Time                  `json:"next_retry_at" db:"next_retry_at"`
	StartedAt      *time.Time                  `json:"started_at" db:"started_at"`
	CompletedAt    *time.Time                  `json:"completed_at" db:"completed_at"`
	CreatedAt      time.Time                   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at" db:"updated_at"`
}

// CanRetry checks if the job can be retried
func (job *WalletProvisioningJob) CanRetry() bool {
	return job.Status == ProvisioningStatusFailed &&
		job.AttemptCount < job.MaxAttempts
}

// MarkStarted marks the job as started
func (job *WalletProvisioningJob) MarkStarted() {
	now := time.Now()
	job.Status = ProvisioningStatusInProgress
	job.StartedAt = &now
	job.AttemptCount++
	job.UpdatedAt = now
}

// MarkCompleted marks the job as completed
func (job *WalletProvisioningJob) MarkCompleted() {
	now := time.Now()
	job.Status = ProvisioningStatusCompleted
	job.CompletedAt = &now
	job.UpdatedAt = now
}

// MarkFailed marks the job as failed
func (job *WalletProvisioningJob) MarkFailed(errorMsg string, retryDelay time.Duration) {
	now := time.Now()
	job.Status = ProvisioningStatusFailed
	job.ErrorMessage = &errorMsg
	job.UpdatedAt = now

	if job.CanRetry() {
		nextRetry := now.Add(retryDelay)
		job.NextRetryAt = &nextRetry
		job.Status = ProvisioningStatusRetry
	}
}

// AddCircleRequest adds a Circle API request/response to the log
func (job *WalletProvisioningJob) AddCircleRequest(operation string, request, response any) {
	if job.CircleRequests == nil {
		job.CircleRequests = make(map[string]any)
	}

	if requests, ok := job.CircleRequests["requests"].([]map[string]any); ok {
		job.CircleRequests["requests"] = append(requests, map[string]any{
			"timestamp": time.Now(),
			"operation": operation,
			"request":   request,
			"response":  response,
		})
	} else {
		job.CircleRequests["requests"] = []map[string]any{{
			"timestamp": time.Now(),
			"operation": operation,
			"request":   request,
			"response":  response,
		}}
	}
}

// === API Request/Response Models ===

// WalletAddressesRequest represents request for wallet addresses
type WalletAddressesRequest struct {
	Chain *WalletChain `json:"chain,omitempty" validate:"omitempty"`
}

// WalletAddressResponse represents a single wallet address
type WalletAddressResponse struct {
	Chain   WalletChain `json:"chain"`
	Address string      `json:"address"`
	Status  string      `json:"status"`
}

// WalletAddressesResponse represents response with wallet addresses
type WalletAddressesResponse struct {
	Wallets []WalletAddressResponse `json:"wallets"`
}

// WalletStatusResponse represents wallet status for all chains
type WalletStatusResponse struct {
	UserID          uuid.UUID                      `json:"userId"`
	TotalWallets    int                            `json:"totalWallets"`
	ReadyWallets    int                            `json:"readyWallets"`
	PendingWallets  int                            `json:"pendingWallets"`
	FailedWallets   int                            `json:"failedWallets"`
	WalletsByChain  map[string]WalletChainStatus   `json:"walletsByChain"`
	ProvisioningJob *WalletProvisioningJobResponse `json:"provisioningJob,omitempty"`
}

// WalletChainStatus represents status for a specific chain
type WalletChainStatus struct {
	Chain     WalletChain `json:"chain"`
	Address   *string     `json:"address,omitempty"`
	Status    string      `json:"status"`
	CreatedAt *time.Time  `json:"createdAt,omitempty"`
	Error     *string     `json:"error,omitempty"`
}

// WalletProvisioningJobResponse represents provisioning job status
type WalletProvisioningJobResponse struct {
	ID           uuid.UUID  `json:"id"`
	Status       string     `json:"status"`
	Progress     string     `json:"progress"`
	AttemptCount int        `json:"attemptCount"`
	MaxAttempts  int        `json:"maxAttempts"`
	ErrorMessage *string    `json:"errorMessage,omitempty"`
	NextRetryAt  *time.Time `json:"nextRetryAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// === Circle API Models ===

// CircleWalletSetRequest represents Circle wallet set creation request
type CircleWalletSetRequest struct {
	IdempotencyKey         string `json:"idempotencyKey,omitempty"`
	Name                   string `json:"name"`
	EntitySecretCiphertext string `json:"entitySecretCiphertext"`
}

// CircleWalletSetResponse represents Circle wallet set response
type CircleWalletSetResponse struct {
	WalletSet CircleWalletSetData `json:"walletSet"`
}

// UnmarshalJSON normalizes Circle wallet set responses that may wrap data
func (r *CircleWalletSetResponse) UnmarshalJSON(data []byte) error {
	type alias CircleWalletSetResponse
	aux := struct {
		Data      *alias               `json:"data"`
		WalletSet *CircleWalletSetData `json:"walletSet"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch {
	case aux.Data != nil && aux.Data.WalletSet.ID != "":
		r.WalletSet = aux.Data.WalletSet
	case aux.WalletSet != nil && aux.WalletSet.ID != "":
		r.WalletSet = *aux.WalletSet
	default:
		r.WalletSet = CircleWalletSetData{}
	}

	return nil
}

// CircleWalletSetData represents Circle wallet set data
type CircleWalletSetData struct {
	ID          string    `json:"id"`
	CustodyType string    `json:"custodyType"`
	Name        string    `json:"name"`
	CreatedDate time.Time `json:"createDate"`
	UpdatedDate time.Time `json:"updateDate"`
}

// CircleWalletCreateRequest represents Circle wallet creation request
type CircleWalletCreateRequest struct {
	IdempotencyKey         string   `json:"idempotencyKey,omitempty"`
	WalletSetID            string   `json:"walletSetId"`
	Blockchains            []string `json:"blockchains"`
	AccountType            string   `json:"accountType"`
	EntitySecretCiphertext string   `json:"entitySecretCiphertext"`
}

// CircleWalletCreateResponse represents Circle wallet creation response
type CircleWalletCreateResponse struct {
	Wallet CircleWalletData `json:"wallet"`
}

// UnmarshalJSON normalizes Circle wallet responses that may wrap data
func (r *CircleWalletCreateResponse) UnmarshalJSON(data []byte) error {
	type alias CircleWalletCreateResponse
	aux := struct {
		Data   *alias           `json:"data"`
		Wallet CircleWalletData `json:"wallet"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch {
	case aux.Data != nil && aux.Data.Wallet.ID != "":
		r.Wallet = aux.Data.Wallet
	case aux.Wallet.ID != "":
		r.Wallet = aux.Wallet
	default:
		r.Wallet = CircleWalletData{}
	}

	return nil
}

// CircleWalletData represents Circle wallet data
type CircleWalletData struct {
	ID          string                `json:"id"`
	State       string                `json:"state"`
	WalletSetId string                `json:"walletSetId"`
	CustodyType string                `json:"custodyType"`
	Addresses   []CircleWalletAddress `json:"addresses"`
	CreatedDate time.Time             `json:"createDate"`
	UpdatedDate time.Time             `json:"updateDate"`
}

// CircleWalletAddress represents a wallet address for a specific blockchain
type CircleWalletAddress struct {
	Address    string `json:"address"`
	Blockchain string `json:"blockchain"`
	Chain      string `json:"chain,omitempty"`
}

// CircleErrorResponse represents Circle API error response
type CircleErrorResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Errors  []CircleFieldError `json:"errors,omitempty"`
}

// CircleFieldError represents field-specific error
type CircleFieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements error interface
func (e CircleErrorResponse) Error() string {
	if len(e.Errors) > 0 {
		var details []string
		for _, fieldErr := range e.Errors {
			details = append(details, fmt.Sprintf("%s: %s", fieldErr.Field, fieldErr.Message))
		}
		return fmt.Sprintf("Circle API error %d: %s (%s)", e.Code, e.Message, strings.Join(details, ", "))
	}
	return fmt.Sprintf("Circle API error %d: %s", e.Code, e.Message)
}
