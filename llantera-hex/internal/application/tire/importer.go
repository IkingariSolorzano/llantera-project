package tireapp

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"

	"github.com/llantera/hex/internal/domain/tire"
)

// ImportFromCSV lee un archivo delimitado por ';' siguiendo el layout del inventario
// y utiliza el servicio para hacer upserts de cada fila válida.
func (s *Service) ImportFromCSV(ctx context.Context, path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("abrir archivo de inventario: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	if _, err := reader.Read(); err != nil {
		if err == io.EOF {
			return 0, fmt.Errorf("el archivo %s no contiene datos", path)
		}
		return 0, fmt.Errorf("leer encabezado: %w", err)
	}

	priceColumns, err := s.priceColumns.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("cargar columnas de precio: %w", err)
	}
	codeToID := make(map[string]int, len(priceColumns))
	for _, c := range priceColumns {
		code := strings.ToLower(strings.TrimSpace(c.Codigo))
		if code == "" {
			continue
		}
		codeToID[code] = c.ID
	}

	var processed int
	line := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			return processed, fmt.Errorf("fila %d: %w", line, err)
		}
		if len(record) == 0 {
			continue
		}

		cmd, err := buildCommandFromRow(record)
		if err != nil {
			return processed, fmt.Errorf("fila %d: %w", line, err)
		}

		entity, err := s.UpsertFromMeasurement(ctx, cmd)
		if err != nil {
			return processed, fmt.Errorf("fila %d sku %s: %w", line, cmd.SKU, err)
		}
		if err := s.updateInventoryAndPrices(ctx, entity, record, codeToID); err != nil {
			return processed, fmt.Errorf("fila %d sku %s: %w", line, cmd.SKU, err)
		}
		processed++
	}

	return processed, nil
}

func buildMedidaOriginalFromFields(
	ancho int,
	perfil *int,
	rin float64,
	construccion string,
	capas string,
	uso string,
	indiceCarga string,
	indiceVelocidad string,
	modelo string,
) string {
	// Normalizar textos de entrada
	constr := strings.ToUpper(strings.TrimSpace(construccion))
	capasNorm := strings.TrimSpace(capas)
	usoNorm := strings.ToUpper(strings.TrimSpace(uso))
	indiceCargaNorm := strings.TrimSpace(indiceCarga)
	indiceVelNorm := strings.ToUpper(strings.TrimSpace(indiceVelocidad))
	modeloNorm := strings.TrimSpace(modelo)

	isDiagonal := constr == "DIAGONAL" || constr == "D" || constr == "-"
	isRadial := constr == "RADIAL" || constr == "R"
	_ = isRadial // mantenido por simetría con la lógica del formulario

	medidaBase := ""

	if ancho > 0 {
		rinStr := ""
		if rin > 0 {
			if rin == float64(int(rin)) {
				rinStr = fmt.Sprintf("%d", int(rin))
			} else {
				rinStr = fmt.Sprintf("%g", rin)
			}
		}

		if perfil != nil && *perfil > 0 {
			perfilStr := fmt.Sprintf("/%d", *perfil)
			sep := "R"
			if isDiagonal {
				sep = "-"
			}
			rinPart := ""
			if rinStr != "" {
				rinPart = sep + rinStr
			}
			medidaBase = fmt.Sprintf("%d%s%s", ancho, perfilStr, rinPart)
		} else if isDiagonal {
			rinPart := ""
			if rinStr != "" {
				rinPart = "-" + rinStr
			}
			medidaBase = fmt.Sprintf("%d%s", ancho, rinPart)
		} else {
			rinPart := ""
			if rinStr != "" {
				rinPart = "X" + rinStr
			}
			medidaBase = fmt.Sprintf("%d%s", ancho, rinPart)
		}
	}

	usoCap := ""
	if usoNorm != "" && capasNorm != "" {
		usoCap = usoNorm + "-" + capasNorm
	} else if usoNorm != "" {
		usoCap = usoNorm
	} else if capasNorm != "" {
		usoCap = capasNorm
	}

	cargaVel := ""
	if indiceCargaNorm != "" && indiceVelNorm != "" {
		cargaVel = indiceCargaNorm + indiceVelNorm
	} else if indiceCargaNorm != "" {
		cargaVel = indiceCargaNorm
	} else if indiceVelNorm != "" {
		cargaVel = indiceVelNorm
	}

	parts := make([]string, 0, 4)
	if medidaBase != "" {
		parts = append(parts, medidaBase)
	}
	if usoCap != "" {
		parts = append(parts, usoCap)
	}
	if cargaVel != "" {
		parts = append(parts, cargaVel)
	}
	if modeloNorm != "" {
		parts = append(parts, modeloNorm)
	}

	if len(parts) == 0 {
		return ""
	}

	// Compactar espacios múltiples, equivalente a .replace(/\s+/g, ' ') y .trim() en el frontend
	joined := strings.Join(parts, " ")
	return strings.Join(strings.Fields(joined), " ")
}

// ImportFromXLSX procesa un archivo XLSX con el layout de catálogo admin.
// Solo el SKU es obligatorio. Los demás campos son opcionales y solo se actualizan
// si la columna existe en el Excel y tiene un valor.
// Para llantas existentes, solo actualiza los campos presentes en el Excel.
// Para llantas nuevas, crea la llanta con los campos proporcionados.
func (s *Service) ImportFromXLSX(ctx context.Context, data []byte) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("el archivo XLSX está vacío")
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("abrir archivo XLSX: %w", err)
	}

	sheetName := "Catalogo"
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return 0, fmt.Errorf("el archivo XLSX no contiene hojas")
	}
	found := false
	for _, name := range sheets {
		if name == sheetName {
			found = true
			break
		}
	}
	if !found {
		sheetName = sheets[0]
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("leer filas de XLSX: %w", err)
	}
	if len(rows) == 0 {
		return 0, fmt.Errorf("el archivo XLSX no contiene datos")
	}

	header := rows[0]
	colIndex := make(map[string]int, len(header))
	for i, name := range header {
		key := strings.TrimSpace(strings.ToLower(name))
		if key == "" {
			continue
		}
		colIndex[key] = i
	}

	// Verificar que exista la columna SKU
	if _, hasSKU := colIndex["sku"]; !hasSKU {
		return 0, fmt.Errorf("el archivo XLSX debe contener una columna 'sku'")
	}

	// Helper para obtener valor de celda
	getCell := func(row []string, key string) string {
		idx, ok := colIndex[key]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}

	// Helper para verificar si una columna existe en el Excel
	hasColumn := func(key string) bool {
		_, ok := colIndex[key]
		return ok
	}

	// Mapear columnas de precio por código para leerlas del XLSX.
	priceCodeToColIndex := map[string]int{}
	if s.priceColumns != nil {
		cols, err := s.priceColumns.List(ctx)
		if err != nil {
			return 0, fmt.Errorf("cargar columnas de precio: %w", err)
		}
		for _, c := range cols {
			code := strings.ToLower(strings.TrimSpace(c.Codigo))
			if code == "" {
				continue
			}
			if idx, ok := colIndex[code]; ok {
				priceCodeToColIndex[code] = idx
			}
		}
	}

	processed := 0

	for i, row := range rows[1:] {
		line := i + 2 // considerando encabezado en la fila 1
		sku := getCell(row, "sku")
		if sku == "" {
			// Filas sin SKU se omiten silenciosamente.
			continue
		}

		// Buscar si la llanta ya existe
		existingTire, err := s.tiRes.GetBySKU(ctx, sku)
		if err != nil && err != tire.ErrTireNotFound {
			return processed, fmt.Errorf("fila %d sku %s: error buscando llanta: %w", line, sku, err)
		}

		isNew := existingTire == nil

		// Construir comando con valores existentes o nuevos
		cmd := tire.UpsertCommand{SKU: sku}

		// Para cada campo, usar el valor del Excel si la columna existe,
		// o mantener el valor existente si la llanta ya existe
		if hasColumn("marca") {
			cmd.MarcaNombre = getCell(row, "marca")
		} else if !isNew && existingTire.MarcaID > 0 {
			// Mantener marca existente - se resuelve en UpsertFromMeasurement
		}

		if hasColumn("modelo") {
			cmd.Modelo = getCell(row, "modelo")
		} else if !isNew {
			cmd.Modelo = existingTire.Modelo
		}

		if hasColumn("ancho") {
			anchoStr := getCell(row, "ancho")
			if anchoStr != "" {
				cmd.Ancho = atoi(anchoStr)
			}
		} else if !isNew {
			cmd.Ancho = existingTire.Ancho
		}

		if hasColumn("perfil") {
			perfilStr := getCell(row, "perfil")
			if perfilStr != "" {
				v := atoi(perfilStr)
				if v > 0 {
					cmd.Perfil = &v
				}
			}
		} else if !isNew {
			cmd.Perfil = existingTire.Perfil
		}

		if hasColumn("rin") {
			rinStr := getCell(row, "rin")
			if rinStr != "" {
				cmd.Rin = atof(rinStr)
			}
		} else if !isNew {
			cmd.Rin = existingTire.Rin
		}

		if hasColumn("construccion") {
			cmd.Construccion = getCell(row, "construccion")
		} else if !isNew {
			cmd.Construccion = existingTire.Construccion
		}

		if hasColumn("tipo_tubo") {
			cmd.TipoTubo = getCell(row, "tipo_tubo")
		} else if !isNew {
			cmd.TipoTubo = existingTire.TipoTubo
		}

		if hasColumn("calificacion_capas") {
			cmd.CalificacionCapas = getCell(row, "calificacion_capas")
		} else if !isNew {
			cmd.CalificacionCapas = existingTire.CalificacionCapas
		}

		if hasColumn("indice_carga") {
			cmd.IndiceCarga = getCell(row, "indice_carga")
		} else if !isNew {
			cmd.IndiceCarga = existingTire.IndiceCarga
		}

		if hasColumn("indice_velocidad") {
			cmd.IndiceVelocidad = getCell(row, "indice_velocidad")
		} else if !isNew {
			cmd.IndiceVelocidad = existingTire.IndiceVelocidad
		}

		if hasColumn("uso") {
			cmd.AbreviaturaUso = getCell(row, "uso")
		} else if !isNew {
			cmd.AbreviaturaUso = existingTire.AbreviaturaUso
		}

		if hasColumn("descripcion") {
			cmd.Descripcion = getCell(row, "descripcion")
		} else if !isNew {
			cmd.Descripcion = existingTire.Descripcion
		}

		if hasColumn("url_imagen") {
			cmd.URLImagen = getCell(row, "url_imagen")
		} else if !isNew {
			cmd.URLImagen = existingTire.URLImagen
		}

		// Precio público
		if hasColumn("precio_publico") {
			precioStr := getCell(row, "precio_publico")
			if precioStr != "" {
				cmd.PrecioPublico = parsePrice(precioStr)
			}
		} else if !isNew {
			cmd.PrecioPublico = existingTire.PrecioPublico
		}

		// Construir mapa de precios por código a partir de las columnas dinámicas.
		precios := make(map[string]*float64)
		for code, idx := range priceCodeToColIndex {
			if idx >= len(row) {
				continue
			}
			raw := strings.TrimSpace(row[idx])
			if raw == "" {
				continue
			}
			value := parsePrice(raw)
			if value <= 0 {
				continue
			}
			v := value
			precios[code] = &v
		}

		// Si precio_publico viene vacío, intentar tomarlo de la columna de lista.
		if cmd.PrecioPublico <= 0 {
			if v, ok := precios["lista"]; ok && v != nil && *v > 0 {
				cmd.PrecioPublico = *v
			}
		}

		// Construir medida original si hay suficientes datos
		if cmd.Ancho > 0 || cmd.Rin > 0 {
			cmd.MedidaOriginal = buildMedidaOriginalFromFields(
				cmd.Ancho, cmd.Perfil, cmd.Rin, cmd.Construccion,
				cmd.CalificacionCapas, cmd.AbreviaturaUso,
				cmd.IndiceCarga, cmd.IndiceVelocidad, cmd.Modelo,
			)
		} else if !isNew {
			cmd.MedidaOriginal = existingTire.MedidaOriginal
		}

		// Ejecutar upsert
		if _, err := s.UpsertFromMeasurement(ctx, cmd); err != nil {
			return processed, fmt.Errorf("fila %d sku %s: %w", line, sku, err)
		}

		// Inventario y precios vía UpdateAdmin para reutilizar la lógica existente
		// (incluyendo recálculo de columnas derivadas).
		var cantidadPtr *int
		if hasColumn("cantidad") {
			cantidadStr := getCell(row, "cantidad")
			if cantidadStr != "" {
				v := atoi(cantidadStr)
				cantidadPtr = &v
			}
		}

		if cantidadPtr != nil || len(precios) > 0 {
			if _, err := s.updateAdminInternal(ctx, sku, cantidadPtr, precios, false); err != nil {
				return processed, fmt.Errorf("fila %d sku %s: %w", line, sku, err)
			}
		}

		processed++
	}

	// Recalcular columnas derivadas una sola vez al final para todas las llantas importadas.
	if s.priceColumns != nil && s.prices != nil {
		cols, err := s.priceColumns.List(ctx)
		if err != nil {
			return processed, fmt.Errorf("cargar columnas de precio: %w", err)
		}
		for i := range cols {
			col := &cols[i]
			mode := strings.ToLower(strings.TrimSpace(col.Mode))
			if mode != "derived" {
				continue
			}
			if err := s.recalculateDerivedColumn(ctx, col); err != nil {
				return processed, err
			}
		}
	}

	return processed, nil
}

func buildCommandFromRow(row []string) (tire.UpsertCommand, error) {
	if len(row) < 17 {
		return tire.UpsertCommand{}, fmt.Errorf("fila incompleta: se esperaban al menos 17 columnas, llegaron %d", len(row))
	}

	sku := strings.TrimSpace(row[0])
	if sku == "" {
		return tire.UpsertCommand{}, fmt.Errorf("sku vacío")
	}

	medida := strings.TrimSpace(row[1])
	if medida == "" {
		return tire.UpsertCommand{}, fmt.Errorf("medida vacía para sku %s", sku)
	}

	data := parseMeasurement(medida)
	modelo := strings.TrimSpace(cleanModel(data.remainder))
	if modelo == "" {
		modelo = strings.TrimSpace(medida)
	}

	alias := strings.TrimSpace(row[13])
	marca := normalizeBrand(alias, modelo)
	tipoNormalizado := normalizeType(row[14], modelo)
	if tipoNormalizado == "" {
		tipoNormalizado = "Otros"
	}

	abreviatura := strings.ToUpper(strings.TrimSpace(row[16]))
	if abreviatura == "" {
		abreviatura = strings.ToUpper(strings.TrimSpace(row[15]))
	}

	width := data.width
	if width == 0 {
		width = atoi(extractFirstNumber(medida))
	}
	if width == 0 {
		return tire.UpsertCommand{}, fmt.Errorf("no se pudo determinar el ancho para sku %s", sku)
	}

	rim := data.rim
	if rim == 0 {
		rim = atof(row[15])
	}
	if rim == 0 {
		return tire.UpsertCommand{}, fmt.Errorf("no se pudo determinar el rin para sku %s", sku)
	}

	precio := parsePrice(row[7])
	if precio == 0 {
		precio = parsePrice(row[8])
	}
	if precio == 0 {
		precio = parsePrice(row[9])
	}

	descripcion := strings.TrimSpace(fmt.Sprintf("%s %s", medida, modelo))

	cmd := tire.UpsertCommand{
		SKU:               sku,
		MarcaNombre:       marca,
		AliasMarca:        strings.ToUpper(alias),
		Modelo:            modelo,
		Ancho:             width,
		Perfil:            data.profile,
		Rin:               rim,
		Construccion:      defaultConstruction(data.construction, medida),
		TipoTubo:          data.tubeType,
		CalificacionCapas: data.plyRating,
		IndiceCarga:       data.loadIndex,
		IndiceVelocidad:   data.speedIndex,
		TipoNormalizado:   tipoNormalizado,
		AbreviaturaUso:    abreviatura,
		Descripcion:       descripcion,
		PrecioPublico:     precio,
		MedidaOriginal:    medida,
	}

	return cmd, nil
}

func (s *Service) updateInventoryAndPrices(ctx context.Context, entity *tire.Tire, row []string, codeToID map[string]int) error {
	if s.inventory == nil && s.prices == nil {
		return nil
	}

	// 1) Inventario: columna 2 = CANT
	if s.inventory != nil && len(row) > 2 {
		cantidad := atoi(row[2])
		inv := &tire.Inventory{
			LlantaID:    entity.ID,
			Cantidad:    cantidad,
			StockMinimo: 4,
		}
		if err := s.inventory.Upsert(ctx, inv); err != nil {
			return fmt.Errorf("actualizar inventario: %w", err)
		}
	}

	// 2) Precios por tipo
	if s.prices != nil {
		var prices []tire.TirePrice

		addPrice := func(code string, idx int) {
			if idx >= len(row) {
				return
			}
			colID, ok := codeToID[code]
			if !ok {
				return
			}
			value := parsePrice(row[idx])
			if value <= 0 {
				return
			}
			prices = append(prices, tire.TirePrice{
				LlantaID:        entity.ID,
				ColumnaPrecioID: colID,
				Precio:          value,
			})
		}

		// Layout inventario.csv:
		// 0 CODIGO; 1 MEDIDA; 2 CANT;
		// 3 MAY -6%; 4 MAY -3%; 5 MAYOREO; 6 EMPPRESA; 7 P.LISTA; 8 P.LIST -10; 9 EFEC; 11 PROMO 4X3
		addPrice("mayoreo_6", 3)
		addPrice("mayoreo_3", 4)
		addPrice("mayoreo", 5)
		addPrice("empresa", 6)
		addPrice("lista", 7)
		addPrice("lista_10", 8)
		addPrice("efectivo", 9)

		if len(prices) > 0 {
			if err := s.prices.UpsertMany(ctx, prices); err != nil {
				return fmt.Errorf("actualizar precios: %w", err)
			}
		}
	}

	return nil
}

func cleanModel(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	return strings.Join(strings.Fields(trimmed), " ")
}

func defaultConstruction(construction, medida string) string {
	if construction != "" {
		return construction
	}
	upper := strings.ToUpper(medida)
	if strings.Contains(upper, "R") {
		return "R"
	}
	if strings.Contains(upper, "-") {
		return "D"
	}
	return ""
}

func extractFirstNumber(input string) string {
	for _, token := range strings.FieldsFunc(input, func(r rune) bool {
		return !(r >= '0' && r <= '9')
	}) {
		if token != "" {
			return token
		}
	}
	return ""
}
