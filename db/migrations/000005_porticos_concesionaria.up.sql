BEGIN;

ALTER TABLE porticos
    ADD COLUMN IF NOT EXISTS concesionaria VARCHAR(120) NULL;

CREATE INDEX IF NOT EXISTS idx_porticos_concesionaria
    ON porticos(concesionaria);

COMMIT;
