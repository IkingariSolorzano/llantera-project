package cart

import (
	"context"

	"github.com/llantera/hex/internal/domain/cart"
)

type Service struct {
	repo cart.Repository
}

func NewService(repo cart.Repository) *Service {
	return &Service{repo: repo}
}

// GetCart obtiene el carrito del usuario con detalles de productos
func (s *Service) GetCart(ctx context.Context, userID string, priceLevel string) (*cart.CartWithDetails, error) {
	if priceLevel == "" {
		priceLevel = "public"
	}
	return s.repo.GetCartWithDetails(ctx, userID, priceLevel)
}

// AddItem agrega un item al carrito
func (s *Service) AddItem(ctx context.Context, userID string, req cart.AddItemRequest) (*cart.CartItem, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return s.repo.AddItem(ctx, userID, req)
}

// UpdateItemQuantity actualiza la cantidad de un item
func (s *Service) UpdateItemQuantity(ctx context.Context, userID string, tireSKU string, quantity int) (*cart.CartItem, error) {
	if quantity < 0 {
		return nil, cart.ErrInvalidQuantity
	}
	if quantity == 0 {
		return nil, s.repo.RemoveItem(ctx, userID, tireSKU)
	}
	return s.repo.UpdateItemQuantity(ctx, userID, tireSKU, quantity)
}

// RemoveItem elimina un item del carrito
func (s *Service) RemoveItem(ctx context.Context, userID string, tireSKU string) error {
	return s.repo.RemoveItem(ctx, userID, tireSKU)
}

// ClearCart vacía el carrito
func (s *Service) ClearCart(ctx context.Context, userID string) error {
	return s.repo.ClearCart(ctx, userID)
}

// GetItemBySKU obtiene un item específico del carrito
func (s *Service) GetItemBySKU(ctx context.Context, userID string, tireSKU string) (*cart.CartItem, error) {
	return s.repo.GetItemBySKU(ctx, userID, tireSKU)
}
