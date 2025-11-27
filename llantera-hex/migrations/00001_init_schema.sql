-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE empresas (
    id                    SERIAL PRIMARY KEY,
    clave                 VARCHAR(120) NOT NULL UNIQUE,
    razon_social          VARCHAR(255) NOT NULL,
    rfc                   VARCHAR(30) DEFAULT '',
    direccion             TEXT         DEFAULT '',
    correos               TEXT[]       DEFAULT '{}',
    telefonos             TEXT[]       DEFAULT '{}',
    contacto_principal_id UUID,
    creado_en             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actualizado_en        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE niveles_precios (
    id                   SERIAL PRIMARY KEY,
    codigo               VARCHAR(32) NOT NULL UNIQUE,
    nombre               VARCHAR(120) NOT NULL,
    descripcion          TEXT,
    porcentaje_descuento NUMERIC(5,2) NOT NULL DEFAULT 0,
    columna_precio       VARCHAR(64)  NOT NULL DEFAULT 'public',
    puede_ver_ofertas    BOOLEAN      NOT NULL DEFAULT FALSE,
    creado_en            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    actualizado_en       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE usuarios (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    correo              VARCHAR(180) NOT NULL UNIQUE,
    nombre              VARCHAR(255) NOT NULL,
    primer_nombre       VARCHAR(120) NOT NULL,
    primer_apellido     VARCHAR(120) NOT NULL,
    segundo_apellido    VARCHAR(120) DEFAULT '',
    telefono            VARCHAR(30)  DEFAULT '',
    activo              BOOLEAN      NOT NULL DEFAULT TRUE,
    empresa_id          INTEGER REFERENCES empresas(id) ON DELETE SET NULL,
    url_imagen_perfil   TEXT,
    hash_contrasena     TEXT         NOT NULL,
    rol                 VARCHAR(30)  NOT NULL DEFAULT 'customer',
    nivel_codigo        VARCHAR(32)  NOT NULL DEFAULT 'public',
    nivel_precio_id     INTEGER REFERENCES niveles_precios(id),
    creado_en           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    actualizado_en      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX usuarios_empresa_id_idx ON usuarios(empresa_id);
CREATE INDEX usuarios_nivel_precio_idx ON usuarios(nivel_precio_id);
CREATE INDEX usuarios_correo_trgm_idx ON usuarios USING GIN (correo gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS usuarios_correo_trgm_idx;
DROP INDEX IF EXISTS usuarios_nivel_precio_idx;
DROP INDEX IF EXISTS usuarios_empresa_id_idx;
DROP TABLE IF EXISTS usuarios;
DROP TABLE IF EXISTS niveles_precios;
DROP TABLE IF EXISTS empresas;
