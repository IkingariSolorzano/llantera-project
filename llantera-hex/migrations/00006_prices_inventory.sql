-- +goose Up

-- 1. Inventario por llanta
CREATE TABLE inventario_llantas (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    llanta_id      UUID NOT NULL REFERENCES llantas(id) ON DELETE CASCADE,
    cantidad       INT  NOT NULL DEFAULT 0,
    stock_minimo   INT  NOT NULL DEFAULT 4,
    creado_en      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (llanta_id)
);

-- 2. Catálogo de tipos de precio
CREATE TABLE columnas_precios (
    id             SERIAL PRIMARY KEY,
    codigo         VARCHAR(64)  NOT NULL UNIQUE,
    nombre         VARCHAR(120) NOT NULL,
    descripcion    TEXT,
    orden_visual   INT          NOT NULL DEFAULT 0,
    creado_en      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    actualizado_en TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 3. Precios por llanta
CREATE TABLE llantas_precios (
    llanta_id         UUID NOT NULL REFERENCES llantas(id) ON DELETE CASCADE,
    columna_precio_id INT  NOT NULL REFERENCES columnas_precios(id) ON DELETE CASCADE,
    precio            NUMERIC(12,2) NOT NULL,
    creado_en         TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    actualizado_en    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    PRIMARY KEY (llanta_id, columna_precio_id)
);

-- 4. Ajuste a niveles de precios
ALTER TABLE niveles_precios
    ADD COLUMN columna_referencia VARCHAR(64);

-- 5. Semilla de columnas de precio
INSERT INTO columnas_precios (codigo, nombre, descripcion, orden_visual)
VALUES
    ('lista',      'Precio de lista',        'Precio público de lista',                       10),
    ('lista_10',   'Precio lista -10%',      'Precio de lista con 10% de descuento',          20),
    ('mayoreo',    'Precio mayoreo',         'Precio para distribuidores',                    30),
    ('mayoreo_6',  'Precio mayoreo -6%',     'Precio mayoreo con 6% de descuento',            40),
    ('mayoreo_3',  'Precio mayoreo -3%',     'Precio mayoreo con 3% de descuento',            50),
    ('empresa',    'Precio empresas',        'Precio para cuentas de empresa',                60),
    ('efectivo',   'Precio efectivo',        'Precio para pago en efectivo',                  70)
ON CONFLICT (codigo) DO NOTHING;

-- 6. Semilla opcional de niveles de precios
INSERT INTO niveles_precios (
    codigo,
    nombre,
    descripcion,
    porcentaje_descuento,
    columna_precio,
    columna_referencia,
    puede_ver_ofertas
)
VALUES
    ('public',       'Público general', 'Clientes sin cuenta',     0, 'lista',     NULL,     TRUE),
    ('empresa',      'Empresas',        'Clientes empresariales',  0, 'empresa',   'lista',  TRUE),
    ('distribuidor', 'Distribuidor',    'Clientes distribuidores', 0, 'mayoreo',   'lista',  TRUE),
    ('mayorista',    'Mayorista',       'Clientes mayoristas',     0, 'mayoreo_6', 'lista',  TRUE)
ON CONFLICT (codigo) DO NOTHING;

-- 7. Inicializar inventario para llantas existentes (stock 0, stock_minimo 4)
INSERT INTO inventario_llantas (llanta_id, cantidad, stock_minimo)
SELECT id, 0, 4
FROM llantas
ON CONFLICT (llanta_id) DO NOTHING;

-- 8. Inicializar precio 'lista' a partir de precio_publico existente
INSERT INTO llantas_precios (llanta_id, columna_precio_id, precio)
SELECT
    l.id,
    cp.id,
    COALESCE(l.precio_publico, 0)
FROM llantas l
CROSS JOIN columnas_precios cp
WHERE cp.codigo = 'lista'
  AND NOT EXISTS (
      SELECT 1
      FROM llantas_precios lp
      WHERE lp.llanta_id = l.id
        AND lp.columna_precio_id = cp.id
  );

-- +goose Down

-- 1. Revertir cambio en niveles de precios
ALTER TABLE niveles_precios
    DROP COLUMN IF EXISTS columna_referencia;

-- 2. Eliminar tablas de precios e inventario
DROP TABLE IF EXISTS llantas_precios;
DROP TABLE IF EXISTS inventario_llantas;
DROP TABLE IF EXISTS columnas_precios;
