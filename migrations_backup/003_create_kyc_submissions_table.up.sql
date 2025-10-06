-- Create kyc_submissions table for tracking KYC submission status
CREATE TABLE IF NOT EXISTS kyc_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_ref VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' CHECK (
        status IN ('pending', 'processing', 'approved', 'rejected')
    ),
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reviewed_at TIMESTAMP,
    rejection_reasons JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_kyc_submissions_user_id ON kyc_submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_kyc_submissions_provider_ref ON kyc_submissions(provider_ref);
CREATE INDEX IF NOT EXISTS idx_kyc_submissions_status ON kyc_submissions(status);
CREATE INDEX IF NOT EXISTS idx_kyc_submissions_submitted_at ON kyc_submissions(submitted_at);

-- Add comment to table
COMMENT ON TABLE kyc_submissions IS 'Tracks KYC document submissions and their review status';