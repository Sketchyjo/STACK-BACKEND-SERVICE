-- Drop triggers
DROP TRIGGER IF EXISTS update_basket_allocations_updated_at ON basket_allocations;
DROP TRIGGER IF EXISTS update_baskets_updated_at ON baskets;
DROP TRIGGER IF EXISTS update_balances_updated_at ON balances;
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;
DROP TRIGGER IF EXISTS update_tokens_updated_at ON tokens;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS basket_allocations;
DROP TABLE IF EXISTS baskets;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS balances;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS users;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";