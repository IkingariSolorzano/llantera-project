package customerrequest

import "context"

// Repository define las operaciones de persistencia para solicitudes.
type Repository interface {
	Create(ctx context.Context, r *CustomerRequest) error
	Update(ctx context.Context, r *CustomerRequest) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*CustomerRequest, error)
	List(ctx context.Context, filter ListFilter) ([]CustomerRequest, int, error)
}

// Service expone los casos de uso hacia la capa de presentación.
type Service interface {
	Create(ctx context.Context, cmd CreateCommand) (*CustomerRequest, error)
	Update(ctx context.Context, cmd UpdateCommand) (*CustomerRequest, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*CustomerRequest, error)
	List(ctx context.Context, filter ListFilter) ([]CustomerRequest, int, error)
}

// CreateCommand contiene los datos necesarios para registrar una nueva solicitud.
type CreateCommand struct {
	FullName          string
	RequestType       string
	Message           string
	Phone             string
	ContactPreference string
	Email             string
}

// UpdateCommand modela los cambios permitidos sobre una solicitud existente.
type UpdateCommand struct {
	ID         string
	Message    string
	Status     *Status
	EmployeeID *string
	Agreement  string
}
