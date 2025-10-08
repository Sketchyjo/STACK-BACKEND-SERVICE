package di

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	domainrepos "github.com/stack-service/stack_service/internal/domain/repositories"
	"github.com/stack-service/stack_service/internal/domain/services"
	"github.com/stack-service/stack_service/internal/domain/services/funding"
	"github.com/stack-service/stack_service/internal/domain/services/investing"
	"github.com/stack-service/stack_service/internal/domain/services/onboarding"
	"github.com/stack-service/stack_service/internal/domain/services/wallet"
	"github.com/stack-service/stack_service/internal/infrastructure/adapters"
	"github.com/stack-service/stack_service/internal/infrastructure/circle"
	"github.com/stack-service/stack_service/internal/infrastructure/config"
	"github.com/stack-service/stack_service/internal/infrastructure/repositories"
	"github.com/stack-service/stack_service/internal/infrastructure/zerog"
	"github.com/stack-service/stack_service/pkg/logger"
	"go.uber.org/zap"
)

// AISummariesRepositoryAdapter adapts the domain repository to the service interface
type AISummariesRepositoryAdapter struct {
	repo domainrepos.AISummaryRepository
}

func (a *AISummariesRepositoryAdapter) CreateSummary(ctx context.Context, summary *services.AISummary) error {
	return a.repo.Create(ctx, summary)
}

func (a *AISummariesRepositoryAdapter) GetLatestSummary(ctx context.Context, userID uuid.UUID) (*services.AISummary, error) {
	return a.repo.GetLatestByUserID(ctx, userID)
}

func (a *AISummariesRepositoryAdapter) GetSummaryByWeek(ctx context.Context, userID uuid.UUID, weekStart time.Time) (*services.AISummary, error) {
	return a.repo.GetByUserAndWeek(ctx, userID, weekStart)
}

func (a *AISummariesRepositoryAdapter) UpdateSummary(ctx context.Context, summary *services.AISummary) error {
	return a.repo.Update(ctx, summary)
}

// Container holds all application dependencies
type Container struct {
	Config *config.Config
	DB     *sql.DB
	Logger *logger.Logger
	ZapLog *zap.Logger

	// Repositories
	UserRepo                  *repositories.UserRepository
	OnboardingFlowRepo        *repositories.OnboardingFlowRepository
	KYCSubmissionRepo         *repositories.KYCSubmissionRepository
	WalletRepo                *repositories.WalletRepository
	WalletSetRepo             *repositories.WalletSetRepository
	WalletProvisioningJobRepo *repositories.WalletProvisioningJobRepository
	DepositRepo               *repositories.DepositRepository
	BalanceRepo               *repositories.BalanceRepository
	FundingEventJobRepo       *repositories.FundingEventJobRepository

	// External Services
	CircleClient *circle.Client
	KYCProvider  *adapters.KYCProvider
	EmailService *adapters.EmailService
	AuditService *adapters.AuditService

	// Domain Services
	OnboardingService *onboarding.Service
	WalletService     *wallet.Service
	FundingService    *funding.Service
	InvestingService  *investing.Service
	AICfoService      *services.AICfoService

	// ZeroG Services
	InferenceGateway  *zerog.InferenceGateway
	StorageClient     *zerog.StorageClient
	NamespaceManager  *zerog.NamespaceManager

	// Additional Repositories for AI-CFO
	AISummariesRepo   domainrepos.AISummaryRepository

	// Workers
	WalletProvisioningScheduler interface{} // Type interface{} to avoid circular dependency, will be set at runtime
	FundingWebhookManager       interface{} // Type interface{} to avoid circular dependency, will be set at runtime
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config, db *sql.DB, log *logger.Logger) (*Container, error) {
	zapLog := log.Zap()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db, zapLog)
	onboardingFlowRepo := repositories.NewOnboardingFlowRepository(db, zapLog)
	kycSubmissionRepo := repositories.NewKYCSubmissionRepository(db, zapLog)
	walletRepo := repositories.NewWalletRepository(db, zapLog)
	walletSetRepo := repositories.NewWalletSetRepository(db, zapLog)
	walletProvisioningJobRepo := repositories.NewWalletProvisioningJobRepository(db, zapLog)
	depositRepo := repositories.NewDepositRepository(db, zapLog)
	balanceRepo := repositories.NewBalanceRepository(db, zapLog)
	fundingEventJobRepo := repositories.NewFundingEventJobRepository(db, log)
	aiSummariesRepo := repositories.NewAISummaryRepository(db, zapLog)

	// Initialize external services
	circleConfig := circle.Config{
		APIKey:      cfg.Circle.APIKey,
		Environment: cfg.Circle.Environment,
	}
	circleClient := circle.NewClient(circleConfig, zapLog)

	// Initialize KYC provider with full configuration
	kycProviderConfig := adapters.KYCProviderConfig{
		APIKey:      cfg.KYC.APIKey,
		APISecret:   cfg.KYC.APISecret,
		BaseURL:     cfg.KYC.BaseURL,
		Environment: cfg.KYC.Environment,
		CallbackURL: cfg.KYC.CallbackURL,
		UserAgent:   cfg.KYC.UserAgent,
	}
	kycProvider := adapters.NewKYCProvider(zapLog, kycProviderConfig)

	// Initialize email service with full configuration
	emailServiceConfig := adapters.EmailServiceConfig{
		APIKey:      cfg.Email.APIKey,
		FromEmail:   cfg.Email.FromEmail,
		FromName:    cfg.Email.FromName,
		Environment: cfg.Email.Environment,
		BaseURL:     cfg.Email.BaseURL,
	}
	emailService := adapters.NewEmailService(zapLog, emailServiceConfig)

	auditService := adapters.NewAuditService(db, zapLog)

	// Initialize ZeroG services
	storageClient, err := zerog.NewStorageClient(&cfg.ZeroG.Storage, zapLog)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ZeroG storage client: %w", err)
	}

	inferenceGateway, err := zerog.NewInferenceGateway(&cfg.ZeroG.Compute, storageClient, zapLog)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ZeroG inference gateway: %w", err)
	}

	namespaceManager := zerog.NewNamespaceManager(storageClient, &cfg.ZeroG.Storage.Namespaces, zapLog)

	container := &Container{
		Config: cfg,
		DB:     db,
		Logger: log,
		ZapLog: zapLog,

		// Repositories
		UserRepo:                  userRepo,
		OnboardingFlowRepo:        onboardingFlowRepo,
		KYCSubmissionRepo:         kycSubmissionRepo,
		WalletRepo:                walletRepo,
		WalletSetRepo:             walletSetRepo,
		WalletProvisioningJobRepo: walletProvisioningJobRepo,
		DepositRepo:               depositRepo,
		BalanceRepo:               balanceRepo,
		FundingEventJobRepo:       fundingEventJobRepo,
		AISummariesRepo:           aiSummariesRepo,

		// External Services
		CircleClient: circleClient,
		KYCProvider:  kycProvider,
		EmailService: emailService,
		AuditService: auditService,

		// ZeroG Services
		InferenceGateway: inferenceGateway,
		StorageClient:    storageClient,
		NamespaceManager: namespaceManager,
	}

	// Initialize domain services with their dependencies
	if err := container.initializeDomainServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize domain services: %w", err)
	}

	return container, nil
}

// initializeDomainServices initializes all domain services with their dependencies
func (c *Container) initializeDomainServices() error {
	// Initialize wallet service first (no dependencies on other domain services)
	c.WalletService = wallet.NewService(
		c.WalletRepo,
		c.WalletSetRepo,
		c.WalletProvisioningJobRepo,
		c.CircleClient,
		c.AuditService,
		c.ZapLog,
	)

	// Initialize onboarding service (depends on wallet service)
	c.OnboardingService = onboarding.NewService(
		c.UserRepo,
		c.OnboardingFlowRepo,
		c.KYCSubmissionRepo,
		c.WalletService, // Domain service dependency
		c.KYCProvider,
		c.EmailService,
		c.AuditService,
		c.ZapLog,
	)

	// Initialize simple wallet repository for funding service
	simpleWalletRepo := repositories.NewSimpleWalletRepository(c.DB, c.Logger)

	// Initialize funding service with dependencies
	c.FundingService = funding.NewService(
		c.DepositRepo,
		c.BalanceRepo,
		simpleWalletRepo,
		c.CircleClient,
		c.Logger,
	)

	// Initialize investing service (placeholder - no dependencies defined yet)
	// TODO: Wire up investing service dependencies when implemented
	c.InvestingService = nil

	// Initialize notification service for AI-CFO
	notificationService, err := services.NewNotificationService(c.EmailService, c.ZapLog)
	if err != nil {
		return fmt.Errorf("failed to initialize notification service: %w", err)
	}

	// Create repository adapter for AI-CFO service
	aiSummariesAdapter := &AISummariesRepositoryAdapter{repo: c.AISummariesRepo}

	// Initialize AI-CFO service
	aicfoService, err := services.NewAICfoService(
		c.InferenceGateway,
		c.StorageClient,
		c.NamespaceManager,
		notificationService,
		nil, // portfolioRepo - TODO: implement when available
		nil, // positionsRepo - TODO: implement when available
		nil, // balanceRepo - TODO: implement GetUserBalance method
		aiSummariesAdapter,
		nil, // userRepo - TODO: implement GetUserPreferences method
		c.ZapLog,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize AI-CFO service: %w", err)
	}
	c.AICfoService = aicfoService

	return nil
}

// GetOnboardingService returns the onboarding service
func (c *Container) GetOnboardingService() *onboarding.Service {
	return c.OnboardingService
}

// GetWalletService returns the wallet service
func (c *Container) GetWalletService() *wallet.Service {
	return c.WalletService
}

// GetFundingService returns the funding service
func (c *Container) GetFundingService() *funding.Service {
	return c.FundingService
}

// GetInvestingService returns the investing service
func (c *Container) GetInvestingService() *investing.Service {
	return c.InvestingService
}

// GetAICfoService returns the AI-CFO service
func (c *Container) GetAICfoService() *services.AICfoService {
	return c.AICfoService
}

// GetInferenceGateway returns the ZeroG inference gateway
func (c *Container) GetInferenceGateway() *zerog.InferenceGateway {
	return c.InferenceGateway
}

// GetStorageClient returns the ZeroG storage client
func (c *Container) GetStorageClient() *zerog.StorageClient {
	return c.StorageClient
}

// GetNamespaceManager returns the ZeroG namespace manager
func (c *Container) GetNamespaceManager() *zerog.NamespaceManager {
	return c.NamespaceManager
}
