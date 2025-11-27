-- +goose Up
CREATE TABLE marcas_llantas (
    id             SERIAL PRIMARY KEY,
    nombre         VARCHAR(120) NOT NULL UNIQUE,
    creado_en      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE alias_marcas (
    id             SERIAL PRIMARY KEY,
    marca_id       INTEGER      NOT NULL REFERENCES marcas_llantas(id) ON DELETE CASCADE,
    alias          VARCHAR(60)  NOT NULL,
    creado_en      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    actualizado_en TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (marca_id, alias)
);

CREATE TABLE tipos_llanta_normalizados (
    id             SERIAL PRIMARY KEY,
    nombre         VARCHAR(120) NOT NULL UNIQUE,
    descripcion    TEXT,
    creado_en      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE llantas (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku                   VARCHAR(120) NOT NULL UNIQUE,
    marca_id              INTEGER      NOT NULL REFERENCES marcas_llantas(id),
    modelo                VARCHAR(180) NOT NULL,
    ancho                 INTEGER      NOT NULL,
    perfil                INTEGER,
    rin                   NUMERIC(5,2) NOT NULL,
    construccion          CHAR(1)      CHECK (construccion IN ('R','D')),
    tipo_tubo             VARCHAR(2)   CHECK (tipo_tubo IN ('TL','TT')),
    calificacion_capas    VARCHAR(20),
    indice_carga          VARCHAR(20),
    indice_velocidad      VARCHAR(5),
    tipo_normalizado_id   INTEGER REFERENCES tipos_llanta_normalizados(id),
    abreviatura_uso       VARCHAR(10),
    descripcion           TEXT,
    precio_publico        NUMERIC(12,2),
    url_imagen            TEXT,
    medida_original       TEXT,
    creado_en             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actualizado_en        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE equivalencias_llanta (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    llanta_base_id     UUID NOT NULL REFERENCES llantas(id) ON DELETE CASCADE,
    llanta_equivalente UUID NOT NULL REFERENCES llantas(id) ON DELETE CASCADE,
    notas              TEXT,
    creado_en          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actualizado_en     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (llanta_base_id, llanta_equivalente)
);

CREATE INDEX llantas_marca_idx ON llantas(marca_id);
CREATE INDEX llantas_tipo_normalizado_idx ON llantas(tipo_normalizado_id);
CREATE INDEX llantas_abreviatura_idx ON llantas(abreviatura_uso);

INSERT INTO marcas_llantas (nombre) VALUES
    ('AB Tires'),
    ('Aurora Tires'),
    ('Bridgestone'),
    ('Dayton'),
    ('Double Coin'),
    ('Firestone'),
    ('Fuzion'),
    ('Goodyear'),
    ('Goodride'),
    ('Hankook'),
    ('Kumho'),
    ('Laufenn'),
    ('OTR Tires'),
    ('Otras Marcas'),
    ('Pirelli'),
    ('Sumitomo'),
    ('Tornel')
ON CONFLICT (nombre) DO NOTHING;

INSERT INTO alias_marcas (marca_id, alias)
SELECT ml.id, am.alias
FROM marcas_llantas ml
JOIN (VALUES
        ('AB Tires','AB'),
        ('Aurora Tires','AURORA'),
        ('Bridgestone','BS'),
        ('Dayton','DAYTON'),
        ('Double Coin','DOUBLE COIN'),
        ('Firestone','FS'),
        ('Fuzion','FUZION'),
        ('Goodyear','GDY'),
        ('Goodyear','GOODYEAR'),
        ('Goodride','GOO'),
        ('Hankook','HAN'),
        ('Hankook','HANKOOK'),
        ('Kumho','KUM'),
        ('Laufenn','LAUFENN'),
        ('OTR Tires','OTR'),
        ('Otras Marcas','OTRAS'),
        ('Pirelli','PIRELLI'),
        ('Sumitomo','SUM'),
        ('Sumitomo','SUMITOMO'),
        ('Tornel','TOR'),
        ('Tornel','TORNEL')
    ) AS am(nombre, alias)
    ON ml.nombre = am.nombre
ON CONFLICT (marca_id, alias) DO NOTHING;

INSERT INTO tipos_llanta_normalizados (nombre) VALUES
    ('Agrícola Convencional'),
    ('Agrícola Radial'),
    ('Camión Convencional'),
    ('Camión Radial'),
    ('Camioneta Convencional'),
    ('Camioneta Radial'),
    ('Industrial Convencional'),
    ('Industrial Radial'),
    ('Pasajero'),
    ('Pasajero Radial (PSR)'),
    ('Light Truck Radial (LTR)'),
    ('Light Truck Convencional (LTS)'),
    ('Moto Convencional'),
    ('Moto Radial'),
    ('Special Trailer (ST)'),
    ('Truck & Bus Radial (TBR)'),
    ('Llanta Temporal'),
    ('Otros')
ON CONFLICT (nombre) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS equivalencias_llanta;
DROP TABLE IF EXISTS llantas;
DROP TABLE IF EXISTS tipos_llanta_normalizados;
DROP TABLE IF EXISTS alias_marcas;
DROP TABLE IF EXISTS marcas_llantas;
