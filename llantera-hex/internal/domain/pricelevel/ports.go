package pricelevel

import (
	"time"
)

type PriceLevel struct {
	ID                 int       `json:"id"`
	Code               string    `json:"code"`
	Name               string    `json:"name"`
	Description        *string   `json:"description"`
	DiscountPercentage float64   `json:"discountPercentage"`
	PriceColumn        string    `json:"priceColumn"`
	ReferenceColumn    *string   `json:"referenceColumn"`
	CanViewOffers      bool      `json:"canViewOffers"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type PriceLevelFilter struct {
	Code   *string `json:"code"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

type PriceLevelRepository interface {
	Create(level *PriceLevel) (*PriceLevel, error)
	GetByID(id int) (*PriceLevel, error)
	GetByCode(code string) (*PriceLevel, error)
	List(filter PriceLevelFilter) ([]*PriceLevel, int, error)
	Update(id int, level *PriceLevel) (*PriceLevel, error)
	Delete(id int) error
	GetUsersCount(id int) (int, error)
	TransferUsers(fromID, toID int) error
}

type PriceLevelService interface {
	Create(level *PriceLevel) (*PriceLevel, error)
	GetByID(id int) (*PriceLevel, error)
	GetByCode(code string) (*PriceLevel, error)
	List(filter PriceLevelFilter) ([]*PriceLevel, int, error)
	Update(id int, level *PriceLevel) (*PriceLevel, error)
	Delete(id int, transferToID *int) error
}
