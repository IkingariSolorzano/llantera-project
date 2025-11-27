package storage

import (
	"context"
	"database/sql"

	"github.com/llantera/hex/internal/domain/cart"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) GetOrCreateCart(ctx context.Context, userID string) (*cart.Cart, error) {
	// Intentar obtener el carrito existente
	var c cart.Cart
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, created_at, updated_at
		FROM user_carts
		WHERE user_id = $1
	`, userID).Scan(&c.ID, &c.UserID, &c.CreatedAt, &c.UpdatedAt)

	if err == sql.ErrNoRows {
		// Crear nuevo carrito
		err = r.db.QueryRowContext(ctx, `
			INSERT INTO user_carts (user_id, created_at, updated_at)
			VALUES ($1, NOW(), NOW())
			RETURNING id, user_id, created_at, updated_at
		`, userID).Scan(&c.ID, &c.UserID, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Obtener items del carrito
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, cart_id, tire_sku, quantity, added_at, updated_at
		FROM cart_items
		WHERE cart_id = $1
		ORDER BY added_at DESC
	`, c.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	c.Items = []cart.CartItem{}
	for rows.Next() {
		var item cart.CartItem
		if err := rows.Scan(&item.ID, &item.CartID, &item.TireSKU, &item.Quantity, &item.AddedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		c.Items = append(c.Items, item)
	}

	return &c, nil
}

func (r *CartRepository) GetCartWithDetails(ctx context.Context, userID string, priceLevel string) (*cart.CartWithDetails, error) {
	c, err := r.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := &cart.CartWithDetails{
		ID:        c.ID,
		UserID:    c.UserID,
		Items:     []cart.CartItemWithDetails{},
		Subtotal:  0,
		ItemCount: 0,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}

	if len(c.Items) == 0 {
		return result, nil
	}

	// Obtener detalles de cada item con precio según nivel
	for _, item := range c.Items {
		var detail cart.CartItemWithDetails
		detail.CartItem = item

		// Obtener información de la llanta
		var tireID string
		var urlImagen sql.NullString
		err := r.db.QueryRowContext(ctx, `
			SELECT l.id, l.medida_original, m.nombre, l.modelo, l.url_imagen
			FROM llantas l
			LEFT JOIN marcas m ON l.marca_id = m.id
			WHERE l.sku = $1
		`, item.TireSKU).Scan(&tireID, &detail.TireMeasure, &detail.TireBrand, &detail.TireModel, &urlImagen)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if urlImagen.Valid {
			detail.TireImage = urlImagen.String
		}

		// Obtener precio público de la llanta
		err = r.db.QueryRowContext(ctx, `
			SELECT COALESCE(precio_publico, 0) FROM llantas WHERE id = $1
		`, tireID).Scan(&detail.Price)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		// Obtener stock
		err = r.db.QueryRowContext(ctx, `
			SELECT COALESCE(cantidad, 0) FROM inventario WHERE llanta_id = $1
		`, tireID).Scan(&detail.Stock)
		if err != nil && err != sql.ErrNoRows {
			detail.Stock = 0
		}

		result.Items = append(result.Items, detail)
		result.Subtotal += detail.Price * float64(item.Quantity)
		result.ItemCount += item.Quantity
	}

	return result, nil
}

func (r *CartRepository) AddItem(ctx context.Context, userID string, req cart.AddItemRequest) (*cart.CartItem, error) {
	c, err := r.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Verificar si el item ya existe
	var existingItem cart.CartItem
	err = r.db.QueryRowContext(ctx, `
		SELECT id, cart_id, tire_sku, quantity, added_at, updated_at
		FROM cart_items
		WHERE cart_id = $1 AND tire_sku = $2
	`, c.ID, req.TireSKU).Scan(&existingItem.ID, &existingItem.CartID, &existingItem.TireSKU, &existingItem.Quantity, &existingItem.AddedAt, &existingItem.UpdatedAt)

	if err == sql.ErrNoRows {
		// Insertar nuevo item
		var newItem cart.CartItem
		err = r.db.QueryRowContext(ctx, `
			INSERT INTO cart_items (cart_id, tire_sku, quantity, added_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
			RETURNING id, cart_id, tire_sku, quantity, added_at, updated_at
		`, c.ID, req.TireSKU, req.Quantity).Scan(&newItem.ID, &newItem.CartID, &newItem.TireSKU, &newItem.Quantity, &newItem.AddedAt, &newItem.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Actualizar timestamp del carrito
		r.db.ExecContext(ctx, `UPDATE user_carts SET updated_at = NOW() WHERE id = $1`, c.ID)

		return &newItem, nil
	} else if err != nil {
		return nil, err
	}

	// Actualizar cantidad existente
	newQuantity := existingItem.Quantity + req.Quantity
	err = r.db.QueryRowContext(ctx, `
		UPDATE cart_items
		SET quantity = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, cart_id, tire_sku, quantity, added_at, updated_at
	`, newQuantity, existingItem.ID).Scan(&existingItem.ID, &existingItem.CartID, &existingItem.TireSKU, &existingItem.Quantity, &existingItem.AddedAt, &existingItem.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Actualizar timestamp del carrito
	r.db.ExecContext(ctx, `UPDATE user_carts SET updated_at = NOW() WHERE id = $1`, c.ID)

	return &existingItem, nil
}

func (r *CartRepository) UpdateItemQuantity(ctx context.Context, userID string, tireSKU string, quantity int) (*cart.CartItem, error) {
	c, err := r.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	if quantity <= 0 {
		// Si la cantidad es 0 o menor, eliminar el item
		return nil, r.RemoveItem(ctx, userID, tireSKU)
	}

	var item cart.CartItem
	err = r.db.QueryRowContext(ctx, `
		UPDATE cart_items
		SET quantity = $1, updated_at = NOW()
		WHERE cart_id = $2 AND tire_sku = $3
		RETURNING id, cart_id, tire_sku, quantity, added_at, updated_at
	`, quantity, c.ID, tireSKU).Scan(&item.ID, &item.CartID, &item.TireSKU, &item.Quantity, &item.AddedAt, &item.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, cart.ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	// Actualizar timestamp del carrito
	r.db.ExecContext(ctx, `UPDATE user_carts SET updated_at = NOW() WHERE id = $1`, c.ID)

	return &item, nil
}

func (r *CartRepository) RemoveItem(ctx context.Context, userID string, tireSKU string) error {
	c, err := r.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE cart_id = $1 AND tire_sku = $2
	`, c.ID, tireSKU)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return cart.ErrItemNotFound
	}

	// Actualizar timestamp del carrito
	r.db.ExecContext(ctx, `UPDATE user_carts SET updated_at = NOW() WHERE id = $1`, c.ID)

	return nil
}

func (r *CartRepository) ClearCart(ctx context.Context, userID string) error {
	c, err := r.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, c.ID)
	if err != nil {
		return err
	}

	// Actualizar timestamp del carrito
	r.db.ExecContext(ctx, `UPDATE user_carts SET updated_at = NOW() WHERE id = $1`, c.ID)

	return nil
}

func (r *CartRepository) GetItemBySKU(ctx context.Context, userID string, tireSKU string) (*cart.CartItem, error) {
	c, err := r.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	var item cart.CartItem
	err = r.db.QueryRowContext(ctx, `
		SELECT id, cart_id, tire_sku, quantity, added_at, updated_at
		FROM cart_items
		WHERE cart_id = $1 AND tire_sku = $2
	`, c.ID, tireSKU).Scan(&item.ID, &item.CartID, &item.TireSKU, &item.Quantity, &item.AddedAt, &item.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, cart.ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// Ensure CartRepository implements cart.Repository
var _ cart.Repository = (*CartRepository)(nil)
