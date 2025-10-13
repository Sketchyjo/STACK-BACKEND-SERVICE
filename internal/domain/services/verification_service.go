package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"go.uber.org/zap"

	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/internal/infrastructure/adapters"
	"github.com/stack-service/stack_service/internal/infrastructure/cache"
	"github.com/stack-service/stack_service/internal/infrastructure/config"
)

const (
	verificationCodeLength  = 6
	verificationCodeTTL     = 10 * time.Minute
	maxVerificationAttempts = 3
	rateLimitWindow         = 1 * time.Minute
	maxSendAttempts         = 5
)

// VerificationService defines the interface for managing verification codes
type VerificationService interface {
	GenerateAndSendCode(ctx context.Context, identifierType, identifier string) (string, error)
	VerifyCode(ctx context.Context, identifierType, identifier, code string) (bool, error)
	CanResendCode(ctx context.Context, identifierType, identifier string) (bool, error)
	RecordSendAttempt(ctx context.Context, identifierType, identifier string) error
}

// verificationService implements VerificationService
type verificationService struct {
	redisClient  cache.RedisClient
	emailService adapters.EmailService
	smsService   adapters.SMSService
	logger       *zap.Logger
	config       *config.Config
}

// NewVerificationService creates a new VerificationService
func NewVerificationService(
	redisClient cache.RedisClient,
	emailService adapters.EmailService,
	smsService adapters.SMSService,
	logger *zap.Logger,
	cfg *config.Config,
) VerificationService {
	return &verificationService{
		redisClient:  redisClient,
		emailService: emailService,
		smsService:   smsService,
		logger:       logger,
		config:       cfg,
	}
}

// GenerateAndSendCode generates a 6-digit code, stores it in Redis, and sends it via email or SMS
func (s *verificationService) GenerateAndSendCode(ctx context.Context, identifierType, identifier string) (string, error) {
	// Check rate limit for sending codes
	sendAttemptsKey := fmt.Sprintf("send_attempts:%s:%s", identifierType, identifier)
	sendAttempts, err := s.redisClient.Incr(ctx, sendAttemptsKey)
	if err != nil {
		s.logger.Error("Failed to increment send attempts counter", zap.Error(err), zap.String("key", sendAttemptsKey))
		return "", fmt.Errorf("failed to check send rate limit: %w", err)
	}
	if sendAttempts == 1 {
		// Set expiration for the first attempt in the window
		if err := s.redisClient.Expire(ctx, sendAttemptsKey, rateLimitWindow); err != nil {
			s.logger.Error("Failed to set expiration for send attempts counter", zap.Error(err), zap.String("key", sendAttemptsKey))
		}
	}
	if sendAttempts > maxSendAttempts {
		s.logger.Warn("Rate limit exceeded for sending verification code", zap.String("identifier", identifier))
		return "", fmt.Errorf("too many verification code send attempts. Please try again after %s", rateLimitWindow.String())
	}

	code, err := generateNumericCode(verificationCodeLength)
	if err != nil {
		s.logger.Error("Failed to generate verification code", zap.Error(err))
		return "", fmt.Errorf("failed to generate verification code: %w", err)
	}

	verificationData := entities.VerificationCodeData{
		Code:      code,
		Attempts:  0,
		ExpiresAt: time.Now().Add(verificationCodeTTL),
		CreatedAt: time.Now(),
	}

	key := fmt.Sprintf("verification:%s:%s", identifierType, identifier)
	if err := s.redisClient.Set(ctx, key, verificationData, verificationCodeTTL); err != nil {
		s.logger.Error("Failed to store verification code in Redis", zap.Error(err), zap.String("key", key))
		return "", fmt.Errorf("failed to store verification code: %w", err)
	}

	var sendErr error
	if identifierType == "email" {
		sendErr = s.emailService.SendVerificationEmail(ctx, identifier, code)
	} else if identifierType == "phone" {
		sendErr = s.smsService.SendVerificationSMS(ctx, identifier, code)
	} else {
		return "", fmt.Errorf("unsupported identifier type: %s", identifierType)
	}

	if sendErr != nil {
		s.logger.Error("Failed to send verification code", zap.Error(sendErr), zap.String("identifier", identifier))
		return "", fmt.Errorf("failed to send verification code: %w", sendErr)
	}

	s.logger.Info("Verification code generated and sent", zap.String("identifier", identifier), zap.String("code", code))
	return code, nil
}

// VerifyCode validates the provided code against the stored one
func (s *verificationService) VerifyCode(ctx context.Context, identifierType, identifier, code string) (bool, error) {
	key := fmt.Sprintf("verification:%s:%s", identifierType, identifier)
	var storedData entities.VerificationCodeData
	err := s.redisClient.Get(ctx, key, &storedData)
	if err != nil {
		if err.Error() == fmt.Sprintf("key '%s' not found: redis: nil", key) { // Specific check for redis.Nil
			s.logger.Warn("Verification code not found or expired", zap.String("identifier", identifier))
			return false, fmt.Errorf("verification code not found or expired")
		}
		s.logger.Error("Failed to retrieve verification code from Redis", zap.Error(err), zap.String("key", key))
		return false, fmt.Errorf("failed to retrieve verification code: %w", err)
	}

	// Increment attempt count
	storedData.Attempts++
	if err := s.redisClient.Set(ctx, key, storedData, storedData.ExpiresAt.Sub(time.Now())); err != nil {
		s.logger.Error("Failed to update verification code attempts in Redis", zap.Error(err), zap.String("key", key))
		// Non-critical error, continue with verification
	}

	if storedData.Attempts > maxVerificationAttempts {
		s.logger.Warn("Too many verification attempts for code", zap.String("identifier", identifier))
		s.redisClient.Del(ctx, key) // Invalidate code after too many attempts
		return false, fmt.Errorf("too many verification attempts. Please request a new code")
	}

	if storedData.Code != code {
		s.logger.Warn("Invalid verification code provided", zap.String("identifier", identifier), zap.Int("attempts", storedData.Attempts))
		return false, fmt.Errorf("invalid verification code")
	}

	// Code is valid, delete it from Redis
	if err := s.redisClient.Del(ctx, key); err != nil {
		s.logger.Error("Failed to delete verification code from Redis after successful verification", zap.Error(err), zap.String("key", key))
		// Non-critical error, but log it
	}

	s.logger.Info("Verification successful", zap.String("identifier", identifier))
	return true, nil
}

// CanResendCode checks if a new verification code can be sent based on rate limits
func (s *verificationService) CanResendCode(ctx context.Context, identifierType, identifier string) (bool, error) {
	sendAttemptsKey := fmt.Sprintf("send_attempts:%s:%s", identifierType, identifier)
	var sendAttemptsStr string
	err := s.redisClient.Get(ctx, sendAttemptsKey, &sendAttemptsStr) // Get as string to check existence
	if err != nil && err.Error() != fmt.Sprintf("key '%s' not found: redis: nil", sendAttemptsKey) {
		s.logger.Error("Failed to check send attempts counter", zap.Error(err), zap.String("key", sendAttemptsKey))
		return false, fmt.Errorf("failed to check resend eligibility: %w", err)
	}

	var currentAttempts int64
	if sendAttemptsStr != "" {
		fmt.Sscanf(sendAttemptsStr, "%d", &currentAttempts)
	}

	return currentAttempts < maxSendAttempts, nil
}

// RecordSendAttempt records a send attempt for rate limiting
func (s *verificationService) RecordSendAttempt(ctx context.Context, identifierType, identifier string) error {
	sendAttemptsKey := fmt.Sprintf("send_attempts:%s:%s", identifierType, identifier)
	sendAttempts, err := s.redisClient.Incr(ctx, sendAttemptsKey)
	if err != nil {
		s.logger.Error("Failed to increment send attempts counter", zap.Error(err), zap.String("key", sendAttemptsKey))
		return fmt.Errorf("failed to record send attempt: %w", err)
	}
	if sendAttempts == 1 {
		// Set expiration for the first attempt in the window
		if err := s.redisClient.Expire(ctx, sendAttemptsKey, rateLimitWindow); err != nil {
			s.logger.Error("Failed to set expiration for send attempts counter", zap.Error(err), zap.String("key", sendAttemptsKey))
		}
	}
	return nil
}

// generateNumericCode generates a random numeric string of specified length
func generateNumericCode(length int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		b[i] = digits[num.Int64()]
	}
	return string(b), nil
}
