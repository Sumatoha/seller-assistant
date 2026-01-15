# Kaspi Seller Assistant Bot

A production-ready Telegram Bot for Kaspi.kz marketplace sellers. Built with Go following Clean Architecture principles.

## Features

- **Telegram Bot Interface** - Intuitive bot commands and menu-driven UI
- **Inventory Tracking** - Predict "Days of Stock" based on sales velocity
- **AI Review Responder** - Automatic AI-generated responses to customer reviews in Russian and Kazakh
- **Price Dumping (Автодемпинг)** - Automatic price monitoring and undercutting competitors by 1₸ every 5 minutes with minimum price protection
- **Kaspi.kz Integration** - Full integration with Kaspi marketplace API
- **Background Worker** - Periodic data synchronization and price updates
- **Encrypted API Keys** - Secure storage of Kaspi API credentials
- **Low Stock Alerts** - Automatic notifications when inventory is running low
- **Auto-Reply Mode** - Toggle automated AI responses to reviews

## Architecture

The project follows Clean Architecture with the following structure:

```
seller-assistant/
├── cmd/
│   ├── bot/          # Telegram bot entry point
│   └── worker/       # Background worker entry point
├── internal/
│   ├── domain/       # Business entities and interfaces
│   ├── repository/   # Data access layer (PostgreSQL)
│   ├── service/      # Business logic
│   ├── telegram/     # Telegram bot handlers
│   ├── marketplace/  # Marketplace API clients
│   └── config/       # Configuration management
├── pkg/
│   ├── crypto/       # Encryption utilities
│   ├── logger/       # Logging utilities
│   └── scheduler/    # Job scheduler
├── migrations/       # Database migrations
└── Dockerfile        # Production-ready container
```

## Tech Stack

- **Language**: Go 1.21+
- **Database**: MongoDB 7+
- **Telegram Bot**: go-telegram-bot-api/telegram-bot-api/v5
- **AI**: OpenAI GPT-4
- **Database Driver**: mongo-driver v1.13+
- **Scheduler**: robfig/cron/v3
- **Logging**: uber-go/zap
- **Encryption**: AES-256-GCM

## Prerequisites

- Go 1.21 or higher
- MongoDB 7 or higher
- Telegram Bot Token (from [@BotFather](https://t.me/botfather))
- OpenAI API Key
- Kaspi.kz API credentials (API Key and Merchant ID)

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/seller-assistant.git
cd seller-assistant
```

### 2. Generate Encryption Key

Generate a secure 32-byte encryption key for storing API credentials:

```bash
# Using OpenSSL
openssl rand -base64 32
```

### 3. Set Up Environment Variables

Copy the example environment file and configure it:

```bash
cp .env.example .env
```

Edit `.env` with your credentials:

```env
# Database Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=seller_assistant

# Telegram Bot
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here

# OpenAI API
OPENAI_API_KEY=your_openai_api_key_here

# Encryption Key (use the key generated in step 2)
ENCRYPTION_KEY=your_32_byte_encryption_key_base64

# Server Configuration
PORT=8080
ENVIRONMENT=production

# Worker Configuration
SYNC_INTERVAL_HOURS=6

# Log Level
LOG_LEVEL=info
```

### 4. Set Up Database

Start MongoDB:

```bash
# Using Docker
docker run -d -p 27017:27017 --name mongodb mongo:7

# Or use your local MongoDB installation
mongod
```

The database and indexes will be created automatically on first run.

### 5. Install Dependencies

```bash
go mod download
```

### 6. Run Locally

Run the bot and worker:

```bash
# Terminal 1 - Run the bot
go run cmd/bot/main.go

# Terminal 2 - Run the worker
go run cmd/worker/main.go
```

## Docker Deployment

### Local Development with Docker Compose

```bash
# Set environment variables
export TELEGRAM_BOT_TOKEN=your_token
export OPENAI_API_KEY=your_key
export ENCRYPTION_KEY=your_encryption_key

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Build Docker Image

```bash
docker build -t seller-assistant:latest .
```

## Deployment to Railway.app

Railway.app provides easy deployment with automatic SSL, monitoring, and scaling.

### Step 1: Prepare Your Repository

1. Push your code to GitHub
2. Make sure `.env` is in `.gitignore` (it already is)

### Step 2: Deploy to Railway

1. Go to [Railway.app](https://railway.app)
2. Click "New Project" → "Deploy from GitHub repo"
3. Select your repository
4. Railway will detect the Dockerfile automatically

### Step 3: Add PostgreSQL Database

1. In your Railway project, click "New"
2. Select "Database" → "PostgreSQL"
3. Railway will automatically create and configure the database

### Step 4: Configure Environment Variables

In Railway project settings, add these environment variables:

```
MONGODB_URI=your_mongodb_connection_uri
MONGODB_DATABASE=seller_assistant
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
OPENAI_API_KEY=your_openai_api_key
ENCRYPTION_KEY=your_32_byte_encryption_key_base64
SYNC_INTERVAL_HOURS=6
LOG_LEVEL=info
ENVIRONMENT=production
```

**Note**: For MongoDB on Railway, use MongoDB Atlas or add a MongoDB service. Railway doesn't provide MongoDB by default.

### Step 5: MongoDB Setup

The application automatically creates indexes on startup. No manual migration needed!

### Step 6: Deploy Worker Service

The Dockerfile builds both the bot and worker. To run the worker:

1. In Railway, click "New" → "Empty Service"
2. Link to the same GitHub repository
3. In Settings → Deploy, set custom start command: `./worker`
4. Add the same environment variables as the bot service

### Step 7: Monitor and Scale

- Railway provides automatic metrics and logging
- View logs in real-time from the Railway dashboard
- Scale vertically by upgrading your plan
- Railway auto-scales based on traffic

## Usage

### Bot Commands

- `/start` - Welcome message and onboarding
- `📊 Dashboard` - View summary of inventory and reviews
- `📦 Low Stock Alerts` - Check products running low
- `⭐ Reviews` - View recent customer reviews
- `🔑 Manage API Keys` - Add/remove marketplace credentials
- `⚙️ Settings` - Configure auto-reply and language
- `ℹ️ Help` - Get help and support

### Adding Your Kaspi API Key

1. Click "🔑 Manage API Keys"
2. Click "Add Kaspi Key"
3. Send your credentials in format: `API_KEY MERCHANT_ID`
4. Keys are encrypted before storage

### Enabling Auto-Reply

1. Go to "⚙️ Settings"
2. Click "Toggle Auto-Reply"
3. AI will automatically respond to new reviews

### Setting Up Price Dumping

1. Go to "💰 Price Dumping"
2. Click "Enable for Product"
3. Send: `SKU MIN_PRICE` (e.g., `ABC123 15000`)
4. System will monitor competitors every 5 minutes and set your price 1₸ lower
5. Your price will never go below the minimum threshold you set

See [PRICE_DUMPING.md](PRICE_DUMPING.md) for detailed documentation.

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `MONGODB_URI` | MongoDB connection URI | mongodb://localhost:27017 | Yes |
| `MONGODB_DATABASE` | MongoDB database name | seller_assistant | Yes |
| `TELEGRAM_BOT_TOKEN` | Telegram Bot API token | - | Yes |
| `OPENAI_API_KEY` | OpenAI API key | - | Yes |
| `ENCRYPTION_KEY` | 32-byte base64-encoded key | - | Yes |
| `PORT` | HTTP server port | 8080 | No |
| `ENVIRONMENT` | Environment (development/production) | development | No |
| `SYNC_INTERVAL_HOURS` | How often to sync marketplace data | 6 | No |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | info | No |

### Kaspi API Configuration

- Obtain API credentials from Kaspi Merchant Cabinet
- You'll need: API Key and Merchant ID
- Format when adding to bot: `API_KEY MERCHANT_ID`
- [Kaspi API Documentation](https://kaspi.kz/merchantcabinet/)

**Note**: The Kaspi API client includes mock implementations. Replace with actual API endpoints based on official Kaspi documentation when available.

## Development

### Running Tests

```bash
go test ./...
```

### Building Binaries

```bash
# Build bot
go build -o bin/bot cmd/bot/main.go

# Build worker
go build -o bin/worker cmd/worker/main.go
```

### Code Structure

- **cmd/**: Application entry points
- **internal/domain/**: Core business entities and repository interfaces
- **internal/repository/mongodb/**: MongoDB repository implementations
- **internal/service/**: Business logic (inventory tracking, AI responses, Kaspi sync)
- **internal/telegram/**: Bot handlers and keyboards
- **internal/marketplace/kaspi/**: Kaspi.kz API client
- **pkg/**: Reusable utilities

## Security

- API keys are encrypted using AES-256-GCM before storage
- Environment variables are used for sensitive configuration
- MongoDB connections support TLS/SSL
- Input validation on all user inputs
- Secure random encryption key generation
- Document-level security with proper indexing

## Monitoring

The application uses structured logging (zap) with the following levels:

- **DEBUG**: Detailed debugging information
- **INFO**: General informational messages
- **WARN**: Warning messages
- **ERROR**: Error messages

Logs are output in JSON format in production for easy parsing.

## Troubleshooting

### Database Connection Issues

```bash
# Test MongoDB connection
mongosh $MONGODB_URI

# Check collections
mongosh $MONGODB_URI --eval "use seller_assistant; show collections"
```

### Bot Not Responding

1. Check if bot token is correct
2. Verify bot is running: `docker-compose ps`
3. Check logs: `docker-compose logs bot`

### Worker Not Syncing

1. Check worker logs: `docker-compose logs worker`
2. Verify marketplace API credentials
3. Check sync interval configuration

## Performance

- **Concurrent requests**: Uses goroutines for parallel processing
- **MongoDB connection pooling**: Automatic connection management
- **Efficient queries**: Uses indexes for all common queries
- **Background processing**: Non-blocking sync operations
- **Automatic index creation**: Indexes created on startup

## Roadmap

- [x] Price dumping with minimum price protection
- [ ] Real Kaspi API integration (currently uses mock data)
- [ ] Configurable price dumping margin (not just -1₸)
- [ ] Price history and analytics
- [ ] Advanced analytics and reporting
- [ ] Multi-user team support
- [ ] Mobile app companion
- [ ] Webhook support for real-time Kaspi updates
- [ ] Custom AI response templates
- [ ] Export reports to Excel/PDF

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions:
- GitHub Issues: [github.com/yourusername/seller-assistant/issues](https://github.com/yourusername/seller-assistant/issues)
- Email: support@yourcompany.com
- Telegram: @your_support_username

## Acknowledgments

- Built with [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- AI powered by [OpenAI](https://openai.com)
- Built specifically for Kaspi.kz sellers in Kazakhstan
