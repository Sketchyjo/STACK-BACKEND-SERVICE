package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/internal/domain/services/funding"
	"github.com/stack-service/stack_service/internal/domain/services/investing"

	// "github.com/stack-service/stack_service/internal/infrastructure/config"
	"github.com/stack-service/stack_service/pkg/logger"
)

// StackHandlers contains the STACK MVP API handlers
type StackHandlers struct {
	fundingService   *funding.Service
	investingService *investing.Service
	logger           *logger.Logger
}

// NewStackHandlers creates handlers for STACK MVP endpoints
func NewStackHandlers(
	fundingService *funding.Service,
	investingService *investing.Service,
	logger *logger.Logger,
) *StackHandlers {
	return &StackHandlers{
		fundingService:   fundingService,
		investingService: investingService,
		logger:           logger,
	}
}

// checkInvestingServiceAvailable returns true if the investing service is available, otherwise sends error response
func (h *StackHandlers) checkInvestingServiceAvailable(c *gin.Context) bool {
	if h.investingService == nil {
		c.JSON(http.StatusServiceUnavailable, entities.ErrorResponse{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Investing service is not available",
		})
		return false
	}
	return true
}

// checkFundingServiceAvailable returns true if the funding service is available, otherwise sends error response
func (h *StackHandlers) checkFundingServiceAvailable(c *gin.Context) bool {
	if h.fundingService == nil {
		c.JSON(http.StatusServiceUnavailable, entities.ErrorResponse{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Funding service is not available",
		})
		return false
	}
	return true
}

// === FUNDING ENDPOINTS ===

// CreateDepositAddress generates or retrieves a deposit address for a specific chain
// POST /funding/deposit/address
func (h *StackHandlers) CreateDepositAddress(c *gin.Context) {
	if !h.checkFundingServiceAvailable(c) {
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	var req entities.DepositAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request format",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	resp, err := h.fundingService.CreateDepositAddress(c.Request.Context(), userID.(uuid.UUID), req.Chain)
	if err != nil {
		h.logger.Error("Failed to create deposit address", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to generate deposit address",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListFundingConfirmations lists recent funding confirmations for the authenticated user
// GET /funding/confirmations
func (h *StackHandlers) ListFundingConfirmations(c *gin.Context) {
	if !h.checkFundingServiceAvailable(c) {
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	// Parse query parameters
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if cursor := c.Query("cursor"); cursor != "" {
		if o, err := strconv.Atoi(cursor); err == nil && o > 0 {
			offset = o
		}
	}

	confirmations, err := h.fundingService.GetFundingConfirmations(c.Request.Context(), userID.(uuid.UUID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get funding confirmations", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve confirmations",
		})
		return
	}

	// Prepare response with pagination
	response := map[string]interface{}{
		"items": confirmations,
	}

	// Add cursor for next page if we got full limit
	if len(confirmations) == limit {
		response["nextCursor"] = strconv.Itoa(offset + limit)
	}

	c.JSON(http.StatusOK, response)
}

// GetBalances returns the user's current balances (buying power, pending, etc.)
// GET /balances
func (h *StackHandlers) GetBalances(c *gin.Context) {
	if !h.checkFundingServiceAvailable(c) {
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	balance, err := h.fundingService.GetBalance(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get balance", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve balance",
		})
		return
	}

	c.JSON(http.StatusOK, balance)
}

// === INVESTING ENDPOINTS ===

// ListBaskets returns all curated baskets
// GET /baskets
func (h *StackHandlers) ListBaskets(c *gin.Context) {
	if h.investingService == nil {
		c.JSON(http.StatusServiceUnavailable, entities.ErrorResponse{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Investing service is not available",
		})
		return
	}

	baskets, err := h.investingService.ListBaskets(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list baskets", "error", err)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve baskets",
		})
		return
	}

	response := map[string]interface{}{
		"items": baskets,
	}

	c.JSON(http.StatusOK, response)
}

// GetBasket returns a single basket by ID
// GET /baskets/{id}
func (h *StackHandlers) GetBasket(c *gin.Context) {
	if !h.checkInvestingServiceAvailable(c) {
		return
	}

	basketIDStr := c.Param("id")
	basketID, err := uuid.Parse(basketIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid basket ID format",
		})
		return
	}

	basket, err := h.investingService.GetBasket(c.Request.Context(), basketID)
	if err != nil {
		if err == investing.ErrBasketNotFound {
			c.JSON(http.StatusNotFound, entities.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: "Basket not found",
			})
			return
		}

		h.logger.Error("Failed to get basket", "error", err, "basket_id", basketID)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve basket",
		})
		return
	}

	c.JSON(http.StatusOK, basket)
}

// CreateOrder creates a new buy/sell order
// POST /orders
func (h *StackHandlers) CreateOrder(c *gin.Context) {
	if !h.checkInvestingServiceAvailable(c) {
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	var req entities.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request format",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	order, err := h.investingService.CreateOrder(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		switch err {
		case investing.ErrBasketNotFound:
			c.JSON(http.StatusBadRequest, entities.ErrorResponse{
				Code:    "BAD_REQUEST",
				Message: "Invalid basket ID",
			})
			return
		case investing.ErrInvalidAmount:
			c.JSON(http.StatusBadRequest, entities.ErrorResponse{
				Code:    "BAD_REQUEST",
				Message: "Invalid amount",
			})
			return
		case investing.ErrInsufficientFunds:
			c.JSON(http.StatusConflict, entities.ErrorResponse{
				Code:    "INSUFFICIENT_FUNDS",
				Message: "Insufficient buying power",
			})
			return
		case investing.ErrInsufficientPosition:
			c.JSON(http.StatusConflict, entities.ErrorResponse{
				Code:    "INSUFFICIENT_POSITION",
				Message: "Insufficient position to sell",
			})
			return
		}

		h.logger.Error("Failed to create order", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create order",
		})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// ListOrders returns order history for the authenticated user
// GET /orders
func (h *StackHandlers) ListOrders(c *gin.Context) {
	if !h.checkInvestingServiceAvailable(c) {
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	// Parse query parameters
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if cursor := c.Query("cursor"); cursor != "" {
		if o, err := strconv.Atoi(cursor); err == nil && o > 0 {
			offset = o
		}
	}

	var status *entities.OrderStatus
	if statusStr := c.Query("status"); statusStr != "" {
		orderStatus := entities.OrderStatus(statusStr)
		status = &orderStatus
	}

	orders, err := h.investingService.ListOrders(c.Request.Context(), userID.(uuid.UUID), limit, offset, status)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve orders",
		})
		return
	}

	// Prepare response with pagination
	response := map[string]interface{}{
		"items": orders,
	}

	// Add cursor for next page if we got full limit
	if len(orders) == limit {
		response["nextCursor"] = strconv.Itoa(offset + limit)
	}

	c.JSON(http.StatusOK, response)
}

// GetOrder returns order by ID
// GET /orders/{id}
func (h *StackHandlers) GetOrder(c *gin.Context) {
	if !h.checkInvestingServiceAvailable(c) {
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid order ID format",
		})
		return
	}

	order, err := h.investingService.GetOrder(c.Request.Context(), userID.(uuid.UUID), orderID)
	if err != nil {
		if err == investing.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, entities.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: "Order not found",
			})
			return
		}

		h.logger.Error("Failed to get order", "error", err, "user_id", userID, "order_id", orderID)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve order",
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetPortfolio returns current portfolio positions and valuations
// GET /portfolio
func (h *StackHandlers) GetPortfolio(c *gin.Context) {
	if !h.checkInvestingServiceAvailable(c) {
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	portfolio, err := h.investingService.GetPortfolio(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get portfolio", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve portfolio",
		})
		return
	}

	c.JSON(http.StatusOK, portfolio)
}

// GetPortfolioOverview returns comprehensive portfolio overview with balance and performance
// GET /portfolio/overview
func (h *StackHandlers) GetPortfolioOverview(c *gin.Context) {
	if !h.checkInvestingServiceAvailable(c) {
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, entities.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	overview, err := h.investingService.GetPortfolioOverview(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get portfolio overview", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve portfolio overview",
		})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// === WEBHOOK ENDPOINTS ===

// ChainDepositWebhook handles inbound chain webhook for deposits/confirmations
// POST /webhooks/chain-deposit
func (h *StackHandlers) ChainDepositWebhook(c *gin.Context) {
	if !h.checkFundingServiceAvailable(c) {
		return
	}

	var webhook entities.ChainDepositWebhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		h.logger.Error("Invalid chain deposit webhook payload", "error", err)
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid webhook payload",
		})
		return
	}

	if err := h.fundingService.ProcessChainDeposit(c.Request.Context(), &webhook); err != nil {
		h.logger.Error("Failed to process chain deposit", "error", err, "tx_hash", webhook.TxHash)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "PROCESSING_ERROR",
			Message: "Failed to process deposit",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// BrokerageFillWebhook handles inbound brokerage fills/exec reports
// POST /webhooks/brokerage-fills
func (h *StackHandlers) BrokerageFillWebhook(c *gin.Context) {
	if !h.checkInvestingServiceAvailable(c) {
		return
	}

	var webhook entities.BrokerageFillWebhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		h.logger.Error("Invalid brokerage fill webhook payload", "error", err)
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid webhook payload",
		})
		return
	}

	if err := h.investingService.ProcessBrokerageFill(c.Request.Context(), &webhook); err != nil {
		h.logger.Error("Failed to process brokerage fill", "error", err, "order_id", webhook.OrderID)
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "PROCESSING_ERROR",
			Message: "Failed to process fill",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}
