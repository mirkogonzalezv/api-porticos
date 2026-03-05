BEGIN;

CREATE TABLE IF NOT EXISTS pasos_portico (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_supabase_user_id UUID NOT NULL,
    vehiculo_id UUID NOT NULL REFERENCES vehiculos(id) ON DELETE RESTRICT,
    portico_id UUID NOT NULL REFERENCES porticos(id) ON DELETE RESTRICT,
    fecha_hora_paso TIMESTAMPTZ NOT NULL,
    latitud NUMERIC(9,6) NULL CHECK (latitud >= -90 AND latitud <= 90),
    longitud NUMERIC(9,6) NULL CHECK (longitud >= -180 AND longitud <= 180),
    monto_cobrado INTEGER NOT NULL CHECK (monto_cobrado >= 0),
    moneda CHAR(3) NOT NULL DEFAULT 'CLP',
    fuente VARCHAR(20) NOT NULL DEFAULT 'mobile' CHECK (fuente IN ('mobile', 'backend', 'batch')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pasos_owner_fecha
    ON pasos_portico (owner_supabase_user_id, fecha_hora_paso DESC);

CREATE INDEX IF NOT EXISTS idx_pasos_owner_vehiculo_fecha
    ON pasos_portico (owner_supabase_user_id, vehiculo_id, fecha_hora_paso DESC);

CREATE INDEX IF NOT EXISTS idx_pasos_owner_portico_fecha
    ON pasos_portico (owner_supabase_user_id, portico_id, fecha_hora_paso DESC);

COMMIT;
