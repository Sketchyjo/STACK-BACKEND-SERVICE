package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// EmailSender defines the contract required for sending rich email content
type EmailSender interface {
	SendCustomEmail(ctx context.Context, to, subject, htmlContent, textContent string) error
}

// NotificationService handles sending notifications to users
type NotificationService struct {
	emailService EmailSender
	logger       *zap.Logger
	tracer       trace.Tracer
	metrics      *NotificationMetrics
}

// NotificationMetrics contains observability metrics for notifications
type NotificationMetrics struct {
	NotificationsSent   metric.Int64Counter
	NotificationErrors  metric.Int64Counter
	NotificationLatency metric.Float64Histogram
}

// WeeklySummaryNotification represents a weekly summary notification
type WeeklySummaryNotification struct {
	UserID      uuid.UUID
	Email       string
	WeekStart   time.Time
	WeekEnd     time.Time
	SummaryID   uuid.UUID
	SummaryMD   string
	ArtifactURI string
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	emailService EmailSender,
	logger *zap.Logger,
) (*NotificationService, error) {
	tracer := otel.Tracer("notification-service")
	meter := otel.Meter("notification-service")

	// Initialize metrics
	metrics, err := initNotificationMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize notification metrics: %w", err)
	}

	return &NotificationService{
		emailService: emailService,
		logger:       logger,
		tracer:       tracer,
		metrics:      metrics,
	}, nil
}

// SendWeeklySummaryNotification sends an email notification for a new weekly summary
func (n *NotificationService) SendWeeklySummaryNotification(ctx context.Context, notification *WeeklySummaryNotification) error {
	startTime := time.Now()
	ctx, span := n.tracer.Start(ctx, "notification.send_weekly_summary", trace.WithAttributes(
		attribute.String("user_id", notification.UserID.String()),
		attribute.String("email", notification.Email),
		attribute.String("week_start", notification.WeekStart.Format("2006-01-02")),
	))
	defer span.End()

	n.logger.Info("Sending weekly summary notification",
		zap.String("user_id", notification.UserID.String()),
		zap.String("email", notification.Email),
		zap.String("week_start", notification.WeekStart.Format("2006-01-02")),
	)

	// Create email content
	subject := n.generateEmailSubject(notification)
	htmlBody := n.generateEmailHTML(notification)
	textBody := n.generateEmailText(notification)

	// Send email using a custom method that accesses the private sendEmail
	err := n.sendWeeklySummaryEmail(ctx, notification.Email, subject, htmlBody, textBody)

	// Record metrics
	duration := time.Since(startTime)
	n.metrics.NotificationLatency.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("type", "weekly_summary"),
		attribute.String("channel", "email"),
		attribute.Bool("success", err == nil),
	))

	if err != nil {
		span.RecordError(err)
		n.metrics.NotificationErrors.Add(ctx, 1, metric.WithAttributes(
			attribute.String("type", "weekly_summary"),
			attribute.String("channel", "email"),
		))
		n.logger.Error("Failed to send weekly summary notification",
			zap.Error(err),
			zap.String("user_id", notification.UserID.String()),
			zap.String("email", notification.Email),
		)
		return fmt.Errorf("failed to send weekly summary notification: %w", err)
	}

	n.metrics.NotificationsSent.Add(ctx, 1, metric.WithAttributes(
		attribute.String("type", "weekly_summary"),
		attribute.String("channel", "email"),
	))

	n.logger.Info("Weekly summary notification sent successfully",
		zap.String("user_id", notification.UserID.String()),
		zap.String("email", notification.Email),
		zap.Duration("duration", duration),
	)

	return nil
}

// SendPushNotification sends a push notification (placeholder for future implementation)
func (n *NotificationService) SendPushNotification(ctx context.Context, userID uuid.UUID, title, body string) error {
	ctx, span := n.tracer.Start(ctx, "notification.send_push", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("title", title),
	))
	defer span.End()

	// Placeholder for push notification implementation
	n.logger.Info("Push notification sent (placeholder)",
		zap.String("user_id", userID.String()),
		zap.String("title", title),
		zap.String("body", body),
	)

	n.metrics.NotificationsSent.Add(ctx, 1, metric.WithAttributes(
		attribute.String("type", "push"),
		attribute.String("channel", "mobile"),
	))

	return nil
}

// generateEmailSubject creates the email subject for weekly summary notifications
func (n *NotificationService) generateEmailSubject(notification *WeeklySummaryNotification) string {
	return fmt.Sprintf("Your Weekly Portfolio Summary - %s to %s",
		notification.WeekStart.Format("Jan 2"),
		notification.WeekEnd.Format("Jan 2, 2006"),
	)
}

// generateEmailHTML creates the HTML email body
func (n *NotificationService) generateEmailHTML(notification *WeeklySummaryNotification) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Weekly Portfolio Summary</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .summary-content { background-color: #ffffff; padding: 20px; border: 1px solid #dee2e6; border-radius: 8px; margin-bottom: 20px; }
        .cta-button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 4px; margin: 10px 0; }
        .disclaimer { font-size: 12px; color: #6c757d; margin-top: 20px; padding-top: 20px; border-top: 1px solid #dee2e6; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #6c757d; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📊 Weekly Portfolio Summary</h1>
        <p><strong>Week of %s - %s</strong></p>
        <p>Your AI-powered portfolio analysis is ready!</p>
    </div>

    <div class="summary-content">
        <h2>New Summary Available</h2>
        <p>Your personalized weekly portfolio summary has been generated using advanced AI analysis. This summary includes:</p>
        <ul>
            <li>Portfolio performance overview</li>
            <li>Risk assessment and insights</li>
            <li>Asset allocation analysis</li>
            <li>Market context and trends</li>
            <li>Key areas for consideration</li>
        </ul>
        
        <p>
            <a href="#" class="cta-button">View Your Summary</a>
        </p>
    </div>

    <div class="disclaimer">
        <p><strong>Important Disclaimer:</strong> This analysis is for informational purposes only and does not constitute financial advice. All investment decisions should be made based on your own research and consideration of your financial situation. Past performance does not guarantee future results.</p>
    </div>

    <div class="footer">
        <p>© 2024 STACK Portfolio Management Platform</p>
        <p>This email was sent to you because you have enabled weekly portfolio summary notifications.</p>
    </div>
</body>
</html>`,
		notification.WeekStart.Format("January 2, 2006"),
		notification.WeekEnd.Format("January 2, 2006"),
	)
}

// generateEmailText creates the plain text email body
func (n *NotificationService) generateEmailText(notification *WeeklySummaryNotification) string {
	return fmt.Sprintf(`
STACK Portfolio Management - Weekly Summary

Week of %s - %s

Your AI-powered portfolio analysis is ready!

Your personalized weekly portfolio summary has been generated using advanced AI analysis. This summary includes:

- Portfolio performance overview
- Risk assessment and insights  
- Asset allocation analysis
- Market context and trends
- Key areas for consideration

View your summary in the STACK app or web platform.

IMPORTANT DISCLAIMER: This analysis is for informational purposes only and does not constitute financial advice. All investment decisions should be made based on your own research and consideration of your financial situation. Past performance does not guarantee future results.

---
© 2024 STACK Portfolio Management Platform
This email was sent to you because you have enabled weekly portfolio summary notifications.
`,
		notification.WeekStart.Format("January 2, 2006"),
		notification.WeekEnd.Format("January 2, 2006"),
	)
}

// sendWeeklySummaryEmail sends an email using the email service's private method
// This is a workaround since the email service doesn't have a public SendEmail method with our desired interface
func (n *NotificationService) sendWeeklySummaryEmail(ctx context.Context, to, subject, htmlContent, textContent string) error {
	if n.emailService == nil {
		return fmt.Errorf("email service not configured")
	}

	if err := n.emailService.SendCustomEmail(ctx, to, subject, htmlContent, textContent); err != nil {
		return fmt.Errorf("failed to dispatch weekly summary email: %w", err)
	}
	return nil
}

// initNotificationMetrics initializes OpenTelemetry metrics
func initNotificationMetrics(meter metric.Meter) (*NotificationMetrics, error) {
	notificationsSent, err := meter.Int64Counter("notifications_sent_total",
		metric.WithDescription("Total number of notifications sent"))
	if err != nil {
		return nil, err
	}

	notificationErrors, err := meter.Int64Counter("notification_errors_total",
		metric.WithDescription("Total number of notification errors"))
	if err != nil {
		return nil, err
	}

	notificationLatency, err := meter.Float64Histogram("notification_latency_seconds",
		metric.WithDescription("Notification sending latency in seconds"))
	if err != nil {
		return nil, err
	}

	return &NotificationMetrics{
		NotificationsSent:   notificationsSent,
		NotificationErrors:  notificationErrors,
		NotificationLatency: notificationLatency,
	}, nil
}
