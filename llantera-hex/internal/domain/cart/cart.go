package cart

import (
	"errors"
	"time"
)

var (
	ErrCartNotFound      = errors.New("carrito no encontrado")
	ErrItemNotFound      = errors.New("item no encontrado en el carrito")
	ErrInvalidQuantity   = errors.New("cantidad inválida")
	ErrInsufficientStock = errors.New("stock insuficiente")
)

// CartItem representa un item en el carrito
type CartItem struct {
	ID        int       `json:"id"`
	CartID    int       `json:"cartId"`
	TireSKU   string    `json:"tireSku"`
	Quantity  int       `json:"quantity"`
	AddedAt   time.Time `json:"addedAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Cart representa el carrito de un usuario
type Cart struct {
	ID        int        `json:"id"`
	UserID    string     `json:"userId"`
	Items     []CartItem `json:"items"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// CartItemWithDetails incluye información adicional del producto
type CartItemWithDetails struct {
	CartItem
	TireMeasure string  `json:"tireMeasure"`
	TireBrand   string  `json:"tireBrand,omitempty"`
	TireModel   string  `json:"tireModel,omitempty"`
	TireImage   string  `json:"tireImage,omitempty"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

// CartWithDetails incluye items con detalles de productos
type CartWithDetails struct {
	ID        int                   `json:"id"`
	UserID    string                `json:"userId"`
	Items     []CartItemWithDetails `json:"items"`
	Subtotal  float64               `json:"subtotal"`
	ItemCount int                   `json:"itemCount"`
	CreatedAt time.Time             `json:"createdAt"`
	UpdatedAt time.Time             `json:"updatedAt"`
}

// AddItemRequest representa la solicitud para agregar un item
type AddItemRequest struct {
	TireSKU  string `json:"tireSku"`
	Quantity int    `json:"quantity"`
}

// UpdateItemRequest representa la solicitud para actualizar cantidad
type UpdateItemRequest struct {
	Quantity int `json:"quantity"`
}

// Validate valida la solicitud de agregar item
func (r *AddItemRequest) Validate() error {
	if r.TireSKU == "" {
		return errors.New("el SKU es requerido")
	}
	if r.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	return nil
}

// Validate valida la solicitud de actualizar item
func (r *UpdateItemRequest) Validate() error {
	if r.Quantity < 0 {
		return ErrInvalidQuantity
	}
	return nil
}
