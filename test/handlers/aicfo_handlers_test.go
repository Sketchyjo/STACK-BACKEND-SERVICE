package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stack-service/stack_service/internal/api/handlers"
	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/internal/domain/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// MockAICfoService is a mock implementation of the AI-CFO service
type MockAICfoService struct {
	mock.Mock
}

func (m *MockAICfoService) GenerateWeeklySummary(ctx context.Context, userID uuid.UUID, weekStart time.Time) (*services.AISummary, error) {
	args := m.Called(ctx, userID, weekStart)
	return args.Get(0).(*services.AISummary), args.Error(1)
}

func (m *MockAICfoService) PerformOnDemandAnalysis(ctx context.Context, userID uuid.UUID, analysisType string, parameters map[string]interface{}) (*entities.InferenceResult, error) {
	args := m.Called(ctx, userID, analysisType, parameters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.InferenceResult), args.Error(1)
}

func (m *MockAICfoService) GetLatestSummary(ctx context.Context, userID uuid.UUID) (*services.AISummary, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AISummary), args.Error(1)
}

func (m *MockAICfoService) GetHealthStatus(ctx context.Context) (*entities.HealthStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.HealthStatus), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func setupAICfoHandler(mockService *MockAICfoService) *handlers.AICfoHandler {
	logger := zaptest.NewLogger(nil)
	return handlers.NewAICfoHandler(mockService, logger)
}

func TestAICfoHandler_GetLatestSummary(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		mockSetup      func(*MockAICfoService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful summary retrieval",
			userID: uuid.New(),
			mockSetup: func(m *MockAICfoService) {
				summary := &services.AISummary{
					ID:          uuid.New(),
					UserID:      uuid.New(),
					WeekStart:   time.Now().Truncate(24 * time.Hour),
					SummaryMD:   "# Weekly Summary\n\nYour portfolio performed well this week.",
					CreatedAt:   time.Now(),
					ArtifactURI: "0g://ai-summaries/abc123",
				}
				m.On("GetLatestSummary", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(summary, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "no summaries found",
			userID: uuid.New(),
			mockSetup: func(m *MockAICfoService) {
				m.On("GetLatestSummary", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, fmt.Errorf("no summaries found for user"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:   "service error",
			userID: uuid.New(),
			mockSetup: func(m *MockAICfoService) {
				m.On("GetLatestSummary", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, fmt.Errorf("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAICfoService{}
			handler := setupAICfoHandler(mockService)
			router := setupTestRouter()
			
			// Apply mock setup
			tt.mockSetup(mockService)
			
			// Setup route
			router.GET("/summary/latest", func(c *gin.Context) {
				// Simulate auth middleware setting user_id
				c.Set("user_id", tt.userID)
				handler.GetLatestSummary(c)
			})
			
			// Make request
			req := httptest.NewRequest(http.MethodGet, "/summary/latest", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// Assert response
			if tt.expectedError != "" {
				var errorResp handlers.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResp)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResp.Code)
			} else {
				var summaryResp handlers.SummaryResponse
				err := json.Unmarshal(w.Body.Bytes(), &summaryResp)
				require.NoError(t, err)
				assert.NotEmpty(t, summaryResp.ID)
				assert.NotEmpty(t, summaryResp.Content)
				assert.NotEmpty(t, summaryResp.Title)
			}
			
			// Verify mocks
			mockService.AssertExpectations(t)
		})
	}
}

func TestAICfoHandler_GetLatestSummary_Unauthorized(t *testing.T) {
	mockService := &MockAICfoService{}
	handler := setupAICfoHandler(mockService)
	router := setupTestRouter()
	
	// Setup route without setting user_id (simulating missing auth)
	router.GET("/summary/latest", handler.GetLatestSummary)
	
	// Make request
	req := httptest.NewRequest(http.MethodGet, "/summary/latest", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var errorResp handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", errorResp.Code)
}

func TestAICfoHandler_AnalyzeOnDemand(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		request        handlers.AnalysisRequest
		mockSetup      func(*MockAICfoService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful risk analysis",
			userID: uuid.New(),
			request: handlers.AnalysisRequest{
				AnalysisType: entities.AnalysisTypeRisk,
				Parameters: map[string]interface{}{
					"timeframe": "30d",
				},
			},
			mockSetup: func(m *MockAICfoService) {
				result := &entities.InferenceResult{
					RequestID:      uuid.New().String(),
					Content:        "Your portfolio shows moderate risk exposure...",
					ContentType:    "text/markdown",
					TokensUsed:     150,
					ProcessingTime: 2 * time.Second,
					CreatedAt:      time.Now(),
					ArtifactURI:    "0g://ai-artifacts/analysis-123",
					Metadata: map[string]interface{}{
						"analysis_type": entities.AnalysisTypeRisk,
						"timeframe":     "30d",
					},
				}
				m.On("PerformOnDemandAnalysis", mock.Anything, mock.AnythingOfType("uuid.UUID"), 
					entities.AnalysisTypeRisk, mock.AnythingOfType("map[string]interface {}")).
					Return(result, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid analysis type",
			userID: uuid.New(),
			request: handlers.AnalysisRequest{
				AnalysisType: "invalid_type",
			},
			mockSetup: func(m *MockAICfoService) {
				// No mock setup needed as validation happens before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ANALYSIS_TYPE",
		},
		{
			name:   "analysis service error",
			userID: uuid.New(),
			request: handlers.AnalysisRequest{
				AnalysisType: entities.AnalysisTypePerformance,
			},
			mockSetup: func(m *MockAICfoService) {
				m.On("PerformOnDemandAnalysis", mock.Anything, mock.AnythingOfType("uuid.UUID"), 
					entities.AnalysisTypePerformance, mock.AnythingOfType("map[string]interface {}")).
					Return(nil, fmt.Errorf("inference service unavailable"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "ANALYSIS_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAICfoService{}
			handler := setupAICfoHandler(mockService)
			router := setupTestRouter()
			
			// Apply mock setup
			tt.mockSetup(mockService)
			
			// Setup route
			router.POST("/analyze", func(c *gin.Context) {
				// Simulate auth middleware setting user_id
				c.Set("user_id", tt.userID)
				handler.AnalyzeOnDemand(c)
			})
			
			// Prepare request body
			requestBody, err := json.Marshal(tt.request)
			require.NoError(t, err)
			
			// Make request
			req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// Assert response
			if tt.expectedError != "" {
				var errorResp handlers.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResp)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResp.Code)
			} else {
				var analysisResp handlers.AnalysisResponse
				err := json.Unmarshal(w.Body.Bytes(), &analysisResp)
				require.NoError(t, err)
				assert.NotEmpty(t, analysisResp.RequestID)
				assert.Equal(t, tt.request.AnalysisType, analysisResp.AnalysisType)
				assert.NotEmpty(t, analysisResp.Content)
				assert.Greater(t, len(analysisResp.Insights), 0)
				assert.Greater(t, len(analysisResp.Recommendations), 0)
			}
			
			// Verify mocks
			mockService.AssertExpectations(t)
		})
	}
}

func TestAICfoHandler_AnalyzeOnDemand_InvalidJSON(t *testing.T) {
	mockService := &MockAICfoService{}
	handler := setupAICfoHandler(mockService)
	router := setupTestRouter()
	
	router.POST("/analyze", func(c *gin.Context) {
		c.Set("user_id", uuid.New())
		handler.AnalyzeOnDemand(c)
	})
	
	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var errorResp handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_REQUEST", errorResp.Code)
}

func TestAICfoHandler_HealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockAICfoService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "healthy service",
			mockSetup: func(m *MockAICfoService) {
				health := &entities.HealthStatus{
					Status:  "healthy",
					Message: "All systems operational",
					Timestamp: time.Now(),
					Services: map[string]entities.ServiceHealth{
						"storage": {
							Status:      "healthy",
							Message:     "Connected",
							LastChecked: time.Now(),
						},
						"inference": {
							Status:      "healthy",
							Message:     "Available",
							LastChecked: time.Now(),
						},
					},
				}
				m.On("GetHealthStatus", mock.Anything).Return(health, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unhealthy service",
			mockSetup: func(m *MockAICfoService) {
				m.On("GetHealthStatus", mock.Anything).
					Return(nil, fmt.Errorf("health check failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "HEALTH_CHECK_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAICfoService{}
			handler := setupAICfoHandler(mockService)
			router := setupTestRouter()
			
			// Apply mock setup
			tt.mockSetup(mockService)
			
			// Setup route
			router.GET("/health", handler.HealthCheck)
			
			// Make request
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// Assert response
			if tt.expectedError != "" {
				var errorResp handlers.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResp)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResp.Code)
			} else {
				var healthResp entities.HealthStatus
				err := json.Unmarshal(w.Body.Bytes(), &healthResp)
				require.NoError(t, err)
				assert.NotEmpty(t, healthResp.Status)
			}
			
			// Verify mocks
			mockService.AssertExpectations(t)
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkAICfoHandler_GetLatestSummary(b *testing.B) {
	mockService := &MockAICfoService{}
	handler := setupAICfoHandler(mockService)
	router := setupTestRouter()
	
	// Setup mock
	summary := &services.AISummary{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		WeekStart:   time.Now().Truncate(24 * time.Hour),
		SummaryMD:   "# Weekly Summary\n\nBenchmark summary content",
		CreatedAt:   time.Now(),
		ArtifactURI: "0g://ai-summaries/benchmark",
	}
	mockService.On("GetLatestSummary", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return(summary, nil)
	
	// Setup route
	router.GET("/summary/latest", func(c *gin.Context) {
		c.Set("user_id", uuid.New())
		handler.GetLatestSummary(c)
	})
	
	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/summary/latest", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkAICfoHandler_AnalyzeOnDemand(b *testing.B) {
	mockService := &MockAICfoService{}
	handler := setupAICfoHandler(mockService)
	router := setupTestRouter()
	
	// Setup mock
	result := &entities.InferenceResult{
		RequestID:      uuid.New().String(),
		Content:        "Benchmark analysis content",
		ContentType:    "text/markdown",
		TokensUsed:     100,
		ProcessingTime: time.Second,
		CreatedAt:      time.Now(),
		ArtifactURI:    "0g://ai-artifacts/benchmark",
		Metadata:       map[string]interface{}{},
	}
	mockService.On("PerformOnDemandAnalysis", mock.Anything, mock.AnythingOfType("uuid.UUID"), 
		mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).
		Return(result, nil)
	
	// Setup route
	router.POST("/analyze", func(c *gin.Context) {
		c.Set("user_id", uuid.New())
		handler.AnalyzeOnDemand(c)
	})
	
	// Prepare request
	request := handlers.AnalysisRequest{
		AnalysisType: entities.AnalysisTypeRisk,
	}
	requestBody, _ := json.Marshal(request)
	
	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}