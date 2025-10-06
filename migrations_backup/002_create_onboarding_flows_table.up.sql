-- Create onboarding_flows table for tracking onboarding step progress
CREATE TABLE IF NOT EXISTS onboarding_flows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    step_type VARCHAR(50) NOT NULL CHECK (
        step_type IN ('email_verification', 'kyc_submission', 'wallet_creation')
    ),
    status VARCHAR(50) DEFAULT 'pending' CHECK (
        status IN ('pending', 'in_progress', 'completed', 'failed', 'skipped')
    ),
    metadata JSONB DEFAULT '{}',
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_onboarding_flows_user_id ON onboarding_flows(user_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_flows_step_type ON onboarding_flows(step_type);
CREATE INDEX IF NOT EXISTS idx_onboarding_flows_status ON onboarding_flows(status);
CREATE INDEX IF NOT EXISTS idx_onboarding_flows_user_step ON onboarding_flows(user_id, step_type);

-- Add comment to table
COMMENT ON TABLE onboarding_flows IS 'Tracks individual onboarding step progress for users';