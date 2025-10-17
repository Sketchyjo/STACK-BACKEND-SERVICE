package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/pkg/crypto"
)

// UserRepository implements the user repository interface using PostgreSQL
type UserRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB, logger *zap.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entities.UserProfile) error {
	query := `
		INSERT INTO users (
			id, email, phone, first_name, last_name, date_of_birth,
			auth_provider_id, email_verified, phone_verified, 
			onboarding_status, kyc_status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		user.FirstName,
		user.LastName,
		user.DateOfBirth,
		user.AuthProviderID,
		user.EmailVerified,
		user.PhoneVerified,
		string(user.OnboardingStatus),
		user.KYCStatus,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("user with email already exists: %w", err)
		}
		r.logger.Error("Failed to create user", zap.Error(err), zap.String("email", user.Email))
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Debug("User created successfully", zap.String("user_id", user.ID.String()))
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.UserProfile, error) {
	query := `
        SELECT id, email, phone, first_name, last_name, date_of_birth,
               auth_provider_id, email_verified, phone_verified,
               onboarding_status, kyc_status, kyc_approved_at, kyc_rejection_reason,
               is_active, created_at, updated_at
        FROM users 
        WHERE id = $1`

	user := &entities.UserProfile{}
	var kycApprovedAt sql.NullTime
	var kycRejectionReason sql.NullString
	var firstName, lastName sql.NullString
	var dateOfBirth sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&firstName,
		&lastName,
		&dateOfBirth,
		&user.AuthProviderID,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.OnboardingStatus,
		&user.KYCStatus,
		&kycApprovedAt,
		&kycRejectionReason,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", id.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if kycApprovedAt.Valid {
		user.KYCApprovedAt = &kycApprovedAt.Time
	}
	if kycRejectionReason.Valid {
		user.KYCRejectionReason = &kycRejectionReason.String
	}
	// firstName, lastName, dateOfBirth fields not yet implemented in UserProfile entity

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.UserProfile, error) {
	query := `
        SELECT id, email, phone, first_name, last_name, date_of_birth,
               auth_provider_id, email_verified, phone_verified,
               onboarding_status, kyc_status, kyc_approved_at, kyc_rejection_reason,
               is_active, created_at, updated_at
        FROM users 
        WHERE email = $1`

	user := &entities.UserProfile{}
	var kycApprovedAt sql.NullTime
	var kycRejectionReason sql.NullString
	var firstName, lastName sql.NullString
	var dateOfBirth sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&firstName,
		&lastName,
		&dateOfBirth,
		&user.AuthProviderID,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.OnboardingStatus,
		&user.KYCStatus,
		&kycApprovedAt,
		&kycRejectionReason,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if kycApprovedAt.Valid {
		user.KYCApprovedAt = &kycApprovedAt.Time
	}
	if kycRejectionReason.Valid {
		user.KYCRejectionReason = &kycRejectionReason.String
	}

	return user, nil
}

// GetByAuthProviderID retrieves a user by auth provider ID
func (r *UserRepository) GetByAuthProviderID(ctx context.Context, authProviderID string) (*entities.UserProfile, error) {
	query := `
        SELECT id, email, phone, first_name, last_name, date_of_birth,
               auth_provider_id, email_verified, phone_verified,
               onboarding_status, kyc_status, kyc_approved_at, kyc_rejection_reason,
               is_active, created_at, updated_at
        FROM users 
        WHERE auth_provider_id = $1`

	user := &entities.UserProfile{}
	var kycApprovedAt sql.NullTime
	var kycRejectionReason sql.NullString
	var firstName, lastName sql.NullString
	var dateOfBirth sql.NullTime

	err := r.db.QueryRowContext(ctx, query, authProviderID).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&firstName,
		&lastName,
		&dateOfBirth,
		&user.AuthProviderID,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.OnboardingStatus,
		&user.KYCStatus,
		&kycApprovedAt,
		&kycRejectionReason,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by auth provider ID", zap.Error(err), zap.String("auth_provider_id", authProviderID))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if kycApprovedAt.Valid {
		user.KYCApprovedAt = &kycApprovedAt.Time
	}
	if kycRejectionReason.Valid {
		user.KYCRejectionReason = &kycRejectionReason.String
	}

	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *entities.UserProfile) error {
	query := `
		UPDATE users SET 
			email = $2, phone = $3, first_name = $4, last_name = $5, 
			date_of_birth = $6, auth_provider_id = $7, email_verified = $8, 
			phone_verified = $9, onboarding_status = $10, kyc_status = $11, 
			kyc_approved_at = $12, kyc_rejection_reason = $13, updated_at = $14
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		"",  // first_name - not yet implemented in UserProfile
		"",  // last_name - not yet implemented in UserProfile
		nil, // date_of_birth - not yet implemented in UserProfile
		user.AuthProviderID,
		user.EmailVerified,
		user.PhoneVerified,
		string(user.OnboardingStatus),
		user.KYCStatus,
		user.KYCApprovedAt,
		user.KYCRejectionReason,
		time.Now(),
	)

	if err != nil {
		r.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", user.ID.String()))
		return fmt.Errorf("failed to update user: %w", err)
	}

	r.logger.Debug("User updated successfully", zap.String("user_id", user.ID.String()))
	return nil
}

// UpdateOnboardingStatus updates the onboarding status
func (r *UserRepository) UpdateOnboardingStatus(ctx context.Context, userID uuid.UUID, status entities.OnboardingStatus) error {
	query := `UPDATE users SET onboarding_status = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, userID, string(status), time.Now())
	if err != nil {
		r.logger.Error("Failed to update onboarding status", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to update onboarding status: %w", err)
	}

	r.logger.Debug("Onboarding status updated", zap.String("user_id", userID.String()), zap.String("status", string(status)))
	return nil
}

// UpdateKYCStatus updates the KYC status and related fields
func (r *UserRepository) UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status string, approvedAt *time.Time, rejectionReason *string) error {
	query := `
		UPDATE users SET 
			kyc_status = $2, 
			kyc_approved_at = $3, 
			kyc_rejection_reason = $4, 
			updated_at = $5 
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, userID, status, approvedAt, rejectionReason, time.Now())
	if err != nil {
		r.logger.Error("Failed to update KYC status", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to update KYC status: %w", err)
	}

	r.logger.Debug("KYC status updated", zap.String("user_id", userID.String()), zap.String("status", status))
	return nil
}

// UpdateKYCProvider sets the KYC provider reference and status
func (r *UserRepository) UpdateKYCProvider(ctx context.Context, userID uuid.UUID, providerRef string, status entities.KYCStatus) error {
	query := `
		UPDATE users SET 
			kyc_provider_ref = $2,
			kyc_status = $3,
			updated_at = $4
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, userID, providerRef, string(status), time.Now())
	if err != nil {
		r.logger.Error("Failed to update KYC provider reference", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to update KYC provider reference: %w", err)
	}

	r.logger.Info("Updated KYC provider reference",
		zap.String("user_id", userID.String()),
		zap.String("provider_ref", providerRef),
		zap.String("status", string(status)))

	return nil
}

// === Authentication-related methods ===

// CreateUserFromAuth creates a new user with authentication data
func (r *UserRepository) CreateUserFromAuth(ctx context.Context, req *entities.RegisterRequest) (*entities.User, error) {
	// Hash password
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		r.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user := &entities.User{
		ID:               uuid.New(),
		Email:            req.Email,
		Phone:            req.Phone,
		PasswordHash:     passwordHash,
		EmailVerified:    false, // Will be set to true after email verification
		PhoneVerified:    false, // Will be set to true after phone verification (if phone provided)
		OnboardingStatus: entities.OnboardingStatusStarted,
		KYCStatus:        string(entities.KYCStatusPending),
		Role:             "user",
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Insert into database
	query := `
		INSERT INTO users (
			id, email, phone, password_hash, auth_provider_id, 
			email_verified, phone_verified, onboarding_status, kyc_status,
			role, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	_, err = r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.AuthProviderID,
		user.EmailVerified,
		user.PhoneVerified,
		string(user.OnboardingStatus),
		user.KYCStatus,
		user.Role,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			r.logger.Warn("User registration failed - email already exists", zap.String("email", req.Email))
			return nil, fmt.Errorf("user with email already exists")
		}
		r.logger.Error("Failed to create user", zap.Error(err), zap.String("email", req.Email))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info("User created successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email))

	return user, nil
}

// GetUserByEmailForLogin retrieves a user by email for login purposes (includes password hash)
func (r *UserRepository) GetUserByEmailForLogin(ctx context.Context, email string) (*entities.User, error) {
	query := `
		SELECT id, email, phone, password_hash, auth_provider_id,
		       email_verified, phone_verified, onboarding_status, kyc_status,
		       kyc_provider_ref, kyc_submitted_at, kyc_approved_at, kyc_rejection_reason,
		       role, is_active, last_login_at, created_at, updated_at
		FROM users 
		WHERE email = $1 AND is_active = true`

	user := &entities.User{}
	var kycSubmittedAt, kycApprovedAt, lastLoginAt sql.NullTime
	var kycRejectionReason, kycProviderRef sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.AuthProviderID,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.OnboardingStatus,
		&user.KYCStatus,
		&kycProviderRef,
		&kycSubmittedAt,
		&kycApprovedAt,
		&kycRejectionReason,
		&user.Role,
		&user.IsActive,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by email for login", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Handle nullable fields
	if kycProviderRef.Valid {
		user.KYCProviderRef = &kycProviderRef.String
	}
	if kycSubmittedAt.Valid {
		user.KYCSubmittedAt = &kycSubmittedAt.Time
	}
	if kycApprovedAt.Valid {
		user.KYCApprovedAt = &kycApprovedAt.Time
	}
	if kycRejectionReason.Valid {
		user.KYCRejectionReason = &kycRejectionReason.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// PhoneExists checks if a phone number already exists
func (r *UserRepository) PhoneExists(ctx context.Context, phone string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1 AND is_active = true)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, phone).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check phone existence", zap.Error(err), zap.String("phone", phone))
		return false, fmt.Errorf("failed to check phone existence: %w", err)
	}

	r.logger.Debug("Checked phone existence", zap.String("phone", phone), zap.Bool("exists", exists))
	return exists, nil
}

// GetUserByPhoneForLogin retrieves a user by phone for login purposes (includes password hash)
func (r *UserRepository) GetUserByPhoneForLogin(ctx context.Context, phone string) (*entities.User, error) {
	query := `
		SELECT id, email, phone, password_hash, auth_provider_id,
		       email_verified, phone_verified, onboarding_status, kyc_status,
		       kyc_provider_ref, kyc_submitted_at, kyc_approved_at, kyc_rejection_reason,
		       role, is_active, last_login_at, created_at, updated_at
		FROM users 
		WHERE phone = $1 AND is_active = true`

	user := &entities.User{}
	var kycSubmittedAt, kycApprovedAt, lastLoginAt sql.NullTime
	var kycRejectionReason, kycProviderRef sql.NullString

	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.AuthProviderID,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.OnboardingStatus,
		&user.KYCStatus,
		&kycProviderRef,
		&kycSubmittedAt,
		&kycApprovedAt,
		&kycRejectionReason,
		&user.Role,
		&user.IsActive,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by phone for login", zap.Error(err), zap.String("phone", phone))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Handle nullable fields
	if kycProviderRef.Valid {
		user.KYCProviderRef = &kycProviderRef.String
	}
	if kycSubmittedAt.Valid {
		user.KYCSubmittedAt = &kycSubmittedAt.Time
	}
	if kycApprovedAt.Valid {
		user.KYCApprovedAt = &kycApprovedAt.Time
	}
	if kycRejectionReason.Valid {
		user.KYCRejectionReason = &kycRejectionReason.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetUserEntityByID retrieves a user as User entity by ID (excludes sensitive fields like password)
func (r *UserRepository) GetUserEntityByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	query := `
		SELECT id, email, phone, auth_provider_id,
		       email_verified, phone_verified, onboarding_status, kyc_status,
		       kyc_provider_ref, kyc_submitted_at, kyc_approved_at, kyc_rejection_reason,
		       role, is_active, last_login_at, created_at, updated_at
		FROM users 
		WHERE id = $1 AND is_active = true`

	user := &entities.User{}
	var kycSubmittedAt, kycApprovedAt, lastLoginAt sql.NullTime
	var kycRejectionReason, kycProviderRef sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.AuthProviderID,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.OnboardingStatus,
		&user.KYCStatus,
		&kycProviderRef,
		&kycSubmittedAt,
		&kycApprovedAt,
		&kycRejectionReason,
		&user.Role,
		&user.IsActive,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", id.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Handle nullable fields
	if kycProviderRef.Valid {
		user.KYCProviderRef = &kycProviderRef.String
	}
	if kycSubmittedAt.Valid {
		user.KYCSubmittedAt = &kycSubmittedAt.Time
	}
	if kycApprovedAt.Valid {
		user.KYCApprovedAt = &kycApprovedAt.Time
	}
	if kycRejectionReason.Valid {
		user.KYCRejectionReason = &kycRejectionReason.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login_at = $2, updated_at = $2 WHERE id = $1`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, now)
	if err != nil {
		r.logger.Error("Failed to update last login", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// ValidatePassword validates a user's password
func (r *UserRepository) ValidatePassword(plainPassword, hashedPassword string) bool {
	return crypto.ValidatePassword(plainPassword, hashedPassword)
}

// EmailExists checks if an email is already registered
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1 AND is_active = true`

	err := r.db.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to check email existence", zap.Error(err), zap.String("email", email))
		return false, fmt.Errorf("failed to check email: %w", err)
	}

	return count > 0, nil
}

// UpdatePassword updates the user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, newHash, now)
	if err != nil {
		r.logger.Error("Failed to update password", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to update password: %w", err)
	}
	r.logger.Debug("Password updated", zap.String("user_id", userID.String()))
	return nil
}

// DeactivateUser sets is_active to false for the given user
func (r *UserRepository) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET is_active = false, updated_at = $2 WHERE id = $1`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, now)
	if err != nil {
		r.logger.Error("Failed to deactivate user", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to deactivate user: %w", err)
	}
	r.logger.Info("User deactivated", zap.String("user_id", userID.String()))
	return nil
}
