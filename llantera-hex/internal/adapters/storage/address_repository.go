package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/llantera/hex/internal/domain/address"
)

type AddressRepository struct {
	db *sql.DB
}

func NewAddressRepository(db *sql.DB) *AddressRepository {
	return &AddressRepository{db: db}
}

func (r *AddressRepository) Create(ctx context.Context, userID string, req address.CreateRequest) (*address.Address, error) {
	// Si es default, quitar default de las demás
	if req.IsDefault {
		_, err := r.db.ExecContext(ctx,
			`UPDATE customer_addresses SET is_default = FALSE WHERE user_id = $1`,
			userID)
		if err != nil {
			return nil, err
		}
	}

	query := `
		INSERT INTO customer_addresses (
			user_id, alias, street, exterior_number, interior_number,
			neighborhood, postal_code, city, state, reference, phone, is_default
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, user_id, alias, street, exterior_number, interior_number,
			neighborhood, postal_code, city, state, reference, phone, is_default,
			created_at, updated_at
	`

	var addr address.Address
	var interiorNumber, reference sql.NullString

	err := r.db.QueryRowContext(ctx, query,
		userID, req.Alias, req.Street, req.ExteriorNumber, nullString(req.InteriorNumber),
		req.Neighborhood, req.PostalCode, req.City, req.State, nullString(req.Reference),
		req.Phone, req.IsDefault,
	).Scan(
		&addr.ID, &addr.UserID, &addr.Alias, &addr.Street, &addr.ExteriorNumber,
		&interiorNumber, &addr.Neighborhood, &addr.PostalCode, &addr.City, &addr.State,
		&reference, &addr.Phone, &addr.IsDefault, &addr.CreatedAt, &addr.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	addr.InteriorNumber = interiorNumber.String
	addr.Reference = reference.String

	return &addr, nil
}

func (r *AddressRepository) GetByID(ctx context.Context, id int) (*address.Address, error) {
	query := `
		SELECT id, user_id, alias, street, exterior_number, interior_number,
			neighborhood, postal_code, city, state, reference, phone, is_default,
			created_at, updated_at
		FROM customer_addresses
		WHERE id = $1
	`

	var addr address.Address
	var interiorNumber, reference sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&addr.ID, &addr.UserID, &addr.Alias, &addr.Street, &addr.ExteriorNumber,
		&interiorNumber, &addr.Neighborhood, &addr.PostalCode, &addr.City, &addr.State,
		&reference, &addr.Phone, &addr.IsDefault, &addr.CreatedAt, &addr.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, address.ErrNotFound
		}
		return nil, err
	}

	addr.InteriorNumber = interiorNumber.String
	addr.Reference = reference.String

	return &addr, nil
}

func (r *AddressRepository) ListByUserID(ctx context.Context, userID string) ([]address.Address, error) {
	query := `
		SELECT id, user_id, alias, street, exterior_number, interior_number,
			neighborhood, postal_code, city, state, reference, phone, is_default,
			created_at, updated_at
		FROM customer_addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []address.Address
	for rows.Next() {
		var addr address.Address
		var interiorNumber, reference sql.NullString

		err := rows.Scan(
			&addr.ID, &addr.UserID, &addr.Alias, &addr.Street, &addr.ExteriorNumber,
			&interiorNumber, &addr.Neighborhood, &addr.PostalCode, &addr.City, &addr.State,
			&reference, &addr.Phone, &addr.IsDefault, &addr.CreatedAt, &addr.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		addr.InteriorNumber = interiorNumber.String
		addr.Reference = reference.String
		addresses = append(addresses, addr)
	}

	return addresses, rows.Err()
}

func (r *AddressRepository) Update(ctx context.Context, id int, req address.UpdateRequest) (*address.Address, error) {
	// Obtener dirección actual
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Si se está estableciendo como default, quitar default de las demás
	if req.IsDefault != nil && *req.IsDefault {
		_, err := r.db.ExecContext(ctx,
			`UPDATE customer_addresses SET is_default = FALSE WHERE user_id = $1 AND id != $2`,
			current.UserID, id)
		if err != nil {
			return nil, err
		}
	}

	// Aplicar cambios
	if req.Alias != nil {
		current.Alias = *req.Alias
	}
	if req.Street != nil {
		current.Street = *req.Street
	}
	if req.ExteriorNumber != nil {
		current.ExteriorNumber = *req.ExteriorNumber
	}
	if req.InteriorNumber != nil {
		current.InteriorNumber = *req.InteriorNumber
	}
	if req.Neighborhood != nil {
		current.Neighborhood = *req.Neighborhood
	}
	if req.PostalCode != nil {
		current.PostalCode = *req.PostalCode
	}
	if req.City != nil {
		current.City = *req.City
	}
	if req.State != nil {
		current.State = *req.State
	}
	if req.Reference != nil {
		current.Reference = *req.Reference
	}
	if req.Phone != nil {
		current.Phone = *req.Phone
	}
	if req.IsDefault != nil {
		current.IsDefault = *req.IsDefault
	}

	query := `
		UPDATE customer_addresses SET
			alias = $2, street = $3, exterior_number = $4, interior_number = $5,
			neighborhood = $6, postal_code = $7, city = $8, state = $9,
			reference = $10, phone = $11, is_default = $12
		WHERE id = $1
		RETURNING updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		id, current.Alias, current.Street, current.ExteriorNumber,
		nullString(current.InteriorNumber), current.Neighborhood, current.PostalCode,
		current.City, current.State, nullString(current.Reference),
		current.Phone, current.IsDefault,
	).Scan(&current.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return current, nil
}

func (r *AddressRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM customer_addresses WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return address.ErrNotFound
	}

	return nil
}

func (r *AddressRepository) SetDefault(ctx context.Context, userID string, addressID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Quitar default de todas
	_, err = tx.ExecContext(ctx,
		`UPDATE customer_addresses SET is_default = FALSE WHERE user_id = $1`,
		userID)
	if err != nil {
		return err
	}

	// Establecer la nueva default
	result, err := tx.ExecContext(ctx,
		`UPDATE customer_addresses SET is_default = TRUE WHERE id = $1 AND user_id = $2`,
		addressID, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return address.ErrNotFound
	}

	return tx.Commit()
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
