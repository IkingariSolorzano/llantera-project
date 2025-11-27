-- +goose Up
-- Inserta la empresa base "Llantera de Occidente"
INSERT INTO empresas (
    id,
    clave,
    razon_social,
    rfc,
    direccion,
    correos,
    telefonos,
    contacto_principal_id,
    creado_en,
    actualizado_en
) VALUES (
    1,
    'LDO01',
    'Llantera de Occidente',
    'HUGS550910TY1',
    'Av. Periodismo José Tocaven Lavín 2649, Col. Carlos María de Bustamante, C.P. 58197, Morelia, Michoacán, México',
    ARRAY[]::text[],
    ARRAY['+52 443 326 1312'],
    '3b6a9005-1c43-4f86-a628-1fbb182bd874',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE SET
    clave = EXCLUDED.clave,
    razon_social = EXCLUDED.razon_social,
    rfc = EXCLUDED.rfc,
    direccion = EXCLUDED.direccion,
    correos = EXCLUDED.correos,
    telefonos = EXCLUDED.telefonos,
    contacto_principal_id = EXCLUDED.contacto_principal_id,
    actualizado_en = NOW();

-- +goose Down
DELETE FROM empresas WHERE id = 1;
