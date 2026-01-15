package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/seller-assistant/internal/config"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/internal/repository/mongodb"
	"github.com/yourusername/seller-assistant/internal/service"
	"github.com/yourusername/seller-assistant/pkg/crypto"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"github.com/yourusername/seller-assistant/pkg/scheduler"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	if err := logger.Init(cfg.LogLevel); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting Seller Assistant Worker",
		zap.String("environment", cfg.Environment),
		zap.Int("sync_interval_hours", cfg.SyncIntervalHours),
	)

	// Initialize MongoDB
	db, err := mongodb.NewDB(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		logger.Log.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer db.Close()

	logger.Log.Info("MongoDB connected successfully")

	// Initialize encryptor
	encryptor, err := crypto.NewEncryptor(cfg.EncryptionKey)
	if err != nil {
		logger.Log.Fatal("Failed to initialize encryptor", zap.Error(err))
	}

	// Initialize repositories
	userRepo := mongodb.NewUserRepository(db)
	kaspiKeyRepo := mongodb.NewKaspiKeyRepository(db)
	productRepo := mongodb.NewProductRepository(db)
	salesHistoryRepo := mongodb.NewSalesHistoryRepository(db)
	reviewRepo := mongodb.NewReviewRepository(db)
	lowStockAlertRepo := mongodb.NewLowStockAlertRepository(db)

	// Initialize services
	inventoryService := service.NewInventoryService(
		productRepo,
		salesHistoryRepo,
		lowStockAlertRepo,
	)

	aiResponder := service.NewAIResponderService(
		cfg.OpenAIAPIKey,
		reviewRepo,
	)

	syncService := service.NewKaspiSyncService(
		kaspiKeyRepo,
		productRepo,
		salesHistoryRepo,
		reviewRepo,
		encryptor,
		inventoryService,
	)

	// TEMPORARILY DISABLED - Price Dumping
	/*
	priceDumpingService := service.NewPriceDumpingService(
		kaspiKeyRepo,
		productRepo,
		encryptor,
	)
	*/

	// Initialize scheduler
	sched := scheduler.New()

	// Schedule Kaspi sync
	err = sched.AddIntervalJob(cfg.SyncIntervalHours, func() {
		logger.Log.Info("Starting scheduled Kaspi sync")

		// Sync all Kaspi accounts
		if err := syncService.SyncAll(); err != nil {
			logger.Log.Error("Kaspi sync failed", zap.Error(err))
		}

		// Get all users with auto-reply enabled
		users, err := getUsersWithAutoReply(userRepo)
		if err != nil {
			logger.Log.Error("Failed to get users with auto-reply", zap.Error(err))
			return
		}

		for _, user := range users {
			// Process AI responses for pending reviews
			if err := aiResponder.ProcessPendingReviews(user.TelegramID, true); err != nil {
				logger.Log.Error("Failed to process pending reviews",
					zap.Int64("user_id", user.TelegramID),
					zap.Error(err),
				)
			}

			// Process low stock alerts
			if err := inventoryService.ProcessLowStockAlerts(user.TelegramID, 7); err != nil {
				logger.Log.Error("Failed to process low stock alerts",
					zap.Int64("user_id", user.TelegramID),
					zap.Error(err),
				)
			}
		}

		logger.Log.Info("Scheduled sync completed")
	})

	if err != nil {
		logger.Log.Fatal("Failed to schedule sync job", zap.Error(err))
	}

	// TEMPORARILY DISABLED - Price Dumping
	// Schedule price dumping (every 5 minutes)
	// err = sched.AddJob("*/5 * * * *", func() {
	// 	logger.Log.Info("Starting price dumping cycle")
	//
	// 	if err := priceDumpingService.ProcessAllUsers(); err != nil {
	// 		logger.Log.Error("Price dumping failed", zap.Error(err))
	// 	}
	//
	// 	logger.Log.Info("Price dumping cycle completed")
	// })
	//
	// if err != nil {
	// 	logger.Log.Fatal("Failed to schedule price dumping job", zap.Error(err))
	// }

	// Run initial sync immediately
	logger.Log.Info("Running initial sync...")
	if err := syncService.SyncAll(); err != nil {
		logger.Log.Error("Initial sync failed", zap.Error(err))
	}

	// TEMPORARILY DISABLED - Price Dumping
	// Run initial price dumping
	// logger.Log.Info("Running initial price dumping...")
	// if err := priceDumpingService.ProcessAllUsers(); err != nil {
	// 	logger.Log.Error("Initial price dumping failed", zap.Error(err))
	// }

	// Start scheduler
	sched.Start()
	logger.Log.Info("Worker started, scheduler is running")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down worker...")
	sched.Stop()
	logger.Log.Info("Worker stopped gracefully")
}

func getUsersWithAutoReply(userRepo domain.UserRepository) ([]domain.User, error) {
	// This would need a new method in the repository
	// For now, we'll return an empty slice
	// You can implement GetUsersWithAutoReply() in the user repository
	return []domain.User{}, nil
}
