package billing

import (
	"context"

	"github.com/llantera/hex/internal/domain/billing"
)

// Service define las operaciones de negocio para datos de facturación
type Service struct {
	repo billing.Repository
}

// NewService crea una nueva instancia del servicio
func NewService(repo billing.Repository) *Service {
	return &Service{repo: repo}
}

// Create crea nuevos datos de facturación
func (s *Service) Create(ctx context.Context, userID string, req billing.CreateRequest) (*billing.BillingInfo, error) {
	// Validaciones básicas
	if req.RFC == "" || req.RazonSocial == "" || req.RegimenFiscal == "" ||
		req.UsoCFDI == "" || req.PostalCode == "" {
		return nil, billing.ErrValidation
	}

	if len(req.PostalCode) != 5 {
		return nil, billing.ErrValidation
	}

	return s.repo.Create(ctx, userID, req)
}

// GetByID obtiene datos de facturación por su ID
func (s *Service) GetByID(ctx context.Context, id int) (*billing.BillingInfo, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByUser lista los datos de facturación de un usuario
func (s *Service) ListByUser(ctx context.Context, userID string) ([]billing.BillingInfo, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// GetDefaultByUser obtiene los datos de facturación predeterminados del usuario
func (s *Service) GetDefaultByUser(ctx context.Context, userID string) (*billing.BillingInfo, error) {
	return s.repo.GetDefaultByUserID(ctx, userID)
}

// Update actualiza datos de facturación
func (s *Service) Update(ctx context.Context, id int, userID string, req billing.UpdateRequest) (*billing.BillingInfo, error) {
	// Verificar que los datos pertenecen al usuario
	info, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if info.UserID != userID {
		return nil, billing.ErrNotFound
	}

	// Validaciones
	if req.PostalCode != nil && len(*req.PostalCode) != 5 {
		return nil, billing.ErrValidation
	}

	return s.repo.Update(ctx, id, req)
}

// Delete elimina datos de facturación
func (s *Service) Delete(ctx context.Context, id int, userID string) error {
	// Verificar que los datos pertenecen al usuario
	info, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if info.UserID != userID {
		return billing.ErrNotFound
	}

	return s.repo.Delete(ctx, id)
}

// SetDefault establece datos de facturación como predeterminados
func (s *Service) SetDefault(ctx context.Context, userID string, billingID int) error {
	return s.repo.SetDefault(ctx, userID, billingID)
}
