-- +goose Up
ALTER TABLE usuarios
    DROP COLUMN IF EXISTS nombre,
    DROP COLUMN IF EXISTS nivel_codigo;

-- +goose Down
ALTER TABLE usuarios
    ADD COLUMN IF NOT EXISTS nombre VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS nivel_codigo VARCHAR(32) NOT NULL DEFAULT 'public';

UPDATE usuarios
SET nombre = TRIM(CONCAT_WS(' ', primer_nombre, primer_apellido, segundo_apellido))
WHERE nombre = '';

UPDATE usuarios u
SET nivel_codigo = COALESCE(np.codigo, 'public')
FROM niveles_precios np
WHERE u.nivel_precio_id = np.id;
