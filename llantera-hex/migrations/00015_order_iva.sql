-- +goose Up
-- Agregar campo IVA a la tabla de pedidos

ALTER TABLE orders ADD COLUMN IF NOT EXISTS iva DECIMAL(10, 2) DEFAULT 0;

-- Actualizar pedidos existentes para calcular el IVA (16%)
UPDATE orders SET iva = subtotal * 0.16 WHERE iva = 0 OR iva IS NULL;

-- Actualizar total para incluir IVA en pedidos existentes
UPDATE orders SET total = subtotal + iva + shipping_cost WHERE total = subtotal + shipping_cost;

-- +goose Down
ALTER TABLE orders DROP COLUMN IF EXISTS iva;
