package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/llantera/hex/internal/domain/tire"
)

// TireHandler expone endpoints HTTP para gestionar llantas.
type TireHandler struct {
	service tire.TireService
}

func NewTireHandler(service tire.TireService) *TireHandler {
	return &TireHandler{service: service}
}

// HandleCollection atiende rutas sin identificador (/api/tires).
func (h *TireHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

// HandleAdmin atiende el listado de administración (/api/tires/admin).
// Devuelve la llanta junto con inventario y todos los precios registrados.
func (h *TireHandler) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	q := r.URL.Query()

	filter := tire.TireFilter{
		Search: strings.TrimSpace(q.Get("search")),
		Limit:  parseLimit(q.Get("limit")),
		Offset: parseOffset(q.Get("offset")),
		Sort:   q.Get("sort"),
	}

	if v := strings.TrimSpace(q.Get("marcaId")); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "marcaId debe ser numérico")
			return
		}
		filter.MarcaID = &id
	}

	if v := strings.TrimSpace(q.Get("tipoId")); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "tipoId debe ser numérico")
			return
		}
		filter.TipoID = &id
	}

	if v := strings.TrimSpace(q.Get("abreviatura")); v != "" {
		filter.Abreviatura = strings.ToUpper(v)
	}

	if v := strings.TrimSpace(q.Get("ancho")); v != "" {
		ancho, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ancho debe ser numérico")
			return
		}
		filter.Ancho = &ancho
	}

	if v := strings.TrimSpace(q.Get("perfil")); v != "" {
		perfil, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "perfil debe ser numérico")
			return
		}
		filter.Perfil = &perfil
	}

	if v := strings.TrimSpace(q.Get("rin")); v != "" {
		rin, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "rin debe ser numérico")
			return
		}
		filter.Rin = &rin
	}

	if v := strings.TrimSpace(q.Get("construccion")); v != "" {
		filter.Construccion = v
	}

	if v := strings.TrimSpace(q.Get("capas")); v != "" {
		filter.CalificacionCapas = v
	}

	if v := strings.TrimSpace(q.Get("indiceCarga")); v != "" {
		filter.IndiceCarga = v
	}

	if v := strings.TrimSpace(q.Get("indiceVelocidad")); v != "" {
		filter.IndiceVelocidad = v
	}

	if v := strings.TrimSpace(q.Get("inStock")); v != "" {
		lower := strings.ToLower(v)
		if lower == "1" || lower == "true" || lower == "yes" || lower == "on" {
			filter.InStockOnly = true
		}
	}

	if v := strings.TrimSpace(q.Get("ancho")); v != "" {
		ancho, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ancho debe ser numérico")
			return
		}
		filter.Ancho = &ancho
	}

	if v := strings.TrimSpace(q.Get("perfil")); v != "" {
		perfil, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "perfil debe ser numérico")
			return
		}
		filter.Perfil = &perfil
	}

	if v := strings.TrimSpace(q.Get("rin")); v != "" {
		rin, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "rin debe ser numérico")
			return
		}
		filter.Rin = &rin
	}

	if v := strings.TrimSpace(q.Get("construccion")); v != "" {
		filter.Construccion = v
	}

	if v := strings.TrimSpace(q.Get("capas")); v != "" {
		filter.CalificacionCapas = v
	}

	if v := strings.TrimSpace(q.Get("indiceCarga")); v != "" {
		filter.IndiceCarga = v
	}

	if v := strings.TrimSpace(q.Get("indiceVelocidad")); v != "" {
		filter.IndiceVelocidad = v
	}

	if v := strings.TrimSpace(q.Get("ancho")); v != "" {
		ancho, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ancho debe ser numérico")
			return
		}
		filter.Ancho = &ancho
	}

	if v := strings.TrimSpace(q.Get("perfil")); v != "" {
		perfil, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "perfil debe ser numérico")
			return
		}
		filter.Perfil = &perfil
	}

	if v := strings.TrimSpace(q.Get("rin")); v != "" {
		rin, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "rin debe ser numérico")
			return
		}
		filter.Rin = &rin
	}

	if v := strings.TrimSpace(q.Get("construccion")); v != "" {
		filter.Construccion = v
	}

	if v := strings.TrimSpace(q.Get("capas")); v != "" {
		filter.CalificacionCapas = v
	}

	if v := strings.TrimSpace(q.Get("indiceCarga")); v != "" {
		filter.IndiceCarga = v
	}

	if v := strings.TrimSpace(q.Get("indiceVelocidad")); v != "" {
		filter.IndiceVelocidad = v
	}

	items, total, err := h.service.AdminList(r.Context(), filter)
	if err != nil {
		respondTireError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": items,
		"meta": map[string]int{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

// HandleAdminExport genera un archivo XLSX con el catálogo admin de llantas.
func (h *TireHandler) HandleAdminExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	q := r.URL.Query()

	filter := tire.TireFilter{
		Search: strings.TrimSpace(q.Get("search")),
		Limit:  0,
		Offset: 0,
		Sort:   q.Get("sort"),
	}

	if v := strings.TrimSpace(q.Get("marcaId")); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "marcaId debe ser numérico")
			return
		}
		filter.MarcaID = &id
	}

	if v := strings.TrimSpace(q.Get("tipoId")); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "tipoId debe ser numérico")
			return
		}
		filter.TipoID = &id
	}

	if v := strings.TrimSpace(q.Get("abreviatura")); v != "" {
		filter.Abreviatura = strings.ToUpper(v)
	}

	data, err := h.service.ExportAdmin(r.Context(), filter)
	if err != nil {
		respondTireError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\"catalogo_llantas.xlsx\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// HandleAdminImport procesa la carga masiva de llantas desde un archivo XLSX.
// Espera un multipart/form-data con un campo "file" que contenga el archivo.
func (h *TireHandler) HandleAdminImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	// Limitar memoria usada por el parseo del multipart.
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "no se pudo leer el archivo enviado")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "no se recibió ningún archivo")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "no se pudo leer el contenido del archivo")
		return
	}

	count, err := h.service.ImportFromXLSX(r.Context(), data)
	if err != nil {
		respondTireError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"processed": count,
	})
}

// HandleAdminResource atiende rutas con identificador para administración (/api/tires/admin/{sku}).
// Actualmente sólo permite actualizar inventario y precios vía PUT.
func (h *TireHandler) HandleAdminResource(w http.ResponseWriter, r *http.Request) {
	sku := strings.TrimSpace(extractResourceID(r.URL.Path, "/api/tires/admin"))
	if sku == "" {
		writeError(w, http.StatusBadRequest, "SKU inválido")
		return
	}

	switch r.Method {
	case http.MethodPut:
		updateAdmin(w, r, sku, h.service)
	default:
		w.Header().Set("Allow", "PUT")
		writeError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

// HandleCatalog atiende el listado de catálogo (/api/catalog/tires).
func (h *TireHandler) HandleCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	q := r.URL.Query()

	filter := tire.TireFilter{
		Search: strings.TrimSpace(q.Get("search")),
		Limit:  parseLimit(q.Get("limit")),
		Offset: parseOffset(q.Get("offset")),
		Sort:   q.Get("sort"),
	}

	if v := strings.TrimSpace(q.Get("marcaId")); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "marcaId debe ser numérico")
			return
		}
		filter.MarcaID = &id
	}

	if v := strings.TrimSpace(q.Get("tipoId")); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "tipoId debe ser numérico")
			return
		}
		filter.TipoID = &id
	}

	if v := strings.TrimSpace(q.Get("abreviatura")); v != "" {
		filter.Abreviatura = strings.ToUpper(v)
	}

	if v := strings.TrimSpace(q.Get("ancho")); v != "" {
		ancho, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ancho debe ser numérico")
			return
		}
		filter.Ancho = &ancho
	}

	if v := strings.TrimSpace(q.Get("perfil")); v != "" {
		perfil, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "perfil debe ser numérico")
			return
		}
		filter.Perfil = &perfil
	}

	if v := strings.TrimSpace(q.Get("rin")); v != "" {
		rin, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "rin debe ser numérico")
			return
		}
		filter.Rin = &rin
	}

	if v := strings.TrimSpace(q.Get("construccion")); v != "" {
		filter.Construccion = v
	}

	if v := strings.TrimSpace(q.Get("capas")); v != "" {
		filter.CalificacionCapas = v
	}

	if v := strings.TrimSpace(q.Get("indiceCarga")); v != "" {
		filter.IndiceCarga = v
	}

	if v := strings.TrimSpace(q.Get("indiceVelocidad")); v != "" {
		filter.IndiceVelocidad = v
	}

	level := strings.TrimSpace(q.Get("level"))

	items, total, err := h.service.ListCatalog(r.Context(), filter, level)
	if err != nil {
		respondTireError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": items,
		"meta": map[string]int{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

// HandleResource atiende rutas con identificador (/api/tires/{sku}).
func (h *TireHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	sku := strings.TrimSpace(extractResourceID(r.URL.Path, "/api/tires"))
	if sku == "" {
		writeError(w, http.StatusBadRequest, "SKU inválido")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, r, sku)
	case http.MethodPut:
		h.update(w, r, sku)
	case http.MethodDelete:
		h.delete(w, r, sku)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

func (h *TireHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := tire.TireFilter{
		Search: strings.TrimSpace(q.Get("search")),
		Limit:  parseLimit(q.Get("limit")),
		Offset: parseOffset(q.Get("offset")),
		Sort:   q.Get("sort"),
	}

	if v := strings.TrimSpace(q.Get("marcaId")); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "marcaId debe ser numérico")
			return
		}
		filter.MarcaID = &id
	}

	if v := strings.TrimSpace(q.Get("tipoId")); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "tipoId debe ser numérico")
			return
		}
		filter.TipoID = &id
	}

	if v := strings.TrimSpace(q.Get("abreviatura")); v != "" {
		filter.Abreviatura = strings.ToUpper(v)
	}

	items, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		respondTireError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": items,
		"meta": map[string]int{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

func (h *TireHandler) create(w http.ResponseWriter, r *http.Request) {
	var payload tirePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	if strings.TrimSpace(payload.SKU) == "" {
		writeError(w, http.StatusBadRequest, "SKU es obligatorio")
		return
	}

	cmd := payload.toCommand()
	created, err := h.service.UpsertFromMeasurement(r.Context(), cmd)
	if err != nil {
		respondTireError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *TireHandler) get(w http.ResponseWriter, r *http.Request, sku string) {
	entity, err := h.service.Get(r.Context(), sku)
	if err != nil {
		respondTireError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, entity)
}

func (h *TireHandler) update(w http.ResponseWriter, r *http.Request, sku string) {
	var payload tirePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	payload.SKU = sku
	cmd := payload.toCommand()

	updated, err := h.service.UpsertFromMeasurement(r.Context(), cmd)
	if err != nil {
		respondTireError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *TireHandler) delete(w http.ResponseWriter, r *http.Request, sku string) {
	if err := h.service.Delete(r.Context(), sku); err != nil {
		respondTireError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func respondTireError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, tire.ErrTireNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

type tirePayload struct {
	SKU               string  `json:"sku"`
	MarcaNombre       string  `json:"marcaNombre"`
	AliasMarca        string  `json:"aliasMarca"`
	Modelo            string  `json:"modelo"`
	Ancho             int     `json:"ancho"`
	Perfil            *int    `json:"perfil"`
	Rin               float64 `json:"rin"`
	Construccion      string  `json:"construccion"`
	TipoTubo          string  `json:"tipoTubo"`
	CalificacionCapas string  `json:"calificacionCapas"`
	IndiceCarga       string  `json:"indiceCarga"`
	IndiceVelocidad   string  `json:"indiceVelocidad"`
	TipoNormalizado   string  `json:"tipoNormalizado"`
	AbreviaturaUso    string  `json:"abreviaturaUso"`
	Descripcion       string  `json:"descripcion"`
	PrecioPublico     float64 `json:"precioPublico"`
	URLImagen         string  `json:"urlImagen"`
	MedidaOriginal    string  `json:"medidaOriginal"`
}

// adminUpdatePayload representa el cuerpo del PUT /api/tires/admin/{sku}.
type adminUpdatePayload struct {
	Cantidad *int                `json:"cantidad"`
	Precios  map[string]*float64 `json:"precios"`
}

func updateAdmin(w http.ResponseWriter, r *http.Request, sku string, service tire.TireService) {
	var payload adminUpdatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	updated, err := service.UpdateAdmin(r.Context(), sku, payload.Cantidad, payload.Precios)
	if err != nil {
		respondTireError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (p tirePayload) toCommand() tire.UpsertCommand {
	return tire.UpsertCommand{
		SKU:               strings.TrimSpace(p.SKU),
		MarcaNombre:       p.MarcaNombre,
		AliasMarca:        p.AliasMarca,
		Modelo:            p.Modelo,
		Ancho:             p.Ancho,
		Perfil:            p.Perfil,
		Rin:               p.Rin,
		Construccion:      p.Construccion,
		TipoTubo:          p.TipoTubo,
		CalificacionCapas: p.CalificacionCapas,
		IndiceCarga:       p.IndiceCarga,
		IndiceVelocidad:   p.IndiceVelocidad,
		TipoNormalizado:   p.TipoNormalizado,
		AbreviaturaUso:    p.AbreviaturaUso,
		Descripcion:       p.Descripcion,
		PrecioPublico:     p.PrecioPublico,
		URLImagen:         p.URLImagen,
		MedidaOriginal:    p.MedidaOriginal,
	}
}
