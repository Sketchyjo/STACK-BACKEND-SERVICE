package onboarding

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stack-service/stack_service/internal/domain/entities"
	"go.uber.org/zap"
)

// Service handles onboarding operations - user creation, KYC flow, wallet provisioning
type Service struct {
	userRepo           UserRepository
	onboardingFlowRepo OnboardingFlowRepository
	kycSubmissionRepo  KYCSubmissionRepository
	walletService      WalletService
	kycProvider        KYCProvider
	emailService       EmailService
	auditService       AuditService
	logger             *zap.Logger
}

// Repository interfaces
type UserRepository interface {
	Create(ctx context.Context, user *entities.UserProfile) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.UserProfile, error)
	GetByEmail(ctx context.Context, email string) (*entities.UserProfile, error)
	GetByAuthProviderID(ctx context.Context, authProviderID string) (*entities.UserProfile, error)
	Update(ctx context.Context, user *entities.UserProfile) error
	UpdateOnboardingStatus(ctx context.Context, userID uuid.UUID, status entities.OnboardingStatus) error
	UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status string, approvedAt *time.Time, rejectionReason *string) error
}

type OnboardingFlowRepository interface {
	Create(ctx context.Context, flow *entities.OnboardingFlow) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.OnboardingFlow, error)
	GetByUserAndStep(ctx context.Context, userID uuid.UUID, step entities.OnboardingStepType) (*entities.OnboardingFlow, error)
	Update(ctx context.Context, flow *entities.OnboardingFlow) error
	GetCompletedSteps(ctx context.Context, userID uuid.UUID) ([]entities.OnboardingStepType, error)
}

type KYCSubmissionRepository interface {
	Create(ctx context.Context, submission *entities.KYCSubmission) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.KYCSubmission, error)
	GetByProviderRef(ctx context.Context, providerRef string) (*entities.KYCSubmission, error)
	Update(ctx context.Context, submission *entities.KYCSubmission) error
	GetLatestByUserID(ctx context.Context, userID uuid.UUID) (*entities.KYCSubmission, error)
}

// External service interfaces
type WalletService interface {
	CreateWalletsForUser(ctx context.Context, userID uuid.UUID, chains []entities.WalletChain) error
	GetWalletStatus(ctx context.Context, userID uuid.UUID) (*entities.WalletStatusResponse, error)
}

type KYCProvider interface {
	SubmitKYC(ctx context.Context, userID uuid.UUID, documents []entities.KYCDocument, personalInfo *entities.KYCPersonalInfo) (string, error)
	GetKYCStatus(ctx context.Context, providerRef string) (*entities.KYCSubmission, error)
	GenerateKYCURL(ctx context.Context, userID uuid.UUID) (string, error)
}

type EmailService interface {
	SendVerificationEmail(ctx context.Context, email, verificationToken string) error
	SendKYCStatusEmail(ctx context.Context, email string, status entities.KYCStatus, rejectionReasons []string) error
	SendWelcomeEmail(ctx context.Context, email string) error
}

type AuditService interface {
	LogOnboardingEvent(ctx context.Context, userID uuid.UUID, action, entity string, before, after interface{}) error
}

// NewService creates a new onboarding service
func NewService(
	userRepo UserRepository,
	onboardingFlowRepo OnboardingFlowRepository,
	kycSubmissionRepo KYCSubmissionRepository,
	walletService WalletService,
	kycProvider KYCProvider,
	emailService EmailService,
	auditService AuditService,
	logger *zap.Logger,
) *Service {
	return &Service{
		userRepo:           userRepo,
		onboardingFlowRepo: onboardingFlowRepo,
		kycSubmissionRepo:  kycSubmissionRepo,
		walletService:      walletService,
		kycProvider:        kycProvider,
		emailService:       emailService,
		auditService:       auditService,
		logger:             logger,
	}
}

// StartOnboarding initiates the onboarding process for a new user
func (s *Service) StartOnboarding(ctx context.Context, req *entities.OnboardingStartRequest) (*entities.OnboardingStartResponse, error) {
	s.logger.Info("Starting onboarding process", zap.String("email", req.Email))

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		s.logger.Info("User already exists, returning existing onboarding status",
			zap.String("email", req.Email),
			zap.String("userId", existingUser.ID.String()),
			zap.String("status", string(existingUser.OnboardingStatus)))

		return &entities.OnboardingStartResponse{
			UserID:           existingUser.ID,
			OnboardingStatus: existingUser.OnboardingStatus,
			NextStep:         s.determineNextStep(existingUser),
		}, nil
	}

	// Create new user
	user := &entities.UserProfile{
		ID:               uuid.New(),
		Email:            req.Email,
		Phone:            req.Phone,
		EmailVerified:    false,
		PhoneVerified:    false,
		OnboardingStatus: entities.OnboardingStatusStarted,
		KYCStatus:        string(entities.KYCStatusPending),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err), zap.String("email", req.Email))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create initial onboarding flow steps
	if err := s.createInitialOnboardingSteps(ctx, user.ID); err != nil {
		s.logger.Error("Failed to create onboarding steps", zap.Error(err), zap.String("userId", user.ID.String()))
		return nil, fmt.Errorf("failed to create onboarding steps: %w", err)
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(ctx, user.Email, user.ID.String()); err != nil {
		s.logger.Warn("Failed to send verification email", zap.Error(err), zap.String("email", user.Email))
		// Don't fail onboarding start if email fails
	}

	// Log audit event
	if err := s.auditService.LogOnboardingEvent(ctx, user.ID, "onboarding_started", "user", nil, user); err != nil {
		s.logger.Warn("Failed to log audit event", zap.Error(err))
	}

	s.logger.Info("Onboarding started successfully",
		zap.String("userId", user.ID.String()),
		zap.String("email", user.Email))

	return &entities.OnboardingStartResponse{
		UserID:           user.ID,
		OnboardingStatus: user.OnboardingStatus,
		NextStep:         entities.StepEmailVerification,
	}, nil
}

// GetOnboardingStatus returns the current onboarding status for a user
func (s *Service) GetOnboardingStatus(ctx context.Context, userID uuid.UUID) (*entities.OnboardingStatusResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get completed steps
	completedSteps, err := s.onboardingFlowRepo.GetCompletedSteps(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get completed steps", zap.Error(err), zap.String("userId", userID.String()))
		completedSteps = []entities.OnboardingStepType{}
	}

	// Get wallet status if KYC is approved
	var walletStatus *entities.WalletStatusSummary
	if user.OnboardingStatus == entities.OnboardingStatusKYCApproved ||
		user.OnboardingStatus == entities.OnboardingStatusWalletsPending ||
		user.OnboardingStatus == entities.OnboardingStatusCompleted {

		walletStatusResp, err := s.walletService.GetWalletStatus(ctx, userID)
		if err != nil {
			s.logger.Warn("Failed to get wallet status", zap.Error(err), zap.String("userId", userID.String()))
		} else {
			walletStatus = &entities.WalletStatusSummary{
				TotalWallets:    walletStatusResp.TotalWallets,
				CreatedWallets:  walletStatusResp.ReadyWallets,
				PendingWallets:  walletStatusResp.PendingWallets,
				FailedWallets:   walletStatusResp.FailedWallets,
				SupportedChains: []string{"ETH", "SOL", "APTOS"},
				WalletsByChain:  make(map[string]string),
			}

			for chain, status := range walletStatusResp.WalletsByChain {
				walletStatus.WalletsByChain[chain] = status.Status
			}
		}
	}

	// Determine current step and required actions
	currentStep := s.determineCurrentStep(user, completedSteps)
	requiredActions := s.determineRequiredActions(user, completedSteps)
	canProceed := s.canProceed(user, completedSteps)

	return &entities.OnboardingStatusResponse{
		UserID:           user.ID,
		OnboardingStatus: user.OnboardingStatus,
		KYCStatus:        user.KYCStatus,
		CurrentStep:      currentStep,
		CompletedSteps:   completedSteps,
		WalletStatus:     walletStatus,
		CanProceed:       canProceed,
		RequiredActions:  requiredActions,
	}, nil
}

// SubmitKYC handles KYC document submission
func (s *Service) SubmitKYC(ctx context.Context, userID uuid.UUID, req *entities.KYCSubmitRequest) error {
	s.logger.Info("Submitting KYC documents", zap.String("userId", userID.String()))

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if !user.CanStartKYC() {
		return fmt.Errorf("user cannot start KYC process")
	}

	// Submit to KYC provider
	providerRef, err := s.kycProvider.SubmitKYC(ctx, userID, req.Documents, req.PersonalInfo)
	if err != nil {
		return fmt.Errorf("failed to submit KYC to provider: %w", err)
	}

	// Create KYC submission record
	submission := &entities.KYCSubmission{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       "jumio", // TODO: make configurable
		ProviderRef:    providerRef,
		SubmissionType: req.DocumentType,
		Status:         entities.KYCStatusProcessing,
		VerificationData: map[string]any{
			"document_type": req.DocumentType,
			"documents":     req.Documents,
			"personal_info": req.PersonalInfo,
			"metadata":      req.Metadata,
		},
		SubmittedAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.kycSubmissionRepo.Create(ctx, submission); err != nil {
		return fmt.Errorf("failed to create KYC submission record: %w", err)
	}

	// Update user status
	now := time.Now()
	user.OnboardingStatus = entities.OnboardingStatusKYCPending
	user.KYCProviderRef = &providerRef
	user.KYCSubmittedAt = &now
	user.UpdatedAt = now

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	// Update onboarding flow
	if err := s.markStepCompleted(ctx, userID, entities.StepKYCSubmission, map[string]any{
		"provider_ref": providerRef,
		"submitted_at": now,
	}); err != nil {
		s.logger.Warn("Failed to mark KYC submission step as completed", zap.Error(err))
	}

	// Log audit event
	if err := s.auditService.LogOnboardingEvent(ctx, userID, "kyc_submitted", "kyc_submission", nil, submission); err != nil {
		s.logger.Warn("Failed to log audit event", zap.Error(err))
	}

	s.logger.Info("KYC submitted successfully",
		zap.String("userId", userID.String()),
		zap.String("providerRef", providerRef))

	return nil
}

// ProcessKYCCallback processes KYC provider callbacks
func (s *Service) ProcessKYCCallback(ctx context.Context, providerRef string, status entities.KYCStatus, rejectionReasons []string) error {
	s.logger.Info("Processing KYC callback",
		zap.String("providerRef", providerRef),
		zap.String("status", string(status)))

	// Get KYC submission
	submission, err := s.kycSubmissionRepo.GetByProviderRef(ctx, providerRef)
	if err != nil {
		return fmt.Errorf("failed to get KYC submission: %w", err)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, submission.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update submission
	submission.MarkReviewed(status, rejectionReasons)
	if err := s.kycSubmissionRepo.Update(ctx, submission); err != nil {
		return fmt.Errorf("failed to update KYC submission: %w", err)
	}

	// Update user based on KYC result
	var newOnboardingStatus entities.OnboardingStatus
	var kycApprovedAt *time.Time
	var kycRejectionReason *string

	switch status {
	case entities.KYCStatusApproved:
		newOnboardingStatus = entities.OnboardingStatusKYCApproved
		now := time.Now()
		kycApprovedAt = &now

		// Mark KYC review step as completed
		if err := s.markStepCompleted(ctx, user.ID, entities.StepKYCReview, map[string]any{
			"status":      string(status),
			"approved_at": now,
		}); err != nil {
			s.logger.Warn("Failed to mark KYC review step as completed", zap.Error(err))
		}

		// Trigger wallet creation
		if err := s.triggerWalletCreation(ctx, user.ID); err != nil {
			s.logger.Error("Failed to trigger wallet creation", zap.Error(err), zap.String("userId", user.ID.String()))
			// Don't fail the KYC approval process if wallet creation fails
		}

	case entities.KYCStatusRejected:
		newOnboardingStatus = entities.OnboardingStatusKYCRejected
		if len(rejectionReasons) > 0 {
			reason := fmt.Sprintf("KYC rejected: %v", rejectionReasons)
			kycRejectionReason = &reason
		}

		// Mark KYC review step as failed
		if err := s.markStepFailed(ctx, user.ID, entities.StepKYCReview, fmt.Sprintf("KYC rejected: %v", rejectionReasons)); err != nil {
			s.logger.Warn("Failed to mark KYC review step as failed", zap.Error(err))
		}

	default:
		// For processing status, no onboarding status change
		s.logger.Info("KYC still processing", zap.String("status", string(status)))
		return nil
	}

	// Update user status
	if err := s.userRepo.UpdateKYCStatus(ctx, user.ID, string(status), kycApprovedAt, kycRejectionReason); err != nil {
		return fmt.Errorf("failed to update user KYC status: %w", err)
	}

	if err := s.userRepo.UpdateOnboardingStatus(ctx, user.ID, newOnboardingStatus); err != nil {
		return fmt.Errorf("failed to update onboarding status: %w", err)
	}

	// Send status email
	if err := s.emailService.SendKYCStatusEmail(ctx, user.Email, status, rejectionReasons); err != nil {
		s.logger.Warn("Failed to send KYC status email", zap.Error(err))
	}

	// Log audit event
	if err := s.auditService.LogOnboardingEvent(ctx, user.ID, "kyc_reviewed", "kyc_submission",
		map[string]any{"status": "processing"},
		map[string]any{"status": string(status), "rejection_reasons": rejectionReasons}); err != nil {
		s.logger.Warn("Failed to log audit event", zap.Error(err))
	}

	s.logger.Info("KYC callback processed successfully",
		zap.String("userId", user.ID.String()),
		zap.String("status", string(status)))

	return nil
}

// ProcessWalletCreationComplete handles wallet creation completion
func (s *Service) ProcessWalletCreationComplete(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("Processing wallet creation completion", zap.String("userId", userID.String()))

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.OnboardingStatus != entities.OnboardingStatusWalletsPending {
		s.logger.Warn("User is not in wallets pending status",
			zap.String("userId", userID.String()),
			zap.String("status", string(user.OnboardingStatus)))
		return nil
	}

	// Mark wallet creation step as completed
	if err := s.markStepCompleted(ctx, userID, entities.StepWalletCreation, map[string]any{
		"completed_at": time.Now(),
	}); err != nil {
		s.logger.Warn("Failed to mark wallet creation step as completed", zap.Error(err))
	}

	// Mark onboarding as completed
	if err := s.markStepCompleted(ctx, userID, entities.StepOnboardingComplete, map[string]any{
		"completed_at": time.Now(),
	}); err != nil {
		s.logger.Warn("Failed to mark onboarding complete step as completed", zap.Error(err))
	}

	// Update user status
	if err := s.userRepo.UpdateOnboardingStatus(ctx, userID, entities.OnboardingStatusCompleted); err != nil {
		return fmt.Errorf("failed to update onboarding status: %w", err)
	}

	// Send welcome email
	if err := s.emailService.SendWelcomeEmail(ctx, user.Email); err != nil {
		s.logger.Warn("Failed to send welcome email", zap.Error(err))
	}

	// Log audit event
	if err := s.auditService.LogOnboardingEvent(ctx, userID, "onboarding_completed", "user",
		map[string]any{"status": string(entities.OnboardingStatusWalletsPending)},
		map[string]any{"status": string(entities.OnboardingStatusCompleted)}); err != nil {
		s.logger.Warn("Failed to log audit event", zap.Error(err))
	}

	s.logger.Info("Onboarding completed successfully", zap.String("userId", userID.String()))

	return nil
}

// Helper methods

func (s *Service) createInitialOnboardingSteps(ctx context.Context, userID uuid.UUID) error {
	steps := []entities.OnboardingStepType{
		entities.StepRegistration,
		entities.StepEmailVerification,
		entities.StepKYCSubmission,
		entities.StepKYCReview,
		entities.StepWalletCreation,
		entities.StepOnboardingComplete,
	}

	for _, step := range steps {
		flow := &entities.OnboardingFlow{
			ID:        uuid.New(),
			UserID:    userID,
			Step:      step,
			Status:    entities.StepStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mark registration as completed since user was just created
		if step == entities.StepRegistration {
			flow.MarkCompleted(map[string]any{
				"registration_completed": true,
			})
		}

		if err := s.onboardingFlowRepo.Create(ctx, flow); err != nil {
			return fmt.Errorf("failed to create step %s: %w", step, err)
		}
	}

	return nil
}

func (s *Service) markStepCompleted(ctx context.Context, userID uuid.UUID, step entities.OnboardingStepType, data map[string]any) error {
	flow, err := s.onboardingFlowRepo.GetByUserAndStep(ctx, userID, step)
	if err != nil {
		return fmt.Errorf("failed to get onboarding flow step: %w", err)
	}

	flow.MarkCompleted(data)
	return s.onboardingFlowRepo.Update(ctx, flow)
}

func (s *Service) markStepFailed(ctx context.Context, userID uuid.UUID, step entities.OnboardingStepType, errorMsg string) error {
	flow, err := s.onboardingFlowRepo.GetByUserAndStep(ctx, userID, step)
	if err != nil {
		return fmt.Errorf("failed to get onboarding flow step: %w", err)
	}

	flow.MarkFailed(errorMsg)
	return s.onboardingFlowRepo.Update(ctx, flow)
}

func (s *Service) triggerWalletCreation(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("Triggering wallet creation for user", zap.String("userId", userID.String()))

	// Update user status to wallets pending
	if err := s.userRepo.UpdateOnboardingStatus(ctx, userID, entities.OnboardingStatusWalletsPending); err != nil {
		return fmt.Errorf("failed to update status to wallets pending: %w", err)
	}

	// Enqueue wallet provisioning job for supported chains
	// The worker will process this asynchronously with retries
	supportedChains := []entities.WalletChain{
		entities.ChainETH,   // or ChainETHSepolia for testing
		entities.ChainSOL,   // or ChainSOLDevnet for testing
		entities.ChainAPTOS, // or ChainAPTOSTestnet for testing
	}

	// This now enqueues a job instead of processing immediately
	// The worker scheduler will pick it up and process with retries and audit logging
	if err := s.walletService.CreateWalletsForUser(ctx, userID, supportedChains); err != nil {
		s.logger.Error("Failed to enqueue wallet provisioning job",
			zap.Error(err),
			zap.String("userId", userID.String()))
		return fmt.Errorf("failed to enqueue wallet provisioning: %w", err)
	}

	s.logger.Info("Wallet provisioning job enqueued successfully",
		zap.String("userId", userID.String()),
		zap.Int("chains_count", len(supportedChains)))

	return nil
}

func (s *Service) determineNextStep(user *entities.UserProfile) entities.OnboardingStepType {
	switch user.OnboardingStatus {
	case entities.OnboardingStatusStarted:
		return entities.StepEmailVerification
	case entities.OnboardingStatusKYCPending:
		return entities.StepKYCReview
	case entities.OnboardingStatusKYCApproved:
		return entities.StepWalletCreation
	case entities.OnboardingStatusKYCRejected:
		return entities.StepKYCSubmission
	case entities.OnboardingStatusWalletsPending:
		return entities.StepWalletCreation
	case entities.OnboardingStatusCompleted:
		return entities.StepOnboardingComplete
	default:
		return entities.StepRegistration
	}
}

func (s *Service) determineCurrentStep(user *entities.UserProfile, completedSteps []entities.OnboardingStepType) *entities.OnboardingStepType {
	// Find the first uncompleted step
	allSteps := []entities.OnboardingStepType{
		entities.StepRegistration,
		entities.StepEmailVerification,
		entities.StepKYCSubmission,
		entities.StepKYCReview,
		entities.StepWalletCreation,
		entities.StepOnboardingComplete,
	}

	completedMap := make(map[entities.OnboardingStepType]bool)
	for _, step := range completedSteps {
		completedMap[step] = true
	}

	for _, step := range allSteps {
		if !completedMap[step] {
			return &step
		}
	}

	// All steps completed
	step := entities.StepOnboardingComplete
	return &step
}

func (s *Service) determineRequiredActions(user *entities.UserProfile, completedSteps []entities.OnboardingStepType) []string {
	var actions []string

	completedMap := make(map[entities.OnboardingStepType]bool)
	for _, step := range completedSteps {
		completedMap[step] = true
	}

	if !user.EmailVerified && !completedMap[entities.StepEmailVerification] {
		actions = append(actions, "Verify your email address")
	}

	if user.OnboardingStatus == entities.OnboardingStatusStarted && user.EmailVerified {
		actions = append(actions, "Complete KYC verification")
	}

	if user.OnboardingStatus == entities.OnboardingStatusKYCRejected {
		actions = append(actions, "Resubmit KYC documents")
	}

	return actions
}

func (s *Service) canProceed(user *entities.UserProfile, completedSteps []entities.OnboardingStepType) bool {
	switch user.OnboardingStatus {
	case entities.OnboardingStatusStarted:
		return user.EmailVerified
	case entities.OnboardingStatusKYCPending:
		return false // Wait for KYC review
	case entities.OnboardingStatusKYCApproved, entities.OnboardingStatusWalletsPending:
		return false // Wait for wallet creation
	case entities.OnboardingStatusKYCRejected:
		return true // Can retry KYC
	case entities.OnboardingStatusCompleted:
		return true // Fully completed
	default:
		return false
	}
}
