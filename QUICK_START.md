# Quick Start Guide

Get the Kaspi Seller Assistant Bot running in 5 minutes!

## Prerequisites

- Go 1.21+
- MongoDB 7+
- Telegram Bot Token
- OpenAI API Key
- Kaspi API Key and Merchant ID

## Option 1: Docker (Recommended)

```bash
# 1. Clone and setup
git clone <your-repo>
cd seller-assistant
cp .env.example .env

# 2. Generate encryption key
openssl rand -base64 32

# 3. Edit .env and add your credentials
nano .env

# 4. Start everything
export TELEGRAM_BOT_TOKEN=your_token
export OPENAI_API_KEY=your_key
export ENCRYPTION_KEY=your_encryption_key
docker-compose up -d

# 5. View logs
docker-compose logs -f
```

Done! Your bot is now running at `@your_bot_username` on Telegram.

## Option 2: Local Development

```bash
# 1. Run setup script
./scripts/setup.sh

# 2. Edit .env with your credentials
nano .env

# 3. Start MongoDB
docker run -d -p 27017:27017 --name mongodb mongo:7
# Database and indexes auto-created on first run

# 4. Run bot and worker (in separate terminals)
make run-bot     # Terminal 1
make run-worker  # Terminal 2
```

## Option 3: Deploy to Railway.app

```bash
# 1. Push to GitHub
git init
git add .
git commit -m "Initial commit"
git push origin main

# 2. Go to railway.app
# - Create new project from GitHub repo
# - Add PostgreSQL database
# - Add environment variables (see Railway section in README)

# 3. Deploy worker as separate service
# - Add new service from same repo
# - Set custom command: ./worker
```

## Environment Variables Checklist

Make sure you have these set:

- ✅ `MONGODB_URI` - MongoDB connection URI (default: mongodb://localhost:27017)
- ✅ `MONGODB_DATABASE` - Database name (default: seller_assistant)
- ✅ `TELEGRAM_BOT_TOKEN` - From @BotFather
- ✅ `OPENAI_API_KEY` - From OpenAI dashboard
- ✅ `ENCRYPTION_KEY` - Generated with `openssl rand -base64 32`

Optional:
- `SYNC_INTERVAL_HOURS` (default: 6)
- `LOG_LEVEL` (default: info)
- `ENVIRONMENT` (default: development)

## Getting Your Credentials

### Telegram Bot Token
1. Open Telegram and message [@BotFather](https://t.me/botfather)
2. Send `/newbot`
3. Follow instructions to create your bot
4. Copy the token provided

### OpenAI API Key
1. Go to [platform.openai.com](https://platform.openai.com)
2. Sign up or log in
3. Go to API Keys section
4. Create new secret key
5. Copy the key (you won't see it again!)

### Encryption Key
```bash
openssl rand -base64 32
```

## First Steps After Deployment

1. Open Telegram and find your bot
2. Send `/start`
3. Click "🔑 Manage API Keys"
4. Add your Kaspi API credentials (API_KEY MERCHANT_ID)
5. Wait for initial sync (~2-5 minutes)
6. Check "📊 Dashboard" to see your data

## Testing the Bot

Quick test commands:
```
/start          → Welcome message
📊 Dashboard    → View your inventory summary
⚙️ Settings     → Toggle auto-reply, change language
ℹ️ Help         → Get help information
```

## Troubleshooting

### Bot not responding?
```bash
# Check if bot is running
docker-compose ps

# Check logs
docker-compose logs bot
```

### Database errors?
```bash
# Check MongoDB is running
mongosh $MONGODB_URI

# Check collections
mongosh $MONGODB_URI --eval "use seller_assistant; show collections"
```

### Worker not syncing?
```bash
# Check worker logs
docker-compose logs worker

# Restart worker
docker-compose restart worker
```

## Useful Commands

```bash
make help          # Show all available commands
make build         # Build binaries
make test          # Run tests
make docker-up     # Start with Docker
make docker-logs   # View logs
make gen-key       # Generate encryption key
```

## What's Next?

- Add your Kaspi API credentials in the bot
- Enable auto-reply in Settings for automatic review responses
- Monitor your dashboard for low stock alerts
- Customize AI response language (RU/KK/EN)

## Support

- 📖 Full docs: See README.md
- 🐛 Issues: GitHub Issues
- 💬 Telegram: @your_support_username

## Production Checklist

Before going live:

- [ ] Use strong encryption key (32 bytes)
- [ ] Set `ENVIRONMENT=production`
- [ ] Enable PostgreSQL SSL (`?sslmode=require`)
- [ ] Set appropriate `LOG_LEVEL` (info or warn)
- [ ] Configure backup for database
- [ ] Set up monitoring/alerts
- [ ] Test all bot commands
- [ ] Test marketplace integrations
- [ ] Verify auto-reply works correctly

## Architecture

```
┌─────────────────┐
│  Telegram User  │
└────────┬────────┘
         │
    ┌────▼────┐
    │   Bot   │ ← Handles user interactions
    └────┬────┘
         │
    ┌────▼────────┐
    │  Services   │ ← Business logic
    └────┬────────┘
         │
    ┌────▼─────────┐
    │   MongoDB    │ ← Data storage
    └──────────────┘

┌──────────┐
│  Worker  │ ← Syncs Kaspi data every 6h
└────┬─────┘
     │
┌────▼────────────┐
│  Kaspi.kz API  │
└─────────────────┘
```

Happy selling! 🚀
