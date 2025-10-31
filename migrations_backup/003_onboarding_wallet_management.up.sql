-- Onboarding & Managed Wallet Creation Migration
-- This migration adds support for KYC onboarding flow and Circle wallet integration

-- Update users table to support onboarding flow
ALTER TABLE users
    ADD COLUMN auth_provider_id VARCHAR(200), -- Auth provider reference (Cognito/Auth0)
    ADD COLUMN phone VARCHAR(20),
    ADD COLUMN phone_verified BOOLEAN DEFAULT FALSE,
    ADD COLUMN onboarding_status VARCHAR(20) DEFAULT 'started' CHECK (onboarding_status IN ('started', 'kyc_pending', 'kyc_approved', 'kyc_rejected', 'wallets_pending', 'completed')),
    ADD COLUMN kyc_provider_ref VARCHAR(200), -- KYC provider reference
    ADD COLUMN kyc_submitted_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN kyc_approved_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN kyc_rejection_reason TEXT,
    -- Drop old fields that don't align with managed wallet approach
    ALTER COLUMN username DROP NOT NULL,
    ALTER COLUMN first_name DROP NOT NULL,
    ALTER COLUMN last_name DROP NOT NULL,
    ALTER COLUMN password_hash DROP NOT NULL;

-- Create index for onboarding queries
CREATE INDEX idx_users_auth_provider_id ON users(auth_provider_id);
CREATE INDEX idx_users_onboarding_status ON users(onboarding_status);
CREATE INDEX idx_users_phone ON users(phone);

-- Create wallet_sets table for Circle integration
CREATE TABLE wallet_sets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    circle_wallet_set_id VARCHAR(200) NOT NULL UNIQUE, -- Circle's walletSetId
    entity_secret_ciphertext TEXT NOT NULL, -- Circle's entity secret
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallet_sets_circle_id ON wallet_sets(circle_wallet_set_id);

-- Update wallets table for Circle integration
-- Drop the old wallet table from migration 002 and recreate with proper Circle support
DROP TABLE IF EXISTS wallets CASCADE;

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chain VARCHAR(20) NOT NULL CHECK (chain IN ('ETH', 'ETH-SEPOLIA', 'SOL', 'SOL-DEVNET', 'APTOS', 'APTOS-TESTNET')),
    address VARCHAR(100) NOT NULL,
    circle_wallet_id VARCHAR(200) NOT NULL UNIQUE, -- Circle's wallet ID
    wallet_set_id UUID NOT NULL REFERENCES wallet_sets(id),
    account_type VARCHAR(10) NOT NULL DEFAULT 'EOA' CHECK (account_type IN ('EOA', 'SCA')),
    status VARCHAR(20) NOT NULL DEFAULT 'creating' CHECK (status IN ('creating', 'live', 'failed')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id, chain)
);

-- Create indexes for wallets
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_chain ON wallets(chain);
CREATE INDEX idx_wallets_address ON wallets(address);
CREATE INDEX idx_wallets_circle_wallet_id ON wallets(circle_wallet_id);
CREATE INDEX idx_wallets_status ON wallets(status);
CREATE INDEX idx_wallets_wallet_set_id ON wallets(wallet_set_id);

-- Create onboarding_flows table to track the onboarding process
CREATE TABLE onboarding_flows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    step VARCHAR(50) NOT NULL CHECK (step IN ('registration', 'email_verification', 'phone_verification', 'kyc_submission', 'kyc_review', 'wallet_creation', 'completed')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'failed', 'skipped')),
    data JSONB, -- Store step-specific data
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_onboarding_flows_user_id ON onboarding_flows(user_id);
CREATE INDEX idx_onboarding_flows_step ON onboarding_flows(step);
CREATE INDEX idx_onboarding_flows_status ON onboarding_flows(status);

-- Create kyc_submissions table for KYC tracking
CREATE TABLE kyc_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL, -- e.g., 'jumio', 'onfido'
    provider_ref VARCHAR(200) NOT NULL, -- Provider's reference ID
    submission_type VARCHAR(20) NOT NULL CHECK (submission_type IN ('identity', 'address', 'full')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'approved', 'rejected', 'expired')),
    verification_data JSONB, -- Store verification details
    rejection_reasons TEXT[],
    submitted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_submissions_user_id ON kyc_submissions(user_id);
CREATE INDEX idx_kyc_submissions_provider_ref ON kyc_submissions(provider_ref);
CREATE INDEX idx_kyc_submissions_status ON kyc_submissions(status);

-- Create wallet_provisioning_jobs table for tracking async wallet creation
CREATE TABLE wallet_provisioning_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chains TEXT[] NOT NULL, -- Array of chains to provision ['ETH', 'SOL', 'APTOS']
    status VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'in_progress', 'completed', 'failed', 'retry')),
    attempt_count INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    circle_requests JSONB, -- Store Circle API request/response logs
    error_message TEXT,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallet_provisioning_jobs_user_id ON wallet_provisioning_jobs(user_id);
CREATE INDEX idx_wallet_provisioning_jobs_status ON wallet_provisioning_jobs(status);
CREATE INDEX idx_wallet_provisioning_jobs_next_retry_at ON wallet_provisioning_jobs(next_retry_at);

-- Update audit_logs table to support onboarding and wallet events
-- Add specific action types for onboarding flow
INSERT INTO audit_logs (id, actor, action, entity, after, occurred_at) VALUES 
(uuid_generate_v4(), NULL, 'MIGRATION', 'onboarding_schema', '{"version": "003", "description": "Added onboarding and wallet management tables"}'::jsonb, NOW());

-- Create triggers for updated_at columns
CREATE TRIGGER update_wallet_sets_updated_at BEFORE UPDATE ON wallet_sets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_wallets_updated_at BEFORE UPDATE ON wallets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_onboarding_flows_updated_at BEFORE UPDATE ON onboarding_flows FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_kyc_submissions_updated_at BEFORE UPDATE ON kyc_submissions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_wallet_provisioning_jobs_updated_at BEFORE UPDATE ON wallet_provisioning_jobs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert a default wallet set for the application (to be configured with actual Circle credentials)
-- Note: This should be updated with real Circle wallet set ID during deployment
INSERT INTO wallet_sets (id, circle_wallet_set_id, entity_secret_ciphertext, status)
VALUES (
    '00000000-0000-0000-0000-000000000001'::uuid,
    'placeholder-wallet-set-id',
    'placeholder-entity-secret',
    'active'
);