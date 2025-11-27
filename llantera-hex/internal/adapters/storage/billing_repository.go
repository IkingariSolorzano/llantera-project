package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/llantera/hex/internal/domain/billing"
)

type BillingRepository struct {
	db *sql.DB
}

func NewBillingRepository(db *sql.DB) *BillingRepository {
	return &BillingRepository{db: db}
}

func (r *BillingRepository) Create(ctx context.Context, userID string, req billing.CreateRequest) (*billing.BillingInfo, error) {
	// Si es default, quitar default de los demás
	if req.IsDefault {
		_, err := r.db.ExecContext(ctx,
			`UPDATE customer_billing_info SET is_default = FALSE WHERE user_id = $1`,
			userID)
		if err != nil {
			return nil, err
		}
	}

	query := `
		INSERT INTO customer_billing_info (
			user_id, rfc, razon_social, regimen_fiscal, uso_cfdi, postal_code, email, is_default
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, rfc, razon_social, regimen_fiscal, uso_cfdi, postal_code, email, is_default, created_at, updated_at
	`

	var info billing.BillingInfo
	var email sql.NullString

	err := r.db.QueryRowContext(ctx, query,
		userID, req.RFC, req.RazonSocial, req.RegimenFiscal, req.UsoCFDI, req.PostalCode,
		nullString(req.Email), req.IsDefault,
	).Scan(
		&info.ID, &info.UserID, &info.RFC, &info.RazonSocial, &info.RegimenFiscal,
		&info.UsoCFDI, &info.PostalCode, &email, &info.IsDefault, &info.CreatedAt, &info.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	info.Email = email.String
	return &info, nil
}

func (r *BillingRepository) GetByID(ctx context.Context, id int) (*billing.BillingInfo, error) {
	query := `
		SELECT id, user_id, rfc, razon_social, regimen_fiscal, uso_cfdi, postal_code, email, is_default, created_at, updated_at
		FROM customer_billing_info
		WHERE id = $1
	`

	var info billing.BillingInfo
	var email sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&info.ID, &info.UserID, &info.RFC, &info.RazonSocial, &info.RegimenFiscal,
		&info.UsoCFDI, &info.PostalCode, &email, &info.IsDefault, &info.CreatedAt, &info.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}

	info.Email = email.String
	return &info, nil
}

func (r *BillingRepository) ListByUserID(ctx context.Context, userID string) ([]billing.BillingInfo, error) {
	query := `
		SELECT id, user_id, rfc, razon_social, regimen_fiscal, uso_cfdi, postal_code, email, is_default, created_at, updated_at
		FROM customer_billing_info
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var infos []billing.BillingInfo
	for rows.Next() {
		var info billing.BillingInfo
		var email sql.NullString

		err := rows.Scan(
			&info.ID, &info.UserID, &info.RFC, &info.RazonSocial, &info.RegimenFiscal,
			&info.UsoCFDI, &info.PostalCode, &email, &info.IsDefault, &info.CreatedAt, &info.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		info.Email = email.String
		infos = append(infos, info)
	}

	return infos, rows.Err()
}

func (r *BillingRepository) GetDefaultByUserID(ctx context.Context, userID string) (*billing.BillingInfo, error) {
	query := `
		SELECT id, user_id, rfc, razon_social, regimen_fiscal, uso_cfdi, postal_code, email, is_default, created_at, updated_at
		FROM customer_billing_info
		WHERE user_id = $1 AND is_default = TRUE
		LIMIT 1
	`

	var info billing.BillingInfo
	var email sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&info.ID, &info.UserID, &info.RFC, &info.RazonSocial, &info.RegimenFiscal,
		&info.UsoCFDI, &info.PostalCode, &email, &info.IsDefault, &info.CreatedAt, &info.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}

	info.Email = email.String
	return &info, nil
}

func (r *BillingRepository) Update(ctx context.Context, id int, req billing.UpdateRequest) (*billing.BillingInfo, error) {
	// Obtener datos actuales
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Si se está estableciendo como default, quitar default de los demás
	if req.IsDefault != nil && *req.IsDefault {
		_, err := r.db.ExecContext(ctx,
			`UPDATE customer_billing_info SET is_default = FALSE WHERE user_id = $1 AND id != $2`,
			current.UserID, id)
		if err != nil {
			return nil, err
		}
	}

	// Aplicar cambios
	if req.RFC != nil {
		current.RFC = *req.RFC
	}
	if req.RazonSocial != nil {
		current.RazonSocial = *req.RazonSocial
	}
	if req.RegimenFiscal != nil {
		current.RegimenFiscal = *req.RegimenFiscal
	}
	if req.UsoCFDI != nil {
		current.UsoCFDI = *req.UsoCFDI
	}
	if req.PostalCode != nil {
		current.PostalCode = *req.PostalCode
	}
	if req.Email != nil {
		current.Email = *req.Email
	}
	if req.IsDefault != nil {
		current.IsDefault = *req.IsDefault
	}

	query := `
		UPDATE customer_billing_info SET
			rfc = $2, razon_social = $3, regimen_fiscal = $4, uso_cfdi = $5,
			postal_code = $6, email = $7, is_default = $8
		WHERE id = $1
		RETURNING updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		id, current.RFC, current.RazonSocial, current.RegimenFiscal, current.UsoCFDI,
		current.PostalCode, nullString(current.Email), current.IsDefault,
	).Scan(&current.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return current, nil
}

func (r *BillingRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM customer_billing_info WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return billing.ErrNotFound
	}

	return nil
}

func (r *BillingRepository) SetDefault(ctx context.Context, userID string, billingID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Quitar default de todos
	_, err = tx.ExecContext(ctx,
		`UPDATE customer_billing_info SET is_default = FALSE WHERE user_id = $1`,
		userID)
	if err != nil {
		return err
	}

	// Establecer el nuevo default
	result, err := tx.ExecContext(ctx,
		`UPDATE customer_billing_info SET is_default = TRUE WHERE id = $1 AND user_id = $2`,
		billingID, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return billing.ErrNotFound
	}

	return tx.Commit()
}

// Ensure BillingRepository implements billing.Repository
var _ billing.Repository = (*BillingRepository)(nil)
