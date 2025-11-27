-- +goose Up
-- Agregar campo de modalidad de pago a pedidos
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_mode VARCHAR(50) DEFAULT 'contado';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_installments INTEGER DEFAULT 1;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_notes TEXT;

-- Comentarios descriptivos
COMMENT ON COLUMN orders.payment_mode IS 'Modalidad de pago: contado, credito, parcialidades, anticipo';
COMMENT ON COLUMN orders.payment_installments IS 'Número de parcialidades (solo aplica si payment_mode = parcialidades)';
COMMENT ON COLUMN orders.payment_notes IS 'Notas adicionales sobre el pago';

-- +goose Down
ALTER TABLE orders DROP COLUMN IF EXISTS payment_mode;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_installments;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_notes;
