package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/llantera/hex/internal/domain/user"
)

const (
	userFullNameExpr      = "LOWER(CONCAT_WS(' ', u.primer_nombre, u.primer_apellido, u.segundo_apellido))"
	userFullNameOrderExpr = "CONCAT_WS(' ', u.primer_nombre, u.primer_apellido, u.segundo_apellido)"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

var _ user.Repository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	const query = `
        INSERT INTO usuarios (
            id, correo, primer_nombre, primer_apellido, segundo_apellido,
            telefono, domicilio_calle, domicilio_numero, domicilio_colonia, domicilio_codigo_postal,
            puesto, activo, empresa_id, url_imagen_perfil, hash_contrasena,
            rol, nivel_precio_id, creado_en, actualizado_en
        ) VALUES (
            $1, $2, $3, $4, $5,
            $6, $7, $8, $9, $10,
            $11, $12, $13, $14, $15,
            $16, $17, $18, $19
        )
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		u.ID,
		u.Email,
		u.FirstName,
		u.FirstLastName,
		u.SecondLastName,
		u.Phone,
		u.AddressStreet,
		u.AddressNumber,
		u.AddressNeighborhood,
		u.AddressPostalCode,
		u.JobTitle,
		u.Active,
		nullableInt(u.CompanyID),
		nullString(u.ProfileImageURL),
		u.PasswordHash,
		string(u.Role),
		nullableInt(u.PriceLevelID),
		u.CreatedAt,
		u.UpdatedAt,
	)
	if err != nil {
		log.Printf("[UserRepository] error creando usuario en DB (correo=%s): %v", u.Email, err)
	}
	return err
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	const query = `
        UPDATE usuarios SET
            correo = $2,
            primer_nombre = $3,
            primer_apellido = $4,
            segundo_apellido = $5,
            telefono = $6,
            domicilio_calle = $7,
            domicilio_numero = $8,
            domicilio_colonia = $9,
            domicilio_codigo_postal = $10,
            puesto = $11,
            activo = $12,
            empresa_id = $13,
            url_imagen_perfil = $14,
            rol = $15,
            nivel_precio_id = $16,
            actualizado_en = $17
        WHERE id = $1
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		u.ID,
		u.Email,
		u.FirstName,
		u.FirstLastName,
		u.SecondLastName,
		u.Phone,
		u.AddressStreet,
		u.AddressNumber,
		u.AddressNeighborhood,
		u.AddressPostalCode,
		u.JobTitle,
		u.Active,
		nullableInt(u.CompanyID),
		nullString(u.ProfileImageURL),
		string(u.Role),
		nullableInt(u.PriceLevelID),
		u.UpdatedAt,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM usuarios WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	const query = `
        SELECT u.id, u.correo, u.primer_nombre, u.primer_apellido, u.segundo_apellido,
               u.telefono,
               u.domicilio_calle, u.domicilio_numero, u.domicilio_colonia, u.domicilio_codigo_postal,
               u.puesto,
               u.activo, u.empresa_id, u.url_imagen_perfil, u.hash_contrasena, u.rol,
               COALESCE(np.codigo, 'public') AS nivel_codigo,
               u.nivel_precio_id, u.creado_en, u.actualizado_en
        FROM usuarios u
        LEFT JOIN niveles_precios np ON u.nivel_precio_id = np.id
        WHERE u.id = $1
    `
	return r.fetchOne(ctx, query, id)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	const query = `
        SELECT u.id, u.correo, u.primer_nombre, u.primer_apellido, u.segundo_apellido,
               u.telefono,
               u.domicilio_calle, u.domicilio_numero, u.domicilio_colonia, u.domicilio_codigo_postal,
               u.puesto,
               u.activo, u.empresa_id, u.url_imagen_perfil, u.hash_contrasena, u.rol,
               COALESCE(np.codigo, 'public') AS nivel_codigo,
               u.nivel_precio_id, u.creado_en, u.actualizado_en
        FROM usuarios u
        LEFT JOIN niveles_precios np ON u.nivel_precio_id = np.id
        WHERE LOWER(u.correo) = LOWER($1)
    `
	return r.fetchOne(ctx, query, email)
}

func (r *UserRepository) fetchOne(ctx context.Context, query string, arg interface{}) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, query, arg)

	var (
		companyID sql.NullInt64
		profile   sql.NullString
		priceID   sql.NullInt64
		level     string
	)

	u := &user.User{}
	if err := row.Scan(
		&u.ID,
		&u.Email,
		&u.FirstName,
		&u.FirstLastName,
		&u.SecondLastName,
		&u.Phone,
		&u.AddressStreet,
		&u.AddressNumber,
		&u.AddressNeighborhood,
		&u.AddressPostalCode,
		&u.JobTitle,
		&u.Active,
		&companyID,
		&profile,
		&u.PasswordHash,
		&u.Role,
		&level,
		&priceID,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}

	if companyID.Valid {
		v := int(companyID.Int64)
		u.CompanyID = &v
	}
	if profile.Valid {
		u.ProfileImageURL = profile.String
	}
	if priceID.Valid {
		v := int(priceID.Int64)
		u.PriceLevelID = &v
	}

	u.Level = user.PriceLevel(level)
	u.Name = u.FullName()
	if strings.TrimSpace(u.Name) == "" {
		u.Name = u.Email
	}

	return u, nil
}

func (r *UserRepository) List(ctx context.Context, filter user.ListFilter) ([]user.User, int, error) {
	where, args := buildUserFilters(filter)

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM usuarios u %s", where)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortField, sortDir := userSort(filter.Sort)
	limitPlaceholder := fmt.Sprintf("$%d", len(args)+1)
	offsetPlaceholder := fmt.Sprintf("$%d", len(args)+2)

	listQuery := fmt.Sprintf(`
        SELECT u.id, u.correo, u.primer_nombre, u.primer_apellido, u.segundo_apellido,
               u.telefono,
               u.domicilio_calle, u.domicilio_numero, u.domicilio_colonia, u.domicilio_codigo_postal,
               u.puesto,
               u.activo, u.empresa_id, u.url_imagen_perfil, u.rol,
               COALESCE(np.codigo, 'public') AS nivel_codigo,
               u.nivel_precio_id, u.creado_en, u.actualizado_en
        FROM usuarios u
        LEFT JOIN niveles_precios np ON u.nivel_precio_id = np.id
        %s
        ORDER BY %s %s
        LIMIT %s OFFSET %s
    `, where, sortField, sortDir, limitPlaceholder, offsetPlaceholder)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []user.User
	for rows.Next() {
		var (
			companyID sql.NullInt64
			profile   sql.NullString
			priceID   sql.NullInt64
			item      user.User
			level     string
		)

		if err := rows.Scan(
			&item.ID,
			&item.Email,
			&item.FirstName,
			&item.FirstLastName,
			&item.SecondLastName,
			&item.Phone,
			&item.AddressStreet,
			&item.AddressNumber,
			&item.AddressNeighborhood,
			&item.AddressPostalCode,
			&item.JobTitle,
			&item.Active,
			&companyID,
			&profile,
			&item.Role,
			&level,
			&priceID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if companyID.Valid {
			v := int(companyID.Int64)
			item.CompanyID = &v
		}
		if profile.Valid {
			item.ProfileImageURL = profile.String
		}
		if priceID.Valid {
			v := int(priceID.Int64)
			item.PriceLevelID = &v
		}

		item.Level = user.PriceLevel(level)
		item.Name = item.FullName()
		if strings.TrimSpace(item.Name) == "" {
			item.Name = item.Email
		}

		result = append(result, item)
	}

	return result, total, nil
}

func buildUserFilters(filter user.ListFilter) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		clauses = append(clauses, fmt.Sprintf("(LOWER(u.correo) LIKE $%d OR %s LIKE $%d)", len(args)+1, userFullNameExpr, len(args)+2))
		args = append(args, like, like)
	}

	if filter.CompanyID != nil {
		clauses = append(clauses, fmt.Sprintf("u.empresa_id = $%d", len(args)+1))
		args = append(args, *filter.CompanyID)
	}

	if filter.Role != nil {
		clauses = append(clauses, fmt.Sprintf("u.rol = $%d", len(args)+1))
		args = append(args, string(*filter.Role))
	}

	if filter.Active != nil {
		clauses = append(clauses, fmt.Sprintf("u.activo = $%d", len(args)+1))
		args = append(args, *filter.Active)
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	return where, args
}

func userSort(sort string) (string, string) {
	field := "u.creado_en"
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
	case "nombre", "name":
		field = userFullNameOrderExpr
	case "correo", "email":
		field = "u.correo"
	case "creado", "created_at", "createdat":
		field = "u.creado_en"
	default:
		field = "u.creado_en"
	}

	return field, dir
}

func nullableInt(value *int) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func nullString(value string) interface{} {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
