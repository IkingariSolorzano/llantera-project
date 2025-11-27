package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/llantera/hex/internal/domain/company"
)

type CompanyRepository struct {
	db *sql.DB
}

func NewCompanyRepository(db *sql.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

var _ company.Repository = (*CompanyRepository)(nil)

func (r *CompanyRepository) Create(ctx context.Context, c *company.Company) error {
	const query = `
        INSERT INTO empresas (
            clave, razon_social, rfc, direccion, correos, telefonos, contacto_principal_id,
            creado_en, actualizado_en
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id
    `

	emails := pq.StringArray(normalizeSlice(c.Emails))
	phones := pq.StringArray(normalizeSlice(c.Phones))

	return r.db.QueryRowContext(
		ctx,
		query,
		strings.TrimSpace(c.KeyName),
		strings.TrimSpace(c.SocialReason),
		strings.TrimSpace(c.RFC),
		strings.TrimSpace(c.Address),
		emails,
		phones,
		nullableStringPtr(c.MainContactID),
		c.CreatedAt,
		c.UpdatedAt,
	).Scan(&c.ID)
}

func (r *CompanyRepository) Update(ctx context.Context, c *company.Company) error {
	const query = `
        UPDATE empresas SET
            clave = $2,
            razon_social = $3,
            rfc = $4,
            direccion = $5,
            correos = $6,
            telefonos = $7,
            contacto_principal_id = $8,
            actualizado_en = $9
        WHERE id = $1
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		c.ID,
		strings.TrimSpace(c.KeyName),
		strings.TrimSpace(c.SocialReason),
		strings.TrimSpace(c.RFC),
		strings.TrimSpace(c.Address),
		pq.StringArray(normalizeSlice(c.Emails)),
		pq.StringArray(normalizeSlice(c.Phones)),
		nullableStringPtr(c.MainContactID),
		c.UpdatedAt,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return company.ErrNotFound
	}
	return nil
}

func (r *CompanyRepository) Delete(ctx context.Context, id int) error {
	const query = `DELETE FROM empresas WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return company.ErrNotFound
	}
	return nil
}

func (r *CompanyRepository) GetByID(ctx context.Context, id int) (*company.Company, error) {
	const query = `
        SELECT id, clave, razon_social, rfc, direccion, correos, telefonos,
               contacto_principal_id, creado_en, actualizado_en
        FROM empresas
        WHERE id = $1
    `

	var (
		emails      pq.StringArray
		phones      pq.StringArray
		mainContact sql.NullString
	)

	c := &company.Company{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.KeyName,
		&c.SocialReason,
		&c.RFC,
		&c.Address,
		&emails,
		&phones,
		&mainContact,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, company.ErrNotFound
		}
		return nil, err
	}

	c.Emails = []string(emails)
	c.Phones = []string(phones)
	if mainContact.Valid {
		value := mainContact.String
		c.MainContactID = &value
	}

	return c, nil
}

func (r *CompanyRepository) List(ctx context.Context, filter company.ListFilter) ([]company.Company, int, error) {
	where, args := buildCompanyFilters(filter)

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM empresas c %s", where)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortField, sortDir := companySort(filter.Sort)
	limitPlaceholder := fmt.Sprintf("$%d", len(args)+1)
	offsetPlaceholder := fmt.Sprintf("$%d", len(args)+2)

	query := fmt.Sprintf(`
        SELECT id, clave, razon_social, rfc, direccion, correos, telefonos,
               contacto_principal_id, creado_en, actualizado_en
        FROM empresas c
        %s
        ORDER BY %s %s
        LIMIT %s OFFSET %s
    `, where, sortField, sortDir, limitPlaceholder, offsetPlaceholder)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []company.Company
	for rows.Next() {
		var (
			emails      pq.StringArray
			phones      pq.StringArray
			mainContact sql.NullString
			item        company.Company
		)

		if err := rows.Scan(
			&item.ID,
			&item.KeyName,
			&item.SocialReason,
			&item.RFC,
			&item.Address,
			&emails,
			&phones,
			&mainContact,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		item.Emails = []string(emails)
		item.Phones = []string(phones)
		if mainContact.Valid {
			value := mainContact.String
			item.MainContactID = &value
		}

		results = append(results, item)
	}

	return results, total, nil
}

func buildCompanyFilters(filter company.ListFilter) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		clauses = append(clauses, fmt.Sprintf("(LOWER(c.clave) LIKE $%d OR LOWER(c.razon_social) LIKE $%d)", len(args)+1, len(args)+2))
		args = append(args, like, like)
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	return where, args
}

func companySort(sort string) (string, string) {
	field := "c.creado_en"
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
	case "nombre", "key_name":
		field = "c.clave"
	case "razon", "social_reason":
		field = "c.razon_social"
	case "creado", "created_at":
		field = "c.creado_en"
	default:
		field = "c.creado_en"
	}

	return field, dir
}

func normalizeSlice(values []string) []string {
	var out []string
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func nullableStringPtr(value *string) interface{} {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
