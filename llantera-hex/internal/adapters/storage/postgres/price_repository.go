package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/llantera/hex/internal/domain/tire"
)

type PriceColumnRepository struct {
	db *sql.DB
}

func (r *PriceRepository) ListByColumnID(ctx context.Context, columnID int) ([]tire.TirePrice, error) {
	const query = `
SELECT llanta_id, columna_precio_id, precio, creado_en, actualizado_en
FROM llantas_precios
WHERE columna_precio_id = $1
`

	rows, err := r.db.QueryContext(ctx, query, columnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []tire.TirePrice
	for rows.Next() {
		var entity tire.TirePrice
		if err := rows.Scan(
			&entity.LlantaID,
			&entity.ColumnaPrecioID,
			&entity.Precio,
			&entity.CreadoEn,
			&entity.ActualizadoEn,
		); err != nil {
			return nil, err
		}
		result = append(result, entity)
	}

	return result, nil
}

func NewPriceColumnRepository(db *sql.DB) *PriceColumnRepository {
	return &PriceColumnRepository{db: db}
}

var _ tire.PriceColumnRepository = (*PriceColumnRepository)(nil)

func (r *PriceColumnRepository) GetByCode(ctx context.Context, code string) (*tire.PriceColumn, error) {
	const query = `
SELECT id, codigo, nombre, descripcion, orden_visual, activo, es_publico,
       COALESCE(modo_calculo, ''), codigo_base, COALESCE(operacion, ''), cantidad,
       creado_en, actualizado_en
FROM columnas_precios
WHERE codigo = $1
`

	row := r.db.QueryRowContext(ctx, query, code)
	var entity tire.PriceColumn
	if err := row.Scan(
		&entity.ID,
		&entity.Codigo,
		&entity.Nombre,
		&entity.Descripcion,
		&entity.OrdenVisual,
		&entity.Activo,
		&entity.EsPublico,
		&entity.Mode,
		&entity.BaseCode,
		&entity.Operation,
		&entity.Amount,
		&entity.CreadoEn,
		&entity.ActualizadoEn,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (r *PriceColumnRepository) List(ctx context.Context) ([]tire.PriceColumn, error) {
	const query = `
SELECT id, codigo, nombre, descripcion, orden_visual, activo, es_publico,
       COALESCE(modo_calculo, ''), codigo_base, COALESCE(operacion, ''), cantidad,
       creado_en, actualizado_en
FROM columnas_precios
ORDER BY orden_visual ASC, id ASC
`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []tire.PriceColumn
	for rows.Next() {
		var entity tire.PriceColumn
		if err := rows.Scan(
			&entity.ID,
			&entity.Codigo,
			&entity.Nombre,
			&entity.Descripcion,
			&entity.OrdenVisual,
			&entity.Activo,
			&entity.EsPublico,
			&entity.Mode,
			&entity.BaseCode,
			&entity.Operation,
			&entity.Amount,
			&entity.CreadoEn,
			&entity.ActualizadoEn,
		); err != nil {
			return nil, err
		}
		result = append(result, entity)
	}

	return result, nil
}

type PriceRepository struct {
	db *sql.DB
}

func NewPriceRepository(db *sql.DB) *PriceRepository {
	return &PriceRepository{db: db}
}

var _ tire.PriceRepository = (*PriceRepository)(nil)

func (r *PriceColumnRepository) GetByID(ctx context.Context, id int) (*tire.PriceColumn, error) {
	const query = `
SELECT id, codigo, nombre, descripcion, orden_visual, activo, es_publico,
       COALESCE(modo_calculo, ''), codigo_base, COALESCE(operacion, ''), cantidad,
       creado_en, actualizado_en
FROM columnas_precios
WHERE id = $1
`

	row := r.db.QueryRowContext(ctx, query, id)
	var entity tire.PriceColumn
	if err := row.Scan(
		&entity.ID,
		&entity.Codigo,
		&entity.Nombre,
		&entity.Descripcion,
		&entity.OrdenVisual,
		&entity.Activo,
		&entity.EsPublico,
		&entity.Mode,
		&entity.BaseCode,
		&entity.Operation,
		&entity.Amount,
		&entity.CreadoEn,
		&entity.ActualizadoEn,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &entity, nil
}

func (r *PriceColumnRepository) Create(ctx context.Context, column *tire.PriceColumn) error {
	const query = `
INSERT INTO columnas_precios (
    codigo, nombre, descripcion, orden_visual, activo, es_publico,
    modo_calculo, codigo_base, operacion, cantidad
)
VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, NULLIF($9, ''), $10)
RETURNING id, creado_en, actualizado_en
`

	return r.db.QueryRowContext(
		ctx,
		query,
		column.Codigo,
		column.Nombre,
		column.Descripcion,
		column.OrdenVisual,
		column.Activo,
		column.EsPublico,
		column.Mode,
		column.BaseCode,
		column.Operation,
		column.Amount,
	).Scan(&column.ID, &column.CreadoEn, &column.ActualizadoEn)
}

func (r *PriceColumnRepository) Update(ctx context.Context, column *tire.PriceColumn) error {
	const query = `
UPDATE columnas_precios SET
    nombre = $2,
    descripcion = $3,
    orden_visual = $4,
    activo = $5,
    es_publico = $6,
    modo_calculo = NULLIF($7, ''),
    codigo_base = $8,
    operacion = NULLIF($9, ''),
    cantidad = $10,
    actualizado_en = NOW()
WHERE id = $1
`

	result, err := r.db.ExecContext(
		ctx,
		query,
		column.ID,
		column.Nombre,
		column.Descripcion,
		column.OrdenVisual,
		column.Activo,
		column.EsPublico,
		column.Mode,
		column.BaseCode,
		column.Operation,
		column.Amount,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return tire.ErrPriceColumnNotFound
	}
	return nil
}

func (r *PriceColumnRepository) Delete(ctx context.Context, id int) error {
	const query = `
DELETE FROM columnas_precios WHERE id = $1
`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return tire.ErrPriceColumnNotFound
	}
	return nil
}

func (r *PriceRepository) UpsertMany(ctx context.Context, prices []tire.TirePrice) error {
	if len(prices) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	const query = `
INSERT INTO llantas_precios (llanta_id, columna_precio_id, precio)
VALUES ($1, $2, $3)
ON CONFLICT (llanta_id, columna_precio_id) DO UPDATE
SET precio = EXCLUDED.precio,
    actualizado_en = NOW()
`

	for _, p := range prices {
		if _, errExec := tx.ExecContext(ctx, query, p.LlantaID, p.ColumnaPrecioID, p.Precio); errExec != nil {
			err = errExec
			return err
		}
	}

	return nil
}

func (r *PriceRepository) ListByTireID(ctx context.Context, llantaID string) ([]tire.TirePrice, error) {
	const query = `
SELECT llanta_id, columna_precio_id, precio, creado_en, actualizado_en
FROM llantas_precios
WHERE llanta_id = $1
`

	rows, err := r.db.QueryContext(ctx, query, llantaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []tire.TirePrice
	for rows.Next() {
		var entity tire.TirePrice
		if err := rows.Scan(
			&entity.LlantaID,
			&entity.ColumnaPrecioID,
			&entity.Precio,
			&entity.CreadoEn,
			&entity.ActualizadoEn,
		); err != nil {
			return nil, err
		}
		result = append(result, entity)
	}

	return result, nil
}
