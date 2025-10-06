package zerog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/internal/infrastructure/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// InferenceGateway implements the ZeroGInferenceGateway interface
type InferenceGateway struct {
	config        *config.ZeroGComputeConfig
	storageClient entities.ZeroGStorageClient
	logger        *zap.Logger
	tracer        trace.Tracer
	metrics       *InferenceMetrics
	httpClient    *http.Client
}

// InferenceMetrics contains observability metrics for inference operations
type InferenceMetrics struct {
	RequestsTotal     metric.Int64Counter
	RequestDuration   metric.Float64Histogram
	RequestErrors     metric.Int64Counter
	TokensUsed        metric.Int64Counter
	TokensGenerated   metric.Int64Counter
ActiveRequests    metric.Int64UpDownCounter
	ModelUsage        metric.Int64Counter
}

// OpenAICompatibleRequest represents an OpenAI-compatible API request
type OpenAICompatibleRequest struct {
	Model       string                   `json:"model"`
	Messages    []ChatMessage            `json:"messages"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Temperature float64                  `json:"temperature,omitempty"`
	TopP        float64                  `json:"top_p,omitempty"`
	FrequencyPenalty float64             `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64             `json:"presence_penalty,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
	User        string                   `json:"user,omitempty"`
}

// ChatMessage represents a chat message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"` // message content
}

// OpenAICompatibleResponse represents an OpenAI-compatible API response
type OpenAICompatibleResponse struct {
	ID                string           `json:"id"`
	Object            string           `json:"object"`
	Created           int64            `json:"created"`
	Model             string           `json:"model"`
	Choices           []Choice         `json:"choices"`
	Usage             *Usage           `json:"usage,omitempty"`
	SystemFingerprint string           `json:"system_fingerprint,omitempty"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message,omitempty"`
	FinishReason string       `json:"finish_reason,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewInferenceGateway creates a new 0G inference gateway
func NewInferenceGateway(cfg *config.ZeroGComputeConfig, storageClient entities.ZeroGStorageClient, logger *zap.Logger) (*InferenceGateway, error) {
	if cfg == nil {
		return nil, fmt.Errorf("compute config is required")
	}

	if cfg.BrokerEndpoint == "" {
		return nil, fmt.Errorf("broker endpoint is required")
	}

	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("private key is required")
	}

	tracer := otel.Tracer("zerog-inference")
	meter := otel.Meter("zerog-inference")

	// Initialize metrics
	metrics, err := initInferenceMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Create HTTP client with timeouts
	httpClient := &http.Client{
		Timeout: 60 * time.Second, // TODO: Make configurable
		Transport: &http.Transport{
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
		},
	}

	gateway := &InferenceGateway{
		config:        cfg,
		storageClient: storageClient,
		logger:        logger,
		tracer:        tracer,
		metrics:       metrics,
		httpClient:    httpClient,
	}

	logger.Info("0G inference gateway initialized",
		zap.String("broker_endpoint", cfg.BrokerEndpoint),
		zap.String("default_model", cfg.ModelConfig.DefaultModel),
	)

	return gateway, nil
}

// GenerateWeeklySummary creates an AI-generated weekly portfolio summary
func (g *InferenceGateway) GenerateWeeklySummary(ctx context.Context, request *entities.WeeklySummaryRequest) (*entities.InferenceResult, error) {
	startTime := time.Now()
	ctx, span := g.tracer.Start(ctx, "inference.generate_weekly_summary", trace.WithAttributes(
		attribute.String("user_id", request.UserID.String()),
		attribute.String("week_start", request.WeekStart.Format("2006-01-02")),
	))
	defer span.End()

	g.metrics.RequestsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", "weekly_summary"),
		attribute.String("model", g.config.ModelConfig.DefaultModel),
	))
	g.metrics.ActiveRequests.Add(ctx, 1)
	defer g.metrics.ActiveRequests.Add(ctx, -1)

	g.logger.Info("Generating weekly summary",
		zap.String("user_id", request.UserID.String()),
		zap.String("week_start", request.WeekStart.Format("2006-01-02")),
		zap.Float64("total_value", request.PortfolioData.TotalValue),
	)

	// Build the prompt for weekly summary
	prompt, err := g.buildWeeklySummaryPrompt(request)
	if err != nil {
		span.RecordError(err)
		g.recordError(ctx, "weekly_summary", err)
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// Make inference request
	result, err := g.makeInferenceRequest(ctx, "weekly_summary", prompt, request.UserID.String())
	if err != nil {
		span.RecordError(err)
		g.recordError(ctx, "weekly_summary", err)
		return nil, fmt.Errorf("inference request failed: %w", err)
	}

	// Store detailed analysis as artifact
	if err := g.storeArtifact(ctx, result, request.UserID.String(), "weekly_summary"); err != nil {
		g.logger.Warn("Failed to store artifact", zap.Error(err))
		// Don't fail the entire operation if artifact storage fails
	}

	// Record success metrics
	duration := time.Since(startTime)
	g.metrics.RequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("operation", "weekly_summary"),
		attribute.Bool("success", true),
	))

	g.logger.Info("Weekly summary generated successfully",
		zap.String("user_id", request.UserID.String()),
		zap.String("request_id", result.RequestID),
		zap.Int("tokens_used", result.TokensUsed),
		zap.Duration("duration", duration),
	)

	return result, nil
}

// AnalyzeOnDemand performs on-demand portfolio analysis
func (g *InferenceGateway) AnalyzeOnDemand(ctx context.Context, request *entities.AnalysisRequest) (*entities.InferenceResult, error) {
	startTime := time.Now()
	ctx, span := g.tracer.Start(ctx, "inference.analyze_on_demand", trace.WithAttributes(
		attribute.String("user_id", request.UserID.String()),
		attribute.String("analysis_type", request.AnalysisType),
	))
	defer span.End()

	g.metrics.RequestsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", "on_demand_analysis"),
		attribute.String("analysis_type", request.AnalysisType),
		attribute.String("model", g.config.ModelConfig.DefaultModel),
	))
	g.metrics.ActiveRequests.Add(ctx, 1)
	defer g.metrics.ActiveRequests.Add(ctx, -1)

	g.logger.Info("Performing on-demand analysis",
		zap.String("user_id", request.UserID.String()),
		zap.String("analysis_type", request.AnalysisType),
		zap.Float64("total_value", request.PortfolioData.TotalValue),
	)

	// Build the prompt for on-demand analysis
	prompt, err := g.buildAnalysisPrompt(request)
	if err != nil {
		span.RecordError(err)
		g.recordError(ctx, "on_demand_analysis", err)
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// Make inference request
	result, err := g.makeInferenceRequest(ctx, "on_demand_analysis", prompt, request.UserID.String())
	if err != nil {
		span.RecordError(err)
		g.recordError(ctx, "on_demand_analysis", err)
		return nil, fmt.Errorf("inference request failed: %w", err)
	}

	// Store detailed analysis as artifact
	if err := g.storeArtifact(ctx, result, request.UserID.String(), request.AnalysisType); err != nil {
		g.logger.Warn("Failed to store artifact", zap.Error(err))
		// Don't fail the entire operation if artifact storage fails
	}

	// Record success metrics
	duration := time.Since(startTime)
	g.metrics.RequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("operation", "on_demand_analysis"),
		attribute.String("analysis_type", request.AnalysisType),
		attribute.Bool("success", true),
	))

	g.logger.Info("On-demand analysis completed successfully",
		zap.String("user_id", request.UserID.String()),
		zap.String("analysis_type", request.AnalysisType),
		zap.String("request_id", result.RequestID),
		zap.Int("tokens_used", result.TokensUsed),
		zap.Duration("duration", duration),
	)

	return result, nil
}

// HealthCheck verifies connectivity to 0G compute network
func (g *InferenceGateway) HealthCheck(ctx context.Context) (*entities.HealthStatus, error) {
	startTime := time.Now()
	ctx, span := g.tracer.Start(ctx, "inference.health_check")
	defer span.End()

	g.logger.Debug("Performing 0G compute health check")

	// TODO: Implement actual health check against 0G compute network
	// For now, return a mock healthy status
	status := &entities.HealthStatus{
		Status:      entities.HealthStatusHealthy,
		Latency:     time.Since(startTime),
		Version:     "1.0.0", // TODO: Get actual version
		Uptime:      24 * time.Hour, // TODO: Get actual uptime
		Metrics: map[string]interface{}{
			"available_models": []string{g.config.ModelConfig.DefaultModel},
			"provider_id":      g.config.ProviderID,
			"active_requests":  0, // TODO: Get actual metrics
		},
		LastChecked: time.Now(),
		Errors:      []string{},
	}

	g.logger.Info("0G compute health check completed",
		zap.String("status", status.Status),
		zap.Duration("latency", status.Latency),
	)

	return status, nil
}

// GetServiceInfo returns information about available inference services
func (g *InferenceGateway) GetServiceInfo(ctx context.Context) (*entities.ServiceInfo, error) {
	ctx, span := g.tracer.Start(ctx, "inference.get_service_info")
	defer span.End()

	// TODO: Implement actual service discovery
	// For now, return mock service info
	serviceInfo := &entities.ServiceInfo{
		ProviderID:  g.config.ProviderID,
		ServiceName: "0G AI-CFO Service",
		Models: []entities.ModelInfo{
			{
				ModelID:     g.config.ModelConfig.DefaultModel,
				Name:        "GPT-4 Compatible Model",
				Description: "Large language model optimized for financial analysis",
				MaxTokens:   g.config.ModelConfig.MaxTokens,
				Version:     "1.0.0",
				UpdatedAt:   time.Now(),
			},
		},
		Pricing: &entities.PricingInfo{
			Currency:      "USD",
			BaseRate:      0.01,
			TokenRate:     0.002,
			MinimumCharge: 0.001,
		},
		Capabilities: []string{
			"weekly_summaries",
			"portfolio_analysis",
			"risk_assessment",
			"performance_evaluation",
		},
		Status: entities.HealthStatusHealthy,
		Metadata: map[string]interface{}{
			"endpoint":       g.config.BrokerEndpoint,
			"max_tokens":     g.config.ModelConfig.MaxTokens,
			"temperature":    g.config.ModelConfig.Temperature,
			"auto_topup":     g.config.Funding.AutoTopup,
		},
	}

	return serviceInfo, nil
}

// makeInferenceRequest makes a request to the 0G compute network
func (g *InferenceGateway) makeInferenceRequest(ctx context.Context, operation string, prompt string, userID string) (*entities.InferenceResult, error) {
	requestID := uuid.New().String()
	
	// Build OpenAI-compatible request
	apiRequest := &OpenAICompatibleRequest{
		Model: g.config.ModelConfig.DefaultModel,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are an AI-powered Chief Financial Officer (AI-CFO) providing professional investment analysis and portfolio insights. Provide clear, actionable insights while avoiding specific trading recommendations.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:        g.config.ModelConfig.MaxTokens,
		Temperature:      g.config.ModelConfig.Temperature,
		TopP:             g.config.ModelConfig.TopP,
		FrequencyPenalty: g.config.ModelConfig.FrequencyPenalty,
		PresencePenalty:  g.config.ModelConfig.PresencePenalty,
		Stream:           false,
		User:             userID,
	}

	// TODO: Implement actual 0G compute network request
	// For now, simulate the request and return mock response
	result := g.simulateInferenceResponse(requestID, apiRequest, operation)
	
	// Record token usage metrics
	g.metrics.TokensUsed.Add(ctx, int64(result.TokensUsed))
	g.metrics.ModelUsage.Add(ctx, 1, metric.WithAttributes(
		attribute.String("model", result.Model),
		attribute.String("operation", operation),
	))

	return result, nil
}

// simulateInferenceResponse creates a mock inference response for development
func (g *InferenceGateway) simulateInferenceResponse(requestID string, request *OpenAICompatibleRequest, operation string) *entities.InferenceResult {
	// Generate mock response based on operation type
	var content string
	switch operation {
	case "weekly_summary":
		content = g.generateMockWeeklySummary()
	case "on_demand_analysis":
		content = g.generateMockAnalysis()
	default:
		content = "Analysis completed successfully."
	}

	return &entities.InferenceResult{
		RequestID:      requestID,
		Content:        content,
		ContentType:    "text/markdown",
		Metadata: map[string]interface{}{
			"model_version": "1.0.0",
			"provider":      "0G Network",
			"operation":     operation,
		},
		TokensUsed:     250, // Mock value
		ProcessingTime: 2 * time.Second, // Mock value
		Model:          request.Model,
		CreatedAt:      time.Now(),
		ArtifactURI:    "", // Will be set by storeArtifact
	}
}

// buildWeeklySummaryPrompt constructs the prompt for weekly summary generation
func (g *InferenceGateway) buildWeeklySummaryPrompt(request *entities.WeeklySummaryRequest) (string, error) {
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString("# Weekly Portfolio Summary Request\n\n")
	promptBuilder.WriteString(fmt.Sprintf("**Week Period**: %s to %s\n\n", 
		request.WeekStart.Format("January 2, 2006"), 
		request.WeekEnd.Format("January 2, 2006")))
	
	// Portfolio performance data
	if request.PortfolioData != nil {
		promptBuilder.WriteString("## Portfolio Performance\n")
		promptBuilder.WriteString(fmt.Sprintf("- **Total Value**: $%.2f\n", request.PortfolioData.TotalValue))
		promptBuilder.WriteString(fmt.Sprintf("- **Week Change**: $%.2f (%.2f%%)\n", 
			request.PortfolioData.WeekChange, request.PortfolioData.WeekChangePct))
		promptBuilder.WriteString(fmt.Sprintf("- **Total Return**: $%.2f (%.2f%%)\n", 
			request.PortfolioData.TotalReturn, request.PortfolioData.TotalReturnPct))
		
		// Position details
		if len(request.PortfolioData.Positions) > 0 {
			promptBuilder.WriteString("\n### Position Performance\n")
			for _, pos := range request.PortfolioData.Positions {
				promptBuilder.WriteString(fmt.Sprintf("- **%s**: $%.2f (%.2f%% of portfolio), P&L: $%.2f (%.2f%%)\n",
					pos.BasketName, pos.CurrentValue, pos.Weight*100, pos.UnrealizedPL, pos.UnrealizedPLPct))
			}
		}
		
		// Risk metrics
		if request.PortfolioData.RiskMetrics != nil {
			promptBuilder.WriteString("\n### Risk Analysis\n")
			promptBuilder.WriteString(fmt.Sprintf("- **Volatility**: %.2f%%\n", request.PortfolioData.RiskMetrics.Volatility*100))
			promptBuilder.WriteString(fmt.Sprintf("- **Max Drawdown**: %.2f%%\n", request.PortfolioData.RiskMetrics.MaxDrawdown*100))
			promptBuilder.WriteString(fmt.Sprintf("- **Diversification Score**: %.2f/1.0\n", request.PortfolioData.RiskMetrics.Diversification))
		}
	}
	
	// User preferences
	if request.Preferences != nil {
		promptBuilder.WriteString(fmt.Sprintf("\n## User Preferences\n"))
		promptBuilder.WriteString(fmt.Sprintf("- **Risk Tolerance**: %s\n", request.Preferences.RiskTolerance))
		promptBuilder.WriteString(fmt.Sprintf("- **Preferred Style**: %s\n", request.Preferences.PreferredStyle))
		if len(request.Preferences.FocusAreas) > 0 {
			promptBuilder.WriteString(fmt.Sprintf("- **Focus Areas**: %s\n", strings.Join(request.Preferences.FocusAreas, ", ")))
		}
	}
	
	promptBuilder.WriteString("\n---\n\n")
	promptBuilder.WriteString("Please provide a comprehensive weekly portfolio summary in markdown format that includes:\n\n")
	promptBuilder.WriteString("1. **Executive Summary** - Key highlights and overall performance\n")
	promptBuilder.WriteString("2. **Performance Analysis** - Detailed breakdown of gains/losses and contributing factors\n")
	promptBuilder.WriteString("3. **Risk Assessment** - Current risk exposure and portfolio health\n")
	promptBuilder.WriteString("4. **Market Context** - Relevant market conditions that affected performance\n")
	promptBuilder.WriteString("5. **Key Observations** - Notable trends or changes in the portfolio\n")
	promptBuilder.WriteString("6. **Looking Ahead** - Considerations for the upcoming week\n\n")
	promptBuilder.WriteString("**Important**: Focus on educational insights and avoid specific trading recommendations. Include appropriate disclaimers about investment risks.\n")
	
	return promptBuilder.String(), nil
}

// buildAnalysisPrompt constructs the prompt for on-demand analysis
func (g *InferenceGateway) buildAnalysisPrompt(request *entities.AnalysisRequest) (string, error) {
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString(fmt.Sprintf("# On-Demand Portfolio Analysis: %s\n\n", strings.Title(request.AnalysisType)))
	
	// Portfolio data
	if request.PortfolioData != nil {
		promptBuilder.WriteString("## Current Portfolio\n")
		promptBuilder.WriteString(fmt.Sprintf("- **Total Value**: $%.2f\n", request.PortfolioData.TotalValue))
		promptBuilder.WriteString(fmt.Sprintf("- **Total Return**: $%.2f (%.2f%%)\n", 
			request.PortfolioData.TotalReturn, request.PortfolioData.TotalReturnPct))
		
		if len(request.PortfolioData.Positions) > 0 {
			promptBuilder.WriteString("\n### Current Positions\n")
			for _, pos := range request.PortfolioData.Positions {
				promptBuilder.WriteString(fmt.Sprintf("- **%s**: $%.2f (%.2f%%)\n",
					pos.BasketName, pos.CurrentValue, pos.Weight*100))
			}
		}
	}
	
	// Analysis-specific instructions
	promptBuilder.WriteString("\n---\n\n")
	switch request.AnalysisType {
	case entities.AnalysisTypeDiversification:
		promptBuilder.WriteString("Please analyze the portfolio's diversification across:\n")
		promptBuilder.WriteString("- Asset classes and sectors\n- Geographic exposure\n- Risk concentration\n- Recommendations for improving diversification\n")
	case entities.AnalysisTypeRisk:
		promptBuilder.WriteString("Please provide a comprehensive risk analysis including:\n")
		promptBuilder.WriteString("- Current risk levels and exposure\n- Risk-adjusted performance metrics\n- Stress testing scenarios\n- Risk mitigation recommendations\n")
	case entities.AnalysisTypePerformance:
		promptBuilder.WriteString("Please analyze portfolio performance including:\n")
		promptBuilder.WriteString("- Performance attribution by holdings\n- Benchmark comparison\n- Historical performance trends\n- Performance optimization insights\n")
	case entities.AnalysisTypeAllocation:
		promptBuilder.WriteString("Please review portfolio allocation including:\n")
		promptBuilder.WriteString("- Current vs target allocation\n- Allocation efficiency analysis\n- Rebalancing opportunities\n- Strategic allocation recommendations\n")
	default:
		promptBuilder.WriteString("Please provide a comprehensive analysis of the requested aspect of the portfolio.\n")
	}
	
	promptBuilder.WriteString("\n**Important**: Provide actionable insights while avoiding specific trading recommendations. Include appropriate investment disclaimers.\n")
	
	return promptBuilder.String(), nil
}

// storeArtifact stores detailed analysis data as an artifact in 0G storage
func (g *InferenceGateway) storeArtifact(ctx context.Context, result *entities.InferenceResult, userID string, analysisType string) error {
	if g.storageClient == nil {
		return fmt.Errorf("storage client not available")
	}

	// Create detailed artifact data
	artifactData := map[string]interface{}{
		"request_id":      result.RequestID,
		"user_id":         userID,
		"analysis_type":   analysisType,
		"content":         result.Content,
		"metadata":        result.Metadata,
		"tokens_used":     result.TokensUsed,
		"processing_time": result.ProcessingTime.Seconds(),
		"model":           result.Model,
		"created_at":      result.CreatedAt.UTC().Format(time.RFC3339),
		"version":         "1.0",
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(artifactData)
	if err != nil {
		return fmt.Errorf("failed to serialize artifact data: %w", err)
	}

	// Store in 0G storage
	metadata := map[string]string{
		"content_type":   "application/json",
		"user_id":        userID,
		"analysis_type":  analysisType,
		"request_id":     result.RequestID,
		"created_at":     result.CreatedAt.UTC().Format(time.RFC3339),
	}

	storageResult, err := g.storageClient.Store(ctx, entities.NamespaceAIArtifacts, jsonData, metadata)
	if err != nil {
		return fmt.Errorf("failed to store artifact: %w", err)
	}

	// Update result with artifact URI
	result.ArtifactURI = storageResult.URI

	g.logger.Info("Artifact stored successfully",
		zap.String("request_id", result.RequestID),
		zap.String("artifact_uri", result.ArtifactURI),
		zap.Int64("size", storageResult.Size),
	)

	return nil
}

// generateMockWeeklySummary generates a mock weekly summary for development
func (g *InferenceGateway) generateMockWeeklySummary() string {
	return `# Weekly Portfolio Summary

## Executive Summary
Your portfolio showed solid performance this week with a **+2.4%** gain, outperforming the broader market. Strong technology positions drove most of the positive returns.

## Performance Analysis
- **Total Return**: +$1,240 (+2.4%)
- **Best Performer**: Tech Growth basket (+4.1%)
- **Underperformer**: Conservative Income (-0.3%)

### Key Contributors
- Technology sector strength boosted returns
- Growth positions benefited from market optimism
- Defensive positions provided stability during volatility

## Risk Assessment
Your portfolio maintains a **moderate risk profile** with good diversification:
- Volatility: 12.3% (within target range)
- Max Drawdown: -3.2% (acceptable)
- Diversification Score: 0.78/1.0 (well diversified)

## Market Context
- Technology rally continued on AI enthusiasm
- Interest rate concerns remained subdued
- Economic data showed resilient growth

## Key Observations
- Your risk-adjusted returns remain strong
- Technology allocation is paying off
- Conservative positions provided good downside protection

## Looking Ahead
- Monitor technology valuations for sustainability
- Consider rebalancing if tech allocation exceeds targets
- Economic data releases may create volatility

---
*This analysis is for informational purposes only and should not be considered as investment advice. All investments carry risk of loss.*`
}

// generateMockAnalysis generates a mock on-demand analysis for development
func (g *InferenceGateway) generateMockAnalysis() string {
	return `# Portfolio Analysis

## Current Assessment
Your portfolio demonstrates solid fundamentals with room for optimization in several key areas.

## Key Findings
- **Diversification**: Good spread across asset classes with minor concentration in growth stocks
- **Risk Level**: Moderate risk profile aligned with stated risk tolerance
- **Performance**: Above-average returns with reasonable volatility

## Recommendations
1. **Rebalancing Opportunity**: Consider reducing technology exposure by 2-3%
2. **Risk Management**: Current drawdown protection appears adequate
3. **Cost Efficiency**: Review expense ratios on underperforming positions

## Next Steps
- Monitor quarterly rebalancing triggers
- Review risk metrics monthly
- Consider tax-loss harvesting opportunities

---
*This analysis is for educational purposes only. Please consult with a financial advisor for personalized investment advice.*`
}

// recordError records error metrics
func (g *InferenceGateway) recordError(ctx context.Context, operation string, err error) {
	var errorCode string
	if zeroGErr, ok := err.(*entities.ZeroGError); ok {
		errorCode = zeroGErr.Code
	} else {
		errorCode = entities.ErrorCodeInternalError
	}

	g.metrics.RequestErrors.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("error_code", errorCode),
	))
}

// initInferenceMetrics initializes OpenTelemetry metrics for inference operations
func initInferenceMetrics(meter metric.Meter) (*InferenceMetrics, error) {
	requestsTotal, err := meter.Int64Counter("zerog_inference_requests_total",
		metric.WithDescription("Total number of inference requests"))
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Float64Histogram("zerog_inference_request_duration_seconds",
		metric.WithDescription("Duration of inference requests in seconds"))
	if err != nil {
		return nil, err
	}

	requestErrors, err := meter.Int64Counter("zerog_inference_request_errors_total",
		metric.WithDescription("Total number of inference request errors"))
	if err != nil {
		return nil, err
	}

	tokensUsed, err := meter.Int64Counter("zerog_inference_tokens_used_total",
		metric.WithDescription("Total tokens used for inference"))
	if err != nil {
		return nil, err
	}

	tokensGenerated, err := meter.Int64Counter("zerog_inference_tokens_generated_total",
		metric.WithDescription("Total tokens generated by inference"))
	if err != nil {
		return nil, err
	}

activeRequests, err := meter.Int64UpDownCounter("zerog_inference_active_requests",
		metric.WithDescription("Number of active inference requests"))
	if err != nil {
		return nil, err
	}

	modelUsage, err := meter.Int64Counter("zerog_inference_model_usage_total",
		metric.WithDescription("Total usage count by model"))
	if err != nil {
		return nil, err
	}

	return &InferenceMetrics{
		RequestsTotal:   requestsTotal,
		RequestDuration: requestDuration,
		RequestErrors:   requestErrors,
		TokensUsed:      tokensUsed,
		TokensGenerated: tokensGenerated,
		ActiveRequests:  activeRequests,
		ModelUsage:      modelUsage,
	}, nil
}