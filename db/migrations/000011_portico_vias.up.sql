BEGIN;

CREATE TABLE IF NOT EXISTS portico_vias (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portico_id UUID NOT NULL REFERENCES porticos(id) ON DELETE CASCADE,
    way_name VARCHAR(120) NOT NULL,
    direction_deg NUMERIC(6,2) NOT NULL DEFAULT 0,
    center_line GEOGRAPHY(LINESTRING, 4326),
    entry_line GEOGRAPHY(LINESTRING, 4326),
    exit_line GEOGRAPHY(LINESTRING, 4326),
    entry_distance_m NUMERIC(8,2) NOT NULL DEFAULT 0,
    exit_distance_m NUMERIC(8,2) NOT NULL DEFAULT 0,
    auto_calculate BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_portico_vias_portico_id ON portico_vias(portico_id);
CREATE INDEX IF NOT EXISTS idx_portico_vias_center_line ON portico_vias USING GIST(center_line);

COMMIT;
