package routes

import (
	// "database/sql"

	"github.com/stack-service/stack_service/internal/api/handlers"
	"github.com/stack-service/stack_service/internal/api/middleware"

	// "github.com/stack-service/stack_service/internal/infrastructure/config"
	"github.com/stack-service/stack_service/internal/infrastructure/di"
	// "github.com/stack-service/stack_service/pkg/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all application routes
func SetupRoutes(container *di.Container) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(container.Logger))
	router.Use(middleware.Recovery(container.Logger))
	router.Use(middleware.CORS(container.Config.Server.AllowedOrigins))
	router.Use(middleware.RateLimit(container.Config.Server.RateLimitPerMin))
	router.Use(middleware.SecurityHeaders())

	// Health check (no auth required)
	router.GET("/health", handlers.HealthCheck())
	router.GET("/metrics", handlers.Metrics())

	// Swagger documentation (development only)
	if container.Config.Environment != "production" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Initialize handlers with services from DI container
	fundingHandlers := handlers.NewFundingHandlers(container.GetFundingService(), container.Logger)
	// investingHandlers := handlers.NewInvestingHandlers(container.GetInvestingService(), container.Logger)
	onboardingHandlers := handlers.NewOnboardingHandlers(container.GetOnboardingService(), container.ZapLog)
	walletHandlers := handlers.NewWalletHandlers(container.GetWalletService(), container.ZapLog)

	// Initialize AI-CFO and ZeroG handlers
	aicfoHandlers := handlers.NewAICfoHandler(container.GetAICfoService(), container.ZapLog)
	zeroGHandlers := handlers.NewZeroGHandler(
		container.GetStorageClient(),
		container.GetInferenceGateway(),
		container.GetNamespaceManager(),
		container.ZapLog,
	)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		auth := v1.Group("/auth")
		{

			// New signup flow with verification
			authSignupHandlers := handlers.NewAuthSignupHandlers(
				container.DB,
				container.Config,
				container.ZapLog,
				*container.UserRepo,
				container.GetVerificationService(),
				*container.GetOnboardingJobService(),
			)
			auth.POST("/register", authSignupHandlers.Register)
			auth.POST("/verify-code", authSignupHandlers.VerifyCode)
			auth.POST("/resend-code", authSignupHandlers.ResendCode)

			auth.POST("/login", handlers.Login(container.DB, container.Config, container.Logger))
			auth.POST("/refresh", handlers.RefreshToken(container.DB, container.Config, container.Logger))
			auth.POST("/logout", handlers.Logout(container.DB, container.Config, container.Logger))
			auth.POST("/forgot-password", handlers.ForgotPassword(container.DB, container.Config, container.Logger))
			auth.POST("/reset-password", handlers.ResetPassword(container.DB, container.Config, container.Logger))
			auth.POST("/verify-email", handlers.VerifyEmail(container.DB, container.Config, container.Logger))
		}
  
		// Onboarding routes - OpenAPI spec compliant
		onboarding := v1.Group("/onboarding")
		onboarding.Use(middleware.MockAuthMiddleware()) // Use mock auth for development
		{
			onboarding.POST("/start", onboardingHandlers.StartOnboarding)
			onboarding.GET("/status", onboardingHandlers.GetOnboardingStatus)
			onboarding.POST("/kyc/submit", onboardingHandlers.SubmitKYC)
		}

		// KYC provider webhooks (no auth required for external callbacks)
		kyc := v1.Group("/kyc")
		{
			kyc.POST("/callback/:provider_ref", onboardingHandlers.ProcessKYCCallback)
		}

		// Protected routes (auth required)
		protected := v1.Group("/")
		protected.Use(middleware.Authentication(container.Config, container.Logger))
		{
			// User management
			users := protected.Group("/users")
			{
				users.GET("/me", handlers.GetProfile(container.DB, container.Config, container.Logger))
				users.PUT("/me", handlers.UpdateProfile(container.DB, container.Config, container.Logger))
				users.POST("/me/change-password", handlers.ChangePassword(container.DB, container.Config, container.Logger))
				users.DELETE("/me", handlers.DeleteAccount(container.DB, container.Config, container.Logger))
				users.POST("/me/enable-2fa", handlers.Enable2FA(container.DB, container.Config, container.Logger))
				users.POST("/me/disable-2fa", handlers.Disable2FA(container.DB, container.Config, container.Logger))
			}

			// Funding routes (OpenAPI spec compliant)
			funding := protected.Group("/funding")
			{
				funding.POST("/deposit/address", fundingHandlers.CreateDepositAddress)
				funding.GET("/confirmations", fundingHandlers.GetFundingConfirmations)
			}

			// Balance routes (part of funding but separate for clarity)
			protected.GET("/balances", fundingHandlers.GetBalances)

			// Investing routes (OpenAPI spec compliant) - Commented out until service is implemented
			// investing := protected.Group("/investing")
			// {
			// 	investing.GET("/baskets", investingHandlers.GetBaskets)
			// 	investing.GET("/baskets/:basketId", investingHandlers.GetBasket)
			// 	investing.POST("/orders", investingHandlers.CreateOrder)
			// 	investing.GET("/orders", investingHandlers.GetOrders)
			// 	investing.GET("/orders/:orderId", investingHandlers.GetOrder)
			// 	investing.GET("/portfolio", investingHandlers.GetPortfolio)
			// }

			// Wallet routes (OpenAPI spec compliant)
			wallet := protected.Group("/wallet")
			{
				wallet.GET("/addresses", walletHandlers.GetWalletAddresses)
				wallet.GET("/status", walletHandlers.GetWalletStatus)
			}

			// Investment baskets
			baskets := protected.Group("/baskets")
			{
				baskets.GET("/", handlers.GetBaskets(container.DB, container.Config, container.Logger))
				baskets.POST("/", handlers.CreateBasket(container.DB, container.Config, container.Logger))
				baskets.GET("/:id", handlers.GetBasket(container.DB, container.Config, container.Logger))
				baskets.PUT("/:id", handlers.UpdateBasket(container.DB, container.Config, container.Logger))
				baskets.DELETE("/:id", handlers.DeleteBasket(container.DB, container.Config, container.Logger))
				baskets.POST("/:id/invest", handlers.InvestInBasket(container.DB, container.Config, container.Logger))
				baskets.POST("/:id/withdraw", handlers.WithdrawFromBasket(container.DB, container.Config, container.Logger))
			}

			// Curated baskets (public/featured)
			curated := protected.Group("/curated")
			{
				curated.GET("/baskets", handlers.GetCuratedBaskets(container.DB, container.Config, container.Logger))
				curated.GET("/baskets/:id", handlers.GetCuratedBasket(container.DB, container.Config, container.Logger))
				curated.POST("/baskets/:id/invest", handlers.InvestInCuratedBasket(container.DB, container.Config, container.Logger))
			}

			// Copy trading
			copy := protected.Group("/copy")
			{
				copy.GET("/traders", handlers.GetTopTraders(container.DB, container.Config, container.Logger))
				copy.GET("/traders/:id", handlers.GetTrader(container.DB, container.Config, container.Logger))
				copy.POST("/traders/:id/follow", handlers.FollowTrader(container.DB, container.Config, container.Logger))
				copy.DELETE("/traders/:id/unfollow", handlers.UnfollowTrader(container.DB, container.Config, container.Logger))
				copy.GET("/following", handlers.GetFollowedTraders(container.DB, container.Config, container.Logger))
				copy.GET("/followers", handlers.GetFollowers(container.DB, container.Config, container.Logger))
			}

			// Cards and payments
			cards := protected.Group("/cards")
			{
				cards.GET("/", handlers.GetCards(container.DB, container.Config, container.Logger))
				cards.POST("/", handlers.CreateCard(container.DB, container.Config, container.Logger))
				cards.GET("/:id", handlers.GetCard(container.DB, container.Config, container.Logger))
				cards.PUT("/:id", handlers.UpdateCard(container.DB, container.Config, container.Logger))
				cards.DELETE("/:id", handlers.DeleteCard(container.DB, container.Config, container.Logger))
				cards.POST("/:id/freeze", handlers.FreezeCard(container.DB, container.Config, container.Logger))
				cards.POST("/:id/unfreeze", handlers.UnfreezeCard(container.DB, container.Config, container.Logger))
				cards.GET("/:id/transactions", handlers.GetCardTransactions(container.DB, container.Config, container.Logger))
			}

			// Transactions
			transactions := protected.Group("/transactions")
			{
				transactions.GET("/", handlers.GetTransactions(container.DB, container.Config, container.Logger))
				transactions.GET("/:id", handlers.GetTransaction(container.DB, container.Config, container.Logger))
				transactions.POST("/deposit", handlers.Deposit(container.DB, container.Config, container.Logger))
				transactions.POST("/withdraw", handlers.Withdraw(container.DB, container.Config, container.Logger))
				transactions.POST("/transfer", handlers.Transfer(container.DB, container.Config, container.Logger))
				transactions.POST("/swap", handlers.SwapTokens(container.DB, container.Config, container.Logger))
			}

			// Analytics and portfolio
			analytics := protected.Group("/analytics")
			{
				analytics.GET("/portfolio", handlers.GetPortfolioAnalytics(container.DB, container.Config, container.Logger))
				analytics.GET("/performance", handlers.GetPerformanceMetrics(container.DB, container.Config, container.Logger))
				analytics.GET("/allocation", handlers.GetAssetAllocation(container.DB, container.Config, container.Logger))
				analytics.GET("/history", handlers.GetPortfolioHistory(container.DB, container.Config, container.Logger))
			}

			// Notifications
			notifications := protected.Group("/notifications")
			{
				notifications.GET("/", handlers.GetNotifications(container.DB, container.Config, container.Logger))
				notifications.PUT("/:id/read", handlers.MarkNotificationRead(container.DB, container.Config, container.Logger))
				notifications.PUT("/read-all", handlers.MarkAllNotificationsRead(container.DB, container.Config, container.Logger))
				notifications.DELETE("/:id", handlers.DeleteNotification(container.DB, container.Config, container.Logger))
			}
		}

		// Admin routes (admin auth required)
		admin := v1.Group("/admin")
		admin.Use(middleware.Authentication(container.Config, container.Logger))
		admin.Use(middleware.AdminAuth(container.DB, container.Logger))
		{
			admin.GET("/users", handlers.GetAllUsers(container.DB, container.Config, container.Logger))
			admin.GET("/users/:id", handlers.GetUserByID(container.DB, container.Config, container.Logger))
			admin.PUT("/users/:id/status", handlers.UpdateUserStatus(container.DB, container.Config, container.Logger))
			admin.GET("/transactions", handlers.GetAllTransactions(container.DB, container.Config, container.Logger))
			admin.GET("/analytics/system", handlers.GetSystemAnalytics(container.DB, container.Config, container.Logger))
			admin.POST("/baskets/curated", handlers.CreateCuratedBasket(container.DB, container.Config, container.Logger))
			admin.PUT("/baskets/curated/:id", handlers.UpdateCuratedBasket(container.DB, container.Config, container.Logger))

			// Wallet admin routes
			admin.POST("/wallet/create", walletHandlers.CreateWalletsForUser)
			admin.POST("/wallet/retry-provisioning", walletHandlers.RetryWalletProvisioning)
			admin.GET("/wallet/health", walletHandlers.HealthCheck)
		}

		// Webhooks (external systems) - OpenAPI spec compliant
		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("/chain-deposit", fundingHandlers.ChainDepositWebhook)
			// webhooks.POST("/brokerage-fill", investingHandlers.BrokerageFillWebhook) // Commented out until service is implemented
			// Legacy webhook handlers (to be deprecated)
			webhooks.POST("/payment", handlers.PaymentWebhook(container.DB, container.Config, container.Logger))
			webhooks.POST("/blockchain", handlers.BlockchainWebhook(container.DB, container.Config, container.Logger))
			webhooks.POST("/cards", handlers.CardWebhook(container.DB, container.Config, container.Logger))
		}
	}

	// Setup ZeroG and AI routes
	SetupZeroGRoutes(router, zeroGHandlers, aicfoHandlers, container.ZapLog)

	return router
}
