package routes

import (
	"database/sql"

	"github.com/stack-service/stack_service/internal/api/handlers"
	"github.com/stack-service/stack_service/internal/api/middleware"
	"github.com/stack-service/stack_service/internal/infrastructure/config"
	"github.com/stack-service/stack_service/pkg/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupStackRoutes configures STACK MVP routes matching OpenAPI specification
func SetupStackRoutes(db *sql.DB, cfg *config.Config, log *logger.Logger) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(log))
	router.Use(middleware.Recovery(log))
	router.Use(middleware.CORS(cfg.Server.AllowedOrigins))
	router.Use(middleware.RateLimit(cfg.Server.RateLimitPerMin))
	router.Use(middleware.SecurityHeaders())

	// Health check (no auth required)
	router.GET("/health", handlers.HealthCheck())
	router.GET("/metrics", handlers.Metrics())

	// Swagger documentation (development only)
	if cfg.Environment != "production" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Initialize STACK handlers
	// TODO: Wire up actual service dependencies
	stackHandlers := handlers.NewStackHandlers(
		nil, // funding service - needs implementation
		nil, // investing service - needs implementation
		log,
	)

	// API v1 routes matching OpenAPI specification
	v1 := router.Group("/v1")
	{
		// === FUNDING ENDPOINTS ===
		funding := v1.Group("/funding")
		funding.Use(middleware.Authentication(cfg, log)) // JWT required
		{
			funding.POST("/deposit/address", stackHandlers.CreateDepositAddress)
			funding.GET("/confirmations", stackHandlers.ListFundingConfirmations)
		}

		// Balance endpoint (separate from funding per OpenAPI)
		balances := v1.Group("/balances")
		balances.Use(middleware.Authentication(cfg, log))
		{
			balances.GET("", stackHandlers.GetBalances)
		}

		// === INVESTING ENDPOINTS ===
		baskets := v1.Group("/baskets")
		baskets.Use(middleware.Authentication(cfg, log))
		{
			baskets.GET("", stackHandlers.ListBaskets)
			baskets.GET("/:id", stackHandlers.GetBasket)
		}

		orders := v1.Group("/orders")
		orders.Use(middleware.Authentication(cfg, log))
		{
			orders.POST("", stackHandlers.CreateOrder)
			orders.GET("", stackHandlers.ListOrders)
			orders.GET("/:id", stackHandlers.GetOrder)
		}

		portfolio := v1.Group("/portfolio")
		portfolio.Use(middleware.Authentication(cfg, log))
		{
			portfolio.GET("", stackHandlers.GetPortfolio)
		}

		// === WEBHOOK ENDPOINTS (No auth - validated via signature) ===
		webhooks := v1.Group("/webhooks")
		{
			// Chain deposit webhook (from Circle, blockchain nodes, etc.)
			webhooks.POST("/chain-deposit", stackHandlers.ChainDepositWebhook)

			// Brokerage fill webhook (from brokerage partner)
			webhooks.POST("/brokerage-fills", stackHandlers.BrokerageFillWebhook)
		}

		// === AUTHENTICATION ENDPOINTS ===
		// NOTE: These are not in the MVP OpenAPI spec but needed for a complete system
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.Login(db, cfg, log))
			auth.POST("/refresh", handlers.RefreshToken(db, cfg, log))
			auth.POST("/logout", handlers.Logout(db, cfg, log))
		}
	}

	return router
}
