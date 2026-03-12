BEGIN;

DROP INDEX IF EXISTS idx_porticos_concesionaria;
ALTER TABLE porticos DROP COLUMN IF EXISTS concesionaria;

COMMIT;
