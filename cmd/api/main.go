package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/yourusername/seller-assistant/internal/api"
	"github.com/yourusername/seller-assistant/internal/api/middleware"
	"github.com/yourusername/seller-assistant/internal/config"
	"github.com/yourusername/seller-assistant/internal/repository/mongodb"
	"github.com/yourusername/seller-assistant/internal/service"
	"github.com/yourusername/seller-assistant/pkg/crypto"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	if err := logger.Init(cfg.LogLevel); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting Kaspi Seller Assistant API Server",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
	)

	// Initialize MongoDB
	db, err := mongodb.NewDB(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		logger.Log.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer db.Close()

	userRepo := mongodb.NewUserRepository(db)
	kaspiKeyRepo := mongodb.NewKaspiKeyRepository(db)
	productRepo := mongodb.NewProductRepository(db)
	reviewRepo := mongodb.NewReviewRepository(db)

	// Initialize encryptor
	encryptor, err := crypto.NewEncryptor(cfg.EncryptionKey)
	if err != nil {
		logger.Log.Fatal("Failed to initialize encryptor", zap.Error(err))
	}

	// Initialize JWT middleware
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-change-in-production"
		logger.Log.Warn("Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}
	middleware.InitJWTSecret(jwtSecret)

	// Initialize services
	aiResponder := service.NewAIResponderService(cfg.OpenAIAPIKey, reviewRepo)
	// priceDumpingService := service.NewPriceDumpingService(kaspiKeyRepo, productRepo, encryptor) // Temporarily disabled

	// Setup router
	routerCfg := &api.RouterConfig{
		UserRepo:     userRepo,
		KaspiKeyRepo: kaspiKeyRepo,
		ProductRepo:  productRepo,
		ReviewRepo:   reviewRepo,
		// PriceDumpingService: priceDumpingService, // Temporarily disabled
		AIResponder: aiResponder,
		Encryptor:   encryptor,
	}

	router := api.SetupRouter(routerCfg)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	logger.Log.Info("API Server started successfully",
		zap.String("port", port),
		zap.String("base_url", fmt.Sprintf("http://localhost:%s", port)),
		zap.String("api_version", "v1"),
	)

	logger.Log.Info("Available endpoints:",
		zap.String("health", "GET /health"),
		zap.String("login", "POST /api/v1/auth/login"),
		zap.String("profile", "GET /api/v1/user/profile (auth)"),
		zap.String("products", "GET /api/v1/products (auth)"),
		zap.String("reviews", "GET /api/v1/reviews (auth)"),
		zap.String("dashboard", "GET /api/v1/dashboard/stats (auth)"),
	)

	if err := router.Run(":" + port); err != nil {
		logger.Log.Fatal("Failed to start server", zap.Error(err))
	}
}
