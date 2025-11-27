-- +goose Up
-- Asegurar columna 'apartadas' en inventario_llantas
ALTER TABLE inventario_llantas
    ADD COLUMN IF NOT EXISTS apartadas INT NOT NULL DEFAULT 0;

COMMENT ON COLUMN inventario_llantas.apartadas IS 'Cantidad de llantas apartadas/reservadas por pedidos pendientes';

-- +goose Down
-- Quitar columna 'apartadas' de inventario_llantas
ALTER TABLE inventario_llantas
    DROP COLUMN IF EXISTS apartadas;
