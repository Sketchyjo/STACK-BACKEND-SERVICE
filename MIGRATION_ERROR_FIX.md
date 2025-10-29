# Quick Fix: Migration 014 Error

## The Error You're Seeing

```
check constraint "chk_wallet_chain" of relation "managed_wallets" is violated by some row
```

## What's Happening

The migration is trying to restrict the database to only allow `SOL-DEVNET` wallets, but you have existing wallets on other chains (ETH, MATIC, etc.) in your database.

## Quick Fix (3 Options)

### Option 1: Automated Script (Easiest & Safest)

```bash
./scripts/clean_non_soldevnet_wallets.sh
```

This will:
- ✓ Show you what will be deleted
- ✓ Create a backup automatically
- ✓ Ask for your confirmation
- ✓ Clean the data

Then run:
```bash
make run
```

---

### Option 2: Manual One-Liner (Fastest)

```bash
psql -d stack_service_dev -c "DELETE FROM managed_wallets WHERE chain != 'SOL-DEVNET';" && make run
```

**Warning**: No backup created with this method!

---

### Option 3: Do Nothing (Let Migration Handle It)

Just run `make run` again. The migration now includes a `DELETE` statement that will automatically remove non-SOL-DEVNET wallets.

---

## What Data Will Be Lost?

- Wallets on chains: ETH, ETH-SEPOLIA, MATIC, MATIC-AMOY, AVAX, BASE, BASE-SEPOLIA, APTOS, APTOS-TESTNET
- Only SOL-DEVNET wallets will remain

## How to See What Will Be Deleted

```bash
psql -d stack_service_dev -c "SELECT chain, COUNT(*) FROM managed_wallets WHERE chain != 'SOL-DEVNET' GROUP BY chain;"
```

## Need More Details?

See: `migrations/README_014_MIGRATION.md`
