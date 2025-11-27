package customerrequestapp

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/llantera/hex/internal/domain/customerrequest"
)

// Service implementa los casos de uso del agregado CustomerRequest.
type Service struct {
	repo customerrequest.Repository
}

func NewService(repo customerrequest.Repository) *Service {
	return &Service{repo: repo}
}

var _ customerrequest.Service = (*Service)(nil)

func (s *Service) Create(ctx context.Context, cmd customerrequest.CreateCommand) (*customerrequest.CustomerRequest, error) {
	fullName := strings.TrimSpace(cmd.FullName)
	if fullName == "" {
		return nil, customerrequest.NewValidationError("el nombre completo es obligatorio")
	}

	requestType := strings.TrimSpace(cmd.RequestType)
	if requestType == "" {
		return nil, customerrequest.NewValidationError("el tipo de solicitud es obligatorio")
	}

	now := time.Now().UTC()
	entity := &customerrequest.CustomerRequest{
		ID:                uuid.NewString(),
		FullName:          fullName,
		RequestType:       requestType,
		Message:           strings.TrimSpace(cmd.Message),
		Phone:             strings.TrimSpace(cmd.Phone),
		ContactPreference: strings.TrimSpace(cmd.ContactPreference),
		Email:             strings.TrimSpace(strings.ToLower(cmd.Email)),
		Status:            customerrequest.StatusPending,
		Agreement:         "",
		CreatedAt:         now,
		UpdatedAt:         now,
		AttendedAt:        nil,
	}

	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) Update(ctx context.Context, cmd customerrequest.UpdateCommand) (*customerrequest.CustomerRequest, error) {
	if strings.TrimSpace(cmd.ID) == "" {
		return nil, customerrequest.NewValidationError("el identificador es obligatorio")
	}

	existing, err := s.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		if errors.Is(err, customerrequest.ErrNotFound) {
			return nil, customerrequest.ErrNotFound
		}
		return nil, err
	}
	if existing == nil {
		return nil, customerrequest.ErrNotFound
	}

	if cmd.Message != "" {
		existing.Message = strings.TrimSpace(cmd.Message)
	}
	if cmd.Agreement != "" {
		existing.Agreement = strings.TrimSpace(cmd.Agreement)
	}

	if cmd.Status != nil {
		statusValue := strings.TrimSpace(string(*cmd.Status))
		if statusValue == "" {
			return nil, customerrequest.NewValidationError("el estado no puede estar vacío")
		}
		status := customerrequest.Status(statusValue)
		switch status {
		case customerrequest.StatusPending, customerrequest.StatusViewed, customerrequest.StatusHandled:
			if existing.Status != status && status == customerrequest.StatusHandled && existing.AttendedAt == nil {
				now := time.Now().UTC()
				existing.AttendedAt = &now
			}
			existing.Status = status
		default:
			return nil, customerrequest.NewValidationError("estado inválido")
		}
	}

	if cmd.EmployeeID != nil {
		trimmed := strings.TrimSpace(*cmd.EmployeeID)
		if trimmed == "" {
			existing.EmployeeID = nil
		} else {
			existing.EmployeeID = &trimmed
		}
	}

	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return customerrequest.NewValidationError("el identificador es obligatorio")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, customerrequest.ErrNotFound) {
			return customerrequest.ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*customerrequest.CustomerRequest, error) {
	if strings.TrimSpace(id) == "" {
		return nil, customerrequest.NewValidationError("el identificador es obligatorio")
	}
	cr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, customerrequest.ErrNotFound) {
			return nil, customerrequest.ErrNotFound
		}
		return nil, err
	}
	if cr == nil {
		return nil, customerrequest.ErrNotFound
	}
	return cr, nil
}

func (s *Service) List(ctx context.Context, filter customerrequest.ListFilter) ([]customerrequest.CustomerRequest, int, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.List(ctx, filter)
}
