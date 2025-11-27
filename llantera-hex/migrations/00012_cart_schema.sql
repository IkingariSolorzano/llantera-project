-- +goose Up

-- Tabla de carrito de compras del usuario
CREATE TABLE IF NOT EXISTS user_carts (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES usuarios(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_carts_user_id ON user_carts(user_id);

-- Tabla de items del carrito
CREATE TABLE IF NOT EXISTS cart_items (
    id SERIAL PRIMARY KEY,
    cart_id INTEGER NOT NULL REFERENCES user_carts(id) ON DELETE CASCADE,
    tire_sku VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    added_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(cart_id, tire_sku)
);

CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX idx_cart_items_tire_sku ON cart_items(tire_sku);

-- +goose Down
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS user_carts;
