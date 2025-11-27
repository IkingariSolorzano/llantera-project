-- +goose Up

-- Tabla de direcciones de envío de clientes
CREATE TABLE IF NOT EXISTS customer_addresses (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    alias VARCHAR(100) NOT NULL DEFAULT 'Principal',
    street VARCHAR(255) NOT NULL,
    exterior_number VARCHAR(20) NOT NULL,
    interior_number VARCHAR(20),
    neighborhood VARCHAR(100) NOT NULL,
    postal_code VARCHAR(5) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    reference TEXT,
    phone VARCHAR(15) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_customer_addresses_user_id ON customer_addresses(user_id);

-- Tabla de datos de facturación de clientes
CREATE TABLE IF NOT EXISTS customer_billing_info (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    rfc VARCHAR(13) NOT NULL,
    razon_social VARCHAR(255) NOT NULL,
    regimen_fiscal VARCHAR(10) NOT NULL,
    uso_cfdi VARCHAR(10) NOT NULL,
    postal_code VARCHAR(5) NOT NULL,
    email VARCHAR(255),
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_customer_billing_info_user_id ON customer_billing_info(user_id);

-- Estados de pedido
CREATE TYPE order_status AS ENUM ('solicitado', 'preparando', 'enviado', 'entregado', 'cancelado');

-- Métodos de pago
CREATE TYPE payment_method AS ENUM ('transferencia', 'tarjeta', 'efectivo');

-- Tabla de pedidos
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    order_number VARCHAR(20) NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES usuarios(id),
    status order_status NOT NULL DEFAULT 'solicitado',
    
    -- Dirección de envío (copia al momento del pedido)
    shipping_address_id INTEGER REFERENCES customer_addresses(id),
    shipping_street VARCHAR(255) NOT NULL,
    shipping_exterior_number VARCHAR(20) NOT NULL,
    shipping_interior_number VARCHAR(20),
    shipping_neighborhood VARCHAR(100) NOT NULL,
    shipping_postal_code VARCHAR(5) NOT NULL,
    shipping_city VARCHAR(100) NOT NULL,
    shipping_state VARCHAR(100) NOT NULL,
    shipping_reference TEXT,
    shipping_phone VARCHAR(15) NOT NULL,
    
    -- Pago
    payment_method payment_method NOT NULL,
    
    -- Facturación
    requires_invoice BOOLEAN NOT NULL DEFAULT FALSE,
    billing_info_id INTEGER REFERENCES customer_billing_info(id),
    billing_rfc VARCHAR(13),
    billing_razon_social VARCHAR(255),
    billing_regimen_fiscal VARCHAR(10),
    billing_uso_cfdi VARCHAR(10),
    billing_postal_code VARCHAR(5),
    billing_email VARCHAR(255),
    
    -- Totales
    subtotal DECIMAL(12, 2) NOT NULL DEFAULT 0,
    shipping_cost DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL DEFAULT 0,
    
    -- Archivos de factura
    invoice_xml_path VARCHAR(500),
    invoice_pdf_path VARCHAR(500),
    
    -- Notas
    customer_notes TEXT,
    admin_notes TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    shipped_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);

-- Tabla de items de pedido
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    tire_sku VARCHAR(50) NOT NULL,
    tire_measure VARCHAR(100) NOT NULL,
    tire_brand VARCHAR(100),
    tire_model VARCHAR(100),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(12, 2) NOT NULL,
    subtotal DECIMAL(12, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- +goose Down
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS customer_billing_info;
DROP TABLE IF EXISTS customer_addresses;
DROP TYPE IF EXISTS payment_method;
DROP TYPE IF EXISTS order_status;
