package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Chain represents supported blockchain networks as per OpenAPI spec
type Chain string

const (
	ChainAptos    Chain = "Aptos"
	ChainSolana   Chain = "Solana"
	ChainPolygon  Chain = "polygon"
	ChainStarknet Chain = "starknet"
)

// Stablecoin represents supported stablecoins
type Stablecoin string

const (
	StablecoinUSDC Stablecoin = "USDC"
)

// OrderSide represents buy/sell direction
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderStatus represents order states
type OrderStatus string

const (
	OrderStatusAccepted        OrderStatus = "accepted"
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusFailed          OrderStatus = "failed"
	OrderStatusCanceled        OrderStatus = "canceled"
)

// RiskLevel represents basket risk levels
type RiskLevel string

const (
	RiskLevelConservative RiskLevel = "conservative"
	RiskLevelBalanced     RiskLevel = "balanced"
	RiskLevelGrowth       RiskLevel = "growth"
)

// === Core Domain Entities (as per architecture) ===

// User represents a platform user
type StackUser struct {
	ID             uuid.UUID `json:"id" db:"id"`
	AuthProviderID string    `json:"auth_provider_id" db:"auth_provider_id"`
	Email          string    `json:"email" db:"email"`
	KYCStatus      string    `json:"kyc_status" db:"kyc_status"` // pending, approved, rejected
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Wallet represents a managed blockchain wallet
type Wallet struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Chain       Chain     `json:"chain" db:"chain"`
	Address     string    `json:"address" db:"address"`
	ProviderRef string    `json:"provider_ref" db:"provider_ref"` // Reference to wallet manager
	Status      string    `json:"status" db:"status"`             // active, inactive
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Deposit represents a stablecoin deposit
type Deposit struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	UserID      uuid.UUID       `json:"user_id" db:"user_id"`
	Chain       Chain           `json:"chain" db:"chain"`
	TxHash      string          `json:"tx_hash" db:"tx_hash"`
	Token       Stablecoin      `json:"token" db:"token"`
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	Status      string          `json:"status" db:"status"` // pending, confirmed, failed
	ConfirmedAt *time.Time      `json:"confirmed_at" db:"confirmed_at"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// Balance represents user's buying power and pending deposits
type Balance struct {
	UserID          uuid.UUID       `json:"user_id" db:"user_id"`
	BuyingPower     decimal.Decimal `json:"buying_power" db:"buying_power"`
	PendingDeposits decimal.Decimal `json:"pending_deposits" db:"pending_deposits"`
	Currency        string          `json:"currency" db:"currency"` // USD
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// Basket represents a curated investment basket
type Basket struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Description string            `json:"description" db:"description"`
	RiskLevel   RiskLevel         `json:"risk_level" db:"risk_level"`
	Composition []BasketComponent `json:"composition"` // Stored as JSON in DB
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// BasketComponent represents a component within a basket
type BasketComponent struct {
	Symbol string          `json:"symbol"` // e.g., "VTI"
	Weight decimal.Decimal `json:"weight"` // 0.0 to 1.0
}

// Order represents a basket investment order
type Order struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	UserID       uuid.UUID       `json:"user_id" db:"user_id"`
	BasketID     uuid.UUID       `json:"basket_id" db:"basket_id"`
	Side         OrderSide       `json:"side" db:"side"`
	Amount       decimal.Decimal `json:"amount" db:"amount"`
	Status       OrderStatus     `json:"status" db:"status"`
	BrokerageRef *string         `json:"brokerage_ref" db:"brokerage_ref"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// Position represents a user's position in a basket
type Position struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	UserID      uuid.UUID       `json:"user_id" db:"user_id"`
	BasketID    uuid.UUID       `json:"basket_id" db:"basket_id"`
	Quantity    decimal.Decimal `json:"quantity" db:"quantity"`
	AvgPrice    decimal.Decimal `json:"avg_price" db:"avg_price"`
	MarketValue decimal.Decimal `json:"market_value" db:"market_value"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// PortfolioPerformance represents portfolio performance tracking
type PortfolioPerf struct {
	UserID    uuid.UUID       `json:"user_id" db:"user_id"`
	Date      time.Time       `json:"date" db:"date"`
	NAV       decimal.Decimal `json:"nav" db:"nav"` // Net Asset Value
	PnL       decimal.Decimal `json:"pnl" db:"pnl"` // Profit & Loss
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

// AISummary represents AI-generated weekly summaries
type AISummary struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	WeekStart time.Time `json:"week_start" db:"week_start"`
	Summary   string    `json:"summary" db:"summary_md"` // Markdown content
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AuditLog represents audit trail for compliance
type AuditLog struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	Actor      *uuid.UUID `json:"actor" db:"actor"` // User who performed action
	Action     string     `json:"action" db:"action"`
	Entity     string     `json:"entity" db:"entity"`
	Before     *string    `json:"before" db:"before"` // JSON state before
	After      *string    `json:"after" db:"after"`   // JSON state after
	OccurredAt time.Time  `json:"occurred_at" db:"at"`
}

// === API Request/Response Models ===

// DepositAddressRequest represents request for deposit address
type DepositAddressRequest struct {
	Chain Chain `json:"chain" validate:"required"`
}

// DepositAddressResponse represents deposit address response
type DepositAddressResponse struct {
	Chain   Chain   `json:"chain"`
	Address string  `json:"address"`
	QRCode  *string `json:"qrCode,omitempty"` // Optional QR image URL
}

// FundingConfirmation represents a funding confirmation
type FundingConfirmation struct {
	ID          uuid.UUID  `json:"id"`
	Chain       Chain      `json:"chain"`
	TxHash      string     `json:"txHash"`
	Token       Stablecoin `json:"token"`
	Amount      string     `json:"amount"`
	Status      string     `json:"status"`
	ConfirmedAt time.Time  `json:"confirmedAt"`
}

// FundingConfirmationsPage represents paginated funding confirmations response
type FundingConfirmationsPage struct {
	Items      []*FundingConfirmation `json:"items"`
	NextCursor *string                `json:"nextCursor"`
}

// BalancesResponse represents user balances
type BalancesResponse struct {
	BuyingPower     string `json:"buyingPower"`
	PendingDeposits string `json:"pendingDeposits"`
	Currency        string `json:"currency"`
}

// OrderCreateRequest represents order creation request
type OrderCreateRequest struct {
	BasketID       uuid.UUID `json:"basketId" validate:"required"`
	Side           OrderSide `json:"side" validate:"required"`
	Amount         string    `json:"amount" validate:"required"`
	IdempotencyKey *string   `json:"idempotencyKey,omitempty"`
}

// Portfolio represents a user's complete portfolio
type Portfolio struct {
	Currency   string             `json:"currency"`
	Positions  []PositionResponse `json:"positions"`
	TotalValue string             `json:"totalValue"`
}

// PositionResponse represents a position in portfolio response
type PositionResponse struct {
	BasketID    uuid.UUID `json:"basketId"`
	Quantity    string    `json:"quantity"`
	AvgPrice    string    `json:"avgPrice"`
	MarketValue string    `json:"marketValue"`
}

// === Webhook Models ===

// ChainDepositWebhook represents inbound chain deposit webhook
type ChainDepositWebhook struct {
	Chain     Chain      `json:"chain"`
	Address   string     `json:"address"`
	Token     Stablecoin `json:"token"`
	Amount    string     `json:"amount"`
	TxHash    string     `json:"txHash"`
	BlockTime time.Time  `json:"blockTime"`
	Signature string     `json:"signature"`
}

// BrokerageFillWebhook represents brokerage fill webhook
type BrokerageFillWebhook struct {
	OrderID   uuid.UUID       `json:"orderId"`
	Status    OrderStatus     `json:"status"`
	Fills     []BrokerageFill `json:"fills"`
	Signature string          `json:"signature"`
}

// BrokerageFill represents a single fill in brokerage webhook
type BrokerageFill struct {
	Symbol   string `json:"symbol"`
	Quantity string `json:"quantity"`
	Price    string `json:"price"`
}

// ErrorResponse represents API error responses
type StackErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}
