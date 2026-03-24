BEGIN;

ALTER TABLE porticos
    ADD COLUMN IF NOT EXISTS entry_latitude NUMERIC(10,6) NULL,
    ADD COLUMN IF NOT EXISTS entry_longitude NUMERIC(10,6) NULL,
    ADD COLUMN IF NOT EXISTS exit_latitude NUMERIC(10,6) NULL,
    ADD COLUMN IF NOT EXISTS exit_longitude NUMERIC(10,6) NULL,
    ADD COLUMN IF NOT EXISTS max_crossing_seconds INT NOT NULL DEFAULT 60;

UPDATE porticos
SET entry_latitude = COALESCE(entry_latitude, latitude),
    entry_longitude = COALESCE(entry_longitude, longitude),
    exit_latitude = COALESCE(exit_latitude, latitude),
    exit_longitude = COALESCE(exit_longitude, longitude)
WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

ALTER TABLE pasos_portico
    ADD COLUMN IF NOT EXISTS entry_timestamp TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS exit_timestamp TIMESTAMPTZ NULL;

COMMIT;
