CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    date DATE NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    source VARCHAR(50),
    description TEXT,
    merchant VARCHAR(255),
    category VARCHAR(100),
    subcategory VARCHAR(100),
    currency VARCHAR(10),
    is_transfer BOOLEAN DEFAULT FALSE,
    is_fee BOOLEAN DEFAULT FALSE,
    is_tax BOOLEAN DEFAULT FALSE,
    neutralized BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
