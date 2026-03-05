BEGIN;

DROP INDEX IF EXISTS idx_pasos_owner_portico_fecha;
DROP INDEX IF EXISTS idx_pasos_owner_vehiculo_fecha;
DROP INDEX IF EXISTS idx_pasos_owner_fecha;
DROP TABLE IF EXISTS pasos_portico;

COMMIT;
