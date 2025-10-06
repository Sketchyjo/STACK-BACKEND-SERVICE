package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/internal/infrastructure/zerog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ZeroGHandler handles internal 0G Network operations
type ZeroGHandler struct {
	storageClient     entities.ZeroGStorageClient
	inferenceGateway  entities.ZeroGInferenceGateway
	namespaceManager  *zerog.NamespaceManager
	logger            *zap.Logger
	tracer            trace.Tracer
}

// NewZeroGHandler creates a new 0G handler
func NewZeroGHandler(
	storageClient entities.ZeroGStorageClient,
	inferenceGateway entities.ZeroGInferenceGateway,
	namespaceManager *zerog.NamespaceManager,
	logger *zap.Logger,
) *ZeroGHandler {
	return &ZeroGHandler{
		storageClient:    storageClient,
		inferenceGateway: inferenceGateway,
		namespaceManager: namespaceManager,
		logger:           logger,
		tracer:           otel.Tracer("zerog-handlers"),
	}
}

// HealthRequest represents the health check request
type HealthRequest struct {
	Service string `json:"service"` // "storage", "inference", or "all"
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Overall   string                    `json:"overall"`
	Services  map[string]*ServiceHealth `json:"services"`
	CheckedAt time.Time                 `json:"checked_at"`
}

// ServiceHealth represents the health of a specific service
type ServiceHealth struct {
	Status    string                 `json:"status"`
	Latency   string                 `json:"latency"`
	Version   string                 `json:"version,omitempty"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
	Errors    []string               `json:"errors,omitempty"`
}

// StoreRequest represents a storage request
type StoreRequest struct {
	Namespace string            `json:"namespace" binding:"required"`
	Data      string            `json:"data" binding:"required"`      // Base64 encoded data
	Metadata  map[string]string `json:"metadata,omitempty"`
	UserID    string            `json:"user_id,omitempty"`
}

// StoreResponse represents a storage response
type StoreResponse struct {
	Success     bool      `json:"success"`
	URI         string    `json:"uri,omitempty"`
	ContentHash string    `json:"content_hash,omitempty"`
	Size        int64     `json:"size,omitempty"`
	StoredAt    time.Time `json:"stored_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// GenerateRequest represents an inference request
type GenerateRequest struct {
	Type      string                 `json:"type" binding:"required"` // "weekly_summary" or "analysis"
	UserID    string                 `json:"user_id" binding:"required"`
	WeekStart *time.Time             `json:"week_start,omitempty"`
	AnalysisType string              `json:"analysis_type,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	// Mock portfolio data for testing
	MockPortfolioData *MockPortfolioData `json:"mock_portfolio_data,omitempty"`
}

// MockPortfolioData represents mock portfolio data for testing
type MockPortfolioData struct {
	TotalValue     float64 `json:"total_value"`
	TotalReturn    float64 `json:"total_return"`
	TotalReturnPct float64 `json:"total_return_pct"`
	WeekChange     float64 `json:"week_change"`
	WeekChangePct  float64 `json:"week_change_pct"`
}

// GenerateResponse represents an inference response
type GenerateResponse struct {
	Success        bool                   `json:"success"`
	RequestID      string                 `json:"request_id,omitempty"`
	Content        string                 `json:"content,omitempty"`
	ContentType    string                 `json:"content_type,omitempty"`
	TokensUsed     int                    `json:"tokens_used,omitempty"`
	ProcessingTime string                 `json:"processing_time,omitempty"`
	Model          string                 `json:"model,omitempty"`
	ArtifactURI    string                 `json:"artifact_uri,omitempty"`
	CreatedAt      time.Time              `json:"created_at,omitempty"`
	Error          string                 `json:"error,omitempty"`
}

// HealthCheck performs a health check on 0G services
// @Summary Health check for 0G services
// @Description Performs health checks on 0G storage and inference services
// @Tags internal
// @Accept json
// @Produce json
// @Param request body HealthRequest false "Health check request"
// @Success 200 {object} HealthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /_internal/0g/health [post]
func (h *ZeroGHandler) HealthCheck(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "handler.zerog_health_check")
	defer span.End()

	var req HealthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to checking all services
		req.Service = "all"
	}

	span.SetAttributes(attribute.String("service", req.Service))

	h.logger.Info("Performing 0G health check",
		zap.String("service", req.Service),
		zap.String("request_id", getRequestID(c)),
	)

	response := &HealthResponse{
		Overall:   entities.HealthStatusHealthy,
		Services:  make(map[string]*ServiceHealth),
		CheckedAt: time.Now(),
	}

	// Check storage service
	if req.Service == "all" || req.Service == "storage" {
		storageHealth := h.checkStorageHealth(ctx)
		response.Services["storage"] = storageHealth
		
		if storageHealth.Status != entities.HealthStatusHealthy {
			response.Overall = entities.HealthStatusDegraded
		}
	}

	// Check inference service
	if req.Service == "all" || req.Service == "inference" {
		inferenceHealth := h.checkInferenceHealth(ctx)
		response.Services["inference"] = inferenceHealth
		
		if inferenceHealth.Status != entities.HealthStatusHealthy {
			if response.Overall == entities.HealthStatusHealthy {
				response.Overall = entities.HealthStatusDegraded
			} else {
				response.Overall = entities.HealthStatusUnhealthy
			}
		}
	}

	// Check namespace manager
	if req.Service == "all" || req.Service == "namespace" {
		namespaceHealth := h.checkNamespaceHealth(ctx)
		response.Services["namespace"] = namespaceHealth
		
		if namespaceHealth.Status != entities.HealthStatusHealthy {
			response.Overall = entities.HealthStatusDegraded
		}
	}

	h.logger.Info("0G health check completed",
		zap.String("overall_status", response.Overall),
		zap.Int("services_checked", len(response.Services)),
		zap.String("request_id", getRequestID(c)),
	)

	c.JSON(http.StatusOK, response)
}

// Store stores data in 0G storage
// @Summary Store data in 0G storage
// @Description Stores data in 0G storage with specified namespace and metadata
// @Tags internal
// @Accept json
// @Produce json
// @Param request body StoreRequest true "Storage request"
// @Success 200 {object} StoreResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /_internal/0g/store [post]
func (h *ZeroGHandler) Store(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "handler.zerog_store")
	defer span.End()

	var req StoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid store request", 
			zap.Error(err),
			zap.String("request_id", getRequestID(c)),
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	span.SetAttributes(
		attribute.String("namespace", req.Namespace),
		attribute.Int("data_size", len(req.Data)),
		attribute.String("user_id", req.UserID),
	)

	h.logger.Info("Storing data in 0G storage",
		zap.String("namespace", req.Namespace),
		zap.Int("data_size", len(req.Data)),
		zap.String("user_id", req.UserID),
		zap.String("request_id", getRequestID(c)),
	)

	// Decode base64 data
	data := []byte(req.Data) // For simplicity, assuming plain text for now

	// Add request metadata
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	req.Metadata["request_id"] = getRequestID(c)
	req.Metadata["stored_via"] = "internal_api"
	if req.UserID != "" {
		req.Metadata["user_id"] = req.UserID
	}

	// Store the data
	result, err := h.storageClient.Store(ctx, req.Namespace, data, req.Metadata)
	if err != nil {
		span.RecordError(err)
		h.logger.Error("Failed to store data in 0G storage",
			zap.Error(err),
			zap.String("namespace", req.Namespace),
			zap.String("request_id", getRequestID(c)),
		)
		
		c.JSON(http.StatusInternalServerError, StoreResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	response := StoreResponse{
		Success:     true,
		URI:         result.URI,
		ContentHash: result.Hash,
		Size:        result.Size,
		StoredAt:    result.StoredAt,
	}

	h.logger.Info("Data stored successfully in 0G storage",
		zap.String("uri", result.URI),
		zap.String("content_hash", result.Hash),
		zap.Int64("size", result.Size),
		zap.String("request_id", getRequestID(c)),
	)

	c.JSON(http.StatusOK, response)
}

// Generate performs AI inference using 0G compute network
// @Summary Generate AI analysis
// @Description Performs AI inference for weekly summaries or on-demand analysis
// @Tags internal
// @Accept json
// @Produce json
// @Param request body GenerateRequest true "Generation request"
// @Success 200 {object} GenerateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /_internal/0g/generate [post]
func (h *ZeroGHandler) Generate(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "handler.zerog_generate")
	defer span.End()

	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid generate request", 
			zap.Error(err),
			zap.String("request_id", getRequestID(c)),
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	span.SetAttributes(
		attribute.String("type", req.Type),
		attribute.String("user_id", req.UserID),
		attribute.String("analysis_type", req.AnalysisType),
	)

	h.logger.Info("Performing AI inference",
		zap.String("type", req.Type),
		zap.String("user_id", req.UserID),
		zap.String("analysis_type", req.AnalysisType),
		zap.String("request_id", getRequestID(c)),
	)

	var result *entities.InferenceResult
	var err error

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, GenerateResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}

	switch req.Type {
	case "weekly_summary":
		weekStart := time.Now()
		if req.WeekStart != nil {
			weekStart = *req.WeekStart
		}

		// Build mock request for demonstration
		summaryRequest := &entities.WeeklySummaryRequest{
			UserID:    userID,
			WeekStart: weekStart,
			WeekEnd:   weekStart.AddDate(0, 0, 6),
			PortfolioData: h.buildMockPortfolioMetrics(req.MockPortfolioData),
			Preferences: &entities.UserPreferences{
				RiskTolerance:  "moderate",
				PreferredStyle: "summary",
				FocusAreas:     []string{"performance", "risk"},
				Language:       "en",
			},
		}

		result, err = h.inferenceGateway.GenerateWeeklySummary(ctx, summaryRequest)

	case "analysis":
		if req.AnalysisType == "" {
			req.AnalysisType = entities.AnalysisTypePerformance
		}

		analysisRequest := &entities.AnalysisRequest{
			UserID:        userID,
			AnalysisType:  req.AnalysisType,
			PortfolioData: h.buildMockPortfolioMetrics(req.MockPortfolioData),
			Preferences: &entities.UserPreferences{
				RiskTolerance:  "moderate",
				PreferredStyle: "detailed",
				FocusAreas:     []string{req.AnalysisType},
				Language:       "en",
			},
			Parameters: req.Parameters,
		}

		result, err = h.inferenceGateway.AnalyzeOnDemand(ctx, analysisRequest)

	default:
		c.JSON(http.StatusBadRequest, GenerateResponse{
			Success: false,
			Error:   "Invalid generation type. Must be 'weekly_summary' or 'analysis'",
		})
		return
	}

	if err != nil {
		span.RecordError(err)
		h.logger.Error("Failed to perform AI inference",
			zap.Error(err),
			zap.String("type", req.Type),
			zap.String("user_id", req.UserID),
			zap.String("request_id", getRequestID(c)),
		)
		
		c.JSON(http.StatusInternalServerError, GenerateResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	response := GenerateResponse{
		Success:        true,
		RequestID:      result.RequestID,
		Content:        result.Content,
		ContentType:    result.ContentType,
		TokensUsed:     result.TokensUsed,
		ProcessingTime: result.ProcessingTime.String(),
		Model:          result.Model,
		ArtifactURI:    result.ArtifactURI,
		CreatedAt:      result.CreatedAt,
	}

	h.logger.Info("AI inference completed successfully",
		zap.String("request_id", result.RequestID),
		zap.String("content_type", result.ContentType),
		zap.Int("tokens_used", result.TokensUsed),
		zap.Duration("processing_time", result.ProcessingTime),
		zap.String("artifact_uri", result.ArtifactURI),
		zap.String("api_request_id", getRequestID(c)),
	)

	c.JSON(http.StatusOK, response)
}

// checkStorageHealth checks the health of the storage service
func (h *ZeroGHandler) checkStorageHealth(ctx context.Context) *ServiceHealth {
	start := time.Now()
	
	health, err := h.storageClient.HealthCheck(ctx)
	if err != nil {
		return &ServiceHealth{
			Status:  entities.HealthStatusUnhealthy,
			Latency: time.Since(start).String(),
			Errors:  []string{err.Error()},
		}
	}

	return &ServiceHealth{
		Status:  health.Status,
		Latency: health.Latency.String(),
		Version: health.Version,
		Metrics: health.Metrics,
		Errors:  health.Errors,
	}
}

// checkInferenceHealth checks the health of the inference service
func (h *ZeroGHandler) checkInferenceHealth(ctx context.Context) *ServiceHealth {
	start := time.Now()
	
	health, err := h.inferenceGateway.HealthCheck(ctx)
	if err != nil {
		return &ServiceHealth{
			Status:  entities.HealthStatusUnhealthy,
			Latency: time.Since(start).String(),
			Errors:  []string{err.Error()},
		}
	}

	return &ServiceHealth{
		Status:  health.Status,
		Latency: health.Latency.String(),
		Version: health.Version,
		Metrics: health.Metrics,
		Errors:  health.Errors,
	}
}

// checkNamespaceHealth checks the health of the namespace manager
func (h *ZeroGHandler) checkNamespaceHealth(ctx context.Context) *ServiceHealth {
	start := time.Now()
	
	health, err := h.namespaceManager.HealthCheck(ctx)
	if err != nil {
		return &ServiceHealth{
			Status:  entities.HealthStatusUnhealthy,
			Latency: time.Since(start).String(),
			Errors:  []string{err.Error()},
		}
	}

	return &ServiceHealth{
		Status:  health.Overall,
		Latency: time.Since(start).String(),
		Metrics: map[string]interface{}{
			"namespaces": len(health.Namespaces),
		},
	}
}

// buildMockPortfolioMetrics builds mock portfolio metrics for testing
func (h *ZeroGHandler) buildMockPortfolioMetrics(mockData *MockPortfolioData) *entities.PortfolioMetrics {
	if mockData == nil {
		// Default mock data
		mockData = &MockPortfolioData{
			TotalValue:     50000.0,
			TotalReturn:    2500.0,
			TotalReturnPct: 5.26,
			WeekChange:     750.0,
			WeekChangePct:  1.52,
		}
	}

	return &entities.PortfolioMetrics{
		TotalValue:       mockData.TotalValue,
		TotalReturn:      mockData.TotalReturn,
		TotalReturnPct:   mockData.TotalReturnPct,
		WeekChange:       mockData.WeekChange,
		WeekChangePct:    mockData.WeekChangePct,
		DayChange:        150.0,
		DayChangePct:     0.30,
		MonthChange:      1250.0,
		MonthChangePct:   2.56,
		Positions: []entities.PositionMetrics{
			{
				BasketID:        uuid.New(),
				BasketName:      "Tech Growth",
				Quantity:        100.0,
				AvgPrice:        250.0,
				CurrentValue:    27500.0,
				UnrealizedPL:    2500.0,
				UnrealizedPLPct: 10.0,
				Weight:          0.55,
			},
			{
				BasketID:        uuid.New(),
				BasketName:      "Balanced Growth",
				Quantity:        75.0,
				AvgPrice:        200.0,
				CurrentValue:    15000.0,
				UnrealizedPL:    0.0,
				UnrealizedPLPct: 0.0,
				Weight:          0.30,
			},
			{
				BasketID:        uuid.New(),
				BasketName:      "Conservative Income",
				Quantity:        50.0,
				AvgPrice:        150.0,
				CurrentValue:    7500.0,
				UnrealizedPL:    0.0,
				UnrealizedPLPct: 0.0,
				Weight:          0.15,
			},
		},
		AllocationByBasket: map[string]float64{
			"Tech Growth":        0.55,
			"Balanced Growth":    0.30,
			"Conservative Income": 0.15,
		},
		RiskMetrics: &entities.RiskMetrics{
			Volatility:      0.12,
			Beta:           1.05,
			SharpeRatio:    0.85,
			MaxDrawdown:    0.08,
			VaR:            0.04,
			Diversification: 0.78,
		},
		PerformanceHistory: []entities.PerformancePoint{
			{Date: time.Now().AddDate(0, 0, -6), Value: 47500.0, PnL: 0.0},
			{Date: time.Now().AddDate(0, 0, -5), Value: 48000.0, PnL: 500.0},
			{Date: time.Now().AddDate(0, 0, -4), Value: 48500.0, PnL: 1000.0},
			{Date: time.Now().AddDate(0, 0, -3), Value: 49000.0, PnL: 1500.0},
			{Date: time.Now().AddDate(0, 0, -2), Value: 49250.0, PnL: 1750.0},
			{Date: time.Now().AddDate(0, 0, -1), Value: 49750.0, PnL: 2250.0},
			{Date: time.Now(), Value: mockData.TotalValue, PnL: mockData.TotalReturn},
		},
	}
}