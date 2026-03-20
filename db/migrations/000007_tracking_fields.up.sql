BEGIN;

ALTER TABLE porticos
    ADD COLUMN IF NOT EXISTS bearing_tolerance_deg INT NOT NULL DEFAULT 25,
    ADD COLUMN IF NOT EXISTS entry_radius_meters NUMERIC(8,2) NULL,
    ADD COLUMN IF NOT EXISTS exit_radius_meters NUMERIC(8,2) NULL;

UPDATE porticos
SET entry_radius_meters = COALESCE(entry_radius_meters, detection_radius_meters),
    exit_radius_meters = COALESCE(exit_radius_meters, detection_radius_meters)
WHERE detection_radius_meters IS NOT NULL;

ALTER TABLE pasos_portico
    ADD COLUMN IF NOT EXISTS heading NUMERIC(6,2) NULL,
    ADD COLUMN IF NOT EXISTS speed NUMERIC(8,2) NULL,
    ADD COLUMN IF NOT EXISTS tracking_session_id UUID NULL,
    ADD COLUMN IF NOT EXISTS source_position JSONB NULL;

COMMIT;
