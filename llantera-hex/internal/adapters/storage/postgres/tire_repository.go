package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/llantera/hex/internal/domain/tire"
)

type TireRepository struct {
	db *sql.DB
}

func (r *TireRepository) Delete(ctx context.Context, sku string) error {
	trimmed := strings.TrimSpace(sku)
	if trimmed == "" {
		return tire.ErrTireNotFound
	}

	const query = `
        DELETE FROM llantas
        WHERE LOWER(sku) = LOWER($1)
    `

	result, err := r.db.ExecContext(ctx, query, trimmed)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return tire.ErrTireNotFound
	}

	return nil
}

func NewTireRepository(db *sql.DB) *TireRepository {
	return &TireRepository{db: db}
}

var _ tire.TireRepository = (*TireRepository)(nil)

func (r *TireRepository) Create(ctx context.Context, entity *tire.Tire) error {
	const query = `
        INSERT INTO llantas (
            id, sku, marca_id, modelo, ancho, perfil, rin, construccion, tipo_tubo,
            calificacion_capas, indice_carga, indice_velocidad, tipo_normalizado_id,
            abreviatura_uso, descripcion, precio_publico, url_imagen, medida_original,
            creado_en, actualizado_en
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9,
            $10, $11, $12, $13, $14, $15, $16, $17, $18,
            $19, $20
        )
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		entity.ID,
		strings.TrimSpace(entity.SKU),
		entity.MarcaID,
		strings.TrimSpace(entity.Modelo),
		entity.Ancho,
		nullableIntPtr(entity.Perfil),
		entity.Rin,
		nullableStringValue(entity.Construccion),
		nullableStringValue(entity.TipoTubo),
		nullableStringValue(entity.CalificacionCapas),
		nullableStringValue(entity.IndiceCarga),
		nullableStringValue(entity.IndiceVelocidad),
		nullableIntPtr(entity.TipoNormalizadoID),
		nullableStringValue(entity.AbreviaturaUso),
		nullableStringValue(entity.Descripcion),
		entity.PrecioPublico,
		nullableStringValue(entity.URLImagen),
		nullableStringValue(entity.MedidaOriginal),
		entity.CreadoEn,
		entity.ActualizadoEn,
	)
	return err
}

func (r *TireRepository) Update(ctx context.Context, entity *tire.Tire) error {
	const query = `
        UPDATE llantas SET
            sku = $2,
            marca_id = $3,
            modelo = $4,
            ancho = $5,
            perfil = $6,
            rin = $7,
            construccion = $8,
            tipo_tubo = $9,
            calificacion_capas = $10,
            indice_carga = $11,
            indice_velocidad = $12,
            tipo_normalizado_id = $13,
            abreviatura_uso = $14,
            descripcion = $15,
            precio_publico = $16,
            url_imagen = $17,
            medida_original = $18,
            actualizado_en = $19
        WHERE id = $1
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		entity.ID,
		strings.TrimSpace(entity.SKU),
		entity.MarcaID,
		strings.TrimSpace(entity.Modelo),
		entity.Ancho,
		nullableIntPtr(entity.Perfil),
		entity.Rin,
		nullableStringValue(entity.Construccion),
		nullableStringValue(entity.TipoTubo),
		nullableStringValue(entity.CalificacionCapas),
		nullableStringValue(entity.IndiceCarga),
		nullableStringValue(entity.IndiceVelocidad),
		nullableIntPtr(entity.TipoNormalizadoID),
		nullableStringValue(entity.AbreviaturaUso),
		nullableStringValue(entity.Descripcion),
		entity.PrecioPublico,
		nullableStringValue(entity.URLImagen),
		nullableStringValue(entity.MedidaOriginal),
		entity.ActualizadoEn,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return tire.ErrTireNotFound
	}
	return nil
}

func (r *TireRepository) GetBySKU(ctx context.Context, sku string) (*tire.Tire, error) {
	const query = `
        SELECT id, sku, marca_id, modelo, ancho, perfil, rin, construccion, tipo_tubo,
               calificacion_capas, indice_carga, indice_velocidad, tipo_normalizado_id,
               abreviatura_uso, descripcion, precio_publico, url_imagen, medida_original,
               creado_en, actualizado_en
        FROM llantas
        WHERE LOWER(sku) = LOWER($1)
    `

	entity, err := scanTire(r.db.QueryRowContext(ctx, query, sku))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, tire.ErrTireNotFound
		}
		return nil, err
	}
	return entity, nil
}

func (r *TireRepository) List(ctx context.Context, filter tire.TireFilter) ([]tire.Tire, int, error) {
	where, args := buildTireFilters(filter)

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM llantas t %s", where)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortField, sortDir := tireSort(filter.Sort)
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 10000 {
		limit = 10000
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	limitPlaceholder := fmt.Sprintf("$%d", len(args)+1)
	offsetPlaceholder := fmt.Sprintf("$%d", len(args)+2)

	query := fmt.Sprintf(`
        SELECT id, sku, marca_id, modelo, ancho, perfil, rin, construccion, tipo_tubo,
               calificacion_capas, indice_carga, indice_velocidad, tipo_normalizado_id,
               abreviatura_uso, descripcion, precio_publico, url_imagen, medida_original,
               creado_en, actualizado_en
        FROM llantas t
        %s
        ORDER BY %s %s
        LIMIT %s OFFSET %s
    `, where, sortField, sortDir, limitPlaceholder, offsetPlaceholder)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []tire.Tire
	for rows.Next() {
		entity, err := scanTire(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *entity)
	}

	return items, total, nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanTire(scanner rowScanner) (*tire.Tire, error) {
	var (
		perfil          sql.NullInt64
		construccion    sql.NullString
		tipoTubo        sql.NullString
		calificacion    sql.NullString
		indiceCarga     sql.NullString
		indiceVelocidad sql.NullString
		tipoNormalizado sql.NullInt64
		abreviatura     sql.NullString
		descripcion     sql.NullString
		urlImagen       sql.NullString
		medidaOriginal  sql.NullString
		precio          sql.NullFloat64
	)

	entity := &tire.Tire{}
	if err := scanner.Scan(
		&entity.ID,
		&entity.SKU,
		&entity.MarcaID,
		&entity.Modelo,
		&entity.Ancho,
		&perfil,
		&entity.Rin,
		&construccion,
		&tipoTubo,
		&calificacion,
		&indiceCarga,
		&indiceVelocidad,
		&tipoNormalizado,
		&abreviatura,
		&descripcion,
		&precio,
		&urlImagen,
		&medidaOriginal,
		&entity.CreadoEn,
		&entity.ActualizadoEn,
	); err != nil {
		return nil, err
	}

	if perfil.Valid {
		v := int(perfil.Int64)
		entity.Perfil = &v
	}
	if tipoNormalizado.Valid {
		v := int(tipoNormalizado.Int64)
		entity.TipoNormalizadoID = &v
	}
	if construccion.Valid {
		entity.Construccion = construccion.String
	}
	if tipoTubo.Valid {
		entity.TipoTubo = tipoTubo.String
	}
	if calificacion.Valid {
		entity.CalificacionCapas = calificacion.String
	}
	if indiceCarga.Valid {
		entity.IndiceCarga = indiceCarga.String
	}
	if indiceVelocidad.Valid {
		entity.IndiceVelocidad = indiceVelocidad.String
	}
	if abreviatura.Valid {
		entity.AbreviaturaUso = abreviatura.String
	}
	if descripcion.Valid {
		entity.Descripcion = descripcion.String
	}
	if urlImagen.Valid {
		entity.URLImagen = urlImagen.String
	}
	if medidaOriginal.Valid {
		entity.MedidaOriginal = medidaOriginal.String
	}
	if precio.Valid {
		entity.PrecioPublico = precio.Float64
	}

	return entity, nil
}

func buildTireFilters(filter tire.TireFilter) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		clauses = append(clauses, fmt.Sprintf("(LOWER(t.sku) LIKE $%d OR LOWER(t.modelo) LIKE $%d OR LOWER(t.descripcion) LIKE $%d)", len(args)+1, len(args)+2, len(args)+3))
		args = append(args, like, like, like)
	}

	if filter.MarcaID != nil {
		clauses = append(clauses, fmt.Sprintf("t.marca_id = $%d", len(args)+1))
		args = append(args, *filter.MarcaID)
	}

	if filter.TipoID != nil {
		clauses = append(clauses, fmt.Sprintf("t.tipo_normalizado_id = $%d", len(args)+1))
		args = append(args, *filter.TipoID)
	}

	if abbr := strings.TrimSpace(filter.Abreviatura); abbr != "" {
		clauses = append(clauses, fmt.Sprintf("t.abreviatura_uso = $%d", len(args)+1))
		args = append(args, strings.ToUpper(abbr))
	}

	if filter.Ancho != nil {
		clauses = append(clauses, fmt.Sprintf("t.ancho = $%d", len(args)+1))
		args = append(args, *filter.Ancho)
	}

	if filter.Perfil != nil {
		clauses = append(clauses, fmt.Sprintf("t.perfil = $%d", len(args)+1))
		args = append(args, *filter.Perfil)
	}

	if filter.Rin != nil {
		clauses = append(clauses, fmt.Sprintf("t.rin = $%d", len(args)+1))
		args = append(args, *filter.Rin)
	}

	if v := strings.TrimSpace(filter.Construccion); v != "" {
		clauses = append(clauses, fmt.Sprintf("UPPER(TRIM(t.construccion)) = $%d", len(args)+1))
		args = append(args, strings.ToUpper(v))
	}

	if v := strings.TrimSpace(filter.CalificacionCapas); v != "" {
		clauses = append(clauses, fmt.Sprintf("UPPER(TRIM(t.calificacion_capas)) = $%d", len(args)+1))
		args = append(args, strings.ToUpper(v))
	}

	if v := strings.TrimSpace(filter.IndiceCarga); v != "" {
		clauses = append(clauses, fmt.Sprintf("UPPER(TRIM(t.indice_carga)) = $%d", len(args)+1))
		args = append(args, strings.ToUpper(v))
	}

	if v := strings.TrimSpace(filter.IndiceVelocidad); v != "" {
		clauses = append(clauses, fmt.Sprintf("UPPER(TRIM(t.indice_velocidad)) = $%d", len(args)+1))
		args = append(args, strings.ToUpper(v))
	}

	if filter.InStockOnly {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM inventario_llantas inv WHERE inv.llanta_id = t.id AND inv.cantidad > 0)")
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	return where, args
}

func tireSort(sort string) (string, string) {
	field := "t.creado_en"
	direction := "DESC"

	if sort == "" {
		return field, direction
	}

	dir := "ASC"
	key := strings.TrimSpace(sort)
	if strings.HasPrefix(key, "-") {
		dir = "DESC"
		key = strings.TrimPrefix(key, "-")
	}

	switch strings.ToLower(key) {
	case "sku":
		field = "t.sku"
	case "modelo":
		field = "t.modelo"
	case "precio":
		field = "t.precio_publico"
	case "creado", "created_at", "createdat":
		field = "t.creado_en"
	default:
		field = "t.creado_en"
	}

	return field, dir
}

func nullableIntPtr(value *int) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func nullableStringValue(value string) interface{} {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}
