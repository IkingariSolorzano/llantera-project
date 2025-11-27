-- +goose Up
-- Seed inicial de inventario y usuario administrador

-- Extensión para generar hashes bcrypt dentro de PostgreSQL
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Usuario administrador inicial
-- Password por defecto: Admin123!  (cámbiala en producción)
INSERT INTO usuarios (
    correo,
    nombre,
    primer_nombre,
    primer_apellido,
    segundo_apellido,
    telefono,
    activo,
    empresa_id,
    url_imagen_perfil,
    hash_contrasena,
    rol,
    nivel_codigo,
    nivel_precio_id
) VALUES (
    'admin@llantera.com',
    'Administrador Llantera',
    'Administrador',
    'Llantera',
    '',
    '',
    TRUE,
    NULL,
    NULL,
    crypt('Admin123!', gen_salt('bf')),
    'admin',
    'public',
    NULL
);

-- +goose Down
-- Elimina el usuario administrador sembrado por esta migración
DELETE FROM usuarios WHERE correo = 'admin@llantera.com';