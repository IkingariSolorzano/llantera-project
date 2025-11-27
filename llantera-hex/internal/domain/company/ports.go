package company

import "context"

// Repository define las operaciones de persistencia que la capa de infraestructura
// debe implementar para la entidad Company.
type Repository interface {
	Create(ctx context.Context, c *Company) error
	Update(ctx context.Context, c *Company) error
	Delete(ctx context.Context, id int) error
	GetByID(ctx context.Context, id int) (*Company, error)
	List(ctx context.Context, filter ListFilter) ([]Company, int, error)
}

// Service describe los casos de uso disponibles para la capa de presentación.
type Service interface {
	Create(ctx context.Context, cmd CreateCommand) (*Company, error)
	Update(ctx context.Context, cmd UpdateCommand) (*Company, error)
	Delete(ctx context.Context, id int) error
	Get(ctx context.Context, id int) (*Company, error)
	List(ctx context.Context, filter ListFilter) ([]Company, int, error)
}

// CreateCommand agrupa los datos necesarios para crear una empresa.
type CreateCommand struct {
	KeyName       string
	SocialReason  string
	RFC           string
	Address       string
	Emails        []string
	Phones        []string
	MainContactID *string
}

// UpdateCommand representa la edición de una empresa existente.
type UpdateCommand struct {
	ID            int
	KeyName       string
	SocialReason  string
	RFC           string
	Address       string
	Emails        []string
	Phones        []string
	MainContactID *string
}
