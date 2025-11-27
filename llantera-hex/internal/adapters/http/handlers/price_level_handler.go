package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/llantera/hex/internal/domain/pricelevel"
)

// PriceLevelHandler expone CRUD HTTP para los niveles de precio.
type PriceLevelHandler struct {
	service pricelevel.PriceLevelService
}

func NewPriceLevelHandler(service pricelevel.PriceLevelService) *PriceLevelHandler {
	return &PriceLevelHandler{service: service}
}

// HandleCollection atiende rutas sin identificador (/api/price-levels).
func (h *PriceLevelHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
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

// HandleResource atiende rutas con identificador (/api/price-levels/{id}).
func (h *PriceLevelHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	idStr := extractResourceID(r.URL.Path, "/api/price-levels")
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

func (h *PriceLevelHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := parseLimit(q.Get("limit"))
	offset := parseOffset(q.Get("offset"))

	filter := pricelevel.PriceLevelFilter{
		Limit:  limit,
		Offset: offset,
	}

	if v := strings.TrimSpace(q.Get("code")); v != "" {
		filter.Code = &v
	}

	items, total, err := h.service.List(filter)
	if err != nil {
		respondPriceLevelError(w, err)
		return
	}

	response := make([]priceLevelResponse, 0, len(items))
	for _, lvl := range items {
		if lvl == nil {
			continue
		}
		response = append(response, toPriceLevelResponse(lvl))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": response,
		"meta": map[string]int{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

func (h *PriceLevelHandler) create(w http.ResponseWriter, r *http.Request) {
	var payload priceLevelPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	level := payload.toDomain()

	created, err := h.service.Create(level)
	if err != nil {
		respondPriceLevelError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPriceLevelResponse(created))
}

func (h *PriceLevelHandler) get(w http.ResponseWriter, r *http.Request, id int) {
	level, err := h.service.GetByID(id)
	if err != nil {
		respondPriceLevelError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPriceLevelResponse(level))
}

func (h *PriceLevelHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	var payload priceLevelPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	level := payload.toDomain()

	updated, err := h.service.Update(id, level)
	if err != nil {
		respondPriceLevelError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPriceLevelResponse(updated))
}

func (h *PriceLevelHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	q := r.URL.Query()
	var transferToID *int
	if v := strings.TrimSpace(q.Get("transferToId")); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, "transferToId debe ser un entero positivo")
			return
		}
		transferToID = &parsed
	}

	if err := h.service.Delete(id, transferToID); err != nil {
		respondPriceLevelError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func respondPriceLevelError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	status := http.StatusInternalServerError

	switch {
	case strings.Contains(msg, "no encontrado"):
		status = http.StatusNotFound
	case strings.Contains(msg, "invlido"),
		strings.Contains(msg, "obligatorio"),
		strings.Contains(msg, "ya existe"),
		strings.Contains(msg, "no se puede eliminar"):
		status = http.StatusBadRequest
	default:
		status = http.StatusInternalServerError
	}

	writeError(w, status, err.Error())
}

// priceLevelPayload representa el cuerpo de creacin/actualizacin de niveles de precio.
type priceLevelPayload struct {
	Code               string  `json:"code"`
	Name               string  `json:"name"`
	Description        *string `json:"description"`
	DiscountPercentage float64 `json:"discountPercentage"`
	PriceColumn        string  `json:"priceColumn"`
	ReferenceColumn    *string `json:"referenceColumn"`
	CanViewOffers      bool    `json:"canViewOffers"`
}

func (p *priceLevelPayload) toDomain() *pricelevel.PriceLevel {
	trim := func(s string) string { return strings.TrimSpace(s) }

	var desc *string
	if p.Description != nil {
		val := trim(*p.Description)
		if val != "" {
			desc = &val
		}
	}

	var ref *string
	if p.ReferenceColumn != nil {
		val := trim(*p.ReferenceColumn)
		if val != "" {
			lval := strings.ToLower(val)
			ref = &lval
		}
	}

	return &pricelevel.PriceLevel{
		Code:               strings.ToLower(trim(p.Code)),
		Name:               trim(p.Name),
		Description:        desc,
		DiscountPercentage: p.DiscountPercentage,
		PriceColumn:        strings.ToLower(trim(p.PriceColumn)),
		ReferenceColumn:    ref,
		CanViewOffers:      p.CanViewOffers,
	}
}

// priceLevelResponse representa la vista pblica de un nivel de precio.
type priceLevelResponse struct {
	ID                 int     `json:"id"`
	Code               string  `json:"code"`
	Name               string  `json:"name"`
	Description        *string `json:"description,omitempty"`
	DiscountPercentage float64 `json:"discountPercentage"`
	PriceColumn        string  `json:"priceColumn"`
	ReferenceColumn    *string `json:"referenceColumn,omitempty"`
	CanViewOffers      bool    `json:"canViewOffers"`
}

func toPriceLevelResponse(l *pricelevel.PriceLevel) priceLevelResponse {
	if l == nil {
		return priceLevelResponse{}
	}
	return priceLevelResponse{
		ID:                 l.ID,
		Code:               l.Code,
		Name:               l.Name,
		Description:        l.Description,
		DiscountPercentage: l.DiscountPercentage,
		PriceColumn:        l.PriceColumn,
		ReferenceColumn:    l.ReferenceColumn,
		CanViewOffers:      l.CanViewOffers,
	}
}
