CREATE TABLE IF NOT EXISTS pasos_capturados (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_supabase_user_id UUID NOT NULL,
    vehiculo_id UUID NOT NULL,
    portico_id UUID NOT NULL,
    via_id UUID NULL,
    fecha_hora_inicio TIMESTAMPTZ NOT NULL,
    fecha_hora_fin TIMESTAMPTZ NOT NULL,
    entry_timestamp TIMESTAMPTZ NULL,
    exit_timestamp TIMESTAMPTZ NULL,
    entry_hit BOOLEAN NOT NULL DEFAULT FALSE,
    exit_hit BOOLEAN NOT NULL DEFAULT FALSE,
    heading_avg NUMERIC(6,2) NULL,
    speed_avg NUMERIC(8,2) NULL,
    direccion_paso VARCHAR(3) NULL,
    status VARCHAR(16) NOT NULL,
    source_position JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pasos_capturados_owner ON pasos_capturados(owner_supabase_user_id);
CREATE INDEX IF NOT EXISTS idx_pasos_capturados_vehiculo ON pasos_capturados(vehiculo_id);
CREATE INDEX IF NOT EXISTS idx_pasos_capturados_portico ON pasos_capturados(portico_id);
CREATE INDEX IF NOT EXISTS idx_pasos_capturados_created_at ON pasos_capturados(created_at);
