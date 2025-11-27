package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/llantera/hex/internal/domain/tire"
)

type TireBrandRepository struct {
	db *sql.DB
}

func NewTireBrandRepository(db *sql.DB) *TireBrandRepository {
	return &TireBrandRepository{db: db}
}

var _ tire.BrandRepository = (*TireBrandRepository)(nil)

func (r *TireBrandRepository) GetByName(ctx context.Context, nombre string) (*tire.Brand, error) {
	const query = `
        SELECT id, nombre, creado_en, actualizado_en
        FROM marcas_llantas
        WHERE UPPER(nombre) = UPPER($1)
    `
	brand := &tire.Brand{}
	if err := r.db.QueryRowContext(ctx, query, nombre).Scan(&brand.ID, &brand.Nombre, &brand.CreadoEn, &brand.ActualizadoEn); err != nil {
		if err == sql.ErrNoRows {
			return nil, tire.ErrBrandNotFound
		}
		return nil, err
	}
	return brand, nil
}

func (r *TireBrandRepository) GetByAlias(ctx context.Context, alias string) (*tire.Brand, error) {
	const query = `
        SELECT m.id, m.nombre, m.creado_en, m.actualizado_en
        FROM marcas_llantas m
        JOIN alias_marcas a ON a.marca_id = m.id
        WHERE UPPER(a.alias) = UPPER($1)
    `
	brand := &tire.Brand{}
	if err := r.db.QueryRowContext(ctx, query, alias).Scan(&brand.ID, &brand.Nombre, &brand.CreadoEn, &brand.ActualizadoEn); err != nil {
		if err == sql.ErrNoRows {
			return nil, tire.ErrBrandNotFound
		}
		return nil, err
	}
	return brand, nil
}

func (r *TireBrandRepository) Create(ctx context.Context, marca *tire.Brand, aliases []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const insertBrand = `
        INSERT INTO marcas_llantas (nombre)
        VALUES ($1)
        RETURNING id, creado_en, actualizado_en
    `

	if err := tx.QueryRowContext(ctx, insertBrand, marca.Nombre).Scan(&marca.ID, &marca.CreadoEn, &marca.ActualizadoEn); err != nil {
		return err
	}

	const insertAlias = `
        INSERT INTO alias_marcas (marca_id, alias)
        VALUES ($1, UPPER($2))
        ON CONFLICT (marca_id, alias) DO NOTHING
    `
	for _, alias := range aliases {
		trimmed := strings.TrimSpace(alias)
		if trimmed == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, insertAlias, marca.ID, strings.ToUpper(trimmed)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TireBrandRepository) List(ctx context.Context) ([]tire.Brand, error) {
	const query = `
        SELECT id, nombre, creado_en, actualizado_en
        FROM marcas_llantas
        ORDER BY nombre
    `
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []tire.Brand
	for rows.Next() {
		var b tire.Brand
		if err := rows.Scan(&b.ID, &b.Nombre, &b.CreadoEn, &b.ActualizadoEn); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, nil
}

func (r *TireBrandRepository) GetByID(ctx context.Context, id int) (*tire.Brand, error) {
	const query = `
		SELECT id, nombre, creado_en, actualizado_en
		FROM marcas_llantas
		WHERE id = $1
	`
	brand := &tire.Brand{}
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&brand.ID, &brand.Nombre, &brand.CreadoEn, &brand.ActualizadoEn); err != nil {
		if err == sql.ErrNoRows {
			return nil, tire.ErrBrandNotFound
		}
		return nil, err
	}
	return brand, nil
}

func (r *TireBrandRepository) ListAliases(ctx context.Context, brandID int) ([]string, error) {
	const query = `
		SELECT alias
		FROM alias_marcas
		WHERE marca_id = $1
		ORDER BY alias
	`
	rows, err := r.db.QueryContext(ctx, query, brandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aliases []string
	for rows.Next() {
		var alias string
		if err := rows.Scan(&alias); err != nil {
			return nil, err
		}
		aliases = append(aliases, alias)
	}
	return aliases, nil
}

func (r *TireBrandRepository) Update(ctx context.Context, marca *tire.Brand, aliases []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const updateBrand = `
		UPDATE marcas_llantas
		SET nombre = $1, actualizado_en = NOW()
		WHERE id = $2
	`
	if _, err := tx.ExecContext(ctx, updateBrand, marca.Nombre, marca.ID); err != nil {
		return err
	}

	const deleteAliases = `
		DELETE FROM alias_marcas
		WHERE marca_id = $1
	`
	if _, err := tx.ExecContext(ctx, deleteAliases, marca.ID); err != nil {
		return err
	}

	const insertAlias = `
		INSERT INTO alias_marcas (marca_id, alias)
		VALUES ($1, UPPER($2))
		ON CONFLICT (marca_id, alias) DO NOTHING
	`
	for _, alias := range aliases {
		trimmed := strings.TrimSpace(alias)
		if trimmed == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, insertAlias, marca.ID, strings.ToUpper(trimmed)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TireBrandRepository) Delete(ctx context.Context, id int) error {
	const query = `
		DELETE FROM marcas_llantas
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *TireBrandRepository) HasTires(ctx context.Context, id int) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM llantas WHERE marca_id = $1
		)
	`
	var exists bool
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
