package funding

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/pkg/logger"
)

// Service handles funding operations - deposit addresses, confirmations, balance conversion
type Service struct {
	depositRepo DepositRepository
	balanceRepo BalanceRepository
	walletRepo  WalletRepository
	circleAPI   CircleAdapter
	logger      *logger.Logger
}

// DepositRepository interface for deposit persistence
type DepositRepository interface {
	Create(ctx context.Context, deposit *entities.Deposit) error
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Deposit, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, confirmedAt *time.Time) error
	GetByTxHash(ctx context.Context, txHash string) (*entities.Deposit, error)
}

// BalanceRepository interface for balance management
type BalanceRepository interface {
	Get(ctx context.Context, userID uuid.UUID) (*entities.Balance, error)
	UpdateBuyingPower(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error
	UpdatePendingDeposits(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error
}

// WalletRepository interface for wallet operations
type WalletRepository interface {
	GetByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.Chain) (*entities.Wallet, error)
	GetByAddress(ctx context.Context, address string) (*entities.Wallet, error)
	Create(ctx context.Context, wallet *entities.Wallet) error
}

// CircleAdapter interface for Circle API integration
type CircleAdapter interface {
	GenerateDepositAddress(ctx context.Context, chain entities.Chain, userID uuid.UUID) (string, error)
	ValidateDeposit(ctx context.Context, txHash string, amount decimal.Decimal) (bool, error)
	ConvertToUSD(ctx context.Context, amount decimal.Decimal, token entities.Stablecoin) (decimal.Decimal, error)
}

// NewService creates a new funding service
func NewService(
	depositRepo DepositRepository,
	balanceRepo BalanceRepository,
	walletRepo WalletRepository,
	circleAPI CircleAdapter,
	logger *logger.Logger,
) *Service {
	return &Service{
		depositRepo: depositRepo,
		balanceRepo: balanceRepo,
		walletRepo:  walletRepo,
		circleAPI:   circleAPI,
		logger:      logger,
	}
}

// CreateDepositAddress generates or retrieves deposit address for a chain
func (s *Service) CreateDepositAddress(ctx context.Context, userID uuid.UUID, chain entities.Chain) (*entities.DepositAddressResponse, error) {
	// Check if user already has a wallet for this chain
	wallet, err := s.walletRepo.GetByUserAndChain(ctx, userID, chain)
	if err != nil && err.Error() != "wallet not found" {
		return nil, fmt.Errorf("failed to check existing wallet: %w", err)
	}

	var address string
	if wallet != nil {
		address = wallet.Address
		s.logger.Info("Using existing wallet address", "user_id", userID, "chain", chain, "address", address)
	} else {
		// Generate new address through Circle
		address, err = s.circleAPI.GenerateDepositAddress(ctx, chain, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate deposit address: %w", err)
		}

		// Create wallet record
		wallet = &entities.Wallet{
			ID:          uuid.New(),
			UserID:      userID,
			Chain:       chain,
			Address:     address,
			ProviderRef: fmt.Sprintf("circle-%s", address),
			Status:      "active",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.walletRepo.Create(ctx, wallet); err != nil {
			return nil, fmt.Errorf("failed to create wallet record: %w", err)
		}

		s.logger.Info("Created new wallet address", "user_id", userID, "chain", chain, "address", address)
	}

	return &entities.DepositAddressResponse{
		Chain:   chain,
		Address: address,
		QRCode:  nil, // Could generate QR code URL here
	}, nil
}

// GetFundingConfirmations retrieves recent funding confirmations for user
func (s *Service) GetFundingConfirmations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.FundingConfirmation, error) {
	deposits, err := s.depositRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get deposits: %w", err)
	}

	confirmations := make([]*entities.FundingConfirmation, len(deposits))
	for i, deposit := range deposits {
		var confirmedAt time.Time
		if deposit.ConfirmedAt != nil {
			confirmedAt = *deposit.ConfirmedAt
		}
		confirmations[i] = &entities.FundingConfirmation{
			ID:          deposit.ID,
			Chain:       deposit.Chain,
			TxHash:      deposit.TxHash,
			Token:       deposit.Token,
			Amount:      deposit.Amount.String(),
			Status:      deposit.Status,
			ConfirmedAt: confirmedAt,
		}
	}

	return confirmations, nil
}

// GetBalance returns user's current balance
func (s *Service) GetBalance(ctx context.Context, userID uuid.UUID) (*entities.BalancesResponse, error) {
	balance, err := s.balanceRepo.Get(ctx, userID)
	if err != nil {
		if err.Error() == "balance not found" {
			// Return zero balance for new users
			return &entities.BalancesResponse{
				BuyingPower:     "0.00",
				PendingDeposits: "0.00",
				Currency:        "USD",
			}, nil
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &entities.BalancesResponse{
		BuyingPower:     balance.BuyingPower.String(),
		PendingDeposits: balance.PendingDeposits.String(),
		Currency:        balance.Currency,
	}, nil
}

// ProcessChainDeposit processes incoming chain deposit webhook
func (s *Service) ProcessChainDeposit(ctx context.Context, webhook *entities.ChainDepositWebhook) error {
	s.logger.Info("Processing chain deposit", "chain", webhook.Chain, "tx_hash", webhook.TxHash, "amount", webhook.Amount)

	// Validate the deposit with Circle
	amountFloat, err := strconv.ParseFloat(webhook.Amount, 64)
	if err != nil {
		return fmt.Errorf("invalid deposit amount %q: %w", webhook.Amount, err)
	}
	amount := decimal.NewFromFloat(amountFloat)
	isValid, err := s.circleAPI.ValidateDeposit(ctx, webhook.TxHash, amount)
	if err != nil {
		return fmt.Errorf("failed to validate deposit: %w", err)
	}

	if !isValid {
		s.logger.Warn("Invalid deposit received", "tx_hash", webhook.TxHash)
		return fmt.Errorf("invalid deposit signature or amount")
	}

	// Check if deposit already exists (idempotency check)
	existingDeposit, err := s.depositRepo.GetByTxHash(ctx, webhook.TxHash)
	if err != nil && err.Error() != "deposit not found" {
		return fmt.Errorf("failed to check existing deposit: %w", err)
	}

	if existingDeposit != nil {
		s.logger.Info("Deposit already processed", "tx_hash", webhook.TxHash)
		return nil
	}

	// Find the wallet to get user ID
	wallet, err := s.walletRepo.GetByAddress(ctx, webhook.Address)
	if err != nil {
		return fmt.Errorf("failed to find wallet for address %s: %w", webhook.Address, err)
	}

	// Convert stablecoin to USD buying power
	usdAmount, err := s.circleAPI.ConvertToUSD(ctx, amount, webhook.Token)
	if err != nil {
		return fmt.Errorf("failed to convert to USD: %w", err)
	}

	// Create deposit record
	deposit := &entities.Deposit{
		ID:          uuid.New(),
		UserID:      wallet.UserID,
		Chain:       webhook.Chain,
		TxHash:      webhook.TxHash,
		Token:       webhook.Token,
		Amount:      amount,
		Status:      "confirmed",
		ConfirmedAt: &webhook.BlockTime,
		CreatedAt:   time.Now(),
	}

	if err := s.depositRepo.Create(ctx, deposit); err != nil {
		return fmt.Errorf("failed to create deposit record: %w", err)
	}

	// Update user's buying power
	if err := s.balanceRepo.UpdateBuyingPower(ctx, wallet.UserID, usdAmount); err != nil {
		return fmt.Errorf("failed to update buying power: %w", err)
	}

	s.logger.Info("Deposit processed successfully",
		"user_id", wallet.UserID,
		"amount", webhook.Amount,
		"usd_amount", usdAmount.String(),
		"tx_hash", webhook.TxHash,
	)

	return nil
}
