-- +goose Up

-- Agregar configuración opcional de reglas de cálculo a columnas_precios
ALTER TABLE columnas_precios
    ADD COLUMN IF NOT EXISTS modo_calculo VARCHAR(16),
    ADD COLUMN IF NOT EXISTS codigo_base VARCHAR(64),
    ADD COLUMN IF NOT EXISTS operacion VARCHAR(16),
    ADD COLUMN IF NOT EXISTS cantidad NUMERIC(12,2);

-- +goose Down

ALTER TABLE columnas_precios
    DROP COLUMN IF EXISTS cantidad,
    DROP COLUMN IF EXISTS operacion,
    DROP COLUMN IF EXISTS codigo_base,
    DROP COLUMN IF EXISTS modo_calculo;
