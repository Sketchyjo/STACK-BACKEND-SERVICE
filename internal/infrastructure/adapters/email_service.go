package adapters

import (
	"context"
	"fmt"
	// "net/http"
	// "os"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"

	"github.com/stack-service/stack_service/internal/domain/entities"
)

// EmailServiceConfig holds email service configuration
type EmailServiceConfig struct {
	APIKey      string
	FromEmail   string
	FromName    string
	Environment string // "development", "staging", "production"
	BaseURL     string // For verification links
}

// EmailService implements the email service interface
type EmailService struct {
	logger   *zap.Logger
	config   EmailServiceConfig
	client   *sendgrid.Client
	mockMode bool // Set to true in development/testing
}

// NewEmailService creates a new email service
func NewEmailService(logger *zap.Logger, config EmailServiceConfig) *EmailService {
	mockMode := config.Environment == "development" || config.APIKey == ""

	var client *sendgrid.Client
	if !mockMode {
		client = sendgrid.NewSendClient(config.APIKey)
	}

	return &EmailService{
		logger:   logger,
		config:   config,
		client:   client,
		mockMode: mockMode,
	}
}

// sendEmail is a helper method to send emails via SendGrid or mock
func (e *EmailService) sendEmail(ctx context.Context, to, subject, htmlContent, textContent string) error {
	if e.mockMode {
		e.logger.Info("Email sent successfully (MOCK)",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.String("content_preview", textContent[:min(100, len(textContent))]+"..."))
		return nil
	}

	from := mail.NewEmail(e.config.FromName, e.config.FromEmail)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, textContent, htmlContent)

	// Add timeout to context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	response, err := e.client.SendWithContext(ctxWithTimeout, message)
	if err != nil {
		e.logger.Error("Failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err))
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		e.logger.Error("Email service returned error",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Int("status_code", response.StatusCode),
			zap.String("response_body", response.Body))
		return fmt.Errorf("email service error: status %d, body: %s", response.StatusCode, response.Body)
	}

	e.logger.Info("Email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.Int("status_code", response.StatusCode))

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SendVerificationEmail sends an email verification message
func (e *EmailService) SendVerificationEmail(ctx context.Context, email, verificationToken string) error {
	e.logger.Info("Sending verification email",
		zap.String("email", email),
		zap.String("token", verificationToken))

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", e.config.BaseURL, verificationToken)

	subject := "Verify Your Email Address - Stack Service"

	htmlContent := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Email Verification</title>
		</head>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
			<div style="background-color: #f8f9fa; padding: 30px; border-radius: 8px; text-align: center;">
				<h1 style="color: #333; margin-bottom: 20px;">Welcome to Stack Service!</h1>
				<p style="color: #666; font-size: 16px; line-height: 1.5; margin-bottom: 30px;">
					Thank you for joining Stack Service. To complete your registration and secure your account,
					please verify your email address by clicking the button below.
				</p>
				<a href="%s" 
				   style="display: inline-block; background-color: #007bff; color: white; padding: 15px 30px; 
				          text-decoration: none; border-radius: 5px; font-weight: bold; margin-bottom: 20px;">
					Verify Email Address
				</a>
				<p style="color: #888; font-size: 14px; margin-top: 30px;">
					If you cannot click the button, copy and paste this link into your browser:<br>
					<a href="%s" style="color: #007bff; word-break: break-all;">%s</a>
				</p>
				<p style="color: #888; font-size: 12px; margin-top: 20px;">
					This link will expire in 24 hours. If you did not create an account with Stack Service, 
					please ignore this email.
				</p>
			</div>
		</body>
		</html>
	`, verificationURL, verificationURL, verificationURL)

	textContent := fmt.Sprintf(`
Welcome to Stack Service!

Thank you for joining Stack Service. To complete your registration and secure your account,
please verify your email address by visiting the following link:

%s

This link will expire in 24 hours. If you did not create an account with Stack Service,
please ignore this email.

Best regards,
The Stack Service Team
	`, verificationURL)

	return e.sendEmail(ctx, email, subject, htmlContent, textContent)
}

// SendKYCStatusEmail sends a KYC status update email
func (e *EmailService) SendKYCStatusEmail(ctx context.Context, email string, status entities.KYCStatus, rejectionReasons []string) error {
	e.logger.Info("Sending KYC status email",
		zap.String("email", email),
		zap.String("status", string(status)),
		zap.Strings("rejection_reasons", rejectionReasons))

	var subject, htmlContent, textContent string

	switch status {
	case entities.KYCStatusApproved:
		subject = "✅ KYC Verification Approved - Stack Service"
		htmlContent = e.buildKYCApprovedHTML()
		textContent = e.buildKYCApprovedText()

	case entities.KYCStatusRejected:
		subject = "❌ KYC Verification Requires Additional Information - Stack Service"
		htmlContent = e.buildKYCRejectedHTML(rejectionReasons)
		textContent = e.buildKYCRejectedText(rejectionReasons)

	case entities.KYCStatusProcessing:
		subject = "⏳ KYC Verification In Progress - Stack Service"
		htmlContent = e.buildKYCProcessingHTML()
		textContent = e.buildKYCProcessingText()

	default:
		subject = "KYC Status Update - Stack Service"
		htmlContent = e.buildKYCGenericHTML(string(status))
		textContent = e.buildKYCGenericText(string(status))
	}

	return e.sendEmail(ctx, email, subject, htmlContent, textContent)
}

// SendWelcomeEmail sends a welcome email to a new user
func (e *EmailService) SendWelcomeEmail(ctx context.Context, email string) error {
	e.logger.Info("Sending welcome email",
		zap.String("email", email))

	subject := "🎉 Welcome to Stack Service!"

	htmlContent := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><title>Welcome to Stack Service</title></head>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
			<div style="background-color: #f8f9fa; padding: 30px; border-radius: 8px; text-align: center;">
				<h1 style="color: #333; margin-bottom: 20px;">Welcome to Stack Service! 🎉</h1>
				<p style="color: #666; font-size: 16px; line-height: 1.5; margin-bottom: 30px;">
					Thank you for joining Stack Service! We're excited to have you on board.
					Your account has been successfully created and verified.
				</p>
				<div style="background-color: white; padding: 20px; border-radius: 8px; margin: 20px 0;">
					<h3 style="color: #333; margin-bottom: 15px;">Next Steps:</h3>
					<ul style="text-align: left; color: #666; line-height: 1.8;">
						<li>Complete your KYC verification</li>
						<li>Set up your digital wallets</li>
						<li>Start exploring our platform</li>
					</ul>
				</div>
				<a href="%s/dashboard" 
				   style="display: inline-block; background-color: #28a745; color: white; padding: 15px 30px; 
				          text-decoration: none; border-radius: 5px; font-weight: bold; margin: 20px 0;">
					Get Started
				</a>
				<p style="color: #888; font-size: 12px; margin-top: 30px;">
					If you have any questions, feel free to contact our support team.
				</p>
			</div>
		</body>
		</html>
	`, e.config.BaseURL)

	textContent := fmt.Sprintf(`
Welcome to Stack Service!

Thank you for joining Stack Service! We're excited to have you on board.
Your account has been successfully created and verified.

Next Steps:
- Complete your KYC verification
- Set up your digital wallets  
- Start exploring our platform

Get started by visiting: %s/dashboard

If you have any questions, feel free to contact our support team.

Best regards,
The Stack Service Team
	`, e.config.BaseURL)

	return e.sendEmail(ctx, email, subject, htmlContent, textContent)
}

// KYC Email Templates

func (e *EmailService) buildKYCApprovedHTML() string {
	return fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head><title>KYC Approved</title></head>
	<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
		<div style="background-color: #d4edda; padding: 30px; border-radius: 8px; text-align: center; border: 1px solid #c3e6cb;">
			<h1 style="color: #155724; margin-bottom: 20px;">✅ Verification Complete!</h1>
			<p style="color: #155724; font-size: 16px; line-height: 1.5; margin-bottom: 30px;">
				Congratulations! Your identity verification has been successfully approved.
				You can now proceed to create your digital wallets and access all platform features.
			</p>
			<a href="%s/wallets/create" 
			   style="display: inline-block; background-color: #28a745; color: white; padding: 15px 30px; 
			          text-decoration: none; border-radius: 5px; font-weight: bold;">
				Create Your Wallets
			</a>
		</div>
	</body>
	</html>
	`, e.config.BaseURL)
}

func (e *EmailService) buildKYCApprovedText() string {
	return fmt.Sprintf(`
Verification Complete!

Congratulations! Your identity verification has been successfully approved.
You can now proceed to create your digital wallets and access all platform features.

Create your wallets: %s/wallets/create

Best regards,
The Stack Service Team
	`, e.config.BaseURL)
}

func (e *EmailService) buildKYCRejectedHTML(rejectionReasons []string) string {
	reasons := ""
	for _, reason := range rejectionReasons {
		reasons += fmt.Sprintf("<li style='margin-bottom: 8px;'>%s</li>", reason)
	}

	return fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head><title>KYC Additional Information Required</title></head>
	<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
		<div style="background-color: #fff3cd; padding: 30px; border-radius: 8px; border: 1px solid #ffeaa7;">
			<h1 style="color: #856404; margin-bottom: 20px;">Additional Information Required</h1>
			<p style="color: #856404; font-size: 16px; line-height: 1.5; margin-bottom: 20px;">
				We need additional information to complete your identity verification.
				Please review the following items and resubmit your documents:
			</p>
			<ul style="color: #856404; margin: 20px 0; padding-left: 20px;">%s</ul>
			<a href="%s/kyc/resubmit" 
			   style="display: inline-block; background-color: #ffc107; color: #212529; padding: 15px 30px; 
			          text-decoration: none; border-radius: 5px; font-weight: bold; margin-top: 20px;">
				Resubmit Documents
			</a>
		</div>
	</body>
	</html>
	`, reasons, e.config.BaseURL)
}

func (e *EmailService) buildKYCRejectedText(rejectionReasons []string) string {
	reasons := ""
	for i, reason := range rejectionReasons {
		reasons += fmt.Sprintf("%d. %s\n", i+1, reason)
	}

	return fmt.Sprintf(`
Additional Information Required

We need additional information to complete your identity verification.
Please review the following items and resubmit your documents:

%s
Resubmit documents: %s/kyc/resubmit

Best regards,
The Stack Service Team
	`, reasons, e.config.BaseURL)
}

func (e *EmailService) buildKYCProcessingHTML() string {
	return `
	<!DOCTYPE html>
	<html>
	<head><title>KYC Processing</title></head>
	<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
		<div style="background-color: #cce5ff; padding: 30px; border-radius: 8px; text-align: center; border: 1px solid #99d6ff;">
			<h1 style="color: #004085; margin-bottom: 20px;">⏳ Verification In Progress</h1>
			<p style="color: #004085; font-size: 16px; line-height: 1.5; margin-bottom: 20px;">
				Your identity verification documents are currently being reviewed by our team.
				You will receive an update within 24-48 hours.
			</p>
			<p style="color: #004085; font-size: 14px;">
				Thank you for your patience!
			</p>
		</div>
	</body>
	</html>
	`
}

func (e *EmailService) buildKYCProcessingText() string {
	return `
Verification In Progress

Your identity verification documents are currently being reviewed by our team.
You will receive an update within 24-48 hours.

Thank you for your patience!

Best regards,
The Stack Service Team
	`
}

func (e *EmailService) buildKYCGenericHTML(status string) string {
	return fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head><title>KYC Status Update</title></head>
	<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
		<div style="background-color: #f8f9fa; padding: 30px; border-radius: 8px; border: 1px solid #dee2e6;">
			<h1 style="color: #495057; margin-bottom: 20px;">KYC Status Update</h1>
			<p style="color: #495057; font-size: 16px; line-height: 1.5;">
				Your KYC verification status has been updated to: <strong>%s</strong>
			</p>
		</div>
	</body>
	</html>
	`, status)
}

func (e *EmailService) buildKYCGenericText(status string) string {
	return fmt.Sprintf(`
KYC Status Update

Your KYC verification status has been updated to: %s

Best regards,
The Stack Service Team
	`, status)
}
