package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/internal/domain/services/onboarding"
	"github.com/stack-service/stack_service/internal/domain/services/passcode"
)

// SecurityHandlers manages sensitive security endpoints such as passcodes
type SecurityHandlers struct {
	passcodeService   *passcode.Service
	onboardingService *onboarding.Service
	logger            *zap.Logger
}

// NewSecurityHandlers constructs SecurityHandlers
func NewSecurityHandlers(passcodeService *passcode.Service, onboardingService *onboarding.Service, logger *zap.Logger) *SecurityHandlers {
	return &SecurityHandlers{
		passcodeService:   passcodeService,
		onboardingService: onboardingService,
		logger:            logger,
	}
}

// GetPasscodeStatus returns current passcode configuration for the authenticated user
func (h *SecurityHandlers) GetPasscodeStatus(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := h.getUserID(c)
	if err != nil {
		h.respondWithUserError(c, http.StatusBadRequest, "INVALID_USER_ID", err.Error())
		return
	}

	status, err := h.passcodeService.GetStatus(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to fetch passcode status",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("request_id", getRequestID(c)))
		h.respondWithInternalError(c, "PASSCODE_STATUS_ERROR", "Failed to retrieve passcode status")
		return
	}

	c.JSON(http.StatusOK, status)
}

// CreatePasscode configures a passcode for a user without one
func (h *SecurityHandlers) CreatePasscode(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := h.getUserID(c)
	if err != nil {
		h.respondWithUserError(c, http.StatusBadRequest, "INVALID_USER_ID", err.Error())
		return
	}

	var req entities.PasscodeSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithUserError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request payload")
		return
	}

	if strings.TrimSpace(req.Passcode) == "" || strings.TrimSpace(req.ConfirmPasscode) == "" {
		h.respondWithUserError(c, http.StatusBadRequest, "INVALID_REQUEST", "Passcode and confirmation are required")
		return
	}

	if req.Passcode != req.ConfirmPasscode {
		h.respondWithUserError(c, http.StatusBadRequest, "PASSCODE_MISMATCH", "Passcode and confirmation must match")
		return
	}

	status, err := h.passcodeService.SetPasscode(ctx, userID, req.Passcode)
	if err != nil {
		switch {
		case err == passcode.ErrPasscodeAlreadySet:
			h.respondWithUserError(c, http.StatusConflict, "PASSCODE_EXISTS", "Passcode already configured. Use update endpoint instead.")
		case err == passcode.ErrPasscodeInvalidFormat:
			h.respondWithUserError(c, http.StatusBadRequest, "INVALID_PASSCODE_FORMAT", "Passcode must be 4 digits.")
		default:
			h.logger.Error("Failed to set passcode",
				zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("request_id", getRequestID(c)))
			h.respondWithInternalError(c, "PASSCODE_SETUP_FAILED", "Failed to configure passcode")
		}
		return
	}

	// Trigger wallet creation after passcode creation
	if h.onboardingService != nil {
		if err := h.onboardingService.CompletePasscodeCreation(ctx, userID); err != nil {
			h.logger.Warn("Failed to complete passcode creation in onboarding flow",
				zap.Error(err),
				zap.String("user_id", userID.String()))
			// Don't fail the passcode creation, just log the warning
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Passcode created successfully",
		"status":  status,
	})
}

// UpdatePasscode rotates an existing passcode
func (h *SecurityHandlers) UpdatePasscode(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := h.getUserID(c)
	if err != nil {
		h.respondWithUserError(c, http.StatusBadRequest, "INVALID_USER_ID", err.Error())
		return
	}

	var req entities.PasscodeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithUserError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request payload")
		return
	}

	if req.NewPasscode != req.ConfirmPasscode {
		h.respondWithUserError(c, http.StatusBadRequest, "PASSCODE_MISMATCH", "New passcode and confirmation must match")
		return
	}

	status, err := h.passcodeService.UpdatePasscode(ctx, userID, req.CurrentPasscode, req.NewPasscode)
	if err != nil {
		switch {
		case err == passcode.ErrPasscodeNotSet:
			h.respondWithUserError(c, http.StatusBadRequest, "PASSCODE_NOT_SET", "No passcode configured yet.")
		case err == passcode.ErrPasscodeInvalidFormat:
			h.respondWithUserError(c, http.StatusBadRequest, "INVALID_PASSCODE_FORMAT", "Passcode must be 4 digits.")
		case err == passcode.ErrPasscodeLocked:
			h.respondWithUserError(c, http.StatusLocked, "PASSCODE_LOCKED", "Too many failed attempts. Please try again later.")
		case err == passcode.ErrPasscodeMismatch:
			h.respondWithUserError(c, http.StatusUnauthorized, "INVALID_PASSCODE", "Current passcode is incorrect.")
		case err == passcode.ErrPasscodeSameAsCurrent:
			h.respondWithUserError(c, http.StatusBadRequest, "PASSCODE_UNCHANGED", "New passcode must differ from the current one.")
		default:
			h.logger.Error("Failed to update passcode",
				zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("request_id", getRequestID(c)))
			h.respondWithInternalError(c, "PASSCODE_UPDATE_FAILED", "Failed to update passcode")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Passcode updated successfully",
		"status":  status,
	})
}

// VerifyPasscode validates the passcode and issues a short-lived session token
func (h *SecurityHandlers) VerifyPasscode(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := h.getUserID(c)
	if err != nil {
		h.respondWithUserError(c, http.StatusBadRequest, "INVALID_USER_ID", err.Error())
		return
	}

	var req entities.PasscodeVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithUserError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request payload")
		return
	}

	response, err := h.passcodeService.VerifyPasscode(ctx, userID, req.Passcode)
	if err != nil {
		switch {
		case err == passcode.ErrPasscodeNotSet:
			h.respondWithUserError(c, http.StatusBadRequest, "PASSCODE_NOT_SET", "No passcode configured yet.")
		case err == passcode.ErrPasscodeLocked:
			h.respondWithUserError(c, http.StatusLocked, "PASSCODE_LOCKED", "Too many failed attempts. Please try again later.")
		case err == passcode.ErrPasscodeMismatch:
			h.respondWithUserError(c, http.StatusUnauthorized, "INVALID_PASSCODE", "Passcode verification failed.")
		default:
			h.logger.Error("Failed to verify passcode",
				zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("request_id", getRequestID(c)))
			h.respondWithInternalError(c, "PASSCODE_VERIFY_FAILED", "Failed to verify passcode")
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// RemovePasscode disables the user's passcode
func (h *SecurityHandlers) RemovePasscode(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := h.getUserID(c)
	if err != nil {
		h.respondWithUserError(c, http.StatusBadRequest, "INVALID_USER_ID", err.Error())
		return
	}

	var req entities.PasscodeRemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithUserError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request payload")
		return
	}

	status, err := h.passcodeService.RemovePasscode(ctx, userID, req.Passcode)
	if err != nil {
		switch {
		case err == passcode.ErrPasscodeNotSet:
			h.respondWithUserError(c, http.StatusBadRequest, "PASSCODE_NOT_SET", "No passcode configured yet.")
		case err == passcode.ErrPasscodeLocked:
			h.respondWithUserError(c, http.StatusLocked, "PASSCODE_LOCKED", "Too many failed attempts. Please try again later.")
		case err == passcode.ErrPasscodeMismatch:
			h.respondWithUserError(c, http.StatusUnauthorized, "INVALID_PASSCODE", "Passcode verification failed.")
		default:
			h.logger.Error("Failed to remove passcode",
				zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("request_id", getRequestID(c)))
			h.respondWithInternalError(c, "PASSCODE_REMOVE_FAILED", "Failed to remove passcode")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Passcode removed successfully",
		"status":  status,
	})
}

// Helper methods

func (h *SecurityHandlers) getUserID(c *gin.Context) (uuid.UUID, error) {
	if userIDVal, exists := c.Get("user_id"); exists {
		switch v := userIDVal.(type) {
		case uuid.UUID:
			return v, nil
		case string:
			return uuid.Parse(v)
		}
	}

	userIDQuery := c.Query("user_id")
	if userIDQuery != "" {
		return uuid.Parse(userIDQuery)
	}

	return uuid.Nil, fmt.Errorf("user ID not found in context or query parameters")
}

func (h *SecurityHandlers) respondWithUserError(c *gin.Context, status int, code, message string) {
	c.JSON(status, entities.ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func (h *SecurityHandlers) respondWithInternalError(c *gin.Context, code, message string) {
	c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
		Code:    code,
		Message: message,
		Details: map[string]interface{}{"request_id": getRequestID(c)},
	})
}
