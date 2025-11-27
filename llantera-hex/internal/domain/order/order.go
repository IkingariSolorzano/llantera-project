package order

import (
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("pedido no encontrado")
	ErrInvalidStatus     = errors.New("estado de pedido inválido")
	ErrEmptyCart         = errors.New("el carrito está vacío")
	ErrInsufficientStock = errors.New("stock insuficiente")
	ErrValidation        = errors.New("datos de pedido inválidos")
)

// Status representa el estado de un pedido
type Status string

const (
	StatusSolicitado Status = "solicitado"
	StatusPreparando Status = "preparando"
	StatusEnviado    Status = "enviado"
	StatusEntregado  Status = "entregado"
	StatusCancelado  Status = "cancelado"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusSolicitado, StatusPreparando, StatusEnviado, StatusEntregado, StatusCancelado:
		return true
	}
	return false
}

func (s Status) CanTransitionTo(next Status) bool {
	switch s {
	case StatusSolicitado:
		return next == StatusPreparando || next == StatusCancelado
	case StatusPreparando:
		return next == StatusEnviado || next == StatusCancelado
	case StatusEnviado:
		return next == StatusEntregado || next == StatusCancelado
	case StatusEntregado, StatusCancelado:
		return false
	}
	return false
}

// PaymentMethod representa el método de pago
type PaymentMethod string

const (
	PaymentTransferencia PaymentMethod = "transferencia"
	PaymentTarjeta       PaymentMethod = "tarjeta"
	PaymentEfectivo      PaymentMethod = "efectivo"
)

func (p PaymentMethod) IsValid() bool {
	switch p {
	case PaymentTransferencia, PaymentTarjeta, PaymentEfectivo:
		return true
	}
	return false
}

func (p PaymentMethod) Label() string {
	switch p {
	case PaymentTransferencia:
		return "Transferencia bancaria"
	case PaymentTarjeta:
		return "Tarjeta de crédito/débito"
	case PaymentEfectivo:
		return "Pago en efectivo"
	}
	return string(p)
}

// PaymentMode representa la modalidad de pago
type PaymentMode string

const (
	PaymentModeContado       PaymentMode = "contado"       // Pago único al momento
	PaymentModeCredito       PaymentMode = "credito"       // Pago diferido (clientes con crédito)
	PaymentModeParcialidades PaymentMode = "parcialidades" // Pago en cuotas
	PaymentModeAnticipo      PaymentMode = "anticipo"      // Anticipo + resto contra entrega
)

func (m PaymentMode) IsValid() bool {
	switch m {
	case PaymentModeContado, PaymentModeCredito, PaymentModeParcialidades, PaymentModeAnticipo:
		return true
	}
	return false
}

func (m PaymentMode) Label() string {
	switch m {
	case PaymentModeContado:
		return "Pago de contado"
	case PaymentModeCredito:
		return "Crédito"
	case PaymentModeParcialidades:
		return "Pago en parcialidades"
	case PaymentModeAnticipo:
		return "Anticipo"
	}
	return string(m)
}

// OrderItem representa un producto en el pedido
type OrderItem struct {
	ID          int       `json:"id"`
	OrderID     int       `json:"orderId"`
	TireSKU     string    `json:"tireSku"`
	TireMeasure string    `json:"tireMeasure"`
	TireBrand   string    `json:"tireBrand,omitempty"`
	TireModel   string    `json:"tireModel,omitempty"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unitPrice"`
	Subtotal    float64   `json:"subtotal"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ShippingAddress representa la dirección de envío
type ShippingAddress struct {
	ID             int    `json:"id,omitempty"`
	Street         string `json:"street"`
	ExteriorNumber string `json:"exteriorNumber"`
	InteriorNumber string `json:"interiorNumber,omitempty"`
	Neighborhood   string `json:"neighborhood"`
	PostalCode     string `json:"postalCode"`
	City           string `json:"city"`
	State          string `json:"state"`
	Reference      string `json:"reference,omitempty"`
	Phone          string `json:"phone"`
}

// BillingInfo representa los datos de facturación
type BillingInfo struct {
	ID            int    `json:"id,omitempty"`
	RFC           string `json:"rfc"`
	RazonSocial   string `json:"razonSocial"`
	RegimenFiscal string `json:"regimenFiscal"`
	UsoCFDI       string `json:"usoCfdi"`
	PostalCode    string `json:"postalCode"`
	Email         string `json:"email,omitempty"`
}

// Order representa un pedido completo
type Order struct {
	ID                  int             `json:"id"`
	OrderNumber         string          `json:"orderNumber"`
	UserID              string          `json:"userId"`
	Status              Status          `json:"status"`
	ShippingAddress     ShippingAddress `json:"shippingAddress"`
	PaymentMethod       PaymentMethod   `json:"paymentMethod"`
	PaymentMode         PaymentMode     `json:"paymentMode"`
	PaymentInstallments int             `json:"paymentInstallments,omitempty"`
	PaymentNotes        string          `json:"paymentNotes,omitempty"`
	RequiresInvoice     bool            `json:"requiresInvoice"`
	BillingInfo         *BillingInfo    `json:"billingInfo,omitempty"`
	Items               []OrderItem     `json:"items"`
	Subtotal            float64         `json:"subtotal"`
	IVA                 float64         `json:"iva"`
	ShippingCost        float64         `json:"shippingCost"`
	Total               float64         `json:"total"`
	InvoiceXMLPath      string          `json:"invoiceXmlPath,omitempty"`
	InvoicePDFPath      string          `json:"invoicePdfPath,omitempty"`
	CustomerNotes       string          `json:"customerNotes,omitempty"`
	AdminNotes          string          `json:"adminNotes,omitempty"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
	ShippedAt           *time.Time      `json:"shippedAt,omitempty"`
	DeliveredAt         *time.Time      `json:"deliveredAt,omitempty"`
	CancelledAt         *time.Time      `json:"cancelledAt,omitempty"`
}

// IVARate es la tasa de IVA (16%)
const IVARate = 0.16

// CalculateTotals recalcula los totales del pedido incluyendo IVA
func (o *Order) CalculateTotals() {
	var subtotal float64
	for _, item := range o.Items {
		subtotal += item.Subtotal
	}
	o.Subtotal = subtotal
	o.IVA = subtotal * IVARate
	o.Total = subtotal + o.IVA + o.ShippingCost
}

// ListFilter define los filtros para listar pedidos
type ListFilter struct {
	UserID string
	Status *Status
	Search string
	Limit  int
	Offset int
	Sort   string
}

// CreateOrderRequest representa la solicitud para crear un pedido
type CreateOrderRequest struct {
	Items               []CreateOrderItemRequest `json:"items"`
	ShippingAddress     ShippingAddress          `json:"shippingAddress"`
	PaymentMethod       PaymentMethod            `json:"paymentMethod"`
	PaymentMode         PaymentMode              `json:"paymentMode"`
	PaymentInstallments int                      `json:"paymentInstallments,omitempty"`
	PaymentNotes        string                   `json:"paymentNotes,omitempty"`
	RequiresInvoice     bool                     `json:"requiresInvoice"`
	BillingInfo         *BillingInfo             `json:"billingInfo,omitempty"`
	CustomerNotes       string                   `json:"customerNotes,omitempty"`
	// Totales calculados por el frontend (opcional, se recalculan en backend)
	Subtotal float64 `json:"subtotal,omitempty"`
	IVA      float64 `json:"iva,omitempty"`
	Total    float64 `json:"total,omitempty"`
}

type CreateOrderItemRequest struct {
	TireSKU     string  `json:"tireSku"`
	TireMeasure string  `json:"tireMeasure"`
	TireBrand   string  `json:"tireBrand,omitempty"`
	TireModel   string  `json:"tireModel,omitempty"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
}

// UpdateStatusRequest representa la solicitud para cambiar estado
type UpdateStatusRequest struct {
	Status     Status `json:"status"`
	AdminNotes string `json:"adminNotes,omitempty"`
}
