package address

import (
	"context"

	"github.com/llantera/hex/internal/domain/address"
)

// Service define las operaciones de negocio para direcciones
type Service struct {
	repo address.Repository
}

// NewService crea una nueva instancia del servicio
func NewService(repo address.Repository) *Service {
	return &Service{repo: repo}
}

// Create crea una nueva dirección
func (s *Service) Create(ctx context.Context, userID string, req address.CreateRequest) (*address.Address, error) {
	// Validaciones básicas
	if req.Street == "" || req.ExteriorNumber == "" || req.Neighborhood == "" ||
		req.PostalCode == "" || req.City == "" || req.State == "" || req.Phone == "" {
		return nil, address.ErrValidation
	}

	if len(req.PostalCode) != 5 {
		return nil, address.ErrValidation
	}

	if len(req.Phone) != 10 {
		return nil, address.ErrValidation
	}

	if req.Alias == "" {
		req.Alias = "Principal"
	}

	return s.repo.Create(ctx, userID, req)
}

// GetByID obtiene una dirección por su ID
func (s *Service) GetByID(ctx context.Context, id int) (*address.Address, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByUser lista las direcciones de un usuario
func (s *Service) ListByUser(ctx context.Context, userID string) ([]address.Address, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// Update actualiza una dirección
func (s *Service) Update(ctx context.Context, id int, userID string, req address.UpdateRequest) (*address.Address, error) {
	// Verificar que la dirección pertenece al usuario
	addr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if addr.UserID != userID {
		return nil, address.ErrNotFound
	}

	// Validaciones
	if req.PostalCode != nil && len(*req.PostalCode) != 5 {
		return nil, address.ErrValidation
	}

	if req.Phone != nil && len(*req.Phone) != 10 {
		return nil, address.ErrValidation
	}

	return s.repo.Update(ctx, id, req)
}

// Delete elimina una dirección
func (s *Service) Delete(ctx context.Context, id int, userID string) error {
	// Verificar que la dirección pertenece al usuario
	addr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if addr.UserID != userID {
		return address.ErrNotFound
	}

	return s.repo.Delete(ctx, id)
}

// SetDefault establece una dirección como predeterminada
func (s *Service) SetDefault(ctx context.Context, userID string, addressID int) error {
	return s.repo.SetDefault(ctx, userID, addressID)
}
