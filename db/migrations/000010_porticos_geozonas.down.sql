BEGIN;

ALTER TABLE pasos_portico
    DROP COLUMN IF EXISTS direccion_paso;

ALTER TABLE porticos
    DROP COLUMN IF EXISTS direccion,
    DROP COLUMN IF EXISTS velocidad_maxima,
    DROP COLUMN IF EXISTS zona_de_deteccion,
    DROP COLUMN IF EXISTS vehicle_types,
    DROP COLUMN IF EXISTS is_active;

COMMIT;
