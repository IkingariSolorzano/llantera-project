package postgres

import (
	"context"
	"database/sql"

	"github.com/llantera/hex/internal/domain/tire"
)

type InventoryRepository struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

var _ tire.InventoryRepository = (*InventoryRepository)(nil)

func (r *InventoryRepository) Upsert(ctx context.Context, entity *tire.Inventory) error {
	const query = `
INSERT INTO inventario_llantas (llanta_id, cantidad, apartadas, stock_minimo)
VALUES ($1, $2, $3, $4)
ON CONFLICT (llanta_id) DO UPDATE
SET cantidad = EXCLUDED.cantidad,
    apartadas = EXCLUDED.apartadas,
    stock_minimo = EXCLUDED.stock_minimo,
    actualizado_en = NOW()
RETURNING id, llanta_id, cantidad, COALESCE(apartadas, 0), stock_minimo, creado_en, actualizado_en
`

	row := r.db.QueryRowContext(ctx, query, entity.LlantaID, entity.Cantidad, entity.Apartadas, entity.StockMinimo)
	return scanInventory(row, entity)
}

func (r *InventoryRepository) GetByTireID(ctx context.Context, llantaID string) (*tire.Inventory, error) {
	const query = `
SELECT id, llanta_id, cantidad, COALESCE(apartadas, 0), stock_minimo, creado_en, actualizado_en
FROM inventario_llantas
WHERE llanta_id = $1
`

	row := r.db.QueryRowContext(ctx, query, llantaID)
	inv := &tire.Inventory{}
	if err := scanInventory(row, inv); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return inv, nil
}

// ReserveStock resta de cantidad y suma a apartadas (al crear pedido)
// Flujo: cantidad -= N, apartadas += N
func (r *InventoryRepository) ReserveStock(ctx context.Context, sku string, quantity int) error {
	const query = `
UPDATE inventario_llantas 
SET cantidad = GREATEST(0, cantidad - $2),
    apartadas = COALESCE(apartadas, 0) + $2,
    actualizado_en = NOW()
WHERE llanta_id = (SELECT id FROM llantas WHERE LOWER(sku) = LOWER($1))
`
	_, err := r.db.ExecContext(ctx, query, sku, quantity)
	return err
}

// ReleaseStock devuelve stock al cancelar pedido
// Flujo: cantidad += N, apartadas -= N
func (r *InventoryRepository) ReleaseStock(ctx context.Context, sku string, quantity int) error {
	const query = `
UPDATE inventario_llantas 
SET cantidad = cantidad + $2,
    apartadas = GREATEST(0, COALESCE(apartadas, 0) - $2),
    actualizado_en = NOW()
WHERE llanta_id = (SELECT id FROM llantas WHERE LOWER(sku) = LOWER($1))
`
	_, err := r.db.ExecContext(ctx, query, sku, quantity)
	return err
}

// ConfirmSale libera apartadas al entregar (cantidad ya fue restada al crear pedido)
// Flujo: apartadas -= N (cantidad ya se restó en ReserveStock)
func (r *InventoryRepository) ConfirmSale(ctx context.Context, sku string, quantity int) error {
	const query = `
UPDATE inventario_llantas 
SET apartadas = GREATEST(0, COALESCE(apartadas, 0) - $2),
    actualizado_en = NOW()
WHERE llanta_id = (SELECT id FROM llantas WHERE LOWER(sku) = LOWER($1))
`
	_, err := r.db.ExecContext(ctx, query, sku, quantity)
	return err
}

func scanInventory(row *sql.Row, dst *tire.Inventory) error {
	return row.Scan(
		&dst.ID,
		&dst.LlantaID,
		&dst.Cantidad,
		&dst.Apartadas,
		&dst.StockMinimo,
		&dst.CreadoEn,
		&dst.ActualizadoEn,
	)
}
