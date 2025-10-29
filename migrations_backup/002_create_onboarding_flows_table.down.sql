-- Drop onboarding_flows table and indexes
DROP INDEX IF EXISTS idx_onboarding_flows_user_step;
DROP INDEX IF EXISTS idx_onboarding_flows_status;
DROP INDEX IF EXISTS idx_onboarding_flows_step_type;
DROP INDEX IF EXISTS idx_onboarding_flows_user_id;
DROP TABLE IF EXISTS onboarding_flows;