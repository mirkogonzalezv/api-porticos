BEGIN;

ALTER TABLE porticos
    DROP COLUMN IF EXISTS bearing_tolerance_deg;

COMMIT;
