package billing

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("datos de facturación no encontrados")
	ErrValidation = errors.New("datos de facturación inválidos")
)

// BillingInfo representa los datos de facturación de un cliente
type BillingInfo struct {
	ID            int       `json:"id"`
	UserID        string    `json:"userId"`
	RFC           string    `json:"rfc"`
	RazonSocial   string    `json:"razonSocial"`
	RegimenFiscal string    `json:"regimenFiscal"`
	UsoCFDI       string    `json:"usoCfdi"`
	PostalCode    string    `json:"postalCode"`
	Email         string    `json:"email,omitempty"`
	IsDefault     bool      `json:"isDefault"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// CreateRequest representa la solicitud para crear datos de facturación
type CreateRequest struct {
	RFC           string `json:"rfc"`
	RazonSocial   string `json:"razonSocial"`
	RegimenFiscal string `json:"regimenFiscal"`
	UsoCFDI       string `json:"usoCfdi"`
	PostalCode    string `json:"postalCode"`
	Email         string `json:"email,omitempty"`
	IsDefault     bool   `json:"isDefault"`
}

// UpdateRequest representa la solicitud para actualizar datos de facturación
type UpdateRequest struct {
	RFC           *string `json:"rfc,omitempty"`
	RazonSocial   *string `json:"razonSocial,omitempty"`
	RegimenFiscal *string `json:"regimenFiscal,omitempty"`
	UsoCFDI       *string `json:"usoCfdi,omitempty"`
	PostalCode    *string `json:"postalCode,omitempty"`
	Email         *string `json:"email,omitempty"`
	IsDefault     *bool   `json:"isDefault,omitempty"`
}

// Repository define las operaciones de persistencia para datos de facturación
type Repository interface {
	Create(ctx context.Context, userID string, req CreateRequest) (*BillingInfo, error)
	GetByID(ctx context.Context, id int) (*BillingInfo, error)
	ListByUserID(ctx context.Context, userID string) ([]BillingInfo, error)
	GetDefaultByUserID(ctx context.Context, userID string) (*BillingInfo, error)
	Update(ctx context.Context, id int, req UpdateRequest) (*BillingInfo, error)
	Delete(ctx context.Context, id int) error
	SetDefault(ctx context.Context, userID string, billingID int) error
}
