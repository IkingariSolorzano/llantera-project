package cart

import "context"

// Repository define las operaciones de persistencia del carrito
type Repository interface {
	// GetOrCreateCart obtiene el carrito del usuario o lo crea si no existe
	GetOrCreateCart(ctx context.Context, userID string) (*Cart, error)

	// GetCartWithDetails obtiene el carrito con detalles de productos
	GetCartWithDetails(ctx context.Context, userID string, priceLevel string) (*CartWithDetails, error)

	// AddItem agrega un item al carrito o incrementa la cantidad si ya existe
	AddItem(ctx context.Context, userID string, item AddItemRequest) (*CartItem, error)

	// UpdateItemQuantity actualiza la cantidad de un item
	UpdateItemQuantity(ctx context.Context, userID string, tireSKU string, quantity int) (*CartItem, error)

	// RemoveItem elimina un item del carrito
	RemoveItem(ctx context.Context, userID string, tireSKU string) error

	// ClearCart vacía el carrito
	ClearCart(ctx context.Context, userID string) error

	// GetItemBySKU obtiene un item específico del carrito
	GetItemBySKU(ctx context.Context, userID string, tireSKU string) (*CartItem, error)
}
