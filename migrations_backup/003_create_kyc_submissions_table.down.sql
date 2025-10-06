-- Drop kyc_submissions table and indexes
DROP INDEX IF EXISTS idx_kyc_submissions_submitted_at;
DROP INDEX IF EXISTS idx_kyc_submissions_status;
DROP INDEX IF EXISTS idx_kyc_submissions_provider_ref;
DROP INDEX IF EXISTS idx_kyc_submissions_user_id;
DROP TABLE IF EXISTS kyc_submissions;