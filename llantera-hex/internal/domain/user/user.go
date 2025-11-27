package user

import (
	"errors"
	"strings"
	"time"
)

var (
	// ErrNotFound se utiliza cuando un usuario no existe en el repositorio.
	ErrNotFound = errors.New("usuario no encontrado")
	// ErrEmailAlreadyUsed indica que el email ya está registrado.
	ErrEmailAlreadyUsed   = errors.New("el correo ya se encuentra registrado")
	ErrValidation         = errors.New("datos de usuario inválidos")
	ErrInvalidCredentials = errors.New("credenciales inválidas")
)

// Role define los distintos roles disponibles para los usuarios del sistema.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleCustomer Role = "customer"
	RoleEmployee Role = "employee"
)

// PriceLevel representa el código del nivel de precios/beneficios asociado a un usuario.
type PriceLevel string

const (
	PriceLevelPublic       PriceLevel = "public"
	PriceLevelEmpresa      PriceLevel = "empresa"
	PriceLevelDistribuidor PriceLevel = "distribuidor"
	PriceLevelMayorista    PriceLevel = "mayorista"
	PriceLevelSilver       PriceLevel = "silver"
	PriceLevelGold         PriceLevel = "gold"
	PriceLevelPlatinum     PriceLevel = "platinum"
)

// User concentra la información principal que se almacena para los usuarios.
type User struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	Name                string     `json:"name"`
	FirstName           string     `json:"firstName"`
	FirstLastName       string     `json:"firstLastName"`
	SecondLastName      string     `json:"secondLastName"`
	Phone               string     `json:"phone"`
	AddressStreet       string     `json:"addressStreet"`
	AddressNumber       string     `json:"addressNumber"`
	AddressNeighborhood string     `json:"addressNeighborhood"`
	AddressPostalCode   string     `json:"addressPostalCode"`
	JobTitle            string     `json:"jobTitle"`
	Active              bool       `json:"active"`
	CompanyID           *int       `json:"companyId"`
	ProfileImageURL     string     `json:"profileImageUrl"`
	PasswordHash        string     `json:"-"`
	Role                Role       `json:"role"`
	Level               PriceLevel `json:"level"`
	PriceLevelID        *int       `json:"priceLevelId"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

// FullName regresa el nombre completo calculado a partir de sus partes.
func (u *User) FullName() string {
	parts := []string{u.FirstName, u.FirstLastName, u.SecondLastName}
	return strings.TrimSpace(strings.Join(parts, " "))
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return ErrValidation
}

func NewValidationError(message string) error {
	m := strings.TrimSpace(message)
	if m == "" {
		m = "datos de usuario inválidos"
	}
	return &ValidationError{Message: m}
}

// ListFilter modela los criterios de búsqueda para listar usuarios con paginación y orden.
type ListFilter struct {
	Search    string
	CompanyID *int
	Role      *Role
	Active    *bool
	Limit     int
	Offset    int
	Sort      string
}
