BEGIN;

ALTER TABLE porticos
    ADD COLUMN IF NOT EXISTS tipo VARCHAR(32) NOT NULL DEFAULT 'urbano';

UPDATE porticos
SET tipo = 'urbano'
WHERE tipo IS NULL OR tipo = '';

COMMIT;
