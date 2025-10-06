package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stack-service/stack_service/internal/domain/entities"
	"go.uber.org/zap"
)

// DepositRepository handles deposit persistence operations
type DepositRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewDepositRepository creates a new deposit repository
func NewDepositRepository(db *sql.DB, logger *zap.Logger) *DepositRepository {
	return &DepositRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new deposit record
func (r *DepositRepository) Create(ctx context.Context, deposit *entities.Deposit) error {
	query := `
		INSERT INTO deposits (id, user_id, chain, tx_hash, token, amount, status, confirmed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		deposit.ID,
		deposit.UserID,
		deposit.Chain,
		deposit.TxHash,
		deposit.Token,
		deposit.Amount,
		deposit.Status,
		deposit.ConfirmedAt,
		deposit.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create deposit",
			zap.Error(err),
			zap.String("deposit_id", deposit.ID.String()),
			zap.String("tx_hash", deposit.TxHash),
		)
		return fmt.Errorf("failed to create deposit: %w", err)
	}

	r.logger.Info("deposit created successfully",
		zap.String("deposit_id", deposit.ID.String()),
		zap.String("user_id", deposit.UserID.String()),
		zap.String("tx_hash", deposit.TxHash),
		zap.String("amount", deposit.Amount.String()),
	)

	return nil
}

// GetByUserID retrieves deposits for a specific user with pagination
func (r *DepositRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Deposit, error) {
	query := `
		SELECT id, user_id, chain, tx_hash, token, amount, status, confirmed_at, created_at
		FROM deposits
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error("failed to query deposits by user ID",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return nil, fmt.Errorf("failed to query deposits: %w", err)
	}
	defer rows.Close()

	var deposits []*entities.Deposit
	for rows.Next() {
		deposit := &entities.Deposit{}
		err := rows.Scan(
			&deposit.ID,
			&deposit.UserID,
			&deposit.Chain,
			&deposit.TxHash,
			&deposit.Token,
			&deposit.Amount,
			&deposit.Status,
			&deposit.ConfirmedAt,
			&deposit.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan deposit row",
				zap.Error(err),
				zap.String("user_id", userID.String()),
			)
			return nil, fmt.Errorf("failed to scan deposit: %w", err)
		}
		deposits = append(deposits, deposit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return deposits, nil
}

// UpdateStatus updates the status of a deposit
func (r *DepositRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, confirmedAt *time.Time) error {
	query := `
		UPDATE deposits 
		SET status = $2, confirmed_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, status, confirmedAt)
	if err != nil {
		r.logger.Error("failed to update deposit status",
			zap.Error(err),
			zap.String("deposit_id", id.String()),
			zap.String("status", status),
		)
		return fmt.Errorf("failed to update deposit status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("deposit not found: %s", id.String())
	}

	r.logger.Info("deposit status updated",
		zap.String("deposit_id", id.String()),
		zap.String("status", status),
	)

	return nil
}

// GetByTxHash retrieves a deposit by transaction hash
func (r *DepositRepository) GetByTxHash(ctx context.Context, txHash string) (*entities.Deposit, error) {
	query := `
		SELECT id, user_id, chain, tx_hash, token, amount, status, confirmed_at, created_at
		FROM deposits
		WHERE tx_hash = $1
	`

	deposit := &entities.Deposit{}
	err := r.db.QueryRowContext(ctx, query, txHash).Scan(
		&deposit.ID,
		&deposit.UserID,
		&deposit.Chain,
		&deposit.TxHash,
		&deposit.Token,
		&deposit.Amount,
		&deposit.Status,
		&deposit.ConfirmedAt,
		&deposit.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deposit not found")
		}
		r.logger.Error("failed to get deposit by tx hash",
			zap.Error(err),
			zap.String("tx_hash", txHash),
		)
		return nil, fmt.Errorf("failed to get deposit: %w", err)
	}

	return deposit, nil
}
