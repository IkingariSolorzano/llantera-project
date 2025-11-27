package tire

import "context"

type BrandRepository interface {
	GetByName(ctx context.Context, nombre string) (*Brand, error)
	GetByAlias(ctx context.Context, alias string) (*Brand, error)
	Create(ctx context.Context, marca *Brand, aliases []string) error
	List(ctx context.Context) ([]Brand, error)
	GetByID(ctx context.Context, id int) (*Brand, error)
	ListAliases(ctx context.Context, brandID int) ([]string, error)
	Update(ctx context.Context, marca *Brand, aliases []string) error
	Delete(ctx context.Context, id int) error
	HasTires(ctx context.Context, id int) (bool, error)
}

type NormalizedTypeRepository interface {
	GetByName(ctx context.Context, nombre string) (*NormalizedType, error)
	Create(ctx context.Context, tipo *NormalizedType) error
	List(ctx context.Context) ([]NormalizedType, error)
}

type TireRepository interface {
	Create(ctx context.Context, tire *Tire) error
	Update(ctx context.Context, tire *Tire) error
	GetBySKU(ctx context.Context, sku string) (*Tire, error)
	List(ctx context.Context, filter TireFilter) ([]Tire, int, error)
	Delete(ctx context.Context, sku string) error
}

type EquivalenceRepository interface {
	Create(ctx context.Context, eq *Equivalence) error
	ListByBase(ctx context.Context, llantaBaseID string) ([]Equivalence, error)
}

type TireService interface {
	UpsertFromMeasurement(ctx context.Context, cmd UpsertCommand) (*Tire, error)
	List(ctx context.Context, filter TireFilter) ([]Tire, int, error)
	ListCatalog(ctx context.Context, filter TireFilter, level string) ([]CatalogItem, int, error)
	AdminList(ctx context.Context, filter TireFilter) ([]AdminTire, int, error)
	UpdateAdmin(ctx context.Context, sku string, cantidad *int, precios map[string]*float64) (*AdminTire, error)
	Get(ctx context.Context, sku string) (*Tire, error)
	Delete(ctx context.Context, sku string) error
	ExportAdmin(ctx context.Context, filter TireFilter) ([]byte, error)
	ImportFromXLSX(ctx context.Context, data []byte) (int, error)
}

// AdminTire representa la vista de administración de una llanta,
// incluyendo inventario y precios por columna.
type AdminTire struct {
	Tire      Tire               `json:"tire"`
	Inventory *Inventory         `json:"inventory,omitempty"`
	Prices    map[string]float64 `json:"prices,omitempty"`
	BrandName string             `json:"brandName,omitempty"`
}

type UpsertCommand struct {
	SKU               string
	MarcaNombre       string
	AliasMarca        string
	Modelo            string
	Ancho             int
	Perfil            *int
	Rin               float64
	Construccion      string
	TipoTubo          string
	CalificacionCapas string
	IndiceCarga       string
	IndiceVelocidad   string
	TipoNormalizado   string
	AbreviaturaUso    string
	Descripcion       string
	PrecioPublico     float64
	URLImagen         string
	MedidaOriginal    string
}

type TireFilter struct {
	Search            string
	MarcaID           *int
	TipoID            *int
	Abreviatura       string
	Ancho             *int
	Perfil            *int
	Rin               *float64
	Construccion      string
	CalificacionCapas string
	IndiceCarga       string
	IndiceVelocidad   string
	InStockOnly       bool
	Limit             int
	Offset            int
	Sort              string
}

type InventoryRepository interface {
	Upsert(ctx context.Context, inventory *Inventory) error
	GetByTireID(ctx context.Context, llantaID string) (*Inventory, error)
	// ReserveStock incrementa apartadas para una llanta (al crear pedido)
	ReserveStock(ctx context.Context, sku string, quantity int) error
	// ReleaseStock decrementa apartadas para una llanta (al cancelar pedido)
	ReleaseStock(ctx context.Context, sku string, quantity int) error
	// ConfirmSale resta de cantidad y apartadas (al confirmar venta/entrega)
	ConfirmSale(ctx context.Context, sku string, quantity int) error
}

type PriceColumnRepository interface {
	GetByID(ctx context.Context, id int) (*PriceColumn, error)
	GetByCode(ctx context.Context, code string) (*PriceColumn, error)
	List(ctx context.Context) ([]PriceColumn, error)
	Create(ctx context.Context, column *PriceColumn) error
	Update(ctx context.Context, column *PriceColumn) error
	Delete(ctx context.Context, id int) error
}

type PriceRepository interface {
	UpsertMany(ctx context.Context, prices []TirePrice) error
	ListByTireID(ctx context.Context, llantaID string) ([]TirePrice, error)
	ListByColumnID(ctx context.Context, columnID int) ([]TirePrice, error)
}

type PriceColumnService interface {
	ListColumns(ctx context.Context) ([]PriceColumn, error)
	GetColumn(ctx context.Context, id int) (*PriceColumn, error)
	CreateColumn(ctx context.Context, cmd PriceColumnCreateCommand) (*PriceColumn, error)
	UpdateColumn(ctx context.Context, id int, cmd PriceColumnUpdateCommand) (*PriceColumn, error)
	DeleteColumn(ctx context.Context, id int, transferToCode *string) error
}

type PriceColumnCreateCommand struct {
	Codigo      string
	Nombre      string
	Descripcion string
	OrdenVisual int
	Activo      bool
	EsPublico   bool
	Mode        string
	BaseCode    string
	Operation   string
	Amount      *float64
}

type PriceColumnUpdateCommand struct {
	Nombre      string
	Descripcion string
	OrdenVisual int
	Activo      bool
	EsPublico   bool
	Mode        string
	BaseCode    string
	Operation   string
	Amount      *float64
}
