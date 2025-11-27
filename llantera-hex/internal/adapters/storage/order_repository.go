package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/llantera/hex/internal/domain/order"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *order.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insertar pedido (order_number se genera automáticamente por trigger)
	query := `
		INSERT INTO orders (
			user_id, status, shipping_address_id,
			shipping_street, shipping_exterior_number, shipping_interior_number,
			shipping_neighborhood, shipping_postal_code, shipping_city, shipping_state,
			shipping_reference, shipping_phone, payment_method,
			payment_mode, payment_installments, payment_notes,
			requires_invoice, billing_info_id, billing_rfc, billing_razon_social,
			billing_regimen_fiscal, billing_uso_cfdi, billing_postal_code, billing_email,
			subtotal, iva, shipping_cost, total, customer_notes, order_number
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, ''
		)
		RETURNING id, order_number, created_at, updated_at
	`

	var billingInfoID sql.NullInt64
	var billingRFC, billingRazonSocial, billingRegimen, billingUsoCFDI, billingPostalCode, billingEmail sql.NullString

	if o.BillingInfo != nil {
		if o.BillingInfo.ID > 0 {
			billingInfoID = sql.NullInt64{Int64: int64(o.BillingInfo.ID), Valid: true}
		}
		billingRFC = nullString(o.BillingInfo.RFC)
		billingRazonSocial = nullString(o.BillingInfo.RazonSocial)
		billingRegimen = nullString(o.BillingInfo.RegimenFiscal)
		billingUsoCFDI = nullString(o.BillingInfo.UsoCFDI)
		billingPostalCode = nullString(o.BillingInfo.PostalCode)
		billingEmail = nullString(o.BillingInfo.Email)
	}

	var shippingAddressID sql.NullInt64
	if o.ShippingAddress.ID > 0 {
		shippingAddressID = sql.NullInt64{Int64: int64(o.ShippingAddress.ID), Valid: true}
	}

	// Establecer valor por defecto para PaymentMode si está vacío
	paymentMode := o.PaymentMode
	if paymentMode == "" {
		paymentMode = order.PaymentModeContado
	}

	err = tx.QueryRowContext(ctx, query,
		o.UserID, o.Status, shippingAddressID,
		o.ShippingAddress.Street, o.ShippingAddress.ExteriorNumber,
		nullString(o.ShippingAddress.InteriorNumber), o.ShippingAddress.Neighborhood,
		o.ShippingAddress.PostalCode, o.ShippingAddress.City, o.ShippingAddress.State,
		nullString(o.ShippingAddress.Reference), o.ShippingAddress.Phone, o.PaymentMethod,
		paymentMode, o.PaymentInstallments, nullString(o.PaymentNotes),
		o.RequiresInvoice, billingInfoID, billingRFC, billingRazonSocial,
		billingRegimen, billingUsoCFDI, billingPostalCode, billingEmail,
		o.Subtotal, o.IVA, o.ShippingCost, o.Total, nullString(o.CustomerNotes),
	).Scan(&o.ID, &o.OrderNumber, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return err
	}

	// Insertar items
	for i := range o.Items {
		item := &o.Items[i]
		itemQuery := `
			INSERT INTO order_items (
				order_id, tire_sku, tire_measure, tire_brand, tire_model,
				quantity, unit_price, subtotal
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, created_at
		`
		err = tx.QueryRowContext(ctx, itemQuery,
			o.ID, item.TireSKU, item.TireMeasure,
			nullString(item.TireBrand), nullString(item.TireModel),
			item.Quantity, item.UnitPrice, item.Subtotal,
		).Scan(&item.ID, &item.CreatedAt)
		if err != nil {
			return err
		}
		item.OrderID = o.ID
	}

	return tx.Commit()
}

func (r *OrderRepository) GetByID(ctx context.Context, id int) (*order.Order, error) {
	o, err := r.scanOrder(ctx, `WHERE o.id = $1`, id)
	if err != nil {
		return nil, err
	}

	// Cargar items
	items, err := r.getOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return o, nil
}

func (r *OrderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*order.Order, error) {
	o, err := r.scanOrder(ctx, `WHERE o.order_number = $1`, orderNumber)
	if err != nil {
		return nil, err
	}

	items, err := r.getOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return o, nil
}

func (r *OrderRepository) List(ctx context.Context, filter order.ListFilter) ([]order.Order, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("o.user_id = $%d", argIndex))
		args = append(args, filter.UserID)
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("o.status = $%d::order_status", argIndex))
		args = append(args, string(*filter.Status))
		argIndex++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(o.order_number ILIKE $%d OR u.primer_nombre ILIKE $%d OR u.primer_apellido ILIKE $%d OR u.correo ILIKE $%d)",
			argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Contar total
	countQuery := `SELECT COUNT(*) FROM orders o LEFT JOIN usuarios u ON o.user_id = u.id ` + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		log.Printf("Error en countQuery: %v, query: %s, args: %v", err, countQuery, args)
		return nil, 0, err
	}

	// Ordenamiento
	orderBy := "o.created_at DESC"
	if filter.Sort != "" {
		orderBy = filter.Sort
	}

	// Paginación
	limit := 20
	if filter.Limit > 0 {
		limit = filter.Limit
	}

	query := fmt.Sprintf(`
		SELECT o.id, o.order_number, o.user_id, o.status,
			o.shipping_street, o.shipping_exterior_number, o.shipping_interior_number,
			o.shipping_neighborhood, o.shipping_postal_code, o.shipping_city, o.shipping_state,
			o.shipping_reference, o.shipping_phone, o.payment_method,
			COALESCE(o.payment_mode, 'contado') as payment_mode,
			COALESCE(o.payment_installments, 1) as payment_installments,
			o.payment_notes,
			o.requires_invoice, o.billing_rfc, o.billing_razon_social,
			o.billing_regimen_fiscal, o.billing_uso_cfdi, o.billing_postal_code, o.billing_email,
			o.subtotal, COALESCE(o.iva, 0) as iva, o.shipping_cost, o.total,
			o.invoice_xml_path, o.invoice_pdf_path, o.customer_notes, o.admin_notes,
			o.created_at, o.updated_at, o.shipped_at, o.delivered_at, o.cancelled_at,
			COALESCE(u.primer_nombre || ' ' || u.primer_apellido, u.correo) as user_name,
			u.correo, COALESCE(u.telefono, '') as user_phone
		FROM orders o
		LEFT JOIN usuarios u ON o.user_id = u.id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error en query principal de orders: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var orders []order.Order
	for rows.Next() {
		var o order.Order
		var interiorNumber, reference, billingRFC, billingRazonSocial sql.NullString
		var billingRegimen, billingUsoCFDI, billingPostalCode, billingEmail sql.NullString
		var invoiceXML, invoicePDF, customerNotes, adminNotes, paymentNotes sql.NullString
		var shippedAt, deliveredAt, cancelledAt sql.NullTime
		var userName, userEmail, userPhone string

		err := rows.Scan(
			&o.ID, &o.OrderNumber, &o.UserID, &o.Status,
			&o.ShippingAddress.Street, &o.ShippingAddress.ExteriorNumber, &interiorNumber,
			&o.ShippingAddress.Neighborhood, &o.ShippingAddress.PostalCode,
			&o.ShippingAddress.City, &o.ShippingAddress.State, &reference, &o.ShippingAddress.Phone,
			&o.PaymentMethod, &o.PaymentMode, &o.PaymentInstallments, &paymentNotes,
			&o.RequiresInvoice,
			&billingRFC, &billingRazonSocial, &billingRegimen, &billingUsoCFDI,
			&billingPostalCode, &billingEmail,
			&o.Subtotal, &o.IVA, &o.ShippingCost, &o.Total,
			&invoiceXML, &invoicePDF, &customerNotes, &adminNotes,
			&o.CreatedAt, &o.UpdatedAt, &shippedAt, &deliveredAt, &cancelledAt,
			&userName, &userEmail, &userPhone,
		)
		if err != nil {
			log.Printf("Error escaneando order: %v", err)
			return nil, 0, err
		}

		o.ShippingAddress.InteriorNumber = interiorNumber.String
		o.ShippingAddress.Reference = reference.String
		o.InvoiceXMLPath = invoiceXML.String
		o.InvoicePDFPath = invoicePDF.String
		o.CustomerNotes = customerNotes.String
		o.AdminNotes = adminNotes.String
		o.PaymentNotes = paymentNotes.String

		if shippedAt.Valid {
			o.ShippedAt = &shippedAt.Time
		}
		if deliveredAt.Valid {
			o.DeliveredAt = &deliveredAt.Time
		}
		if cancelledAt.Valid {
			o.CancelledAt = &cancelledAt.Time
		}

		if billingRFC.Valid {
			o.BillingInfo = &order.BillingInfo{
				RFC:           billingRFC.String,
				RazonSocial:   billingRazonSocial.String,
				RegimenFiscal: billingRegimen.String,
				UsoCFDI:       billingUsoCFDI.String,
				PostalCode:    billingPostalCode.String,
				Email:         billingEmail.String,
			}
		}

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Cargar items para cada orden
	for i := range orders {
		items, err := r.getOrderItems(ctx, orders[i].ID)
		if err != nil {
			log.Printf("Error cargando items para orden %d: %v", orders[i].ID, err)
			continue
		}
		orders[i].Items = items
	}

	return orders, total, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id int, status order.Status, adminNotes string) error {
	var timestampField string
	switch status {
	case order.StatusEnviado:
		timestampField = ", shipped_at = NOW()"
	case order.StatusEntregado:
		timestampField = ", delivered_at = NOW()"
	case order.StatusCancelado:
		timestampField = ", cancelled_at = NOW()"
	}

	query := fmt.Sprintf(`
		UPDATE orders SET status = $1, admin_notes = COALESCE($2, admin_notes)%s
		WHERE id = $3
	`, timestampField)

	var notes sql.NullString
	if adminNotes != "" {
		notes = sql.NullString{String: adminNotes, Valid: true}
	}

	result, err := r.db.ExecContext(ctx, query, status, notes, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return order.ErrNotFound
	}

	return nil
}

func (r *OrderRepository) UpdateInvoiceFiles(ctx context.Context, id int, xmlPath, pdfPath string) error {
	query := `
		UPDATE orders SET
			invoice_xml_path = COALESCE($1, invoice_xml_path),
			invoice_pdf_path = COALESCE($2, invoice_pdf_path)
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, nullString(xmlPath), nullString(pdfPath), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return order.ErrNotFound
	}

	return nil
}

func (r *OrderRepository) scanOrder(ctx context.Context, whereClause string, args ...interface{}) (*order.Order, error) {
	query := fmt.Sprintf(`
		SELECT o.id, o.order_number, o.user_id, o.status,
			o.shipping_address_id, o.shipping_street, o.shipping_exterior_number, o.shipping_interior_number,
			o.shipping_neighborhood, o.shipping_postal_code, o.shipping_city, o.shipping_state,
			o.shipping_reference, o.shipping_phone, o.payment_method,
			COALESCE(o.payment_mode, 'contado') as payment_mode,
			COALESCE(o.payment_installments, 1) as payment_installments,
			o.payment_notes,
			o.requires_invoice, o.billing_info_id, o.billing_rfc, o.billing_razon_social,
			o.billing_regimen_fiscal, o.billing_uso_cfdi, o.billing_postal_code, o.billing_email,
			o.subtotal, COALESCE(o.iva, 0) as iva, o.shipping_cost, o.total,
			o.invoice_xml_path, o.invoice_pdf_path, o.customer_notes, o.admin_notes,
			o.created_at, o.updated_at, o.shipped_at, o.delivered_at, o.cancelled_at
		FROM orders o
		%s
	`, whereClause)

	var o order.Order
	var shippingAddressID, billingInfoID sql.NullInt64
	var interiorNumber, reference, billingRFC, billingRazonSocial sql.NullString
	var billingRegimen, billingUsoCFDI, billingPostalCode, billingEmail sql.NullString
	var invoiceXML, invoicePDF, customerNotes, adminNotes, paymentNotes sql.NullString
	var shippedAt, deliveredAt, cancelledAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&o.ID, &o.OrderNumber, &o.UserID, &o.Status,
		&shippingAddressID, &o.ShippingAddress.Street, &o.ShippingAddress.ExteriorNumber, &interiorNumber,
		&o.ShippingAddress.Neighborhood, &o.ShippingAddress.PostalCode,
		&o.ShippingAddress.City, &o.ShippingAddress.State, &reference, &o.ShippingAddress.Phone,
		&o.PaymentMethod, &o.PaymentMode, &o.PaymentInstallments, &paymentNotes,
		&o.RequiresInvoice, &billingInfoID,
		&billingRFC, &billingRazonSocial, &billingRegimen, &billingUsoCFDI,
		&billingPostalCode, &billingEmail,
		&o.Subtotal, &o.IVA, &o.ShippingCost, &o.Total,
		&invoiceXML, &invoicePDF, &customerNotes, &adminNotes,
		&o.CreatedAt, &o.UpdatedAt, &shippedAt, &deliveredAt, &cancelledAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, order.ErrNotFound
		}
		return nil, err
	}

	if shippingAddressID.Valid {
		o.ShippingAddress.ID = int(shippingAddressID.Int64)
	}
	o.ShippingAddress.InteriorNumber = interiorNumber.String
	o.ShippingAddress.Reference = reference.String
	o.InvoiceXMLPath = invoiceXML.String
	o.InvoicePDFPath = invoicePDF.String
	o.CustomerNotes = customerNotes.String
	o.AdminNotes = adminNotes.String
	o.PaymentNotes = paymentNotes.String

	if shippedAt.Valid {
		o.ShippedAt = &shippedAt.Time
	}
	if deliveredAt.Valid {
		o.DeliveredAt = &deliveredAt.Time
	}
	if cancelledAt.Valid {
		o.CancelledAt = &cancelledAt.Time
	}

	if billingRFC.Valid {
		o.BillingInfo = &order.BillingInfo{
			RFC:           billingRFC.String,
			RazonSocial:   billingRazonSocial.String,
			RegimenFiscal: billingRegimen.String,
			UsoCFDI:       billingUsoCFDI.String,
			PostalCode:    billingPostalCode.String,
			Email:         billingEmail.String,
		}
		if billingInfoID.Valid {
			o.BillingInfo.ID = int(billingInfoID.Int64)
		}
	}

	return &o, nil
}

func (r *OrderRepository) getOrderItems(ctx context.Context, orderID int) ([]order.OrderItem, error) {
	query := `
		SELECT id, order_id, tire_sku, tire_measure, tire_brand, tire_model,
			quantity, unit_price, subtotal, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []order.OrderItem
	for rows.Next() {
		var item order.OrderItem
		var brand, model sql.NullString

		err := rows.Scan(
			&item.ID, &item.OrderID, &item.TireSKU, &item.TireMeasure,
			&brand, &model, &item.Quantity, &item.UnitPrice, &item.Subtotal, &item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		item.TireBrand = brand.String
		item.TireModel = model.String
		items = append(items, item)
	}

	return items, rows.Err()
}
