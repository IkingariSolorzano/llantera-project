-- +goose Up

-- Agregar flags de activación y precio público a columnas_precios
ALTER TABLE columnas_precios
    ADD COLUMN IF NOT EXISTS activo BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS es_publico BOOLEAN NOT NULL DEFAULT FALSE;

-- Normalizar valores existentes por si las columnas ya estaban creadas sin default
UPDATE columnas_precios SET activo = TRUE WHERE activo IS NULL;
UPDATE columnas_precios SET es_publico = FALSE WHERE es_publico IS NULL;

-- +goose Down

ALTER TABLE columnas_precios
    DROP COLUMN IF EXISTS es_publico,
    DROP COLUMN IF EXISTS activo;
