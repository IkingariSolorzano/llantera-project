-- +goose Up
-- Agregar columna de llantas apartadas/reservadas en inventario
ALTER TABLE inventario_llantas
    ADD COLUMN apartadas INT NOT NULL DEFAULT 0;

-- Comentario para documentación
COMMENT ON COLUMN inventario_llantas.apartadas IS 'Cantidad de llantas apartadas/reservadas por pedidos pendientes';

-- +goose Down
-- Revertir columna de llantas apartadas/reservadas en inventario
ALTER TABLE inventario_llantas
    DROP COLUMN IF EXISTS apartadas;
