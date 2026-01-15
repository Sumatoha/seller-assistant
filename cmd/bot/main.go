package main

//
//import (
//	"log"
//	"os"
//	"os/signal"
//	"syscall"
//
//	"github.com/yourusername/seller-assistant/internal/config"
//	"github.com/yourusername/seller-assistant/internal/repository/mongodb"
//	"github.com/yourusername/seller-assistant/internal/service"
//	"github.com/yourusername/seller-assistant/internal/telegram"
//	"github.com/yourusername/seller-assistant/pkg/crypto"
//	"github.com/yourusername/seller-assistant/pkg/logger"
//	"go.uber.org/zap"
//)
//
//func main() {
//	// Load configuration
//	cfg, err := config.Load()
//	if err != nil {
//		log.Fatalf("Failed to load config: %v", err)
//	}
//
//	// Initialize logger
//	if err := logger.Init(cfg.LogLevel); err != nil {
//		log.Fatalf("Failed to initialize logger: %v", err)
//	}
//	defer logger.Sync()
//
//	logger.Log.Info("Starting Seller Assistant Bot",
//		zap.String("environment", cfg.Environment),
//	)
//
//	// Initialize MongoDB
//	db, err := mongodb.NewDB(cfg.MongoDBURI, cfg.MongoDBDatabase)
//	if err != nil {
//		logger.Log.Fatal("Failed to connect to MongoDB", zap.Error(err))
//	}
//	defer db.Close()
//
//	logger.Log.Info("MongoDB connected successfully")
//
//	// Initialize encryptor
//	encryptor, err := crypto.NewEncryptor(cfg.EncryptionKey)
//	if err != nil {
//		logger.Log.Fatal("Failed to initialize encryptor", zap.Error(err))
//	}
//
//	// Initialize repositories
//	userRepo := mongodb.NewUserRepository(db)
//	kaspiKeyRepo := mongodb.NewKaspiKeyRepository(db)
//	productRepo := mongodb.NewProductRepository(db)
//	salesHistoryRepo := mongodb.NewSalesHistoryRepository(db)
//	reviewRepo := mongodb.NewReviewRepository(db)
//	lowStockAlertRepo := mongodb.NewLowStockAlertRepository(db)
//
//	// Initialize services
//	inventoryService := service.NewInventoryService(
//		productRepo,
//		salesHistoryRepo,
//		lowStockAlertRepo,
//	)
//
//	aiResponder := service.NewAIResponderService(
//		cfg.OpenAIAPIKey,
//		reviewRepo,
//	)
//
//	syncService := service.NewKaspiSyncService(
//		kaspiKeyRepo,
//		productRepo,
//		salesHistoryRepo,
//		reviewRepo,
//		encryptor,
//		inventoryService,
//	)
//
//	priceDumpingService := service.NewPriceDumpingService(
//		kaspiKeyRepo,
//		productRepo,
//		encryptor,
//	)
//
//	bot, err := telegram.NewBot(
//		cfg.TelegramBotToken,
//		userRepo,
//		kaspiKeyRepo,
//		productRepo,
//		reviewRepo,
//		inventoryService,
//		aiResponder,
//		syncService,
//		priceDumpingService,
//		encryptor,
//	)
//	if err != nil {
//		logger.Log.Fatal("Failed to create bot", zap.Error(err))
//	}
//
//	go func() {
//		if err := bot.Start(); err != nil {
//			logger.Log.Fatal("Bot stopped with error", zap.Error(err))
//		}
//	}()
//
//	quit := make(chan os.Signal, 1)
//	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
//	<-quit
//
//	logger.Log.Info("Shutting down bot...")
//	bot.Stop()
//	logger.Log.Info("Bot stopped gracefully")
//}
