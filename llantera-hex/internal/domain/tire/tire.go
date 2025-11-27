package tire

import (
	"errors"
	"time"
)

var (
	ErrTireNotFound        = errors.New("llanta no encontrada")
	ErrBrandNotFound       = errors.New("marca de llanta no encontrada")
	ErrPriceColumnNotFound = errors.New("columna de precio no encontrada")
)

type Brand struct {
	ID            int       `json:"id"`
	Nombre        string    `json:"nombre"`
	CreadoEn      time.Time `json:"creadoEn"`
	ActualizadoEn time.Time `json:"actualizadoEn"`
}

type BrandAlias struct {
	ID            int       `json:"id"`
	MarcaID       int       `json:"marcaId"`
	Alias         string    `json:"alias"`
	CreadoEn      time.Time `json:"creadoEn"`
	ActualizadoEn time.Time `json:"actualizadoEn"`
}

type NormalizedType struct {
	ID            int       `json:"id"`
	Nombre        string    `json:"nombre"`
	Descripcion   string    `json:"descripcion"`
	CreadoEn      time.Time `json:"creadoEn"`
	ActualizadoEn time.Time `json:"actualizadoEn"`
}

type Tire struct {
	ID                string    `json:"id"`
	SKU               string    `json:"sku"`
	MarcaID           int       `json:"marcaId"`
	Modelo            string    `json:"modelo"`
	Ancho             int       `json:"ancho"`
	Perfil            *int      `json:"perfil"`
	Rin               float64   `json:"rin"`
	Construccion      string    `json:"construccion"`
	TipoTubo          string    `json:"tipoTubo"`
	CalificacionCapas string    `json:"calificacionCapas"`
	IndiceCarga       string    `json:"indiceCarga"`
	IndiceVelocidad   string    `json:"indiceVelocidad"`
	TipoNormalizadoID *int      `json:"tipoNormalizadoId"`
	AbreviaturaUso    string    `json:"abreviaturaUso"`
	Descripcion       string    `json:"descripcion"`
	PrecioPublico     float64   `json:"precioPublico"`
	URLImagen         string    `json:"urlImagen"`
	MedidaOriginal    string    `json:"medidaOriginal"`
	CreadoEn          time.Time `json:"creadoEn"`
	ActualizadoEn     time.Time `json:"actualizadoEn"`
}

type Equivalence struct {
	ID                string    `json:"id"`
	LlantaBaseID      string    `json:"llantaBaseId"`
	LlantaEquivalente string    `json:"llantaEquivalente"`
	Notas             string    `json:"notas"`
	CreadoEn          time.Time `json:"creadoEn"`
	ActualizadoEn     time.Time `json:"actualizadoEn"`
}

type Inventory struct {
	ID            string    `json:"id"`
	LlantaID      string    `json:"llantaId"`
	Cantidad      int       `json:"cantidad"`
	Apartadas     int       `json:"apartadas"`
	StockMinimo   int       `json:"stockMinimo"`
	CreadoEn      time.Time `json:"creadoEn"`
	ActualizadoEn time.Time `json:"actualizadoEn"`
}

// Disponibles retorna la cantidad disponible (cantidad - apartadas)
func (i *Inventory) Disponibles() int {
	return i.Cantidad - i.Apartadas
}

type PriceColumn struct {
	ID            int       `json:"id"`
	Codigo        string    `json:"codigo"`
	Nombre        string    `json:"nombre"`
	Descripcion   string    `json:"descripcion"`
	OrdenVisual   int       `json:"ordenVisual"`
	Activo        bool      `json:"activo"`
	EsPublico     bool      `json:"esPublico"`
	Mode          string    `json:"mode,omitempty"`
	BaseCode      *string   `json:"baseCode,omitempty"`
	Operation     string    `json:"operation,omitempty"`
	Amount        *float64  `json:"amount,omitempty"`
	CreadoEn      time.Time `json:"creadoEn"`
	ActualizadoEn time.Time `json:"actualizadoEn"`
}

type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string {
	return e.msg
}

func NewValidationError(msg string) error {
	if msg == "" {
		msg = "datos de llanta inválidos"
	}
	return &ValidationError{msg: msg}
}

type TirePrice struct {
	LlantaID        string    `json:"llantaId"`
	ColumnaPrecioID int       `json:"columnaPrecioId"`
	Precio          float64   `json:"precio"`
	CreadoEn        time.Time `json:"creadoEn"`
	ActualizadoEn   time.Time `json:"actualizadoEn"`
}

type CatalogItem struct {
	Tire           Tire     `json:"tire"`
	Price          float64  `json:"price"`
	ReferencePrice *float64 `json:"referencePrice,omitempty"`
	PriceCode      string   `json:"priceCode"`
	ReferenceCode  *string  `json:"referenceCode,omitempty"`
	Stock          *int     `json:"stock,omitempty"`
}
