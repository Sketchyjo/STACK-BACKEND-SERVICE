package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/internal/infrastructure/config"
	"github.com/stack-service/stack_service/pkg/auth"
	"github.com/stack-service/stack_service/pkg/crypto"
	"github.com/stack-service/stack_service/pkg/logger"
)

type adminHandler struct {
	db  *sql.DB
	cfg *config.Config
	log *logger.Logger
}

func newAdminHandler(db *sql.DB, cfg *config.Config, log *logger.Logger) *adminHandler {
	return &adminHandler{
		db:  db,
		cfg: cfg,
		log: log,
	}
}

// CreateAdmin handles creation of privileged users.
func CreateAdmin(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.createAdmin
}

// GetAllUsers returns a list of users with optional filters.
func GetAllUsers(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.getAllUsers
}

// GetUserByID returns a user by identifier.
func GetUserByID(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.getUserByID
}

// UpdateUserStatus toggles a user's active state.
func UpdateUserStatus(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.updateUserStatus
}

// GetAllTransactions returns platform transactions relevant to admins.
func GetAllTransactions(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.getAllTransactions
}

// GetSystemAnalytics aggregates high level metrics.
func GetSystemAnalytics(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.getSystemAnalytics
}

// CreateCuratedBasket allows admins to register curated baskets.
func CreateCuratedBasket(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.createCuratedBasket
}

// UpdateCuratedBasket modifies a curated basket definition.
func UpdateCuratedBasket(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.updateCuratedBasket
}

func (h *adminHandler) createAdmin(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req entities.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnw("invalid create admin payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid request payload",
		})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_EMAIL",
			"message": "Email is required",
		})
		return
	}

	if len(req.Password) < max(8, h.cfg.Security.PasswordMinLength) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "WEAK_PASSWORD",
			"message": fmt.Sprintf("Password must be at least %d characters", max(8, h.cfg.Security.PasswordMinLength)),
		})
		return
	}

	adminCount, err := h.countAdmins(ctx)
	if err != nil {
		h.log.Errorw("failed to count admins", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
		return
	}

	desiredRole := entities.AdminRoleAdmin
	if req.Role != nil {
		if !req.Role.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_ROLE",
				"message": "Role must be admin or super_admin",
			})
			return
		}
		desiredRole = *req.Role
	}

	if adminCount == 0 {
		desiredRole = entities.AdminRoleSuperAdmin
	} else {
		if err := h.ensureSuperAdmin(c); err != nil {
			status := http.StatusForbidden
			if errors.Is(err, errUnauthorized) {
				status = http.StatusUnauthorized
			}
			c.JSON(status, gin.H{
				"error":   "ADMIN_PRIVILEGES_REQUIRED",
				"message": err.Error(),
			})
			return
		}
	}

	exists, err := h.emailExists(ctx, req.Email)
	if err != nil {
		h.log.Errorw("failed to check email existence", "error", err, "email", req.Email)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "USER_EXISTS",
			"message": "User already exists with this email",
		})
		return
	}

	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		h.log.Errorw("failed to hash password for admin", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "PASSWORD_HASH_FAILED",
			"message": "Failed to process password",
		})
		return
	}

	now := time.Now().UTC()
	adminID := uuid.New()

	query := `
		INSERT INTO users (
			id, email, password_hash, role, is_active, email_verified, phone_verified,
			onboarding_status, kyc_status, created_at, updated_at, first_name, last_name, phone
		) VALUES (
			$1, $2, $3, $4, true, true, false,
			$5, $6, $7, $8, $9, $10, $11
		)
		RETURNING id, email, role, is_active, onboarding_status, kyc_status, last_login_at, created_at, updated_at`

	onboardingStatus := entities.OnboardingStatusCompleted
	kycStatus := entities.KYCStatusApproved

	var adminResp entities.AdminUserResponse
	var lastLogin sql.NullTime

	err = h.db.QueryRowContext(ctx, query,
		adminID,
		req.Email,
		passwordHash,
		string(desiredRole),
		string(onboardingStatus),
		string(kycStatus),
		now,
		now,
		req.FirstName,
		req.LastName,
		req.Phone,
	).Scan(
		&adminResp.ID,
		&adminResp.Email,
		&adminResp.Role,
		&adminResp.IsActive,
		&adminResp.OnboardingStatus,
		&adminResp.KYCStatus,
		&lastLogin,
		&adminResp.CreatedAt,
		&adminResp.UpdatedAt,
	)

	if err != nil {
		h.log.Errorw("failed to create admin user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "CREATE_FAILED",
			"message": "Failed to create admin",
		})
		return
	}

	if lastLogin.Valid {
		adminResp.LastLoginAt = &lastLogin.Time
	}

	tokenPair, err := auth.GenerateTokenPair(
		adminResp.ID,
		adminResp.Email,
		string(adminResp.Role),
		h.cfg.JWT.Secret,
		h.cfg.JWT.AccessTTL,
		h.cfg.JWT.RefreshTTL,
	)
	if err != nil {
		h.log.Errorw("failed to generate admin session tokens", "error", err, "admin_id", adminResp.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "TOKEN_GENERATION_FAILED",
			"message": "Failed to generate admin session tokens",
		})
		return
	}

	response := entities.AdminCreationResponse{
		AdminUserResponse: adminResp,
		AdminSession: entities.AdminSession{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
			ExpiresAt:    tokenPair.ExpiresAt,
		},
	}

	c.JSON(http.StatusCreated, response)
}

func (h *adminHandler) getAllUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	limit := 50
	if v := strings.TrimSpace(c.DefaultQuery("limit", "50")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	offset := 0
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	var conditions []string
	var args []interface{}

	if roleParam := strings.TrimSpace(c.Query("role")); roleParam != "" {
		if roleParam != "user" && roleParam != "admin" && roleParam != "super_admin" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_ROLE",
				"message": "Role must be user, admin, or super_admin",
			})
			return
		}
		args = append(args, roleParam)
		conditions = append(conditions, fmt.Sprintf("role = $%d", len(args)))
	}

	if isActive := strings.TrimSpace(c.Query("isActive")); isActive != "" {
		active, err := strconv.ParseBool(isActive)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_STATUS",
				"message": "isActive must be a boolean",
			})
			return
		}
		args = append(args, active)
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", len(args)))
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT id, email, role, is_active, onboarding_status, kyc_status, last_login_at, created_at, updated_at
		FROM users`)

	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	queryBuilder.WriteString(" ORDER BY created_at DESC")
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset))

	rows, err := h.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		h.log.Errorw("failed to list users", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to retrieve users",
		})
		return
	}
	defer rows.Close()

	var users []entities.AdminUserResponse
	for rows.Next() {
		var user entities.AdminUserResponse
		var lastLogin sql.NullTime
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Role,
			&user.IsActive,
			&user.OnboardingStatus,
			&user.KYCStatus,
			&lastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			h.log.Errorw("failed to scan user", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "INTERNAL_ERROR",
				"message": "Failed to parse user record",
			})
			return
		}
		if lastLogin.Valid {
			user.LastLoginAt = &lastLogin.Time
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": users,
		"count": len(users),
	})
}

func (h *adminHandler) getUserByID(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_ID",
			"message": "Invalid user ID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, email, role, is_active, onboarding_status, kyc_status, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1`

	var resp entities.AdminUserResponse
	var lastLogin sql.NullTime

	err = h.db.QueryRowContext(ctx, query, userID).Scan(
		&resp.ID,
		&resp.Email,
		&resp.Role,
		&resp.IsActive,
		&resp.OnboardingStatus,
		&resp.KYCStatus,
		&lastLogin,
		&resp.CreatedAt,
		&resp.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "NOT_FOUND",
				"message": "User not found",
			})
			return
		}
		h.log.Errorw("failed to get user by id", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to retrieve user",
		})
		return
	}

	if lastLogin.Valid {
		resp.LastLoginAt = &lastLogin.Time
	}

	c.JSON(http.StatusOK, resp)
}

func (h *adminHandler) updateUserStatus(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_ID",
			"message": "Invalid user ID",
		})
		return
	}

	var req entities.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid request payload",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE users
		SET is_active = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, email, role, is_active, onboarding_status, kyc_status, last_login_at, created_at, updated_at`

	var resp entities.AdminUserResponse
	var lastLogin sql.NullTime

	err = h.db.QueryRowContext(ctx, query, req.IsActive, time.Now().UTC(), userID).Scan(
		&resp.ID,
		&resp.Email,
		&resp.Role,
		&resp.IsActive,
		&resp.OnboardingStatus,
		&resp.KYCStatus,
		&lastLogin,
		&resp.CreatedAt,
		&resp.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "NOT_FOUND",
				"message": "User not found",
			})
			return
		}
		h.log.Errorw("failed to update user status", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to update user status",
		})
		return
	}

	if lastLogin.Valid {
		resp.LastLoginAt = &lastLogin.Time
	}

	c.JSON(http.StatusOK, resp)
}

func (h *adminHandler) getAllTransactions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	limit := 50
	if v := strings.TrimSpace(c.DefaultQuery("limit", "50")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	offset := 0
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	query := `
		SELECT id, user_id, chain, tx_hash, token, amount, status, created_at
		FROM deposits
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := h.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		h.log.Errorw("failed to list transactions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to retrieve transactions",
		})
		return
	}
	defer rows.Close()

	var transactions []entities.AdminTransaction
	for rows.Next() {
		var tx entities.AdminTransaction
		var amount decimal.Decimal
		var chain, txHash, token string

		if err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&chain,
			&txHash,
			&token,
			&amount,
			&tx.Status,
			&tx.CreatedAt,
		); err != nil {
			h.log.Errorw("failed to scan transaction", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "INTERNAL_ERROR",
				"message": "Failed to parse transaction",
			})
			return
		}

		tx.Type = "deposit"
		tx.Amount = amount.String()
		tx.Metadata = map[string]interface{}{
			"chain":  chain,
			"txHash": txHash,
			"token":  token,
		}
		transactions = append(transactions, tx)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": transactions,
		"count": len(transactions),
	})
}

func (h *adminHandler) getSystemAnalytics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		SELECT
			(SELECT COUNT(*) FROM users) AS total_users,
			(SELECT COUNT(*) FROM users WHERE is_active = true) AS active_users,
			(SELECT COUNT(*) FROM users WHERE role IN ('admin','super_admin')) AS total_admins,
			COALESCE((SELECT SUM(amount) FROM deposits WHERE status = 'confirmed'), 0) AS total_deposits,
			(SELECT COUNT(*) FROM deposits WHERE status = 'pending') AS pending_deposits,
			COALESCE((SELECT COUNT(*) FROM wallets), 0) AS total_wallets`

	var analytics entities.SystemAnalytics
	var totalDeposits decimal.Decimal

	err := h.db.QueryRowContext(ctx, query).Scan(
		&analytics.TotalUsers,
		&analytics.ActiveUsers,
		&analytics.TotalAdmins,
		&totalDeposits,
		&analytics.PendingDeposits,
		&analytics.TotalWallets,
	)
	if err != nil {
		h.log.Errorw("failed to load system analytics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to retrieve analytics",
		})
		return
	}

	analytics.TotalDeposits = totalDeposits.String()
	analytics.GeneratedAt = time.Now().UTC()

	c.JSON(http.StatusOK, analytics)
}

func (h *adminHandler) createCuratedBasket(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req entities.CuratedBasketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid request payload",
		})
		return
	}

	if err := h.validateBasketRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "VALIDATION_FAILED",
			"message": err.Error(),
		})
		return
	}

	payload, err := json.Marshal(req.Composition)
	if err != nil {
		h.log.Errorw("failed to marshal basket composition", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to process composition",
		})
		return
	}

	now := time.Now().UTC()
	basketID := uuid.New()

	query := `
		INSERT INTO baskets (id, name, description, risk_level, composition_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, description, risk_level, composition_json, created_at, updated_at`

	var basket entities.Basket
	var compositionRaw []byte

	err = h.db.QueryRowContext(ctx, query,
		basketID,
		req.Name,
		req.Description,
		req.RiskLevel,
		payload,
		now,
		now,
	).Scan(
		&basket.ID,
		&basket.Name,
		&basket.Description,
		&basket.RiskLevel,
		&compositionRaw,
		&basket.CreatedAt,
		&basket.UpdatedAt,
	)

	if err != nil {
		h.log.Errorw("failed to create basket", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "CREATE_FAILED",
			"message": "Failed to create curated basket",
		})
		return
	}

	if err := json.Unmarshal(compositionRaw, &basket.Composition); err != nil {
		h.log.Errorw("failed to unmarshal basket composition", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to process basket composition",
		})
		return
	}

	c.JSON(http.StatusCreated, basket)
}

func (h *adminHandler) updateCuratedBasket(c *gin.Context) {
	basketIDParam := c.Param("id")
	basketID, err := uuid.Parse(basketIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_ID",
			"message": "Invalid basket ID",
		})
		return
	}

	var req entities.CuratedBasketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid request payload",
		})
		return
	}

	if err := h.validateBasketRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "VALIDATION_FAILED",
			"message": err.Error(),
		})
		return
	}

	payload, err := json.Marshal(req.Composition)
	if err != nil {
		h.log.Errorw("failed to marshal basket composition", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to process composition",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE baskets
		SET name = $1,
		    description = $2,
		    risk_level = $3,
		    composition_json = $4,
		    updated_at = $5
		WHERE id = $6
		RETURNING id, name, description, risk_level, composition_json, created_at, updated_at`

	var basket entities.Basket
	var compositionRaw []byte

	err = h.db.QueryRowContext(ctx, query,
		req.Name,
		req.Description,
		req.RiskLevel,
		payload,
		time.Now().UTC(),
		basketID,
	).Scan(
		&basket.ID,
		&basket.Name,
		&basket.Description,
		&basket.RiskLevel,
		&compositionRaw,
		&basket.CreatedAt,
		&basket.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "NOT_FOUND",
				"message": "Basket not found",
			})
			return
		}
		h.log.Errorw("failed to update basket", "error", err, "basket_id", basketID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "UPDATE_FAILED",
			"message": "Failed to update curated basket",
		})
		return
	}

	if err := json.Unmarshal(compositionRaw, &basket.Composition); err != nil {
		h.log.Errorw("failed to unmarshal basket composition", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to process basket composition",
		})
		return
	}

	c.JSON(http.StatusOK, basket)
}

func (h *adminHandler) countAdmins(ctx context.Context) (int64, error) {
	var count int64
	err := h.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users WHERE role IN ('admin','super_admin')`,
	).Scan(&count)
	return count, err
}

func (h *adminHandler) emailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := h.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM users WHERE LOWER(email) = $1
		)`, email,
	).Scan(&exists)
	return exists, err
}

var errUnauthorized = errors.New("authentication required")

func (h *adminHandler) ensureSuperAdmin(c *gin.Context) error {
	if role := c.GetString("user_role"); role != "" {
		if role == string(entities.AdminRoleSuperAdmin) {
			return nil
		}
		return errors.New("super_admin role required")
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return errUnauthorized
	}

	const bearer = "Bearer "
	if !strings.HasPrefix(authHeader, bearer) {
		return errUnauthorized
	}

	token := strings.TrimSpace(authHeader[len(bearer):])
	if token == "" {
		return errUnauthorized
	}

	claims, err := auth.ValidateToken(token, h.cfg.JWT.Secret)
	if err != nil {
		h.log.Warnw("failed to validate token for admin creation", "error", err)
		return errUnauthorized
	}

	if claims.Role != string(entities.AdminRoleSuperAdmin) {
		return errors.New("super_admin role required")
	}

	return nil
}

func (h *adminHandler) validateBasketRequest(req *entities.CuratedBasketRequest) error {
	if len(req.Composition) == 0 {
		return errors.New("composition must contain at least one component")
	}

	if req.RiskLevel != entities.RiskLevelConservative &&
		req.RiskLevel != entities.RiskLevelBalanced &&
		req.RiskLevel != entities.RiskLevelGrowth {
		return fmt.Errorf("invalid riskLevel: %s", req.RiskLevel)
	}

	total := decimal.Zero
	for idx, component := range req.Composition {
		if strings.TrimSpace(component.Symbol) == "" {
			return fmt.Errorf("composition[%d].symbol is required", idx)
		}
		if component.Weight.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("composition[%d].weight must be greater than zero", idx)
		}
		total = total.Add(component.Weight)
	}

	diff := total.Sub(decimal.NewFromInt(1)).Abs()
	if diff.GreaterThan(decimal.NewFromFloat(0.0001)) {
		return errors.New("composition weights must sum to 1.0")
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// CreateWalletSet handles POST /api/v1/admin/wallet-sets
// @Summary Create wallet set
// @Description Creates a new Circle wallet set for managing user wallets
// @Tags admin
// @Accept json
// @Produce json
// @Param request body entities.CreateWalletSetRequest true "Wallet set creation request"
// @Success 201 {object} entities.WalletSet
// @Failure 400 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/wallet-sets [post]
func CreateWalletSet(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.createWalletSet
}

// GetWalletSets handles GET /api/v1/admin/wallet-sets
// @Summary List wallet sets
// @Description Returns a list of all wallet sets with optional pagination
// @Tags admin
// @Produce json
// @Param limit query int false "Number of items per page" default(50)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} entities.WalletSetsListResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/wallet-sets [get]
func GetWalletSets(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.getWalletSets
}

// GetWalletSetByID handles GET /api/v1/admin/wallet-sets/:id
// @Summary Get wallet set by ID
// @Description Returns wallet set details by ID
// @Tags admin
// @Produce json
// @Param id path string true "Wallet Set ID"
// @Success 200 {object} entities.WalletSetDetailResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 404 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/wallet-sets/{id} [get]
func GetWalletSetByID(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.getWalletSetByID
}

// GetAdminWallets handles GET /api/v1/admin/wallets
// @Summary List all wallets (admin)
// @Description Returns a list of all user wallets with optional filters
// @Tags admin
// @Produce json
// @Param limit query int false "Number of items per page" default(50)
// @Param offset query int false "Number of items to skip" default(0)
// @Param user_id query string false "Filter by user ID"
// @Param chain query string false "Filter by blockchain chain"
// @Param account_type query string false "Filter by account type" Enums(EOA,SCA)
// @Param status query string false "Filter by wallet status" Enums(creating,live,failed)
// @Success 200 {object} entities.AdminWalletsListResponse
// @Failure 500 {object} entities.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/wallets [get]
func GetAdminWallets(db *sql.DB, cfg *config.Config, log *logger.Logger) gin.HandlerFunc {
	handler := newAdminHandler(db, cfg, log)
	return handler.getAdminWallets
}

func (h *adminHandler) createWalletSet(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var req entities.CreateWalletSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnw("invalid create wallet set payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid request payload",
		})
		return
	}

	// Validate required fields
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "MISSING_NAME",
			"message": "Wallet set name is required",
		})
		return
	}

	// Entity secret is now generated dynamically, no validation needed

	// Create wallet set in database
	walletSetID := uuid.New()
	now := time.Now().UTC()

	query := `
		INSERT INTO wallet_sets (
			id, name, circle_wallet_set_id, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
		RETURNING id, name, circle_wallet_set_id, status, created_at, updated_at`

	var walletSet entities.WalletSet
	err := h.db.QueryRowContext(ctx, query,
		walletSetID,
		req.Name,
		req.CircleWalletSetID, // This would be empty for new sets
		string(entities.WalletSetStatusActive),
		now,
		now,
	).Scan(
		&walletSet.ID,
		&walletSet.Name,
		&walletSet.CircleWalletSetID,
		&walletSet.Status,
		&walletSet.CreatedAt,
		&walletSet.UpdatedAt,
	)

	if err != nil {
		h.log.Errorw("failed to create wallet set", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "CREATE_FAILED",
			"message": "Failed to create wallet set",
		})
		return
	}

	c.JSON(http.StatusCreated, walletSet)
}

func (h *adminHandler) getWalletSets(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	limit := 50
	if v := strings.TrimSpace(c.DefaultQuery("limit", "50")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	offset := 0
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	query := `
		SELECT id, name, circle_wallet_set_id, status, created_at, updated_at
		FROM wallet_sets
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := h.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		h.log.Errorw("failed to list wallet sets", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to retrieve wallet sets",
		})
		return
	}
	defer rows.Close()

	var walletSets []entities.WalletSet
	for rows.Next() {
		var walletSet entities.WalletSet
		if err := rows.Scan(
			&walletSet.ID,
			&walletSet.Name,
			&walletSet.CircleWalletSetID,
			&walletSet.Status,
			&walletSet.CreatedAt,
			&walletSet.UpdatedAt,
		); err != nil {
			h.log.Errorw("failed to scan wallet set", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "INTERNAL_ERROR",
				"message": "Failed to parse wallet set record",
			})
			return
		}
		walletSets = append(walletSets, walletSet)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": walletSets,
		"count": len(walletSets),
	})
}

func (h *adminHandler) getWalletSetByID(c *gin.Context) {
	walletSetIDParam := c.Param("id")
	walletSetID, err := uuid.Parse(walletSetIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_ID",
			"message": "Invalid wallet set ID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, name, circle_wallet_set_id, status, created_at, updated_at
		FROM wallet_sets
		WHERE id = $1`

	var walletSet entities.WalletSet
	err = h.db.QueryRowContext(ctx, query, walletSetID).Scan(
		&walletSet.ID,
		&walletSet.Name,
		&walletSet.CircleWalletSetID,
		&walletSet.Status,
		&walletSet.CreatedAt,
		&walletSet.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "NOT_FOUND",
				"message": "Wallet set not found",
			})
			return
		}
		h.log.Errorw("failed to get wallet set by id", "error", err, "wallet_set_id", walletSetID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to retrieve wallet set",
		})
		return
	}

	c.JSON(http.StatusOK, walletSet)
}

func (h *adminHandler) getAdminWallets(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	limit := 50
	if v := strings.TrimSpace(c.DefaultQuery("limit", "50")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	offset := 0
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Add filters
	if userIDParam := strings.TrimSpace(c.Query("user_id")); userIDParam != "" {
		userID, err := uuid.Parse(userIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_USER_ID",
				"message": "Invalid user ID format",
			})
			return
		}
		args = append(args, userID)
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		argIndex++
	}

	if chainParam := strings.TrimSpace(c.Query("chain")); chainParam != "" {
		args = append(args, chainParam)
		conditions = append(conditions, fmt.Sprintf("chain = $%d", argIndex))
		argIndex++
	}

	if accountTypeParam := strings.TrimSpace(c.Query("account_type")); accountTypeParam != "" {
		if accountTypeParam != "EOA" && accountTypeParam != "SCA" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_ACCOUNT_TYPE",
				"message": "Account type must be EOA or SCA",
			})
			return
		}
		args = append(args, accountTypeParam)
		conditions = append(conditions, fmt.Sprintf("account_type = $%d", argIndex))
		argIndex++
	}

	if statusParam := strings.TrimSpace(c.Query("status")); statusParam != "" {
		if statusParam != "creating" && statusParam != "live" && statusParam != "failed" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_STATUS",
				"message": "Status must be creating, live, or failed",
			})
			return
		}
		args = append(args, statusParam)
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		argIndex++
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT id, user_id, wallet_set_id, circle_wallet_id, chain, address, account_type, status, created_at, updated_at
		FROM managed_wallets`)

	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	queryBuilder.WriteString(" ORDER BY created_at DESC")
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))

	args = append(args, limit, offset)

	rows, err := h.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		h.log.Errorw("failed to list wallets", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to retrieve wallets",
		})
		return
	}
	defer rows.Close()

	var wallets []entities.ManagedWallet
	for rows.Next() {
		var wallet entities.ManagedWallet
		if err := rows.Scan(
			&wallet.ID,
			&wallet.UserID,
			&wallet.WalletSetID,
			&wallet.CircleWalletID,
			&wallet.Chain,
			&wallet.Address,
			&wallet.AccountType,
			&wallet.Status,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			h.log.Errorw("failed to scan wallet", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "INTERNAL_ERROR",
				"message": "Failed to parse wallet record",
			})
			return
		}
		wallets = append(wallets, wallet)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": wallets,
		"count": len(wallets),
	})
}
