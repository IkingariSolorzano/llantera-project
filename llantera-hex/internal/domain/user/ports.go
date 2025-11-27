package user

import "context"

// Repository define las operaciones de persistencia que la capa de infraestructura
// debe implementar para interactuar con el almacenamiento de usuarios.
type Repository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, filter ListFilter) ([]User, int, error)
}

// Service captura los casos de uso disponibles para la capa de presentación.
type Service interface {
	Create(ctx context.Context, cmd CreateCommand) (*User, error)
	Update(ctx context.Context, cmd UpdateCommand) (*User, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*User, error)
	List(ctx context.Context, filter ListFilter) ([]User, int, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
}

// CreateCommand describe la información necesaria para registrar un usuario.
type CreateCommand struct {
	Email               string
	Password            string
	FirstName           string
	FirstLastName       string
	SecondLastName      string
	Phone               string
	AddressStreet       string
	AddressNumber       string
	AddressNeighborhood string
	AddressPostalCode   string
	JobTitle            string
	Active              bool
	CompanyID           *int
	ProfileImageURL     string
	Role                Role
	Level               PriceLevel
	PriceLevelID        *int
}

// UpdateCommand contiene los campos editables de un usuario existente.
type UpdateCommand struct {
	ID                  string
	Email               string
	FirstName           string
	FirstLastName       string
	SecondLastName      string
	Phone               string
	AddressStreet       string
	AddressNumber       string
	AddressNeighborhood string
	AddressPostalCode   string
	JobTitle            string
	Active              bool
	CompanyID           *int
	ProfileImageURL     string
	Role                Role
	Level               PriceLevel
	PriceLevelID        *int
}
