package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/stack-service/stack_service/internal/domain/entities"
)

// KYCProviderConfig holds KYC provider configuration
type KYCProviderConfig struct {
	APIKey      string
	APISecret   string
	BaseURL     string
	Environment string // "development", "staging", "production"
	CallbackURL string
	UserAgent   string
}

// KYCProvider implements the KYC provider interface
type KYCProvider struct {
	logger     *zap.Logger
	config     KYCProviderConfig
	httpClient *http.Client
	mockMode   bool
}

// NewKYCProvider creates a new KYC provider
func NewKYCProvider(logger *zap.Logger, config KYCProviderConfig) *KYCProvider {
	mockMode := config.Environment == "development" || config.APIKey == ""

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &KYCProvider{
		logger:     logger,
		config:     config,
		httpClient: httpClient,
		mockMode:   mockMode,
	}
}

// Jumio API models
type JumioInitiateRequest struct {
	CustomerInternalReference string `json:"customerInternalReference"`
	UserReference             string `json:"userReference"`
	WorkflowDefinition        struct {
		Key     string `json:"key"`
		Version string `json:"version,omitempty"`
	} `json:"workflowDefinition"`
	CallbackURL   string `json:"callbackUrl,omitempty"`
	SuccessURL    string `json:"successUrl,omitempty"`
	ErrorURL      string `json:"errorUrl,omitempty"`
	TokenLifetime string `json:"tokenLifetime,omitempty"`
}

type JumioInitiateResponse struct {
	Timestamp string `json:"timestamp"`
	Account   struct {
		ID string `json:"id"`
	} `json:"account"`
	WorkflowExecution struct {
		ID                        string `json:"id"`
		Status                    string `json:"status"`
		CustomerInternalReference string `json:"customerInternalReference"`
		UserReference             string `json:"userReference"`
	} `json:"workflowExecution"`
	Web struct {
		Href string `json:"href"`
	} `json:"web"`
	SDK struct {
		Token string `json:"token"`
	} `json:"sdk"`
}

type JumioStatusResponse struct {
	Timestamp string `json:"timestamp"`
	Account   struct {
		ID string `json:"id"`
	} `json:"account"`
	WorkflowExecution struct {
		ID                        string `json:"id"`
		Status                    string `json:"status"`
		CustomerInternalReference string `json:"customerInternalReference"`
		UserReference             string `json:"userReference"`
		DefinitionKey             string `json:"definitionKey"`
		Credentials               []struct {
			ID       string `json:"id"`
			Category string `json:"category"`
			Parts    []struct {
				Classifier string `json:"classifier"`
				Validity   string `json:"validity"`
			} `json:"parts"`
		} `json:"credentials"`
	} `json:"workflowExecution"`
}

type JumioErrorResponse struct {
	Timestamp string `json:"timestamp"`
	TraceID   string `json:"traceId"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Field   string `json:"field,omitempty"`
	} `json:"details,omitempty"`
}

func (e JumioErrorResponse) Error() string {
	return fmt.Sprintf("Jumio API error %s: %s", e.Code, e.Message)
}

// makeJumioRequest makes an HTTP request to Jumio API
func (k *KYCProvider) makeJumioRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, k.config.BaseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication headers
	req.SetBasicAuth(k.config.APIKey, k.config.APISecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", k.config.UserAgent)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// SubmitKYC submits KYC documents to the provider
func (k *KYCProvider) SubmitKYC(ctx context.Context, userID uuid.UUID, documents []entities.KYCDocument, personalInfo *entities.KYCPersonalInfo) (string, error) {
	k.logger.Info("Submitting KYC documents",
		zap.String("user_id", userID.String()),
		zap.Int("document_count", len(documents)))

	if k.mockMode {
		// Mock implementation for development
		providerRef := fmt.Sprintf("kyc_%s_%d", userID.String()[:8], time.Now().Unix())

		documentTypes := make([]string, len(documents))
		for i, doc := range documents {
			documentTypes[i] = string(doc.Type)
		}

		k.logger.Info("KYC submission created successfully (MOCK)",
			zap.String("user_id", userID.String()),
			zap.String("provider_ref", providerRef),
			zap.Strings("document_types", documentTypes))

		return providerRef, nil
	}

	// Real Jumio implementation
	request := JumioInitiateRequest{
		CustomerInternalReference: userID.String(),
		UserReference:             fmt.Sprintf("user_%s", userID.String()[:8]),
		CallbackURL:               k.config.CallbackURL,
		TokenLifetime:             "30m",
	}

	// Set workflow based on document types
	request.WorkflowDefinition.Key = "id_verification"
	if len(documents) > 1 {
		request.WorkflowDefinition.Key = "id_and_identity_verification"
	}

	resp, err := k.makeJumioRequest(ctx, "POST", "/api/v4/workflow/initiate", request)
	if err != nil {
		k.logger.Error("Failed to initiate Jumio workflow",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return "", fmt.Errorf("failed to initiate KYC workflow: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var jumioErr JumioErrorResponse
		if err := json.Unmarshal(respBody, &jumioErr); err == nil {
			k.logger.Error("Jumio API error",
				zap.String("user_id", userID.String()),
				zap.String("error_code", jumioErr.Code),
				zap.String("error_message", jumioErr.Message))
			return "", jumioErr
		}
		return "", fmt.Errorf("KYC API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var jumioResp JumioInitiateResponse
	if err := json.Unmarshal(respBody, &jumioResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	k.logger.Info("KYC workflow initiated successfully",
		zap.String("user_id", userID.String()),
		zap.String("workflow_id", jumioResp.WorkflowExecution.ID),
		zap.String("status", jumioResp.WorkflowExecution.Status))

	return jumioResp.WorkflowExecution.ID, nil
}

// GetKYCStatus retrieves the current KYC status from the provider
func (k *KYCProvider) GetKYCStatus(ctx context.Context, providerRef string) (*entities.KYCSubmission, error) {
	k.logger.Info("Getting KYC status from provider",
		zap.String("provider_ref", providerRef))

	if k.mockMode {
		// Mock response for development
		submission := &entities.KYCSubmission{
			ProviderRef: providerRef,
			Status:      entities.KYCStatusProcessing,
			SubmittedAt: time.Now().Add(-1 * time.Hour),
		}

		k.logger.Info("KYC status retrieved successfully (MOCK)",
			zap.String("provider_ref", providerRef),
			zap.String("status", string(submission.Status)))

		return submission, nil
	}

	// Real Jumio implementation
	endpoint := fmt.Sprintf("/api/v4/workflow/executions/%s", providerRef)
	resp, err := k.makeJumioRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		k.logger.Error("Failed to get KYC status from Jumio",
			zap.String("provider_ref", providerRef),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get KYC status: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var jumioErr JumioErrorResponse
		if err := json.Unmarshal(respBody, &jumioErr); err == nil {
			k.logger.Error("Jumio API error",
				zap.String("provider_ref", providerRef),
				zap.String("error_code", jumioErr.Code),
				zap.String("error_message", jumioErr.Message))
			return nil, jumioErr
		}
		return nil, fmt.Errorf("KYC API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var jumioResp JumioStatusResponse
	if err := json.Unmarshal(respBody, &jumioResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Map Jumio status to our internal status
	var status entities.KYCStatus
	var rejectionReasons []string

	switch jumioResp.WorkflowExecution.Status {
	case "PENDING":
		status = entities.KYCStatusPending
	case "PROCESSING":
		status = entities.KYCStatusProcessing
	case "PASSED":
		status = entities.KYCStatusApproved
	case "FAILED", "REJECTED":
		status = entities.KYCStatusRejected
		// Extract rejection reasons from credentials
		for _, cred := range jumioResp.WorkflowExecution.Credentials {
			for _, part := range cred.Parts {
				if part.Validity != "PASSED" {
					rejectionReasons = append(rejectionReasons, fmt.Sprintf("%s: %s", cred.Category, part.Classifier))
				}
			}
		}
	default:
		status = entities.KYCStatusPending
	}

	submission := &entities.KYCSubmission{
		ProviderRef:      providerRef,
		Status:           status,
		RejectionReasons: rejectionReasons,
		SubmittedAt:      time.Now(), // In real implementation, parse from response
	}

	k.logger.Info("KYC status retrieved successfully",
		zap.String("provider_ref", providerRef),
		zap.String("status", string(submission.Status)),
		zap.Strings("rejection_reasons", rejectionReasons))

	return submission, nil
}

// GenerateKYCURL generates a URL for users to complete KYC verification
func (k *KYCProvider) GenerateKYCURL(ctx context.Context, userID uuid.UUID) (string, error) {
	k.logger.Info("Generating KYC URL",
		zap.String("user_id", userID.String()))

	if k.mockMode {
		// Generate a mock KYC URL for development
		kycURL := fmt.Sprintf("https://mock-kyc-provider.com/verify?user_id=%s&session=%s",
			userID.String(),
			uuid.New().String()[:16])

		k.logger.Info("KYC URL generated successfully (MOCK)",
			zap.String("user_id", userID.String()),
			zap.String("kyc_url", kycURL))

		return kycURL, nil
	}

	// Real implementation: First initiate a workflow, then return the web URL
	request := JumioInitiateRequest{
		CustomerInternalReference: userID.String(),
		UserReference:             fmt.Sprintf("user_%s", userID.String()[:8]),
		CallbackURL:               k.config.CallbackURL,
		TokenLifetime:             "30m",
	}
	request.WorkflowDefinition.Key = "id_verification"

	resp, err := k.makeJumioRequest(ctx, "POST", "/api/v4/workflow/initiate", request)
	if err != nil {
		k.logger.Error("Failed to initiate Jumio workflow for URL generation",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return "", fmt.Errorf("failed to generate KYC URL: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var jumioErr JumioErrorResponse
		if err := json.Unmarshal(respBody, &jumioErr); err == nil {
			k.logger.Error("Jumio API error during URL generation",
				zap.String("user_id", userID.String()),
				zap.String("error_code", jumioErr.Code),
				zap.String("error_message", jumioErr.Message))
			return "", jumioErr
		}
		return "", fmt.Errorf("KYC API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var jumioResp JumioInitiateResponse
	if err := json.Unmarshal(respBody, &jumioResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	k.logger.Info("KYC URL generated successfully",
		zap.String("user_id", userID.String()),
		zap.String("workflow_id", jumioResp.WorkflowExecution.ID),
		zap.String("kyc_url", jumioResp.Web.Href))

	return jumioResp.Web.Href, nil
}
