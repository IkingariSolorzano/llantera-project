package address

import "context"

// Repository define las operaciones de persistencia para direcciones
type Repository interface {
	// Create crea una nueva dirección
	Create(ctx context.Context, userID string, req CreateRequest) (*Address, error)

	// GetByID obtiene una dirección por su ID
	GetByID(ctx context.Context, id int) (*Address, error)

	// ListByUserID lista las direcciones de un usuario
	ListByUserID(ctx context.Context, userID string) ([]Address, error)

	// Update actualiza una dirección
	Update(ctx context.Context, id int, req UpdateRequest) (*Address, error)

	// Delete elimina una dirección
	Delete(ctx context.Context, id int) error

	// SetDefault establece una dirección como predeterminada
	SetDefault(ctx context.Context, userID string, addressID int) error
}
