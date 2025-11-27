package customerrequest

import (
	"errors"
	"strings"
	"time"
)

var (
	// ErrNotFound indica que la solicitud no existe.
	ErrNotFound = errors.New("solicitud no encontrada")
	// ErrValidation agrupa errores de validación de datos.
	ErrValidation = errors.New("datos de solicitud inválidos")
)

// Status representa el estado de la solicitud.
type Status string

const (
	StatusPending Status = "pendiente"
	StatusViewed  Status = "vista"
	StatusHandled Status = "atendida"
)

// CustomerRequest modela la solicitud enviada desde "Quiero ser cliente".
type CustomerRequest struct {
	ID                string     `json:"id"`
	FullName          string     `json:"fullName"`
	RequestType       string     `json:"requestType"`
	Message           string     `json:"message"`
	Phone             string     `json:"phone"`
	ContactPreference string     `json:"contactPreference"`
	Email             string     `json:"email"`
	Status            Status     `json:"status"`
	EmployeeID        *string    `json:"employeeId"`
	Agreement         string     `json:"agreement"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	AttendedAt        *time.Time `json:"attendedAt"`
}

// ValidationError permite distinguir errores de validación.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return ErrValidation
}

// NewValidationError crea un error de validación con mensaje amigable.
func NewValidationError(message string) error {
	m := strings.TrimSpace(message)
	if m == "" {
		m = "datos de solicitud inválidos"
	}
	return &ValidationError{Message: m}
}

// ListFilter permite paginar y filtrar solicitudes.
type ListFilter struct {
	Search     string
	Status     *Status
	EmployeeID *string
	Limit      int
	Offset     int
	Sort       string
}
