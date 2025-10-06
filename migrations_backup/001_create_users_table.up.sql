-- Create users table for user profiles and onboarding tracking
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    date_of_birth DATE,
    auth_provider_id VARCHAR(255),
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    onboarding_status VARCHAR(50) DEFAULT 'started' CHECK (
        onboarding_status IN ('started', 'email_verified', 'kyc_pending', 'kyc_approved', 'kyc_rejected', 'wallets_pending', 'completed')
    ),
    kyc_status VARCHAR(50) DEFAULT 'pending' CHECK (
        kyc_status IN ('pending', 'processing', 'approved', 'rejected')
    ),
    kyc_approved_at TIMESTAMP,
    kyc_rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_auth_provider_id ON users(auth_provider_id);
CREATE INDEX IF NOT EXISTS idx_users_onboarding_status ON users(onboarding_status);
CREATE INDEX IF NOT EXISTS idx_users_kyc_status ON users(kyc_status);

-- Add comment to table
COMMENT ON TABLE users IS 'User profiles and onboarding status tracking';