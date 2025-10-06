package funding

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockDepositRepository struct {
	mock.Mock
}

func (m *MockDepositRepository) Create(ctx context.Context, deposit *entities.Deposit) error {
	args := m.Called(ctx, deposit)
	return args.Error(0)
}

func (m *MockDepositRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Deposit, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entities.Deposit), args.Error(1)
}

func (m *MockDepositRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, confirmedAt *time.Time) error {
	args := m.Called(ctx, id, status, confirmedAt)
	return args.Error(0)
}

func (m *MockDepositRepository) GetByTxHash(ctx context.Context, txHash string) (*entities.Deposit, error) {
	args := m.Called(ctx, txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Deposit), args.Error(1)
}

type MockBalanceRepository struct {
	mock.Mock
}

func (m *MockBalanceRepository) Get(ctx context.Context, userID uuid.UUID) (*entities.Balance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Balance), args.Error(1)
}

func (m *MockBalanceRepository) UpdateBuyingPower(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

func (m *MockBalanceRepository) UpdatePendingDeposits(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) GetByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.Chain) (*entities.Wallet, error) {
	args := m.Called(ctx, userID, chain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetByAddress(ctx context.Context, address string) (*entities.Wallet, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Wallet), args.Error(1)
}

func (m *MockWalletRepository) Create(ctx context.Context, wallet *entities.Wallet) error {
	args := m.Called(ctx, wallet)
	return args.Error(0)
}

type MockCircleAdapter struct {
	mock.Mock
}

func (m *MockCircleAdapter) GenerateDepositAddress(ctx context.Context, chain entities.Chain, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, chain, userID)
	return args.String(0), args.Error(1)
}

func (m *MockCircleAdapter) ValidateDeposit(ctx context.Context, txHash string, amount decimal.Decimal) (bool, error) {
	args := m.Called(ctx, txHash, amount)
	return args.Bool(0), args.Error(1)
}

func (m *MockCircleAdapter) ConvertToUSD(ctx context.Context, amount decimal.Decimal, token entities.Stablecoin) (decimal.Decimal, error) {
	args := m.Called(ctx, amount, token)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

// Test helpers
func createTestService() (*Service, *MockDepositRepository, *MockBalanceRepository, *MockWalletRepository, *MockCircleAdapter) {
	depositRepo := new(MockDepositRepository)
	balanceRepo := new(MockBalanceRepository)
	walletRepo := new(MockWalletRepository)
	circleAPI := new(MockCircleAdapter)
	logger := logger.New("debug", "test")

	service := NewService(depositRepo, balanceRepo, walletRepo, circleAPI, logger)
	return service, depositRepo, balanceRepo, walletRepo, circleAPI
}

func TestCreateDepositAddress_ExistingWallet(t *testing.T) {
	service, _, _, walletRepo, _ := createTestService()

	ctx := context.Background()
	userID := uuid.New()
	chain := entities.ChainSolana

	existingWallet := &entities.Wallet{
		ID:      uuid.New(),
		UserID:  userID,
		Chain:   chain,
		Address: "So1test12345",
		Status:  "active",
	}

	walletRepo.On("GetByUserAndChain", ctx, userID, chain).Return(existingWallet, nil)

	result, err := service.CreateDepositAddress(ctx, userID, chain)

	assert.NoError(t, err)
	assert.Equal(t, chain, result.Chain)
	assert.Equal(t, existingWallet.Address, result.Address)
	walletRepo.AssertExpectations(t)
}

func TestCreateDepositAddress_NewWallet(t *testing.T) {
	service, _, _, walletRepo, circleAPI := createTestService()

	ctx := context.Background()
	userID := uuid.New()
	chain := entities.ChainPolygon
	expectedAddress := "0x1234567890abcdef"

	walletRepo.On("GetByUserAndChain", ctx, userID, chain).Return(nil, &MockError{"wallet not found"})
	circleAPI.On("GenerateDepositAddress", ctx, chain, userID).Return(expectedAddress, nil)
	walletRepo.On("Create", ctx, mock.MatchedBy(func(wallet *entities.Wallet) bool {
		return wallet.UserID == userID && wallet.Chain == chain && wallet.Address == expectedAddress
	})).Return(nil)

	result, err := service.CreateDepositAddress(ctx, userID, chain)

	assert.NoError(t, err)
	assert.Equal(t, chain, result.Chain)
	assert.Equal(t, expectedAddress, result.Address)
	walletRepo.AssertExpectations(t)
	circleAPI.AssertExpectations(t)
}

func TestGetBalance_ExistingBalance(t *testing.T) {
	service, _, balanceRepo, _, _ := createTestService()

	ctx := context.Background()
	userID := uuid.New()

	existingBalance := &entities.Balance{
		UserID:          userID,
		BuyingPower:     decimal.NewFromFloat(100.50),
		PendingDeposits: decimal.NewFromFloat(25.25),
		Currency:        "USD",
	}

	balanceRepo.On("Get", ctx, userID).Return(existingBalance, nil)

	result, err := service.GetBalance(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, "100.5", result.BuyingPower)
	assert.Equal(t, "25.25", result.PendingDeposits)
	assert.Equal(t, "USD", result.Currency)
	balanceRepo.AssertExpectations(t)
}

func TestGetBalance_NonExistingBalance(t *testing.T) {
	service, _, balanceRepo, _, _ := createTestService()

	ctx := context.Background()
	userID := uuid.New()

	balanceRepo.On("Get", ctx, userID).Return(nil, &MockError{"balance not found"})

	result, err := service.GetBalance(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, "0.00", result.BuyingPower)
	assert.Equal(t, "0.00", result.PendingDeposits)
	assert.Equal(t, "USD", result.Currency)
	balanceRepo.AssertExpectations(t)
}

func TestProcessChainDeposit_Success(t *testing.T) {
	service, depositRepo, balanceRepo, walletRepo, circleAPI := createTestService()

	ctx := context.Background()
	userID := uuid.New()
	wallet := &entities.Wallet{
		ID:      uuid.New(),
		UserID:  userID,
		Chain:   entities.ChainSolana,
		Address: "So1test12345",
	}

	webhook := &entities.ChainDepositWebhook{
		Chain:     entities.ChainSolana,
		Address:   wallet.Address,
		Token:     entities.StablecoinUSDC,
		Amount:    "100.0",
		TxHash:    "tx12345",
		BlockTime: time.Now(),
	}

	amount := decimal.NewFromFloat(100.0)
	usdAmount := decimal.NewFromFloat(100.0)

	circleAPI.On("ValidateDeposit", ctx, webhook.TxHash, amount).Return(true, nil)
	depositRepo.On("GetByTxHash", ctx, webhook.TxHash).Return(nil, &MockError{"deposit not found"})
	walletRepo.On("GetByAddress", ctx, webhook.Address).Return(wallet, nil)
	circleAPI.On("ConvertToUSD", ctx, amount, webhook.Token).Return(usdAmount, nil)
	depositRepo.On("Create", ctx, mock.MatchedBy(func(deposit *entities.Deposit) bool {
		return deposit.UserID == userID && deposit.TxHash == webhook.TxHash
	})).Return(nil)
	balanceRepo.On("UpdateBuyingPower", ctx, userID, usdAmount).Return(nil)

	err := service.ProcessChainDeposit(ctx, webhook)

	assert.NoError(t, err)
	circleAPI.AssertExpectations(t)
	depositRepo.AssertExpectations(t)
	walletRepo.AssertExpectations(t)
	balanceRepo.AssertExpectations(t)
}

func TestProcessChainDeposit_DuplicateDeposit(t *testing.T) {
	service, depositRepo, _, _, circleAPI := createTestService()

	ctx := context.Background()
	webhook := &entities.ChainDepositWebhook{
		Chain:     entities.ChainSolana,
		Address:   "So1test12345",
		Token:     entities.StablecoinUSDC,
		Amount:    "100.0",
		TxHash:    "tx12345",
		BlockTime: time.Now(),
	}

	amount := decimal.NewFromFloat(100.0)
	existingDeposit := &entities.Deposit{
		ID:     uuid.New(),
		TxHash: webhook.TxHash,
	}

	circleAPI.On("ValidateDeposit", ctx, webhook.TxHash, amount).Return(true, nil)
	depositRepo.On("GetByTxHash", ctx, webhook.TxHash).Return(existingDeposit, nil)

	err := service.ProcessChainDeposit(ctx, webhook)

	assert.NoError(t, err) // Should not error on duplicate
	circleAPI.AssertExpectations(t)
	depositRepo.AssertExpectations(t)
}

func TestProcessChainDeposit_InvalidDeposit(t *testing.T) {
	service, depositRepo, _, _, circleAPI := createTestService()

	ctx := context.Background()
	webhook := &entities.ChainDepositWebhook{
		Chain:     entities.ChainSolana,
		Address:   "So1test12345",
		Token:     entities.StablecoinUSDC,
		Amount:    "100.0",
		TxHash:    "invalid12345",
		BlockTime: time.Now(),
	}

	amount := decimal.NewFromFloat(100.0)

	circleAPI.On("ValidateDeposit", ctx, webhook.TxHash, amount).Return(false, nil)

	err := service.ProcessChainDeposit(ctx, webhook)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid deposit signature")
	circleAPI.AssertExpectations(t)
	depositRepo.AssertNotCalled(t, "GetByTxHash")
}

func TestGetFundingConfirmations(t *testing.T) {
	service, depositRepo, _, _, _ := createTestService()

	ctx := context.Background()
	userID := uuid.New()
	limit := 20
	offset := 0

	confirmedAt := time.Now()
	deposits := []*entities.Deposit{
		{
			ID:          uuid.New(),
			UserID:      userID,
			Chain:       entities.ChainSolana,
			TxHash:      "tx1",
			Token:       entities.StablecoinUSDC,
			Amount:      decimal.NewFromFloat(100.0),
			Status:      "confirmed",
			ConfirmedAt: &confirmedAt,
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			Chain:       entities.ChainPolygon,
			TxHash:      "tx2",
			Token:       entities.StablecoinUSDC,
			Amount:      decimal.NewFromFloat(50.0),
			Status:      "confirmed",
			ConfirmedAt: &confirmedAt,
		},
	}

	depositRepo.On("GetByUserID", ctx, userID, limit, offset).Return(deposits, nil)

	result, err := service.GetFundingConfirmations(ctx, userID, limit, offset)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, deposits[0].ID, result[0].ID)
	assert.Equal(t, "100", result[0].Amount)
	assert.Equal(t, deposits[1].ID, result[1].ID)
	assert.Equal(t, "50", result[1].Amount)
	depositRepo.AssertExpectations(t)
}

// MockError implements error interface for testing
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}
