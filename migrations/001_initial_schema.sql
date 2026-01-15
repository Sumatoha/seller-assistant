-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    language_code VARCHAR(10) DEFAULT 'ru',
    auto_reply_enabled BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_telegram_id ON users(telegram_id);

-- Marketplace API Keys table (encrypted)
CREATE TABLE IF NOT EXISTS marketplace_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    marketplace_type VARCHAR(50) NOT NULL, -- 'kaspi' or 'wildberries'
    api_key_encrypted TEXT NOT NULL,
    api_secret_encrypted TEXT,
    merchant_id VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, marketplace_type)
);

CREATE INDEX idx_marketplace_keys_user_id ON marketplace_keys(user_id);
CREATE INDEX idx_marketplace_keys_active ON marketplace_keys(is_active);

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    marketplace_key_id BIGINT NOT NULL REFERENCES marketplace_keys(id) ON DELETE CASCADE,
    external_id VARCHAR(255) NOT NULL, -- Product ID from marketplace
    sku VARCHAR(255),
    name TEXT NOT NULL,
    current_stock INTEGER DEFAULT 0,
    price DECIMAL(10, 2),
    currency VARCHAR(10) DEFAULT 'KZT',
    sales_velocity DECIMAL(10, 2) DEFAULT 0, -- Average sales per day
    days_of_stock INTEGER DEFAULT 0, -- Calculated field
    last_sync_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(marketplace_key_id, external_id)
);

CREATE INDEX idx_products_user_id ON products(user_id);
CREATE INDEX idx_products_marketplace_key_id ON products(marketplace_key_id);
CREATE INDEX idx_products_days_of_stock ON products(days_of_stock);

-- Sales History table
CREATE TABLE IF NOT EXISTS sales_history (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    quantity_sold INTEGER DEFAULT 0,
    revenue DECIMAL(10, 2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(product_id, date)
);

CREATE INDEX idx_sales_history_product_id ON sales_history(product_id);
CREATE INDEX idx_sales_history_date ON sales_history(date);

-- Reviews table
CREATE TABLE IF NOT EXISTS reviews (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id BIGINT REFERENCES products(id) ON DELETE SET NULL,
    marketplace_key_id BIGINT NOT NULL REFERENCES marketplace_keys(id) ON DELETE CASCADE,
    external_id VARCHAR(255) NOT NULL, -- Review ID from marketplace
    author_name VARCHAR(255),
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    language VARCHAR(10) DEFAULT 'ru',
    ai_response TEXT,
    ai_response_sent BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(marketplace_key_id, external_id)
);

CREATE INDEX idx_reviews_user_id ON reviews(user_id);
CREATE INDEX idx_reviews_marketplace_key_id ON reviews(marketplace_key_id);
CREATE INDEX idx_reviews_ai_response_sent ON reviews(ai_response_sent);

-- Low Stock Alerts table
CREATE TABLE IF NOT EXISTS low_stock_alerts (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    threshold_days INTEGER DEFAULT 7,
    notified_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_low_stock_alerts_user_id ON low_stock_alerts(user_id);
CREATE INDEX idx_low_stock_alerts_product_id ON low_stock_alerts(product_id);
