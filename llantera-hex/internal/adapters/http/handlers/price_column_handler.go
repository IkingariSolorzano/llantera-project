package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/llantera/hex/internal/domain/tire"
)

// PriceColumnHandler expone CRUD HTTP para las columnas de precio.
type PriceColumnHandler struct {
	service tire.PriceColumnService
}

func NewPriceColumnHandler(service tire.PriceColumnService) *PriceColumnHandler {
	return &PriceColumnHandler{service: service}
}

// HandleCollection atiende rutas sin identificador (/api/price-columns).
func (h *PriceColumnHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
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

// HandleResource atiende rutas con identificador (/api/price-columns/{id}).
func (h *PriceColumnHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	idStr := extractResourceID(r.URL.Path, "/api/price-columns")
	id, err := parsePositiveInt(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, r, id)
	case http.MethodPut:
		h.update(w, r, id)
	case http.MethodDelete:
		h.delete(w, r, id)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

func (h *PriceColumnHandler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListColumns(r.Context())
	if err != nil {
		respondPriceColumnError(w, err)
		return
	}

	response := make([]priceColumnResponse, 0, len(items))
	for _, c := range items {
		response = append(response, toPriceColumnResponse(&c))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": response,
		"meta": map[string]int{
			"total":  len(response),
			"limit":  len(response),
			"offset": 0,
		},
	})
}

func (h *PriceColumnHandler) create(w http.ResponseWriter, r *http.Request) {
	var payload priceColumnPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	cmd := tire.PriceColumnCreateCommand{
		Codigo:      payload.Code,
		Nombre:      payload.Name,
		Descripcion: payload.Description,
		OrdenVisual: payload.VisualOrder,
		Activo:      payload.Active,
		EsPublico:   payload.IsPublic,
		Mode:        payload.Mode,
		BaseCode:    payload.BaseCode,
		Operation:   payload.Operation,
		Amount:      payload.Amount,
	}

	created, err := h.service.CreateColumn(r.Context(), cmd)
	if err != nil {
		respondPriceColumnError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPriceColumnResponse(created))
}

func (h *PriceColumnHandler) get(w http.ResponseWriter, r *http.Request, id int) {
	col, err := h.service.GetColumn(r.Context(), id)
	if err != nil {
		respondPriceColumnError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPriceColumnResponse(col))
}

func (h *PriceColumnHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	var payload priceColumnPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	cmd := tire.PriceColumnUpdateCommand{
		Nombre:      payload.Name,
		Descripcion: payload.Description,
		OrdenVisual: payload.VisualOrder,
		Activo:      payload.Active,
		EsPublico:   payload.IsPublic,
		Mode:        payload.Mode,
		BaseCode:    payload.BaseCode,
		Operation:   payload.Operation,
		Amount:      payload.Amount,
	}

	updated, err := h.service.UpdateColumn(r.Context(), id, cmd)
	if err != nil {
		respondPriceColumnError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPriceColumnResponse(updated))
}

func (h *PriceColumnHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	q := r.URL.Query()
	var transferToCode *string
	if v := strings.TrimSpace(q.Get("transferToCode")); v != "" {
		transferToCode = &v
	}

	if err := h.service.DeleteColumn(r.Context(), id, transferToCode); err != nil {
		respondPriceColumnError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func respondPriceColumnError(w http.ResponseWriter, err error) {
	var verr *tire.ValidationError

	switch {
	case errors.Is(err, tire.ErrPriceColumnNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.As(err, &verr):
		writeError(w, http.StatusBadRequest, verr.Error())
	default:
		writeError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

// priceColumnPayload representa el cuerpo de creación/actualización de columnas de precio.
type priceColumnPayload struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	VisualOrder int      `json:"visualOrder"`
	Active      bool     `json:"active"`
	IsPublic    bool     `json:"isPublic"`
	Mode        string   `json:"mode"`
	BaseCode    string   `json:"baseCode"`
	Operation   string   `json:"operation"`
	Amount      *float64 `json:"amount"`
}

// priceColumnResponse representa la vista pública de una columna de precio.
type priceColumnResponse struct {
	ID          int      `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	VisualOrder int      `json:"visualOrder"`
	Active      bool     `json:"active"`
	IsPublic    bool     `json:"isPublic"`
	Mode        string   `json:"mode"`
	BaseCode    string   `json:"baseCode"`
	Operation   string   `json:"operation"`
	Amount      *float64 `json:"amount"`
}

func toPriceColumnResponse(c *tire.PriceColumn) priceColumnResponse {
	if c == nil {
		return priceColumnResponse{}
	}
	baseCode := ""
	if c.BaseCode != nil {
		baseCode = *c.BaseCode
	}
	var amount *float64
	if c.Amount != nil {
		v := *c.Amount
		amount = &v
	}
	return priceColumnResponse{
		ID:          c.ID,
		Code:        c.Codigo,
		Name:        c.Nombre,
		Description: c.Descripcion,
		VisualOrder: c.OrdenVisual,
		Active:      c.Activo,
		IsPublic:    c.EsPublico,
		Mode:        c.Mode,
		BaseCode:    baseCode,
		Operation:   c.Operation,
		Amount:      amount,
	}
}
