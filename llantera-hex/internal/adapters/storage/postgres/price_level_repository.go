package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/llantera/hex/internal/domain/pricelevel"
)

type priceLevelRepository struct {
	db *sql.DB
}

func NewPriceLevelRepository(db *sql.DB) pricelevel.PriceLevelRepository {
	return &priceLevelRepository{db: db}
}

func (r *priceLevelRepository) Create(level *pricelevel.PriceLevel) (*pricelevel.PriceLevel, error) {
	query := `
		INSERT INTO niveles_precios (
			codigo, nombre, descripcion, porcentaje_descuento,
			columna_precio, columna_referencia, puede_ver_ofertas
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, creado_en, actualizado_en`

	var id int
	var createdAt, updatedAt pq.NullTime

	err := r.db.QueryRow(query,
		level.Code, level.Name, level.Description, level.DiscountPercentage,
		level.PriceColumn, level.ReferenceColumn, level.CanViewOffers,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("código de nivel de precio ya existe: %s", level.Code)
		}
		return nil, err
	}

	level.ID = id
	if createdAt.Valid {
		level.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		level.UpdatedAt = updatedAt.Time
	}

	return level, nil
}

func (r *priceLevelRepository) GetByID(id int) (*pricelevel.PriceLevel, error) {
	query := `
		SELECT id, codigo, nombre, descripcion, porcentaje_descuento,
			   columna_precio, columna_referencia, puede_ver_ofertas,
			   creado_en, actualizado_en
		FROM niveles_precios
		WHERE id = $1`

	var level pricelevel.PriceLevel
	var description, referenceColumn sql.NullString
	var createdAt, updatedAt pq.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&level.ID, &level.Code, &level.Name, &description, &level.DiscountPercentage,
		&level.PriceColumn, &referenceColumn, &level.CanViewOffers,
		&createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("nivel de precio no encontrado: %d", id)
		}
		return nil, err
	}

	if description.Valid {
		level.Description = &description.String
	}
	if referenceColumn.Valid {
		level.ReferenceColumn = &referenceColumn.String
	}
	if createdAt.Valid {
		level.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		level.UpdatedAt = updatedAt.Time
	}

	return &level, nil
}

func (r *priceLevelRepository) GetByCode(code string) (*pricelevel.PriceLevel, error) {
	query := `
		SELECT id, codigo, nombre, descripcion, porcentaje_descuento,
			   columna_precio, columna_referencia, puede_ver_ofertas,
			   creado_en, actualizado_en
		FROM niveles_precios
		WHERE codigo = $1`

	var level pricelevel.PriceLevel
	var description, referenceColumn sql.NullString
	var createdAt, updatedAt pq.NullTime

	err := r.db.QueryRow(query, code).Scan(
		&level.ID, &level.Code, &level.Name, &description, &level.DiscountPercentage,
		&level.PriceColumn, &referenceColumn, &level.CanViewOffers,
		&createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("nivel de precio no encontrado: %s", code)
		}
		return nil, err
	}

	if description.Valid {
		level.Description = &description.String
	}
	if referenceColumn.Valid {
		level.ReferenceColumn = &referenceColumn.String
	}
	if createdAt.Valid {
		level.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		level.UpdatedAt = updatedAt.Time
	}

	return &level, nil
}

func (r *priceLevelRepository) List(filter pricelevel.PriceLevelFilter) ([]*pricelevel.PriceLevel, int, error) {
	whereClauses := []string{}
	args := []interface{}{}
	argCount := 0

	if filter.Code != nil && *filter.Code != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("codigo ILIKE $%d", argCount))
		args = append(args, "%"+*filter.Code+"%")
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM niveles_precios %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Data query
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, codigo, nombre, descripcion, porcentaje_descuento,
			   columna_precio, columna_referencia, puede_ver_ofertas,
			   creado_en, actualizado_en
		FROM niveles_precios
		%s
		ORDER BY nombre
		LIMIT %d OFFSET %d`, whereClause, limit, offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var levels []*pricelevel.PriceLevel
	for rows.Next() {
		var level pricelevel.PriceLevel
		var description, referenceColumn sql.NullString
		var createdAt, updatedAt pq.NullTime

		err := rows.Scan(
			&level.ID, &level.Code, &level.Name, &description, &level.DiscountPercentage,
			&level.PriceColumn, &referenceColumn, &level.CanViewOffers,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if description.Valid {
			level.Description = &description.String
		}
		if referenceColumn.Valid {
			level.ReferenceColumn = &referenceColumn.String
		}
		if createdAt.Valid {
			level.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			level.UpdatedAt = updatedAt.Time
		}

		levels = append(levels, &level)
	}

	return levels, total, nil
}

func (r *priceLevelRepository) Update(id int, level *pricelevel.PriceLevel) (*pricelevel.PriceLevel, error) {
	query := `
		UPDATE niveles_precios
		SET codigo = $2, nombre = $3, descripcion = $4, porcentaje_descuento = $5,
			columna_precio = $6, columna_referencia = $7, puede_ver_ofertas = $8,
			actualizado_en = NOW()
		WHERE id = $1
		RETURNING creado_en, actualizado_en`

	var createdAt, updatedAt pq.NullTime

	err := r.db.QueryRow(query,
		id, level.Code, level.Name, level.Description, level.DiscountPercentage,
		level.PriceColumn, level.ReferenceColumn, level.CanViewOffers,
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("código de nivel de precio ya existe: %s", level.Code)
		}
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("nivel de precio no encontrado: %d", id)
		}
		return nil, err
	}

	level.ID = id
	if createdAt.Valid {
		level.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		level.UpdatedAt = updatedAt.Time
	}

	return level, nil
}

func (r *priceLevelRepository) Delete(id int) error {
	query := "DELETE FROM niveles_precios WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("nivel de precio no encontrado: %d", id)
	}

	return nil
}

func (r *priceLevelRepository) GetUsersCount(id int) (int, error) {
	query := "SELECT COUNT(*) FROM usuarios WHERE nivel_precio_id = $1"
	var count int
	err := r.db.QueryRow(query, id).Scan(&count)
	return count, err
}

func (r *priceLevelRepository) TransferUsers(fromID, toID int) error {
	query := "UPDATE usuarios SET nivel_precio_id = $2 WHERE nivel_precio_id = $1"
	_, err := r.db.Exec(query, fromID, toID)
	return err
}
