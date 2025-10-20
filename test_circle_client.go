package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stack-service/stack_service/internal/domain/entities"
	"github.com/stack-service/stack_service/internal/infrastructure/circle"
	"github.com/stack-service/stack_service/internal/infrastructure/config"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Create Circle client configuration
	circleConfig := circle.Config{
		APIKey:                 cfg.Circle.APIKey,
		BaseURL:                cfg.Circle.BaseURL,
		Environment:            cfg.Circle.Environment,
		Timeout:                30 * time.Second,
		EntitySecretCiphertext: cfg.Circle.EntitySecretCiphertext,
		WalletSetsEndpoint:     "/v1/w3s/developer/walletSets",
		WalletsEndpoint:        "/v1/w3s/developer/wallets",
		PublicKeyEndpoint:      "/v1/w3s/config/entity/publicKey",
		BalancesEndpoint:       "/v1/w3s/wallets",
		TransferEndpoint:       "/v1/w3s/developer/transactions/transfer",
	}

	// Create Circle client
	client := circle.NewClient(circleConfig, logger)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Println("🚀 Starting Circle Client Test Suite")
	fmt.Println("=====================================")

	// Test 1: Health Check
	fmt.Println("\n1. Testing Health Check...")
	if err := testHealthCheck(ctx, client, logger); err != nil {
		logger.Error("Health check failed", zap.Error(err))
	} else {
		fmt.Println("✅ Health check passed")
	}

	// Test 2: Get Entity Public Key
	fmt.Println("\n2. Testing Get Entity Public Key...")
	publicKey, err := testGetEntityPublicKey(ctx, client, logger)
	if err != nil {
		logger.Error("Get entity public key failed", zap.Error(err))
	} else {
		fmt.Printf("✅ Entity public key retrieved: %s...\n", publicKey[:20])
	}

	// Test 3: Create Wallet Set
	fmt.Println("\n3. Testing Create Wallet Set...")
	walletSetID, err := testCreateWalletSet(ctx, client, logger)
	if err != nil {
		logger.Error("Create wallet set failed", zap.Error(err))
	} else {
		fmt.Printf("✅ Wallet set created with ID: %s\n", walletSetID)
	}

	// Test 4: Get Wallet Set
	if walletSetID != "" {
		fmt.Println("\n4. Testing Get Wallet Set...")
		if err := testGetWalletSet(ctx, client, walletSetID, logger); err != nil {
			logger.Error("Get wallet set failed", zap.Error(err))
		} else {
			fmt.Println("✅ Wallet set retrieved successfully")
		}
	}

	// Test 5: Create Wallet
	if walletSetID != "" {
		fmt.Println("\n5. Testing Create Wallet...")
		walletID, err := testCreateWallet(ctx, client, walletSetID, logger)
		if err != nil {
			logger.Error("Create wallet failed", zap.Error(err))
		} else {
			fmt.Printf("✅ Wallet created with ID: %s\n", walletID)
		}

		// Test 6: Get Wallet
		if walletID != "" {
			fmt.Println("\n6. Testing Get Wallet...")
			if err := testGetWallet(ctx, client, walletID, logger); err != nil {
				logger.Error("Get wallet failed", zap.Error(err))
			} else {
				fmt.Println("✅ Wallet retrieved successfully")
			}

			// Test 7: Get Wallet Balances
			fmt.Println("\n7. Testing Get Wallet Balances...")
			if err := testGetWalletBalances(ctx, client, walletID, logger); err != nil {
				logger.Error("Get wallet balances failed", zap.Error(err))
			} else {
				fmt.Println("✅ Wallet balances retrieved successfully")
			}

			// Test 8: Transfer Funds (if wallet has funds)
			fmt.Println("\n8. Testing Transfer Funds...")
			if err := testTransferFunds(ctx, client, walletID, logger); err != nil {
				logger.Error("Transfer funds failed", zap.Error(err))
			} else {
				fmt.Println("✅ Transfer funds completed")
			}
		}
	}

	// Test 9: MVP Simulation Functions
	fmt.Println("\n9. Testing MVP Simulation Functions...")
	if err := testMVPSimulationFunctions(ctx, client, logger); err != nil {
		logger.Error("MVP simulation functions failed", zap.Error(err))
	} else {
		fmt.Println("✅ MVP simulation functions completed")
	}

	// Test 10: Get Metrics
	fmt.Println("\n10. Testing Get Metrics...")
	if err := testGetMetrics(client, logger); err != nil {
		logger.Error("Get metrics failed", zap.Error(err))
	} else {
		fmt.Println("✅ Metrics retrieved successfully")
	}

	fmt.Println("\n🎉 Circle Client Test Suite Completed!")
}

func testHealthCheck(ctx context.Context, client *circle.Client, logger *zap.Logger) error {
	return client.HealthCheck(ctx)
}

func testGetEntityPublicKey(ctx context.Context, client *circle.Client, logger *zap.Logger) (string, error) {
	return client.GetEntityPublicKey(ctx)
}

func testCreateWalletSet(ctx context.Context, client *circle.Client, logger *zap.Logger) (string, error) {
	walletSetName := fmt.Sprintf("test-wallet-set-%d", time.Now().Unix())

	response, err := client.CreateWalletSet(ctx, walletSetName, "")
	if err != nil {
		return "", err
	}

	return response.WalletSet.ID, nil
}

func testGetWalletSet(ctx context.Context, client *circle.Client, walletSetID string, logger *zap.Logger) error {
	_, err := client.GetWalletSet(ctx, walletSetID)
	return err
}

func testCreateWallet(ctx context.Context, client *circle.Client, walletSetID string, logger *zap.Logger) (string, error) {
	req := entities.CircleWalletCreateRequest{
		Blockchains: []string{"ETH-SEPOLIA", "MATIC-AMOY"},
		Count:       1,
		AccountType: "EOA",
		WalletSetID: walletSetID,
	}

	response, err := client.CreateWallet(ctx, req)
	if err != nil {
		return "", err
	}

	return response.Wallet.ID, nil
}

func testGetWallet(ctx context.Context, client *circle.Client, walletID string, logger *zap.Logger) error {
	_, err := client.GetWallet(ctx, walletID)
	return err
}

func testGetWalletBalances(ctx context.Context, client *circle.Client, walletID string, logger *zap.Logger) error {
	balances, err := client.GetWalletBalances(ctx, walletID)
	if err != nil {
		return err
	}

	logger.Info("Wallet balances retrieved", zap.Any("balances", balances))
	return nil
}

func testTransferFunds(ctx context.Context, client *circle.Client, walletID string, logger *zap.Logger) error {
	// Note: This will likely fail in sandbox without actual funds, but tests the API call
	req := entities.CircleTransferRequest{
		IDempotencyKey:     uuid.NewString(),
		WalletID:           walletID,
		TokenID:            "USDC", // Assuming USDC token ID
		Amounts:            []string{"1.00"},
		DestinationAddress: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6", // Example address
	}

	_, err := client.TransferFunds(ctx, req)
	return err
}

func testMVPSimulationFunctions(ctx context.Context, client *circle.Client, logger *zap.Logger) error {
	// Test GenerateDepositAddress
	fmt.Println("  - Testing GenerateDepositAddress...")
	address, err := client.GenerateDepositAddress(ctx, entities.ChainETH, uuid.New())
	if err != nil {
		return fmt.Errorf("generate deposit address failed: %w", err)
	}
	fmt.Printf("    Generated address: %s\n", address)

	// Test ValidateDeposit
	fmt.Println("  - Testing ValidateDeposit...")
	valid, err := client.ValidateDeposit(ctx, "0x1234567890abcdef", decimal.NewFromFloat(100.0))
	if err != nil {
		return fmt.Errorf("validate deposit failed: %w", err)
	}
	fmt.Printf("    Deposit validation result: %t\n", valid)

	// Test ConvertToUSD
	fmt.Println("  - Testing ConvertToUSD...")
	usdAmount, err := client.ConvertToUSD(ctx, decimal.NewFromFloat(100.0), entities.StablecoinUSDC)
	if err != nil {
		return fmt.Errorf("convert to USD failed: %w", err)
	}
	fmt.Printf("    USD amount: %s\n", usdAmount.String())

	return nil
}

func testGetMetrics(client *circle.Client, logger *zap.Logger) error {
	metrics := client.GetMetrics()
	logger.Info("Circuit breaker metrics", zap.Any("metrics", metrics))
	return nil
}
