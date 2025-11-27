-- +goose Up

-- Crear secuencia para números de pedido
CREATE SEQUENCE IF NOT EXISTS order_number_seq START 1;

-- Función para generar número de pedido
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION generate_order_number()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.order_number IS NULL OR NEW.order_number = '' THEN
        NEW.order_number := 'PED-' || TO_CHAR(NOW(), 'YYYY') || '-' || LPAD(NEXTVAL('order_number_seq')::TEXT, 6, '0');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Trigger para auto-generar order_number
DROP TRIGGER IF EXISTS trigger_generate_order_number ON orders;
CREATE TRIGGER trigger_generate_order_number
    BEFORE INSERT ON orders
    FOR EACH ROW
    EXECUTE FUNCTION generate_order_number();

-- Modificar la columna para permitir valores vacíos temporalmente (el trigger los llenará)
ALTER TABLE orders ALTER COLUMN order_number SET DEFAULT '';

-- +goose Down
DROP TRIGGER IF EXISTS trigger_generate_order_number ON orders;
DROP FUNCTION IF EXISTS generate_order_number();
DROP SEQUENCE IF EXISTS order_number_seq;
ALTER TABLE orders ALTER COLUMN order_number DROP DEFAULT;
