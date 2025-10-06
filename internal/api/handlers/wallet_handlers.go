package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/internal/domain/services/wallet"
	"github.com/stack-service/stack_service/internal/infrastructure/config"
	"github.com/stack-service/stack_service/pkg/logger"
)

// WalletHandlers contains the wallet-related HTTP handlers
type WalletHandlers struct {
	walletService *wallet.Service
	validator     *validator.Validate
	logger        *zap.Logger
}

// NewWalletHandlers creates a new instance of wallet handlers
func NewWalletHandlers(walletService *wallet.Service, logger *zap.Logger) *WalletHandlers {
	return &WalletHandlers{
		walletService: walletService,
		validator:     validator.New(),
		logger:        logger,
	}
}

// GetWalletAddresses handles GET /wallet/addresses
// @Summary Get wallet addresses
// @Description Returns wallet addresses for the authenticated user, optionally filtered by chain
// @Tags wallet
// @Produce json
// @Param chain query string false "Blockchain network" Enums(ETH,SOL,APTOS)
// @Success 200 {object} entities.WalletAddressesResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 404 {object} entities.ErrorResponse "User not found"
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/wallet/addresses [get]
func (h *WalletHandlers) GetWalletAddresses(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from authenticated context
	userID, err := h.getUserID(c)
	if err != nil {
		h.logger.Warn("Invalid or missing user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "Invalid or missing user ID",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	h.logger.Debug("Getting wallet addresses",
		zap.String("user_id", userID.String()),
		zap.String("request_id", getRequestID(c)))

	// Parse optional chain filter
	var chainFilter *entities.WalletChain
	if chainQuery := c.Query("chain"); chainQuery != "" {
		chain := entities.WalletChain(chainQuery)
		if !chain.IsValid() {
			h.logger.Warn("Invalid chain parameter", zap.String("chain", chainQuery))
			c.JSON(http.StatusBadRequest, entities.ErrorResponse{
				Code:    "INVALID_CHAIN",
				Message: "Invalid blockchain network",
				Details: map[string]interface{}{
					"chain":            chainQuery,
					"supported_chains": []string{"ETH", "ETH-SEPOLIA", "SOL", "SOL-DEVNET", "APTOS", "APTOS-TESTNET"},
				},
			})
			return
		}
		chainFilter = &chain
	}

	// Get wallet addresses
	response, err := h.walletService.GetWalletAddresses(ctx, userID, chainFilter)
	if err != nil {
		h.logger.Error("Failed to get wallet addresses",
			zap.Error(err),
			zap.String("user_id", userID.String()))

		if isUserNotFoundError(err) {
			c.JSON(http.StatusNotFound, entities.ErrorResponse{
				Code:    "USER_NOT_FOUND",
				Message: "User not found",
				Details: map[string]interface{}{"user_id": userID.String()},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "WALLET_RETRIEVAL_FAILED",
			Message: "Failed to retrieve wallet addresses",
			Details: map[string]interface{}{"error": "Internal server error"},
		})
		return
	}

	h.logger.Debug("Retrieved wallet addresses successfully",
		zap.String("user_id", userID.String()),
		zap.Int("wallet_count", len(response.Wallets)))

	c.JSON(http.StatusOK, response)
}

// GetWalletStatus handles GET /wallet/status
// @Summary Get wallet status
// @Description Returns comprehensive wallet status for the authenticated user including provisioning progress
// @Tags wallet
// @Produce json
// @Success 200 {object} entities.WalletStatusResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 404 {object} entities.ErrorResponse "User not found"
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/wallet/status [get]
func (h *WalletHandlers) GetWalletStatus(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from authenticated context
	userID, err := h.getUserID(c)
	if err != nil {
		h.logger.Warn("Invalid or missing user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "Invalid or missing user ID",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	h.logger.Debug("Getting wallet status",
		zap.String("user_id", userID.String()),
		zap.String("request_id", getRequestID(c)))

	// Get wallet status
	response, err := h.walletService.GetWalletStatus(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get wallet status",
			zap.Error(err),
			zap.String("user_id", userID.String()))

		if isUserNotFoundError(err) {
			c.JSON(http.StatusNotFound, entities.ErrorResponse{
				Code:    "USER_NOT_FOUND",
				Message: "User not found",
				Details: map[string]interface{}{"user_id": userID.String()},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "WALLET_STATUS_FAILED",
			Message: "Failed to retrieve wallet status",
			Details: map[string]interface{}{"error": "Internal server error"},
		})
		return
	}

	h.logger.Debug("Retrieved wallet status successfully",
		zap.String("user_id", userID.String()),
		zap.Int("total_wallets", response.TotalWallets),
		zap.Int("ready_wallets", response.ReadyWallets))

	c.JSON(http.StatusOK, response)
}

// CreateWalletsForUser handles POST /wallet/create (Admin only)
// @Summary Create wallets for user
// @Description Manually trigger wallet creation for a user (Admin only)
// @Tags wallet
// @Accept json
// @Produce json
// @Param request body CreateWalletsRequest true "Wallet creation request"
// @Success 202 {object} map[string]interface{} "Wallet creation initiated"
// @Failure 400 {object} entities.ErrorResponse
// @Failure 403 {object} entities.ErrorResponse "Insufficient permissions"
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/wallet/create [post]
func (h *WalletHandlers) CreateWalletsForUser(c *gin.Context) {
	ctx := c.Request.Context()

	h.logger.Info("Manual wallet creation requested",
		zap.String("request_id", getRequestID(c)),
		zap.String("ip", c.ClientIP()))

	var req CreateWalletsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid wallet creation request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid wallet creation request payload",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Wallet creation request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Wallet creation request validation failed",
			Details: map[string]interface{}{"validation_errors": err.Error()},
		})
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.logger.Warn("Invalid user ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "Invalid user ID format",
			Details: map[string]interface{}{"user_id": req.UserID},
		})
		return
	}

	// Validate chains
	var chains []entities.WalletChain
	for _, chainStr := range req.Chains {
		chain := entities.WalletChain(chainStr)
		if !chain.IsValid() {
			h.logger.Warn("Invalid chain in request", zap.String("chain", chainStr))
			c.JSON(http.StatusBadRequest, entities.ErrorResponse{
				Code:    "INVALID_CHAIN",
				Message: "Invalid blockchain network",
				Details: map[string]interface{}{
					"chain":            chainStr,
					"supported_chains": []string{"ETH", "ETH-SEPOLIA", "SOL", "SOL-DEVNET", "APTOS", "APTOS-TESTNET"},
				},
			})
			return
		}
		chains = append(chains, chain)
	}

	// Create wallets
	err = h.walletService.CreateWalletsForUser(ctx, userID, chains)
	if err != nil {
		h.logger.Error("Failed to create wallets for user",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.Strings("chains", req.Chains))

		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "WALLET_CREATION_FAILED",
			Message: "Failed to create wallets for user",
			Details: map[string]interface{}{"error": "Internal server error"},
		})
		return
	}

	h.logger.Info("Wallet creation initiated successfully",
		zap.String("user_id", userID.String()),
		zap.Strings("chains", req.Chains))

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Wallet creation initiated",
		"user_id":    userID.String(),
		"chains":     req.Chains,
		"next_steps": []string{"Check wallet status for progress", "Wallets will be available once provisioning completes"},
	})
}

// RetryWalletProvisioning handles POST /wallet/retry (Admin only)
// @Summary Retry failed wallet provisioning
// @Description Retries failed wallet provisioning jobs (Admin only)
// @Tags wallet
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of jobs to retry" default(10)
// @Success 200 {object} map[string]interface{} "Retry initiated"
// @Failure 400 {object} entities.ErrorResponse
// @Failure 403 {object} entities.ErrorResponse "Insufficient permissions"
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/wallet/retry [post]
func (h *WalletHandlers) RetryWalletProvisioning(c *gin.Context) {
	ctx := c.Request.Context()

	h.logger.Info("Wallet provisioning retry requested",
		zap.String("request_id", getRequestID(c)),
		zap.String("ip", c.ClientIP()))

	// Parse limit parameter
	limit := 10 // default
	if limitQuery := c.Query("limit"); limitQuery != "" {
		if parsedLimit, err := strconv.Atoi(limitQuery); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Retry failed provisioning jobs
	err := h.walletService.RetryFailedWalletProvisioning(ctx, limit)
	if err != nil {
		h.logger.Error("Failed to retry wallet provisioning", zap.Error(err))

		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Code:    "RETRY_FAILED",
			Message: "Failed to retry wallet provisioning",
			Details: map[string]interface{}{"error": "Internal server error"},
		})
		return
	}

	h.logger.Info("Wallet provisioning retry initiated", zap.Int("limit", limit))

	c.JSON(http.StatusOK, gin.H{
		"message": "Wallet provisioning retry initiated",
		"limit":   limit,
	})
}

// HealthCheck handles GET /wallet/health (Admin only)
// @Summary Wallet service health check
// @Description Returns health status of wallet service and Circle integration
// @Tags wallet
// @Produce json
// @Success 200 {object} map[string]interface{} "Health status"
// @Failure 500 {object} entities.ErrorResponse
// @Router /api/v1/wallet/health [get]
func (h *WalletHandlers) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	h.logger.Debug("Wallet service health check requested")

	// Perform health check
	err := h.walletService.HealthCheck(ctx)
	if err != nil {
		h.logger.Error("Wallet service health check failed", zap.Error(err))

		c.JSON(http.StatusServiceUnavailable, entities.ErrorResponse{
			Code:    "HEALTH_CHECK_FAILED",
			Message: "Wallet service health check failed",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Get metrics
	metrics := h.walletService.GetMetrics()

	h.logger.Debug("Wallet service health check passed")

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "wallet",
		"metrics": metrics,
	})
}

// Helper methods

func (h *WalletHandlers) getUserID(c *gin.Context) (uuid.UUID, error) {
	// Try to get from authenticated user context first
	if userIDStr, exists := c.Get("user_id"); exists {
		if userID, ok := userIDStr.(uuid.UUID); ok {
			return userID, nil
		}
		if userIDStr, ok := userIDStr.(string); ok {
			return uuid.Parse(userIDStr)
		}
	}

	// Fallback to query parameter for development/admin use
	userIDQuery := c.Query("user_id")
	if userIDQuery != "" {
		return uuid.Parse(userIDQuery)
	}

	return uuid.Nil, fmt.Errorf("user ID not found in context or query parameters")
}

// Request/Response models

type CreateWalletsRequest struct {
	UserID string   `json:"user_id" validate:"required,uuid"`
	Chains []string `json:"chains" validate:"required,min=1"`
}

// Legacy handler factories for compatibility
func GetWalletAddresses(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "Not implemented yet",
			"message": "Use WalletHandlers.GetWalletAddresses instead",
		})
	}
}

func GetWalletStatus(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "Not implemented yet",
			"message": "Use WalletHandlers.GetWalletStatus instead",
		})
	}
}
