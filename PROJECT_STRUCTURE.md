# Project Structure

Complete file structure for the Seller Assistant Bot project.

```
seller-assistant/
│
├── cmd/                                    # Application entry points
│   ├── bot/
│   │   └── main.go                        # Telegram bot main entry point
│   └── worker/
│       └── main.go                        # Background worker main entry point
│
├── internal/                              # Private application code
│   ├── config/
│   │   └── config.go                      # Configuration management with env vars
│   │
│   ├── domain/                            # Business entities and interfaces
│   │   ├── user.go                        # User entity and repository interface
│   │   ├── marketplace.go                 # Marketplace key entity and interface
│   │   ├── product.go                     # Product, SalesHistory, LowStockAlert entities
│   │   └── review.go                      # Review entity and repository interface
│   │
│   ├── repository/                        # Data access layer
│   │   └── mongodb/
│   │       ├── db.go                      # MongoDB connection and index setup
│   │       ├── user.go                    # User repository implementation
│   │       ├── marketplace.go             # Marketplace key repository
│   │       ├── product.go                 # Product and sales history repositories
│   │       └── review.go                  # Review repository implementation
│   │
│   ├── service/                           # Business logic layer
│   │   ├── inventory.go                   # Inventory tracking and Days of Stock calculation
│   │   ├── ai_responder.go                # AI-powered review response generator
│   │   └── marketplace_sync.go            # Marketplace data synchronization
│   │
│   ├── telegram/                          # Telegram bot implementation
│   │   ├── bot.go                         # Bot initialization and core methods
│   │   ├── handlers.go                    # Command and callback handlers
│   │   └── keyboard.go                    # Telegram keyboard layouts
│   │
│   └── marketplace/                       # Marketplace API clients
│       ├── interface.go                   # Marketplace interface
│       └── kaspi/
│           └── client.go                  # Kaspi.kz API client
│
├── pkg/                                   # Public, reusable packages
│   ├── crypto/
│   │   └── encryption.go                  # AES-256-GCM encryption utilities
│   ├── logger/
│   │   └── logger.go                      # Zap logger wrapper
│   └── scheduler/
│       └── scheduler.go                   # Cron job scheduler wrapper
│
├── migrations/                            # Database migrations (for reference only)
│   └── 001_initial_schema.sql            # SQL schema for reference (MongoDB uses auto-created indexes)
│
├── scripts/                               # Utility scripts
│   └── setup.sh                          # Development environment setup script
│
├── .dockerignore                         # Docker ignore file
├── .env.example                          # Environment variables template
├── .gitignore                            # Git ignore file
├── docker-compose.yml                    # Docker Compose configuration
├── Dockerfile                            # Multi-stage Docker build
├── go.mod                                # Go module dependencies
├── go.sum                                # Go module checksums
├── Makefile                              # Build and development commands
├── PROJECT_STRUCTURE.md                  # This file
└── README.md                             # Main documentation
```

## Key Components

### Entry Points (`cmd/`)
- **bot/main.go**: Starts the Telegram bot service
- **worker/main.go**: Starts the background worker for periodic syncing

### Domain Layer (`internal/domain/`)
- Defines core business entities (User, Product, Review, etc.)
- Defines repository interfaces (contract for data access)
- No external dependencies - pure business logic

### Repository Layer (`internal/repository/mongodb/`)
- Implements domain repository interfaces
- Handles all database operations
- Uses mongo-driver for MongoDB operations
- Includes automatic index creation and connection pooling

### Service Layer (`internal/service/`)
- **inventory.go**:
  - Calculates sales velocity
  - Predicts days of stock
  - Generates low stock alerts

- **ai_responder.go**:
  - Integrates with OpenAI GPT-4
  - Generates contextual review responses
  - Supports multiple languages (RU, KK, EN)

- **marketplace_sync.go**:
  - Syncs products, sales data, and reviews
  - Handles encryption/decryption of API keys
  - Coordinates periodic synchronization

### Telegram Bot (`internal/telegram/`)
- **bot.go**: Core bot structure and methods
- **handlers.go**: All command and callback handlers
- **keyboard.go**: Interactive keyboard layouts

### Marketplace Client (`internal/marketplace/kaspi/`)
- **client.go**: Kaspi.kz API integration
- Implements marketplace interface for products, sales, and reviews
- Mock implementation ready for real Kaspi API

### Utilities (`pkg/`)
- **crypto**: AES-256-GCM encryption for sensitive data
- **logger**: Structured logging with Zap
- **scheduler**: Cron-based job scheduling

## Data Flow

```
User (Telegram)
    ↓
Telegram Bot (handlers.go)
    ↓
Service Layer (business logic)
    ↓
Repository Layer (database)
    ↓
PostgreSQL Database

Background Worker
    ↓
Kaspi Sync Service
    ↓
Kaspi API Client
    ↓
Repository Layer (save data)
    ↓
MongoDB Database
```

## Configuration

All configuration is managed through environment variables:
- Database connections
- API keys (Telegram, OpenAI, Marketplaces)
- Encryption keys
- Worker intervals
- Logging levels

See `.env.example` for all available options.

## Database Collections

The MongoDB database includes the following collections:
- `users` - Telegram users
- `kaspi_keys` - Encrypted Kaspi API credentials (one per user)
- `products` - Product inventory data from Kaspi
- `sales_history` - Historical sales data for velocity calculation
- `reviews` - Customer reviews from Kaspi and AI responses
- `low_stock_alerts` - Stock alert notifications

Indexes are automatically created on startup. See `internal/repository/mongodb/db.go` for index definitions.

## Deployment

The project is designed for easy deployment:
- **Docker**: Single Dockerfile with multi-stage build
- **Docker Compose**: Complete stack (bot, worker, MongoDB)
- **Railway.app**: Direct deployment from GitHub (with MongoDB Atlas)
- **Kubernetes**: Can be easily adapted (Deployment manifests not included)

## Security Features

1. **Encrypted API Keys**: All marketplace credentials encrypted at rest
2. **Environment Variables**: Sensitive config never committed
3. **NoSQL Injection Prevention**: Proper BSON encoding
4. **Input Validation**: All user inputs validated
5. **Secure Defaults**: TLS/SSL support for MongoDB connections
