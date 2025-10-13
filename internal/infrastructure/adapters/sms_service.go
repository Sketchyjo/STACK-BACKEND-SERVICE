package adapters

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SMSConfig holds SMS service configuration
type SMSConfig struct {
	Provider    string // "twilio", "mock"
	APIKey      string
	APISecret   string
	FromNumber  string
	Environment string // "development", "staging", "production"
}

// SMSService implements SMS delivery interface
type SMSService struct {
	logger   *zap.Logger
	config   SMSConfig
	mockMode bool
}

// NewSMSService creates a new SMS service
func NewSMSService(logger *zap.Logger, config SMSConfig) *SMSService {
	mockMode := config.Environment == "development" || config.APIKey == ""

	return &SMSService{
		logger:   logger,
		config:   config,
		mockMode: mockMode,
	}
}

// SendVerificationSMS sends a verification code via SMS
func (s *SMSService) SendVerificationSMS(ctx context.Context, phone, code string) error {
	s.logger.Info("Sending verification SMS",
		zap.String("phone", s.maskPhone(phone)),
		zap.String("code", code))

	if s.mockMode {
		s.logger.Info("SMS sent successfully (MOCK)",
			zap.String("to", s.maskPhone(phone)),
			zap.String("message", fmt.Sprintf("Your Stack verification code is: %s", code)))
		return nil
	}

	// TODO: Implement Twilio integration
	// For now, we'll use mock mode in development
	s.logger.Warn("SMS service not implemented - using mock mode",
		zap.String("provider", s.config.Provider),
		zap.String("phone", s.maskPhone(phone)))

	return nil
}

// SendWelcomeSMS sends a welcome message via SMS
func (s *SMSService) SendWelcomeSMS(ctx context.Context, phone string) error {
	s.logger.Info("Sending welcome SMS",
		zap.String("phone", s.maskPhone(phone)))

	if s.mockMode {
		s.logger.Info("Welcome SMS sent successfully (MOCK)",
			zap.String("to", s.maskPhone(phone)),
			zap.String("message", "Welcome to Stack! Your account is now active."))
		return nil
	}

	// TODO: Implement Twilio integration
	s.logger.Warn("SMS service not implemented - using mock mode",
		zap.String("provider", s.config.Provider),
		zap.String("phone", s.maskPhone(phone)))

	return nil
}

// SendKYCSMS sends KYC status update via SMS
func (s *SMSService) SendKYCSMS(ctx context.Context, phone, status string) error {
	s.logger.Info("Sending KYC status SMS",
		zap.String("phone", s.maskPhone(phone)),
		zap.String("status", status))

	if s.mockMode {
		var message string
		switch status {
		case "approved":
			message = "Great news! Your KYC verification has been approved. You can now start investing."
		case "rejected":
			message = "Your KYC verification needs additional information. Please check your email for details."
		case "processing":
			message = "Your KYC verification is being processed. We'll notify you once it's complete."
		default:
			message = fmt.Sprintf("Your KYC status has been updated to: %s", status)
		}

		s.logger.Info("KYC SMS sent successfully (MOCK)",
			zap.String("to", s.maskPhone(phone)),
			zap.String("message", message))
		return nil
	}

	// TODO: Implement Twilio integration
	s.logger.Warn("SMS service not implemented - using mock mode",
		zap.String("provider", s.config.Provider),
		zap.String("phone", s.maskPhone(phone)))

	return nil
}

// ValidatePhoneNumber validates phone number format (E.164)
func (s *SMSService) ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return fmt.Errorf("phone number is required")
	}

	// Basic E.164 validation: starts with +, followed by 7-15 digits
	if len(phone) < 8 || len(phone) > 16 {
		return fmt.Errorf("phone number must be 8-16 characters")
	}

	if phone[0] != '+' {
		return fmt.Errorf("phone number must start with +")
	}

	// Check if remaining characters are digits
	for i := 1; i < len(phone); i++ {
		if phone[i] < '0' || phone[i] > '9' {
			return fmt.Errorf("phone number must contain only digits after +")
		}
	}

	return nil
}

// NormalizePhoneNumber normalizes phone number to E.164 format
func (s *SMSService) NormalizePhoneNumber(phone string) string {
	// Remove all non-digit characters except +
	normalized := "+"
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			normalized += string(char)
		}
	}

	// If no + was found, add it
	if normalized == "+" {
		normalized = "+" + phone
	}

	return normalized
}

// maskPhone masks phone number for logging (e.g., +1234567890 -> +123****890)
func (s *SMSService) maskPhone(phone string) string {
	if len(phone) < 7 {
		return "****"
	}

	if len(phone) <= 4 {
		return phone[:2] + "****"
	}

	return phone[:3] + "****" + phone[len(phone)-3:]
}

// HealthCheck checks SMS service health
func (s *SMSService) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if s.mockMode {
		s.logger.Debug("SMS service health check (mock mode)")
		return nil
	}

	// TODO: Implement actual health check for Twilio
	s.logger.Debug("SMS service health check (not implemented)")
	return nil
}
