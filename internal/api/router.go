package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/seller-assistant/internal/api/handlers"
	"github.com/yourusername/seller-assistant/internal/api/middleware"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/internal/service"
	"github.com/yourusername/seller-assistant/pkg/crypto"
)

// RouterConfig holds dependencies for router setup
type RouterConfig struct {
	UserRepo           domain.UserRepository
	KaspiKeyRepo       domain.KaspiKeyRepository
	ProductRepo        domain.ProductRepository
	ReviewRepo         domain.ReviewRepository
	AIResponder        *service.AIResponderService
	SyncService        *service.KaspiSyncService
	Encryptor          *crypto.Encryptor
	JWTSecret          string
	JWTExpirationHours int
}

// SetupRouter creates and configures the Gin router
func SetupRouter(cfg *RouterConfig) *gin.Engine {
	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Initialize handlers
		authHandler := handlers.NewAuthHandler(cfg.UserRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
		userHandler := handlers.NewUserHandler(cfg.UserRepo)
		kaspiKeyHandler := handlers.NewKaspiKeyHandler(cfg.KaspiKeyRepo, cfg.Encryptor, cfg.SyncService)
		productHandler := handlers.NewProductHandler(cfg.ProductRepo, nil) // Price dumping disabled
		reviewHandler := handlers.NewReviewHandler(cfg.ReviewRepo, cfg.AIResponder)
		dashboardHandler := handlers.NewDashboardHandler(cfg.ProductRepo, cfg.ReviewRepo)

		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes (auth required)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Auth endpoints
			protected.GET("/auth/me", authHandler.GetMe)

			// User endpoints
			user := protected.Group("/user")
			{
				user.GET("/profile", userHandler.GetProfile)
				user.PATCH("/settings", userHandler.UpdateSettings)
			}

			// Kaspi key endpoints
			kaspiKey := protected.Group("/kaspi-key")
			{
				kaspiKey.GET("", kaspiKeyHandler.GetKey)
				kaspiKey.POST("", kaspiKeyHandler.CreateKey)
				kaspiKey.DELETE("", kaspiKeyHandler.DeleteKey)
				kaspiKey.POST("/sync", kaspiKeyHandler.SyncNow)
			}

			// Product endpoints
			products := protected.Group("/products")
			{
				products.GET("", productHandler.GetProducts)
				products.GET("/low-stock", productHandler.GetLowStockProducts)
				// Temporarily disabled price dumping
				// products.GET("/dumping", productHandler.GetDumpingProducts)
				products.GET("/:id", productHandler.GetProduct)
				// products.POST("/:id/dumping/enable", productHandler.EnableDumping)
				// products.POST("/:id/dumping/disable", productHandler.DisableDumping)
			}

			// Review endpoints
			reviews := protected.Group("/reviews")
			{
				reviews.GET("", reviewHandler.GetReviews)
				reviews.GET("/pending", reviewHandler.GetPendingReviews)
				reviews.GET("/:id", reviewHandler.GetReview)
				reviews.POST("/:id/generate-reply", reviewHandler.GenerateReply)
				reviews.PATCH("/:id/reply", reviewHandler.UpdateReply)
			}

			// Dashboard endpoints
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/stats", dashboardHandler.GetStats)
				dashboard.GET("/overview", dashboardHandler.GetOverview)
			}
		}
	}

	return router
}
