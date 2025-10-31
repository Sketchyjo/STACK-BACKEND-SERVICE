-- Drop users table and indexes
DROP INDEX IF EXISTS idx_users_kyc_status;
DROP INDEX IF EXISTS idx_users_onboarding_status;
DROP INDEX IF EXISTS idx_users_auth_provider_id;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;