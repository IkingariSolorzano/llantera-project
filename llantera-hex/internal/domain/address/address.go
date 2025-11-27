package address

import (
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("dirección no encontrada")
	ErrValidation = errors.New("datos de dirección inválidos")
)

// Address representa una dirección de envío de un cliente
type Address struct {
	ID             int       `json:"id"`
	UserID         string    `json:"userId"`
	Alias          string    `json:"alias"`
	Street         string    `json:"street"`
	ExteriorNumber string    `json:"exteriorNumber"`
	InteriorNumber string    `json:"interiorNumber,omitempty"`
	Neighborhood   string    `json:"neighborhood"`
	PostalCode     string    `json:"postalCode"`
	City           string    `json:"city"`
	State          string    `json:"state"`
	Reference      string    `json:"reference,omitempty"`
	Phone          string    `json:"phone"`
	IsDefault      bool      `json:"isDefault"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CreateRequest representa la solicitud para crear una dirección
type CreateRequest struct {
	Alias          string `json:"alias"`
	Street         string `json:"street"`
	ExteriorNumber string `json:"exteriorNumber"`
	InteriorNumber string `json:"interiorNumber,omitempty"`
	Neighborhood   string `json:"neighborhood"`
	PostalCode     string `json:"postalCode"`
	City           string `json:"city"`
	State          string `json:"state"`
	Reference      string `json:"reference,omitempty"`
	Phone          string `json:"phone"`
	IsDefault      bool   `json:"isDefault"`
}

// UpdateRequest representa la solicitud para actualizar una dirección
type UpdateRequest struct {
	Alias          *string `json:"alias,omitempty"`
	Street         *string `json:"street,omitempty"`
	ExteriorNumber *string `json:"exteriorNumber,omitempty"`
	InteriorNumber *string `json:"interiorNumber,omitempty"`
	Neighborhood   *string `json:"neighborhood,omitempty"`
	PostalCode     *string `json:"postalCode,omitempty"`
	City           *string `json:"city,omitempty"`
	State          *string `json:"state,omitempty"`
	Reference      *string `json:"reference,omitempty"`
	Phone          *string `json:"phone,omitempty"`
	IsDefault      *bool   `json:"isDefault,omitempty"`
}

// FormatFull devuelve la dirección formateada completa
func (a *Address) FormatFull() string {
	addr := a.Street + " #" + a.ExteriorNumber
	if a.InteriorNumber != "" {
		addr += " Int. " + a.InteriorNumber
	}
	addr += ", " + a.Neighborhood + ", C.P. " + a.PostalCode + ", " + a.City + ", " + a.State
	return addr
}
