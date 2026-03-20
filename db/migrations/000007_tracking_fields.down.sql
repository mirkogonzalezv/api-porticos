BEGIN;

ALTER TABLE pasos_portico
    DROP COLUMN IF EXISTS source_position,
    DROP COLUMN IF EXISTS tracking_session_id,
    DROP COLUMN IF EXISTS speed,
    DROP COLUMN IF EXISTS heading;

ALTER TABLE porticos
    DROP COLUMN IF EXISTS exit_radius_meters,
    DROP COLUMN IF EXISTS entry_radius_meters,
    DROP COLUMN IF EXISTS bearing_tolerance_deg;

COMMIT;
