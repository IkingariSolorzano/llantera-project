package tireapp

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/llantera/hex/internal/domain/pricelevel"
	"github.com/llantera/hex/internal/domain/tire"
)

// Service reúne los casos de uso para la entidad llanta.
type Service struct {
	tiRes        tire.TireRepository
	brands       tire.BrandRepository
	types        tire.NormalizedTypeRepository
	inventory    tire.InventoryRepository
	prices       tire.PriceRepository
	priceColumns tire.PriceColumnRepository
	priceLevels  pricelevel.PriceLevelRepository
	nowFunc      func() time.Time
}

func NewService(
	tireRepo tire.TireRepository,
	brandRepo tire.BrandRepository,
	typeRepo tire.NormalizedTypeRepository,
	inventoryRepo tire.InventoryRepository,
	priceRepo tire.PriceRepository,
	priceColumnRepo tire.PriceColumnRepository,
	priceLevelRepo pricelevel.PriceLevelRepository,
) *Service {
	return &Service{
		tiRes:        tireRepo,
		brands:       brandRepo,
		types:        typeRepo,
		inventory:    inventoryRepo,
		prices:       priceRepo,
		priceColumns: priceColumnRepo,
		priceLevels:  priceLevelRepo,
		nowFunc:      func() time.Time { return time.Now().UTC() },
	}
}

var priceColumnCodeRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

var _ tire.TireService = (*Service)(nil)
var _ tire.PriceColumnService = (*Service)(nil)

func (s *Service) UpsertFromMeasurement(ctx context.Context, cmd tire.UpsertCommand) (*tire.Tire, error) {
	sku := strings.TrimSpace(cmd.SKU)
	if sku == "" {
		return nil, fmt.Errorf("el SKU es obligatorio")
	}

	brand, err := s.resolveBrand(ctx, cmd.MarcaNombre, cmd.AliasMarca)
	if err != nil {
		return nil, err
	}

	normalizedTypeID, err := s.resolveType(ctx, cmd.TipoNormalizado)
	if err != nil {
		return nil, err
	}

	now := s.nowFunc()
	entity, err := s.tiRes.GetBySKU(ctx, sku)
	if err != nil && !errors.Is(err, tire.ErrTireNotFound) {
		return nil, err
	}

	isNew := entity == nil
	if isNew {
		entity = &tire.Tire{
			ID:       uuid.NewString(),
			SKU:      sku,
			CreadoEn: now,
		}
	}

	entity.MarcaID = brand.ID
	entity.Modelo = strings.TrimSpace(cmd.Modelo)
	entity.Ancho = cmd.Ancho
	entity.Perfil = cmd.Perfil
	entity.Rin = cmd.Rin
	entity.Construccion = strings.TrimSpace(strings.ToUpper(cmd.Construccion))
	entity.TipoTubo = strings.TrimSpace(strings.ToUpper(cmd.TipoTubo))
	entity.CalificacionCapas = strings.TrimSpace(cmd.CalificacionCapas)
	entity.IndiceCarga = strings.TrimSpace(cmd.IndiceCarga)
	entity.IndiceVelocidad = strings.TrimSpace(cmd.IndiceVelocidad)
	entity.TipoNormalizadoID = normalizedTypeID
	if entity.TipoNormalizadoID != nil && *entity.TipoNormalizadoID == 0 {
		entity.TipoNormalizadoID = nil
	}
	entity.AbreviaturaUso = strings.TrimSpace(strings.ToUpper(cmd.AbreviaturaUso))
	entity.Descripcion = strings.TrimSpace(cmd.Descripcion)
	entity.PrecioPublico = cmd.PrecioPublico
	entity.URLImagen = strings.TrimSpace(cmd.URLImagen)
	entity.MedidaOriginal = strings.TrimSpace(cmd.MedidaOriginal)
	entity.ActualizadoEn = now

	if entity.CreadoEn.IsZero() {
		entity.CreadoEn = now
	}

	if isNew {
		if err := s.tiRes.Create(ctx, entity); err != nil {
			return nil, err
		}
	} else {
		if err := s.tiRes.Update(ctx, entity); err != nil {
			return nil, err
		}
	}

	return entity, nil
}

func (s *Service) List(ctx context.Context, filter tire.TireFilter) ([]tire.Tire, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.tiRes.List(ctx, filter)
}

// ExportAdmin genera un archivo XLSX con el catálogo de llantas en vista de administración.
// Incluye datos técnicos, inventario, precio público y una columna por cada código de precio.
func (s *Service) ExportAdmin(ctx context.Context, filter tire.TireFilter) ([]byte, error) {
	// Cargar nombres de marca para mapear marcaId -> nombre.
	brandNames := map[int]string{}
	if s.brands != nil {
		brands, err := s.brands.List(ctx)
		if err != nil {
			return nil, err
		}
		for _, b := range brands {
			brandNames[b.ID] = b.Nombre
		}
	}

	// Cargar columnas de precio para definir encabezados dinámicos.
	var priceCols []struct {
		Code  string
		Order int
	}
	if s.priceColumns != nil {
		cols, err := s.priceColumns.List(ctx)
		if err != nil {
			return nil, err
		}
		for _, c := range cols {
			code := strings.ToLower(strings.TrimSpace(c.Codigo))
			if code == "" {
				continue
			}
			priceCols = append(priceCols, struct {
				Code  string
				Order int
			}{
				Code:  code,
				Order: c.OrdenVisual,
			})
		}
	}

	sort.Slice(priceCols, func(i, j int) bool {
		if priceCols[i].Order == priceCols[j].Order {
			return priceCols[i].Code < priceCols[j].Code
		}
		return priceCols[i].Order < priceCols[j].Order
	})

	// Crear archivo XLSX y hoja principal.
	f := excelize.NewFile()
	sheetName := "Catalogo"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{
		"sku",
		"marca",
		"modelo",
		"ancho",
		"perfil",
		"construccion",
		"rin",
		"tipo_tubo",
		"calificacion_capas",
		"indice_carga",
		"indice_velocidad",
		"uso",
		"cantidad",
		"stock_minimo",
		"precio_publico",
	}
	for _, pc := range priceCols {
		headers = append(headers, pc.Code)
	}
	headers = append(headers,
		"descripcion",
		"url_imagen",
	)

	for i, h := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheetName, cell, h); err != nil {
			return nil, err
		}
	}

	// Paginación sobre AdminList para exportar todas las llantas.
	rowIndex := 2
	const pageSize = 500
	offset := 0
	for {
		pageFilter := filter
		pageFilter.Limit = pageSize
		pageFilter.Offset = offset

		items, total, err := s.AdminList(ctx, pageFilter)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}

		for _, it := range items {
			t := it.Tire
			inv := it.Inventory
			brandName := brandNames[t.MarcaID]

			values := make([]interface{}, 0, len(headers))
			values = append(values,
				t.SKU,
				brandName,
				t.Modelo,
				t.Ancho,
			)

			if t.Perfil != nil {
				values = append(values, *t.Perfil)
			} else {
				values = append(values, "")
			}

			values = append(values,
				t.Construccion,
				t.Rin,
				t.TipoTubo,
				t.CalificacionCapas,
				t.IndiceCarga,
				t.IndiceVelocidad,
				t.AbreviaturaUso,
			)

			cantidad := 0
			stockMin := 0
			if inv != nil {
				cantidad = inv.Cantidad
				stockMin = inv.StockMinimo
			}
			values = append(values, cantidad, stockMin, t.PrecioPublico)

			for _, pc := range priceCols {
				if it.Prices != nil {
					if v, ok := it.Prices[pc.Code]; ok {
						values = append(values, v)
						continue
					}
				}
				values = append(values, "")
			}

			values = append(values, t.Descripcion, t.URLImagen)

			for colIdx, v := range values {
				cell, err := excelize.CoordinatesToCellName(colIdx+1, rowIndex)
				if err != nil {
					return nil, err
				}
				if err := f.SetCellValue(sheetName, cell, v); err != nil {
					return nil, err
				}
			}

			rowIndex++
		}

		offset += len(items)
		if offset >= total {
			break
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// AdminList devuelve la información de llantas junto con inventario y precios por columna.
func (s *Service) AdminList(ctx context.Context, filter tire.TireFilter) ([]tire.AdminTire, int, error) {
	items, total, err := s.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	if len(items) == 0 {
		return []tire.AdminTire{}, total, nil
	}

	codeByID := map[int]string{}
	if s.priceColumns != nil {
		cols, err := s.priceColumns.List(ctx)
		if err != nil {
			return nil, 0, err
		}
		for _, c := range cols {
			code := strings.ToLower(strings.TrimSpace(c.Codigo))
			if code == "" {
				continue
			}
			codeByID[c.ID] = code
		}
	}

	result := make([]tire.AdminTire, 0, len(items))

	for _, t := range items {
		var inv *tire.Inventory
		if s.inventory != nil {
			entity, err := s.inventory.GetByTireID(ctx, t.ID)
			if err != nil {
				return nil, 0, err
			}
			inv = entity
		}

		pricesMap := map[string]float64{}
		if s.prices != nil && len(codeByID) > 0 {
			rows, err := s.prices.ListByTireID(ctx, t.ID)
			if err != nil {
				return nil, 0, err
			}
			for _, p := range rows {
				code, ok := codeByID[p.ColumnaPrecioID]
				if !ok {
					continue
				}
				pricesMap[code] = p.Precio
			}
		}

		result = append(result, tire.AdminTire{
			Tire:      t,
			Inventory: inv,
			Prices:    pricesMap,
		})
	}

	return result, total, nil
}

// UpdateAdmin actualiza inventario y precios y sincroniza precio público con lista.
// Es la versión pública utilizada por los endpoints HTTP habituales.
func (s *Service) UpdateAdmin(ctx context.Context, sku string, cantidad *int, precios map[string]*float64) (*tire.AdminTire, error) {
	return s.updateAdminInternal(ctx, sku, cantidad, precios, true)
}

// updateAdminInternal concentra la lógica de actualización de inventario y precios.
// El flag recalcDerived permite desactivar el recálculo de columnas derivadas cuando
// se realizan operaciones masivas (por ejemplo, importaciones) y se desea recalcular
// dichas columnas sólo una vez al final del proceso.
func (s *Service) updateAdminInternal(ctx context.Context, sku string, cantidad *int, precios map[string]*float64, recalcDerived bool) (*tire.AdminTire, error) {
	sku = strings.TrimSpace(sku)
	if sku == "" {
		return nil, fmt.Errorf("el SKU es obligatorio")
	}

	t, err := s.tiRes.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}

	now := s.nowFunc()

	// Inventario
	if cantidad != nil && s.inventory != nil {
		inv, err := s.inventory.GetByTireID(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		if inv == nil {
			inv = &tire.Inventory{
				ID:          uuid.NewString(),
				LlantaID:    t.ID,
				StockMinimo: 0,
				CreadoEn:    now,
			}
		}
		inv.Cantidad = *cantidad
		inv.ActualizadoEn = now
		if err := s.inventory.Upsert(ctx, inv); err != nil {
			return nil, err
		}
	}

	// Precios
	if precios != nil && s.prices != nil && s.priceColumns != nil {
		cols, err := s.priceColumns.List(ctx)
		if err != nil {
			return nil, err
		}
		codeToID := make(map[string]int, len(cols))
		for _, c := range cols {
			code := strings.ToLower(strings.TrimSpace(c.Codigo))
			if code == "" {
				continue
			}
			codeToID[code] = c.ID
		}

		var toUpsert []tire.TirePrice
		var updatedListPrice *float64
		changedCodes := make(map[string]struct{})
		for code, value := range precios {
			clean := strings.ToLower(strings.TrimSpace(code))
			if clean == "" {
				continue
			}
			columnID, ok := codeToID[clean]
			if !ok {
				continue
			}
			if value == nil {
				continue
			}
			toUpsert = append(toUpsert, tire.TirePrice{
				LlantaID:        t.ID,
				ColumnaPrecioID: columnID,
				Precio:          *value,
				CreadoEn:        now,
				ActualizadoEn:   now,
			})
			changedCodes[clean] = struct{}{}
			if clean == "lista" {
				v := *value
				updatedListPrice = &v
			}
		}

		if len(toUpsert) > 0 {
			if err := s.prices.UpsertMany(ctx, toUpsert); err != nil {
				return nil, err
			}
		}

		if updatedListPrice != nil {
			t.PrecioPublico = *updatedListPrice
			t.ActualizadoEn = now
			if err := s.tiRes.Update(ctx, t); err != nil {
				return nil, err
			}
		}

		// Recalcular columnas derivadas afectadas por los códigos de precio modificados.
		if recalcDerived && len(changedCodes) > 0 {
			cols, err := s.priceColumns.List(ctx)
			if err != nil {
				return nil, err
			}
			for i := range cols {
				col := &cols[i]
				mode := strings.ToLower(strings.TrimSpace(col.Mode))
				if mode != "derived" || col.BaseCode == nil {
					continue
				}
				baseCode := strings.ToLower(strings.TrimSpace(*col.BaseCode))
				if baseCode == "" {
					continue
				}
				if _, ok := changedCodes[baseCode]; !ok {
					continue
				}
				if err := s.recalculateDerivedColumn(ctx, col); err != nil {
					return nil, err
				}
			}
		}
	}

	// Reconstruir vista admin para devolver datos actualizados
	filter := tire.TireFilter{Search: sku, Limit: 1, Offset: 0}
	items, _, err := s.AdminList(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, tire.ErrTireNotFound
	}
	return &items[0], nil
}

func (s *Service) ListCatalog(ctx context.Context, filter tire.TireFilter, level string) ([]tire.CatalogItem, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	items, total, err := s.tiRes.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	if len(items) == 0 {
		return []tire.CatalogItem{}, total, nil
	}

	// Resolver columnas de precio según el nivel indicado.
	mainCode := "lista"
	refCode := ""
	levelKey := strings.ToLower(strings.TrimSpace(level))

	if s.priceLevels != nil && levelKey != "" {
		lvl, err := s.priceLevels.GetByCode(levelKey)
		if err == nil && lvl != nil {
			if c := strings.TrimSpace(lvl.PriceColumn); c != "" {
				mainCode = strings.ToLower(c)
			}
			if lvl.ReferenceColumn != nil {
				if rc := strings.TrimSpace(*lvl.ReferenceColumn); rc != "" {
					refCode = strings.ToLower(rc)
				}
			}
		} else {
			// Fallback a mapeo estático si el nivel no existe en catálogo.
			mainCode, refCode = mapLevelToPriceColumns(level)
		}
	} else {
		// Sin repositorio de niveles o sin level: usar mapeo estático existente.
		mainCode, refCode = mapLevelToPriceColumns(level)
	}
	mainCodeLower := strings.ToLower(mainCode)
	refCodeLower := strings.ToLower(refCode)

	priceColumns, err := s.priceColumns.List(ctx)
	if err != nil {
		return nil, 0, err
	}
	codeToID := make(map[string]int, len(priceColumns))
	for _, c := range priceColumns {
		code := strings.ToLower(strings.TrimSpace(c.Codigo))
		if code == "" {
			continue
		}
		codeToID[code] = c.ID
	}

	mainID, okMain := codeToID[mainCodeLower]
	refID, okRef := 0, false
	if refCodeLower != "" {
		if id, ok := codeToID[refCodeLower]; ok {
			refID, okRef = id, true
		}
	}

	result := make([]tire.CatalogItem, 0, len(items))

	for _, t := range items {
		var mainPrice *float64
		var refPrice *float64

		if s.prices != nil && (okMain || okRef) {
			prices, err := s.prices.ListByTireID(ctx, t.ID)
			if err != nil {
				return nil, 0, err
			}
			for _, p := range prices {
				if okMain && p.ColumnaPrecioID == mainID {
					v := p.Precio
					mainPrice = &v
				}
				if okRef && p.ColumnaPrecioID == refID {
					v := p.Precio
					refPrice = &v
				}
			}
		}

		if mainPrice == nil {
			if t.PrecioPublico > 0 {
				v := t.PrecioPublico
				mainPrice = &v
			} else {
				v := 0.0
				mainPrice = &v
			}
		}

		item := tire.CatalogItem{
			Tire:      t,
			Price:     *mainPrice,
			PriceCode: mainCode,
		}
		if refPrice != nil && refCode != "" {
			item.ReferencePrice = refPrice
			code := refCode
			item.ReferenceCode = &code
		}
		// Obtener stock del inventario
		if s.inventory != nil {
			inv, err := s.inventory.GetByTireID(ctx, t.ID)
			if err == nil && inv != nil {
				stock := inv.Cantidad
				item.Stock = &stock
			}
		}
		result = append(result, item)
	}

	return result, total, nil
}

func mapLevelToPriceColumns(level string) (string, string) {
	key := strings.ToLower(strings.TrimSpace(level))
	switch key {
	case "empresa":
		return "empresa", "lista"
	case "distribuidor":
		return "mayoreo", "lista"
	case "mayorista":
		return "mayoreo_6", "lista"
	case "public", "":
		fallthrough
	default:
		return "lista", ""
	}
}

func (s *Service) Get(ctx context.Context, sku string) (*tire.Tire, error) {
	if strings.TrimSpace(sku) == "" {
		return nil, fmt.Errorf("el SKU es obligatorio")
	}
	return s.tiRes.GetBySKU(ctx, sku)
}

func (s *Service) Delete(ctx context.Context, sku string) error {
	if strings.TrimSpace(sku) == "" {
		return fmt.Errorf("el SKU es obligatorio")
	}
	return s.tiRes.Delete(ctx, sku)
}

// ListColumns devuelve todas las columnas de precio configuradas.
func (s *Service) ListColumns(ctx context.Context) ([]tire.PriceColumn, error) {
	if s.priceColumns == nil {
		return []tire.PriceColumn{}, nil
	}
	return s.priceColumns.List(ctx)
}

// GetColumn obtiene una columna de precio por su identificador numérico.
func (s *Service) GetColumn(ctx context.Context, id int) (*tire.PriceColumn, error) {
	if id <= 0 {
		return nil, tire.NewValidationError("el identificador es obligatorio")
	}
	if s.priceColumns == nil {
		return nil, fmt.Errorf("repositorio de columnas de precio no configurado")
	}
	col, err := s.priceColumns.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, tire.ErrPriceColumnNotFound
	}
	return col, nil
}

// CreateColumn registra una nueva columna de precio.
func (s *Service) CreateColumn(ctx context.Context, cmd tire.PriceColumnCreateCommand) (*tire.PriceColumn, error) {
	if s.priceColumns == nil {
		return nil, fmt.Errorf("repositorio de columnas de precio no configurado")
	}

	code := strings.TrimSpace(strings.ToLower(cmd.Codigo))
	if code == "" {
		return nil, tire.NewValidationError("el código es obligatorio")
	}
	if !priceColumnCodeRegexp.MatchString(code) {
		return nil, tire.NewValidationError("el código solo puede contener letras, números y guiones bajos, sin espacios")
	}
	name := strings.TrimSpace(cmd.Nombre)
	if name == "" {
		return nil, tire.NewValidationError("el nombre es obligatorio")
	}
	if cmd.OrdenVisual < 0 {
		return nil, tire.NewValidationError("el orden visual no puede ser negativo")
	}

	existing, err := s.priceColumns.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, tire.NewValidationError("ya existe una columna de precio con ese código")
	}

	// Configuración de cálculo
	mode := strings.ToLower(strings.TrimSpace(cmd.Mode))
	if mode == "" {
		mode = "fixed"
	}
	if mode != "fixed" && mode != "derived" {
		return nil, tire.NewValidationError("modo de cálculo inválido")
	}

	var baseCodePtr *string
	var amountPtr *float64
	operation := strings.ToLower(strings.TrimSpace(cmd.Operation))
	if mode == "derived" {
		baseCode := strings.TrimSpace(strings.ToLower(cmd.BaseCode))
		if baseCode == "" {
			return nil, tire.NewValidationError("la columna base es obligatoria para modo derivado")
		}
		if !priceColumnCodeRegexp.MatchString(baseCode) {
			return nil, tire.NewValidationError("la columna base solo puede contener letras, números y guiones bajos, sin espacios")
		}
		if operation == "" {
			operation = "percent"
		}
		switch operation {
		case "add", "subtract", "multiply", "percent":
		default:
			return nil, tire.NewValidationError("la operación de cálculo no es válida")
		}
		if cmd.Amount == nil {
			return nil, tire.NewValidationError("la cantidad de cálculo es obligatoria para modo derivado")
		}
		amount := *cmd.Amount
		amountPtr = &amount
		baseCodePtr = &baseCode

		// Verificar que la columna base exista
		baseCol, err := s.priceColumns.GetByCode(ctx, baseCode)
		if err != nil {
			return nil, err
		}
		if baseCol == nil {
			return nil, tire.NewValidationError("la columna base especificada no existe")
		}
	}

	entity := &tire.PriceColumn{
		Codigo:      code,
		Nombre:      name,
		Descripcion: strings.TrimSpace(cmd.Descripcion),
		OrdenVisual: cmd.OrdenVisual,
		Activo:      cmd.Activo,
		EsPublico:   cmd.EsPublico,
		Mode:        mode,
		BaseCode:    baseCodePtr,
		Operation:   operation,
		Amount:      amountPtr,
	}

	if err := s.priceColumns.Create(ctx, entity); err != nil {
		return nil, err
	}

	// Inicializar precios para la nueva columna.
	if s.prices != nil {
		modeLower := strings.ToLower(strings.TrimSpace(entity.Mode))
		if modeLower == "derived" {
			if err := s.recalculateDerivedColumn(ctx, entity); err != nil {
				return nil, err
			}
		} else if s.tiRes != nil {
			// Columnas fijas: inicializar en 0 para todas las llantas existentes.
			const pageSize = 200
			offset := 0
			for {
				filter := tire.TireFilter{Limit: pageSize, Offset: offset}
				items, total, err := s.tiRes.List(ctx, filter)
				if err != nil {
					return nil, err
				}
				if len(items) == 0 {
					break
				}

				prices := make([]tire.TirePrice, 0, len(items))
				now := s.nowFunc()
				for _, t := range items {
					prices = append(prices, tire.TirePrice{
						LlantaID:        t.ID,
						ColumnaPrecioID: entity.ID,
						Precio:          0,
						CreadoEn:        now,
						ActualizadoEn:   now,
					})
				}

				if err := s.prices.UpsertMany(ctx, prices); err != nil {
					return nil, err
				}

				offset += len(items)
				if offset >= total {
					break
				}
			}
		}
	}

	return entity, nil
}

// UpdateColumn actualiza una columna de precio existente.
func (s *Service) UpdateColumn(ctx context.Context, id int, cmd tire.PriceColumnUpdateCommand) (*tire.PriceColumn, error) {
	if id <= 0 {
		return nil, tire.NewValidationError("el identificador es obligatorio")
	}
	if s.priceColumns == nil {
		return nil, fmt.Errorf("repositorio de columnas de precio no configurado")
	}

	col, err := s.priceColumns.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, tire.ErrPriceColumnNotFound
	}

	name := strings.TrimSpace(cmd.Nombre)
	if name == "" {
		return nil, tire.NewValidationError("el nombre es obligatorio")
	}
	if cmd.OrdenVisual < 0 {
		return nil, tire.NewValidationError("el orden visual no puede ser negativo")
	}

	// Configuración de cálculo
	mode := strings.ToLower(strings.TrimSpace(cmd.Mode))
	if mode == "" {
		mode = "fixed"
	}
	if mode != "fixed" && mode != "derived" {
		return nil, tire.NewValidationError("modo de cálculo inválido")
	}

	var baseCodePtr *string
	var amountPtr *float64
	operation := strings.ToLower(strings.TrimSpace(cmd.Operation))
	if mode == "derived" {
		baseCode := strings.TrimSpace(strings.ToLower(cmd.BaseCode))
		if baseCode == "" {
			return nil, tire.NewValidationError("la columna base es obligatoria para modo derivado")
		}
		if !priceColumnCodeRegexp.MatchString(baseCode) {
			return nil, tire.NewValidationError("la columna base solo puede contener letras, números y guiones bajos, sin espacios")
		}
		if operation == "" {
			operation = "percent"
		}
		switch operation {
		case "add", "subtract", "multiply", "percent":
		default:
			return nil, tire.NewValidationError("la operación de cálculo no es válida")
		}
		if cmd.Amount == nil {
			return nil, tire.NewValidationError("la cantidad de cálculo es obligatoria para modo derivado")
		}
		amount := *cmd.Amount
		amountPtr = &amount
		baseCodePtr = &baseCode

		// Verificar que la columna base exista
		baseCol, err := s.priceColumns.GetByCode(ctx, baseCode)
		if err != nil {
			return nil, err
		}
		if baseCol == nil {
			return nil, tire.NewValidationError("la columna base especificada no existe")
		}
	}

	col.Nombre = name
	col.Descripcion = strings.TrimSpace(cmd.Descripcion)
	col.OrdenVisual = cmd.OrdenVisual
	col.Activo = cmd.Activo
	col.EsPublico = cmd.EsPublico
	col.Mode = mode
	col.BaseCode = baseCodePtr
	col.Operation = operation
	col.Amount = amountPtr

	if err := s.priceColumns.Update(ctx, col); err != nil {
		return nil, err
	}

	// Si la columna es derivada, recalcular sus precios para todas las llantas.
	if strings.ToLower(strings.TrimSpace(col.Mode)) == "derived" && s.prices != nil {
		if err := s.recalculateDerivedColumn(ctx, col); err != nil {
			return nil, err
		}
	}

	return col, nil
}

// DeleteColumn elimina una columna de precio aplicando reglas de negocio adicionales.
// Si existen niveles de precio que usan esta columna (como principal o de referencia),
// se requiere indicar a qué columna deben ser reasignados mediante transferToCode.
func (s *Service) DeleteColumn(ctx context.Context, id int, transferToCode *string) error {
	if id <= 0 {
		return tire.NewValidationError("el identificador es obligatorio")
	}
	if s.priceColumns == nil {
		return fmt.Errorf("repositorio de columnas de precio no configurado")
	}

	// Obtener la columna para validar reglas de negocio.
	col, err := s.priceColumns.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if col == nil {
		return tire.ErrPriceColumnNotFound
	}

	codeLower := strings.ToLower(strings.TrimSpace(col.Codigo))
	// La columna de lista no debe ser eliminable.
	if codeLower == "lista" {
		return tire.NewValidationError("la columna de lista no se puede eliminar")
	}

	// No permitir eliminar una columna que sea base de otras columnas derivadas.
	cols, err := s.priceColumns.List(ctx)
	if err != nil {
		return err
	}
	for _, c := range cols {
		if c.ID == col.ID {
			continue
		}
		mode := strings.ToLower(strings.TrimSpace(c.Mode))
		if mode != "derived" || c.BaseCode == nil {
			continue
		}
		baseCodeLower := strings.ToLower(strings.TrimSpace(*c.BaseCode))
		if baseCodeLower == codeLower {
			return tire.NewValidationError("no se puede eliminar la columna porque es columna base de otras columnas derivadas")
		}
	}

	// Reasignar niveles de precio que usan esta columna, si el repositorio de niveles está disponible.
	if s.priceLevels != nil {
		levelsFilter := pricelevel.PriceLevelFilter{Limit: 500, Offset: 0}
		levels, _, err := s.priceLevels.List(levelsFilter)
		if err != nil {
			return err
		}

		var affected []*pricelevel.PriceLevel
		for _, lvl := range levels {
			if lvl == nil {
				continue
			}
			mainCodeLower := strings.ToLower(strings.TrimSpace(lvl.PriceColumn))
			refMatches := false
			if lvl.ReferenceColumn != nil {
				refLower := strings.ToLower(strings.TrimSpace(*lvl.ReferenceColumn))
				refMatches = refLower == codeLower
			}
			if mainCodeLower == codeLower || refMatches {
				affected = append(affected, lvl)
			}
		}

		if len(affected) > 0 {
			// Se requiere columna destino para reasignar.
			if transferToCode == nil || strings.TrimSpace(*transferToCode) == "" {
				return tire.NewValidationError("no se puede eliminar la columna porque existen niveles de precio que la utilizan; debes indicar una columna de destino")
			}

			destCode := strings.ToLower(strings.TrimSpace(*transferToCode))
			if destCode == codeLower {
				return tire.NewValidationError("la columna de destino debe ser diferente de la que se desea eliminar")
			}

			// Verificar que la columna destino exista.
			destCol, err := s.priceColumns.GetByCode(ctx, destCode)
			if err != nil || destCol == nil {
				return tire.NewValidationError("la columna de destino especificada no existe")
			}

			// Actualizar niveles para apuntar a la nueva columna.
			for _, lvl := range affected {
				if strings.EqualFold(strings.TrimSpace(lvl.PriceColumn), codeLower) {
					lvl.PriceColumn = destCode
				}
				if lvl.ReferenceColumn != nil && strings.EqualFold(strings.TrimSpace(*lvl.ReferenceColumn), codeLower) {
					newRef := destCode
					lvl.ReferenceColumn = &newRef
				}
				if _, err := s.priceLevels.Update(lvl.ID, lvl); err != nil {
					return err
				}
			}
		}
	}

	return s.priceColumns.Delete(ctx, id)
}

func (s *Service) resolveBrand(ctx context.Context, nombre, alias string) (*tire.Brand, error) {
	aliasClean := strings.TrimSpace(strings.ToUpper(alias))
	if aliasClean != "" {
		brand, err := s.brands.GetByAlias(ctx, aliasClean)
		if err == nil && brand != nil {
			return brand, nil
		}
		if err != nil && !errors.Is(err, tire.ErrBrandNotFound) {
			return nil, err
		}
	}

	cleanedName := strings.TrimSpace(nombre)
	if cleanedName == "" {
		if aliasClean != "" {
			cleanedName = aliasClean
		} else {
			cleanedName = "Otras Marcas"
		}
	}

	brand, err := s.brands.GetByName(ctx, cleanedName)
	if err == nil && brand != nil {
		return brand, nil
	}
	if err != nil && !errors.Is(err, tire.ErrBrandNotFound) {
		return nil, err
	}

	entity := &tire.Brand{Nombre: cleanedName}
	aliases := []string{}
	if aliasClean != "" {
		aliases = append(aliases, aliasClean)
	}
	aliases = append(aliases, cleanedName)
	if err := s.brands.Create(ctx, entity, aliases); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *Service) resolveType(ctx context.Context, nombre string) (*int, error) {
	cleaned := strings.TrimSpace(nombre)
	if cleaned == "" {
		return nil, nil
	}

	existing, err := s.types.GetByName(ctx, cleaned)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return &existing.ID, nil
	}

	entity := &tire.NormalizedType{Nombre: cleaned}
	if err := s.types.Create(ctx, entity); err != nil {
		return nil, err
	}
	return &entity.ID, nil
}

// recalculateDerivedColumn recalcula los precios de una columna derivada a partir de su columna base.
func (s *Service) recalculateDerivedColumn(ctx context.Context, column *tire.PriceColumn) error {
	if column == nil {
		return nil
	}
	if s.prices == nil || s.priceColumns == nil {
		return nil
	}
	mode := strings.ToLower(strings.TrimSpace(column.Mode))
	if mode != "derived" {
		return nil
	}
	if column.BaseCode == nil || column.Amount == nil {
		return tire.NewValidationError("configuración de cálculo incompleta para columna derivada")
	}

	baseCode := strings.TrimSpace(strings.ToLower(*column.BaseCode))
	if baseCode == "" {
		return tire.NewValidationError("la columna base es obligatoria para modo derivado")
	}
	baseCol, err := s.priceColumns.GetByCode(ctx, baseCode)
	if err != nil {
		return err
	}
	if baseCol == nil {
		return tire.NewValidationError("la columna base especificada no existe")
	}

	basePrices, err := s.prices.ListByColumnID(ctx, baseCol.ID)
	if err != nil {
		return err
	}
	if len(basePrices) == 0 {
		return nil
	}

	op := strings.ToLower(strings.TrimSpace(column.Operation))
	amount := *column.Amount
	now := s.nowFunc()
	derived := make([]tire.TirePrice, 0, len(basePrices))
	for _, bp := range basePrices {
		price := applyPriceCalculation(bp.Precio, op, amount)
		derived = append(derived, tire.TirePrice{
			LlantaID:        bp.LlantaID,
			ColumnaPrecioID: column.ID,
			Precio:          price,
			CreadoEn:        now,
			ActualizadoEn:   now,
		})
	}

	return s.prices.UpsertMany(ctx, derived)
}

// applyPriceCalculation aplica la operación configurada sobre el precio base.
// Para "percent" se interpreta la cantidad como descuento porcentual: 10 => 10% de descuento.
func applyPriceCalculation(base float64, operation string, amount float64) float64 {
	op := strings.ToLower(strings.TrimSpace(operation))
	switch op {
	case "add":
		return base + amount
	case "subtract":
		return base - amount
	case "multiply":
		return base * amount
	case "percent":
		return base * (1 - amount/100.0)
	default:
		return base
	}
}
