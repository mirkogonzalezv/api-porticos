BEGIN;

CREATE EXTENSION IF NOT EXISTS postgis;

ALTER TABLE porticos
    ADD COLUMN IF NOT EXISTS direccion VARCHAR(4) NOT NULL DEFAULT 'N',
    ADD COLUMN IF NOT EXISTS velocidad_maxima INT NOT NULL DEFAULT 60,
    ADD COLUMN IF NOT EXISTS zona_de_deteccion GEOGRAPHY(POLYGON, 4326),
    ADD COLUMN IF NOT EXISTS vehicle_types JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE porticos
SET direccion = 'N'
WHERE direccion IS NULL OR direccion = '';

UPDATE porticos
SET velocidad_maxima = 60
WHERE velocidad_maxima IS NULL OR velocidad_maxima <= 0;

UPDATE porticos
SET vehicle_types = '[]'::jsonb
WHERE vehicle_types IS NULL;

ALTER TABLE pasos_portico
    ADD COLUMN IF NOT EXISTS direccion_paso VARCHAR(4);

COMMIT;
