package adapters

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/internal/domain/services/investing"
	"go.uber.org/zap"
)

// BrokerageAdapter provides integration with brokerage partner (Alpaca)
// TODO: Implement actual Alpaca API integration
type BrokerageAdapter struct {
	apiKey  string
	baseURL string
	logger  *zap.Logger
}

// NewBrokerageAdapter creates a new brokerage adapter instance
func NewBrokerageAdapter(apiKey, baseURL string, logger *zap.Logger) *BrokerageAdapter {
	return &BrokerageAdapter{
		apiKey:  apiKey,
		baseURL: baseURL,
		logger:  logger,
	}
}

// PlaceOrder submits an order to the brokerage
// TODO: Implement actual API call to Alpaca
func (a *BrokerageAdapter) PlaceOrder(ctx context.Context, basketID uuid.UUID, side entities.OrderSide, amount decimal.Decimal) (*investing.BrokerageOrderResponse, error) {
	a.logger.Info("PlaceOrder called (stub implementation)",
		zap.String("basket_id", basketID.String()),
		zap.String("side", string(side)),
		zap.String("amount", amount.String()),
	)

	// Stub implementation - returns mock accepted order
	// TODO: Replace with actual Alpaca API call
	orderRef := fmt.Sprintf("DW-%s", uuid.New().String()[:8])

	return &investing.BrokerageOrderResponse{
		OrderRef: orderRef,
		Status:   entities.OrderStatusAccepted,
	}, nil
}

// GetOrderStatus retrieves the current status of an order from the brokerage
// TODO: Implement actual API call to Alpaca
func (a *BrokerageAdapter) GetOrderStatus(ctx context.Context, brokerageRef string) (*investing.BrokerageOrderStatus, error) {
	a.logger.Info("GetOrderStatus called (stub implementation)",
		zap.String("brokerage_ref", brokerageRef),
	)

	// Stub implementation - returns mock filled status
	// TODO: Replace with actual Alpaca API call
	return &investing.BrokerageOrderStatus{
		Status: entities.OrderStatusFilled,
		Fills:  []entities.BrokerageFill{},
	}, nil
}

// CancelOrder cancels an order at the brokerage
// TODO: Implement actual API call to Alpaca
func (a *BrokerageAdapter) CancelOrder(ctx context.Context, brokerageRef string) error {
	a.logger.Info("CancelOrder called (stub implementation)",
		zap.String("brokerage_ref", brokerageRef),
	)

	// Stub implementation - returns success
	// TODO: Replace with actual Alpaca API call
	return nil
}
