package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/llantera/hex/internal/domain/customerrequest"
)

// CustomerRequestRepository implementa customerrequest.Repository usando PostgreSQL.
type CustomerRequestRepository struct {
	db *sql.DB
}

func NewCustomerRequestRepository(db *sql.DB) *CustomerRequestRepository {
	return &CustomerRequestRepository{db: db}
}

var _ customerrequest.Repository = (*CustomerRequestRepository)(nil)

func (r *CustomerRequestRepository) Create(ctx context.Context, cr *customerrequest.CustomerRequest) error {
	const query = `
        INSERT INTO solicitudes_clientes (
            id, nombre_completo, tipo_solicitud, mensaje, telefono,
            preferencia_contacto, correo, estado, empleado_id, acuerdo,
            creado_en, actualizado_en, atendido_en
        ) VALUES (
            $1, $2, $3, $4, $5,
            $6, $7, $8, $9, $10,
            $11, $12, $13
        )
    `

	var empleadoID interface{}
	if cr.EmployeeID != nil && strings.TrimSpace(*cr.EmployeeID) != "" {
		empleadoID = strings.TrimSpace(*cr.EmployeeID)
	} else {
		empleadoID = nil
	}

	var atendidoEn interface{}
	if cr.AttendedAt != nil {
		atendidoEn = *cr.AttendedAt
	} else {
		atendidoEn = nil
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		cr.ID,
		cr.FullName,
		cr.RequestType,
		cr.Message,
		cr.Phone,
		cr.ContactPreference,
		cr.Email,
		string(cr.Status),
		empleadoID,
		cr.Agreement,
		cr.CreatedAt,
		cr.UpdatedAt,
		atendidoEn,
	)
	return err
}

func (r *CustomerRequestRepository) Update(ctx context.Context, cr *customerrequest.CustomerRequest) error {
	const query = `
        UPDATE solicitudes_clientes SET
            mensaje = $2,
            telefono = $3,
            preferencia_contacto = $4,
            correo = $5,
            estado = $6,
            empleado_id = $7,
            acuerdo = $8,
            actualizado_en = $9,
            atendido_en = $10
        WHERE id = $1
    `

	var empleadoID interface{}
	if cr.EmployeeID != nil && strings.TrimSpace(*cr.EmployeeID) != "" {
		empleadoID = strings.TrimSpace(*cr.EmployeeID)
	} else {
		empleadoID = nil
	}

	var atendidoEn interface{}
	if cr.AttendedAt != nil {
		atendidoEn = *cr.AttendedAt
	} else {
		atendidoEn = nil
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		cr.ID,
		cr.Message,
		cr.Phone,
		cr.ContactPreference,
		cr.Email,
		string(cr.Status),
		empleadoID,
		cr.Agreement,
		cr.UpdatedAt,
		atendidoEn,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return customerrequest.ErrNotFound
	}
	return nil
}

func (r *CustomerRequestRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM solicitudes_clientes WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return customerrequest.ErrNotFound
	}
	return nil
}

func (r *CustomerRequestRepository) GetByID(ctx context.Context, id string) (*customerrequest.CustomerRequest, error) {
	const query = `
        SELECT id, nombre_completo, tipo_solicitud, mensaje, telefono,
               preferencia_contacto, correo, estado, empleado_id, acuerdo,
               creado_en, actualizado_en, atendido_en
        FROM solicitudes_clientes
        WHERE id = $1
    `

	var (
		empleadoID sql.NullString
		atendidoEn sql.NullTime
		statusStr  string
	)

	cr := &customerrequest.CustomerRequest{}
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cr.ID,
		&cr.FullName,
		&cr.RequestType,
		&cr.Message,
		&cr.Phone,
		&cr.ContactPreference,
		&cr.Email,
		&statusStr,
		&empleadoID,
		&cr.Agreement,
		&cr.CreatedAt,
		&cr.UpdatedAt,
		&atendidoEn,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customerrequest.ErrNotFound
		}
		return nil, err
	}

	if empleadoID.Valid {
		v := empleadoID.String
		cr.EmployeeID = &v
	}
	if atendidoEn.Valid {
		v := atendidoEn.Time
		cr.AttendedAt = &v
	}
	cr.Status = customerrequest.Status(statusStr)

	return cr, nil
}

func (r *CustomerRequestRepository) List(ctx context.Context, filter customerrequest.ListFilter) ([]customerrequest.CustomerRequest, int, error) {
	where, args := buildCustomerRequestFilters(filter)

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM solicitudes_clientes %s", where)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortField, sortDir := customerRequestSort(filter.Sort)
	limitPlaceholder := fmt.Sprintf("$%d", len(args)+1)
	offsetPlaceholder := fmt.Sprintf("$%d", len(args)+2)

	listQuery := fmt.Sprintf(`
        SELECT id, nombre_completo, tipo_solicitud, mensaje, telefono,
               preferencia_contacto, correo, estado, empleado_id, acuerdo,
               creado_en, actualizado_en, atendido_en
        FROM solicitudes_clientes
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

	var results []customerrequest.CustomerRequest
	for rows.Next() {
		var (
			empleadoID sql.NullString
			atendidoEn sql.NullTime
			statusStr  string
			item       customerrequest.CustomerRequest
		)

		if err := rows.Scan(
			&item.ID,
			&item.FullName,
			&item.RequestType,
			&item.Message,
			&item.Phone,
			&item.ContactPreference,
			&item.Email,
			&statusStr,
			&empleadoID,
			&item.Agreement,
			&item.CreatedAt,
			&item.UpdatedAt,
			&atendidoEn,
		); err != nil {
			return nil, 0, err
		}

		if empleadoID.Valid {
			v := empleadoID.String
			item.EmployeeID = &v
		}
		if atendidoEn.Valid {
			v := atendidoEn.Time
			item.AttendedAt = &v
		}
		item.Status = customerrequest.Status(statusStr)

		results = append(results, item)
	}

	return results, total, nil
}

func buildCustomerRequestFilters(filter customerrequest.ListFilter) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		clauses = append(clauses, fmt.Sprintf("(LOWER(nombre_completo) LIKE $%d OR LOWER(correo) LIKE $%d OR LOWER(telefono) LIKE $%d)", len(args)+1, len(args)+2, len(args)+3))
		args = append(args, like, like, like)
	}

	if filter.Status != nil {
		clauses = append(clauses, fmt.Sprintf("estado = $%d", len(args)+1))
		args = append(args, string(*filter.Status))
	}

	if filter.EmployeeID != nil {
		clauses = append(clauses, fmt.Sprintf("empleado_id = $%d", len(args)+1))
		args = append(args, *filter.EmployeeID)
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	return where, args
}

func customerRequestSort(sort string) (string, string) {
	field := "creado_en"
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
	case "estado", "status":
		field = "estado"
	case "creado", "created_at", "createdat":
		field = "creado_en"
	default:
		field = "creado_en"
	}

	return field, dir
}
