package companyapp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/llantera/hex/internal/domain/company"
)

var rfcRegex = regexp.MustCompile("^[A-ZÑ&]{3,4}[0-9]{6}[A-Z0-9]{3}$")

// Service implementa los casos de uso del agregado Company.
type Service struct {
	repo company.Repository
}

func NewService(repo company.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, cmd company.CreateCommand) (*company.Company, error) {
	if strings.TrimSpace(cmd.KeyName) == "" {
		return nil, fmt.Errorf("key name is required")
	}
	if strings.TrimSpace(cmd.SocialReason) == "" {
		return nil, fmt.Errorf("social reason is required")
	}

	rfc := strings.TrimSpace(strings.ToUpper(cmd.RFC))
	if rfc != "" && !rfcRegex.MatchString(rfc) {
		return nil, fmt.Errorf("formato de RFC inválido")
	}

	now := time.Now().UTC()
	entity := &company.Company{
		KeyName:       strings.TrimSpace(cmd.KeyName),
		SocialReason:  strings.TrimSpace(cmd.SocialReason),
		RFC:           rfc,
		Address:       strings.TrimSpace(cmd.Address),
		Emails:        normalizeSlice(cmd.Emails),
		Phones:        normalizeSlice(cmd.Phones),
		MainContactID: cmd.MainContactID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) Update(ctx context.Context, cmd company.UpdateCommand) (*company.Company, error) {
	if cmd.ID <= 0 {
		return nil, fmt.Errorf("invalid company id")
	}
	if strings.TrimSpace(cmd.KeyName) == "" {
		return nil, fmt.Errorf("key name is required")
	}
	if strings.TrimSpace(cmd.SocialReason) == "" {
		return nil, fmt.Errorf("social reason is required")
	}

	rfc := strings.TrimSpace(strings.ToUpper(cmd.RFC))
	if rfc != "" && !rfcRegex.MatchString(rfc) {
		return nil, fmt.Errorf("formato de RFC inválido")
	}

	existing, err := s.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("company not found")
	}

	existing.KeyName = strings.TrimSpace(cmd.KeyName)
	existing.SocialReason = strings.TrimSpace(cmd.SocialReason)
	existing.RFC = rfc
	existing.Address = strings.TrimSpace(cmd.Address)
	existing.Emails = normalizeSlice(cmd.Emails)
	existing.Phones = normalizeSlice(cmd.Phones)
	existing.MainContactID = cmd.MainContactID
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid company id")
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) Get(ctx context.Context, id int) (*company.Company, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid company id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, filter company.ListFilter) ([]company.Company, int, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.List(ctx, filter)
}

func normalizeSlice(values []string) []string {
	var out []string
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
