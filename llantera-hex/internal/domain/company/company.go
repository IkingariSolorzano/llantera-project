package company

import (
	"errors"
	"time"
)

var (
	// ErrNotFound indica que la empresa solicitada no existe.
	ErrNotFound = errors.New("empresa no encontrada")
)

// Company representa la entidad de negocio para las empresas clientes.
type Company struct {
	ID            int       `json:"id"`
	KeyName       string    `json:"keyName"`
	SocialReason  string    `json:"socialReason"`
	RFC           string    `json:"rfc"`
	Address       string    `json:"address"`
	Emails        []string  `json:"emails"`
	Phones        []string  `json:"phones"`
	MainContactID *string   `json:"mainContactId"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ListFilter permite paginar, buscar y ordenar el catálogo de empresas.
type ListFilter struct {
	Search string
	Limit  int
	Offset int
	Sort   string
}
