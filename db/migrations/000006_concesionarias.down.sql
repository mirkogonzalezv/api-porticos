BEGIN;

ALTER TABLE porticos ADD COLUMN IF NOT EXISTS concesionaria VARCHAR(120);

UPDATE porticos p
SET concesionaria = c.nombre
FROM concesionarias c
WHERE p.concesionaria_id = c.id;

CREATE INDEX IF NOT EXISTS idx_porticos_concesionaria
    ON porticos(concesionaria);

DROP INDEX IF EXISTS idx_porticos_concesionaria_id;
ALTER TABLE porticos DROP CONSTRAINT IF EXISTS fk_porticos_concesionaria;
ALTER TABLE porticos DROP COLUMN IF EXISTS concesionaria_id;

DROP INDEX IF EXISTS idx_concesionarias_estado;
DROP TABLE IF EXISTS concesionarias;

COMMIT;
