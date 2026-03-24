BEGIN;

ALTER TABLE pasos_portico
    DROP COLUMN IF EXISTS entry_timestamp,
    DROP COLUMN IF EXISTS exit_timestamp;

ALTER TABLE porticos
    DROP COLUMN IF EXISTS entry_latitude,
    DROP COLUMN IF EXISTS entry_longitude,
    DROP COLUMN IF EXISTS exit_latitude,
    DROP COLUMN IF EXISTS exit_longitude,
    DROP COLUMN IF EXISTS max_crossing_seconds;

COMMIT;
