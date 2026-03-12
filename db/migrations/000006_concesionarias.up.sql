BEGIN;

CREATE TABLE IF NOT EXISTS concesionarias (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo VARCHAR(50) NULL,
    nombre VARCHAR(120) NOT NULL,
    estado VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (estado IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_concesionarias_codigo UNIQUE (codigo),
    CONSTRAINT uq_concesionarias_nombre UNIQUE (nombre)
);

CREATE INDEX IF NOT EXISTS idx_concesionarias_estado
    ON concesionarias(estado);

ALTER TABLE porticos
    ADD COLUMN IF NOT EXISTS concesionaria_id UUID NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'porticos'
          AND column_name = 'concesionaria'
    ) THEN
        INSERT INTO concesionarias (nombre)
        SELECT DISTINCT concesionaria
        FROM porticos
        WHERE concesionaria IS NOT NULL
          AND btrim(concesionaria) <> ''
        ON CONFLICT (nombre) DO NOTHING;

        UPDATE porticos p
        SET concesionaria_id = c.id
        FROM concesionarias c
        WHERE p.concesionaria IS NOT NULL
          AND btrim(p.concesionaria) <> ''
          AND p.concesionaria = c.nombre;
    END IF;
END $$;

ALTER TABLE porticos
    ADD CONSTRAINT fk_porticos_concesionaria
    FOREIGN KEY (concesionaria_id) REFERENCES concesionarias(id)
    ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_porticos_concesionaria_id
    ON porticos(concesionaria_id);

DROP INDEX IF EXISTS idx_porticos_concesionaria;
ALTER TABLE porticos DROP COLUMN IF EXISTS concesionaria;

COMMIT;
