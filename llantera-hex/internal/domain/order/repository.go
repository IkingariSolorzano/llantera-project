package order

import "context"

// Repository define las operaciones de persistencia para pedidos
type Repository interface {
	// Create crea un nuevo pedido
	Create(ctx context.Context, order *Order) error

	// GetByID obtiene un pedido por su ID
	GetByID(ctx context.Context, id int) (*Order, error)

	// GetByOrderNumber obtiene un pedido por su número
	GetByOrderNumber(ctx context.Context, orderNumber string) (*Order, error)

	// List lista pedidos con filtros
	List(ctx context.Context, filter ListFilter) ([]Order, int, error)

	// UpdateStatus actualiza el estado de un pedido
	UpdateStatus(ctx context.Context, id int, status Status, adminNotes string) error

	// UpdateInvoiceFiles actualiza las rutas de archivos de factura
	UpdateInvoiceFiles(ctx context.Context, id int, xmlPath, pdfPath string) error
}
