BEGIN;

DROP TRIGGER IF EXISTS trigger_families_updated_at ON families;
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP TABLE IF EXISTS families CASCADE;

COMMIT;