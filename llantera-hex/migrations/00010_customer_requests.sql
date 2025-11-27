-- +goose Up

-- Tabla para solicitudes de "Quiero ser cliente" desde la landing
CREATE TABLE solicitudes_clientes (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nombre_completo       VARCHAR(255) NOT NULL,
    tipo_solicitud        VARCHAR(100) NOT NULL,
    mensaje               TEXT,
    telefono              VARCHAR(30),
    preferencia_contacto  VARCHAR(20), -- whatsapp | llamada u otro texto libre
    correo                VARCHAR(180),
    estado                VARCHAR(16) NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente','vista','atendida')),
    empleado_id           UUID REFERENCES usuarios(id) ON DELETE SET NULL,
    acuerdo               TEXT, -- qué se acordó al contactar al cliente
    creado_en             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actualizado_en        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    atendido_en           TIMESTAMPTZ
);

CREATE INDEX solicitudes_clientes_estado_idx ON solicitudes_clientes(estado);
CREATE INDEX solicitudes_clientes_empleado_idx ON solicitudes_clientes(empleado_id);
CREATE INDEX solicitudes_clientes_creado_en_idx ON solicitudes_clientes(creado_en);

-- +goose Down

DROP INDEX IF EXISTS solicitudes_clientes_creado_en_idx;
DROP INDEX IF EXISTS solicitudes_clientes_empleado_idx;
DROP INDEX IF EXISTS solicitudes_clientes_estado_idx;
DROP TABLE IF EXISTS solicitudes_clientes;
