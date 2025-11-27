package userapp

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/llantera/hex/internal/domain/user"
)

// Service implementa los casos de uso de usuarios siguiendo los puertos definidos en el dominio.
type Service struct {
	repo user.Repository
}

func NewService(repo user.Repository) *Service {
	return &Service{repo: repo}
}

var _ user.Service = (*Service)(nil)

func (s *Service) Create(ctx context.Context, cmd user.CreateCommand) (*user.User, error) {
	email := strings.TrimSpace(strings.ToLower(cmd.Email))
	if email == "" {
		return nil, user.NewValidationError("el correo es obligatorio")
	}
	if strings.TrimSpace(cmd.Password) == "" {
		return nil, user.NewValidationError("la contraseña es obligatoria")
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, user.ErrNotFound) {
			return nil, err
		}
		// ErrNotFound: no existe otro usuario con ese correo, continuar
	} else if existing != nil {
		return nil, user.ErrEmailAlreadyUsed
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	role := cmd.Role
	if role == "" {
		role = user.RoleCustomer
	}

	level := cmd.Level
	if level == "" {
		level = user.PriceLevelPublic
	}

	priceLevelID := cmd.PriceLevelID
	if role == user.RoleEmployee {
		// Los empleados no requieren ni deben depender de un nivel de precio específico.
		// Se fuerza a nivel público y se ignora cualquier PriceLevelID recibido.
		level = user.PriceLevelPublic
		priceLevelID = nil
	}

	now := time.Now().UTC()
	entity := &user.User{
		ID:                  uuid.NewString(),
		Email:               email,
		FirstName:           strings.TrimSpace(cmd.FirstName),
		FirstLastName:       strings.TrimSpace(cmd.FirstLastName),
		SecondLastName:      strings.TrimSpace(cmd.SecondLastName),
		Phone:               strings.TrimSpace(cmd.Phone),
		AddressStreet:       strings.TrimSpace(cmd.AddressStreet),
		AddressNumber:       strings.TrimSpace(cmd.AddressNumber),
		AddressNeighborhood: strings.TrimSpace(cmd.AddressNeighborhood),
		AddressPostalCode:   strings.TrimSpace(cmd.AddressPostalCode),
		JobTitle:            strings.TrimSpace(cmd.JobTitle),
		Active:              cmd.Active,
		CompanyID:           cmd.CompanyID,
		ProfileImageURL:     strings.TrimSpace(cmd.ProfileImageURL),
		Role:                role,
		Level:               level,
		PriceLevelID:        priceLevelID,
		CreatedAt:           now,
		UpdatedAt:           now,
		PasswordHash:        string(hash),
	}
	if cmd.Active && entity.Active == false {
		entity.Active = true
	}

	entity.Name = entity.FullName()
	if entity.Name == "" {
		entity.Name = entity.Email
	}

	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}
	entity.PasswordHash = ""
	return entity, nil
}

func (s *Service) Update(ctx context.Context, cmd user.UpdateCommand) (*user.User, error) {
	if strings.TrimSpace(cmd.ID) == "" {
		return nil, user.NewValidationError("el identificador es obligatorio")
	}

	existing, err := s.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	if existing == nil {
		return nil, user.ErrNotFound
	}

	if cmd.Email != "" {
		email := strings.TrimSpace(strings.ToLower(cmd.Email))
		if email == "" {
			return nil, user.NewValidationError("el correo no puede estar vacío")
		}
		if email != existing.Email {
			existingByEmail, err := s.repo.GetByEmail(ctx, email)
			if err != nil && !errors.Is(err, user.ErrNotFound) {
				return nil, err
			}
			if existingByEmail != nil && existingByEmail.ID != existing.ID {
				return nil, user.ErrEmailAlreadyUsed
			}
			existing.Email = email
		}
	}

	if cmd.FirstName != "" {
		existing.FirstName = strings.TrimSpace(cmd.FirstName)
	}
	if cmd.FirstLastName != "" {
		existing.FirstLastName = strings.TrimSpace(cmd.FirstLastName)
	}
	if cmd.SecondLastName != "" {
		existing.SecondLastName = strings.TrimSpace(cmd.SecondLastName)
	}
	if cmd.Phone != "" {
		existing.Phone = strings.TrimSpace(cmd.Phone)
	}
	if cmd.AddressStreet != "" {
		existing.AddressStreet = strings.TrimSpace(cmd.AddressStreet)
	}
	if cmd.AddressNumber != "" {
		existing.AddressNumber = strings.TrimSpace(cmd.AddressNumber)
	}
	if cmd.AddressNeighborhood != "" {
		existing.AddressNeighborhood = strings.TrimSpace(cmd.AddressNeighborhood)
	}
	if cmd.AddressPostalCode != "" {
		existing.AddressPostalCode = strings.TrimSpace(cmd.AddressPostalCode)
	}
	if cmd.JobTitle != "" {
		existing.JobTitle = strings.TrimSpace(cmd.JobTitle)
	}

	existing.Active = cmd.Active
	existing.CompanyID = cmd.CompanyID
	existing.ProfileImageURL = strings.TrimSpace(cmd.ProfileImageURL)
	if cmd.Role != "" {
		existing.Role = cmd.Role
	}
	if cmd.Level != "" {
		existing.Level = cmd.Level
	}
	existing.PriceLevelID = cmd.PriceLevelID
	if existing.Role == user.RoleEmployee {
		// Para empleados, no mantener un PriceLevel específico.
		existing.PriceLevelID = nil
		existing.Level = user.PriceLevelPublic
	}
	existing.Name = existing.FullName()
	if existing.Name == "" {
		existing.Name = existing.Email
	}
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	existing.PasswordHash = ""
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return user.NewValidationError("el identificador es obligatorio")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return user.ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*user.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, user.NewValidationError("el identificador es obligatorio")
	}
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	if u == nil {
		return nil, user.ErrNotFound
	}
	u.PasswordHash = ""
	return u, nil
}

func (s *Service) List(ctx context.Context, filter user.ListFilter) ([]user.User, int, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	for i := range items {
		items[i].PasswordHash = ""
	}
	return items, total, nil
}

// Authenticate valida las credenciales del usuario usando el correo y la contraseña.
// Devuelve ErrInvalidCredentials cuando el correo no existe o la contraseña no coincide.
func (s *Service) Authenticate(ctx context.Context, email, password string) (*user.User, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if normalizedEmail == "" || strings.TrimSpace(password) == "" {
		return nil, user.NewValidationError("correo y contraseña son obligatorios")
	}

	u, err := s.repo.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, user.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, user.ErrInvalidCredentials
	}

	if !u.Active {
		// Para no filtrar información, se responde como credenciales inválidas.
		return nil, user.ErrInvalidCredentials
	}

	u.PasswordHash = ""
	return u, nil
}
