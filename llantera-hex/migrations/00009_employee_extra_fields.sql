-- +goose Up

-- Campos adicionales en usuarios para modelar empleados (domicilio y puesto)
ALTER TABLE usuarios
    ADD COLUMN IF NOT EXISTS domicilio_calle VARCHAR(255) DEFAULT '',
    ADD COLUMN IF NOT EXISTS domicilio_numero VARCHAR(50) DEFAULT '',
    ADD COLUMN IF NOT EXISTS domicilio_colonia VARCHAR(180) DEFAULT '',
    ADD COLUMN IF NOT EXISTS domicilio_codigo_postal VARCHAR(20) DEFAULT '',
    ADD COLUMN IF NOT EXISTS puesto VARCHAR(120) DEFAULT '';

-- +goose Down

ALTER TABLE usuarios
    DROP COLUMN IF EXISTS puesto,
    DROP COLUMN IF EXISTS domicilio_codigo_postal,
    DROP COLUMN IF EXISTS domicilio_colonia,
    DROP COLUMN IF EXISTS domicilio_numero,
    DROP COLUMN IF EXISTS domicilio_calle;
