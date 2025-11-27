package postgres

import (
	"context"
	"database/sql"

	"github.com/llantera/hex/internal/domain/tire"
)

type TireTypeRepository struct {
	db *sql.DB
}

func NewTireTypeRepository(db *sql.DB) *TireTypeRepository {
	return &TireTypeRepository{db: db}
}

var _ tire.NormalizedTypeRepository = (*TireTypeRepository)(nil)

func (r *TireTypeRepository) GetByName(ctx context.Context, nombre string) (*tire.NormalizedType, error) {
	const query = `
        SELECT id, nombre, descripcion, creado_en, actualizado_en
        FROM tipos_llanta_normalizados
        WHERE UPPER(nombre) = UPPER($1)
    `
	var desc sql.NullString
	entity := &tire.NormalizedType{}
	if err := r.db.QueryRowContext(ctx, query, nombre).Scan(&entity.ID, &entity.Nombre, &desc, &entity.CreadoEn, &entity.ActualizadoEn); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if desc.Valid {
		entity.Descripcion = desc.String
	}
	return entity, nil
}

func (r *TireTypeRepository) Create(ctx context.Context, tipo *tire.NormalizedType) error {
	const query = `
        INSERT INTO tipos_llanta_normalizados (nombre, descripcion)
        VALUES ($1, $2)
        RETURNING id, creado_en, actualizado_en
    `
	return r.db.QueryRowContext(ctx, query, tipo.Nombre, tipo.Descripcion).Scan(&tipo.ID, &tipo.CreadoEn, &tipo.ActualizadoEn)
}

func (r *TireTypeRepository) List(ctx context.Context) ([]tire.NormalizedType, error) {
	const query = `
        SELECT id, nombre, descripcion, creado_en, actualizado_en
        FROM tipos_llanta_normalizados
        ORDER BY nombre
    `
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []tire.NormalizedType
	for rows.Next() {
		var (
			t    tire.NormalizedType
			desc sql.NullString
		)
		if err := rows.Scan(&t.ID, &t.Nombre, &desc, &t.CreadoEn, &t.ActualizadoEn); err != nil {
			return nil, err
		}
		if desc.Valid {
			t.Descripcion = desc.String
		}
		types = append(types, t)
	}
	return types, nil
}
