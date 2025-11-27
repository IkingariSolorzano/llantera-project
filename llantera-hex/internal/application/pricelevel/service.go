package pricelevel

import (
	"errors"
	"fmt"

	"github.com/llantera/hex/internal/domain/pricelevel"
)

type priceLevelService struct {
	repo pricelevel.PriceLevelRepository
}

func NewPriceLevelService(repo pricelevel.PriceLevelRepository) pricelevel.PriceLevelService {
	return &priceLevelService{repo: repo}
}

func (s *priceLevelService) Create(level *pricelevel.PriceLevel) (*pricelevel.PriceLevel, error) {
	if level.Code == "" {
		return nil, errors.New("código es requerido")
	}
	if level.Name == "" {
		return nil, errors.New("nombre es requerido")
	}
	if level.PriceColumn == "" {
		return nil, errors.New("columna de precio es requerida")
	}

	// Validar que no exista otro con el mismo código
	existing, err := s.repo.GetByCode(level.Code)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("ya existe un nivel de precio con código: %s", level.Code)
	}

	return s.repo.Create(level)
}

func (s *priceLevelService) GetByID(id int) (*pricelevel.PriceLevel, error) {
	if id <= 0 {
		return nil, errors.New("ID inválido")
	}
	return s.repo.GetByID(id)
}

func (s *priceLevelService) GetByCode(code string) (*pricelevel.PriceLevel, error) {
	if code == "" {
		return nil, errors.New("código es requerido")
	}
	return s.repo.GetByCode(code)
}

func (s *priceLevelService) List(filter pricelevel.PriceLevelFilter) ([]*pricelevel.PriceLevel, int, error) {
	return s.repo.List(filter)
}

func (s *priceLevelService) Update(id int, level *pricelevel.PriceLevel) (*pricelevel.PriceLevel, error) {
	if id <= 0 {
		return nil, errors.New("ID inválido")
	}
	if level.Code == "" {
		return nil, errors.New("código es requerido")
	}
	if level.Name == "" {
		return nil, errors.New("nombre es requerido")
	}
	if level.PriceColumn == "" {
		return nil, errors.New("columna de precio es requerida")
	}

	// Verificar que existe
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Si cambió el código, validar que no exista otro con el mismo código
	if level.Code != existing.Code {
		other, err := s.repo.GetByCode(level.Code)
		if err == nil && other != nil && other.ID != id {
			return nil, fmt.Errorf("ya existe otro nivel de precio con código: %s", level.Code)
		}
	}

	return s.repo.Update(id, level)
}

func (s *priceLevelService) Delete(id int, transferToID *int) error {
	if id <= 0 {
		return errors.New("ID inválido")
	}

	// Verificar que existe
	level, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Contar usuarios asignados
	userCount, err := s.repo.GetUsersCount(id)
	if err != nil {
		return err
	}

	if userCount > 0 {
		if transferToID == nil {
			return fmt.Errorf("no se puede eliminar el nivel '%s' porque tiene %d usuarios asignados. Debe especificar un nivel de destino para transferirlos", level.Name, userCount)
		}

		// Verificar que el destino existe
		_, err = s.repo.GetByID(*transferToID)
		if err != nil {
			return fmt.Errorf("nivel de destino no encontrado: %d", *transferToID)
		}

		// Transferir usuarios
		err = s.repo.TransferUsers(id, *transferToID)
		if err != nil {
			return fmt.Errorf("error al transferir usuarios: %v", err)
		}
	}

	// Eliminar el nivel
	return s.repo.Delete(id)
}
