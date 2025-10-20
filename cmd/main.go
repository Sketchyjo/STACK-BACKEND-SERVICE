package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stack-service/stack_service/internal/api/routes"
	"github.com/stack-service/stack_service/internal/infrastructure/config"
	"github.com/stack-service/stack_service/internal/infrastructure/database"
	"github.com/stack-service/stack_service/internal/infrastructure/di"
	"github.com/stack-service/stack_service/internal/workers/funding_webhook"
	walletprovisioning "github.com/stack-service/stack_service/internal/workers/wallet_provisioning"
	"github.com/stack-service/stack_service/pkg/logger"

	"github.com/gin-gonic/gin"
)

// @title Stack Service API
// @version 1.0
// @description GenZ Web3 Multi-Chain Investment Platform API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.stackservice.com/support
// @contact.email support@stackservice.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel, cfg.Environment)

	// Initialize database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(cfg.Database.URL); err != nil {
		log.Fatal("Failed to run migrations", "error", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Build dependency injection container
	container, err := di.NewContainer(cfg, db, log)
	if err != nil {
		log.Fatal("Failed to create DI container", "error", err)
	}

	// Initialize router with DI container
	router := routes.SetupRoutes(container)

	// Initialize wallet provisioning worker and scheduler
	workerConfig := walletprovisioning.DefaultConfig()
	workerConfig.WalletSetNamePrefix = cfg.Circle.DefaultWalletSetName
	workerConfig.ChainsToProvision = container.WalletService.SupportedChains()
	workerConfig.DefaultWalletSetID = cfg.Circle.DefaultWalletSetID

	worker := walletprovisioning.NewWorker(
		container.WalletRepo,
		container.WalletSetRepo,
		container.WalletProvisioningJobRepo,
		container.CircleClient,
		container.AuditService,
		workerConfig,
		log.Zap(),
	)

	schedulerConfig := walletprovisioning.DefaultSchedulerConfig()
	scheduler := walletprovisioning.NewScheduler(
		worker,
		container.WalletProvisioningJobRepo,
		schedulerConfig,
		log.Zap(),
	)

	// Start the scheduler
	if err := scheduler.Start(); err != nil {
		log.Fatal("Failed to start wallet provisioning scheduler", "error", err)
	}
	log.Info("Wallet provisioning scheduler started")

	// Store scheduler in container for access by handlers
	container.WalletProvisioningScheduler = scheduler

	// Initialize funding webhook workers
	processorConfig := funding_webhook.DefaultProcessorConfig()
	reconciliationConfig := funding_webhook.DefaultReconciliationConfig()

	webhookManager, err := funding_webhook.NewManager(
		processorConfig,
		reconciliationConfig,
		container.FundingEventJobRepo,
		container.DepositRepo,
		container.FundingService,
		container.AuditService,
		log,
	)
	if err != nil {
		log.Fatal("Failed to create webhook manager", "error", err)
	}

	// Start the webhook manager
	if err := webhookManager.Start(context.Background()); err != nil {
		log.Fatal("Failed to start webhook manager", "error", err)
	}
	log.Info("Funding webhook workers started")

	// Store webhook manager in container for access by handlers
	container.FundingWebhookManager = webhookManager

	// Create server
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in goroutine
	go func() {
		log.Info("Starting server", "port", cfg.Server.Port, "environment", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Stop the wallet provisioning scheduler
	log.Info("Stopping wallet provisioning scheduler...")
	if err := scheduler.Stop(); err != nil {
		log.Warn("Error stopping scheduler", "error", err)
	}

	// Stop the funding webhook manager
	log.Info("Stopping funding webhook manager...")
	if webhookMgr, ok := container.FundingWebhookManager.(*funding_webhook.Manager); ok {
		if err := webhookMgr.Shutdown(30 * time.Second); err != nil {
			log.Warn("Error stopping webhook manager", "error", err)
		}
	}

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited")
}
