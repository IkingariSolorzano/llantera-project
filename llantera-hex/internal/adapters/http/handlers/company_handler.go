package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/llantera/hex/internal/domain/company"
)

// CompanyHandler expone CRUD HTTP para las empresas del sistema.
type CompanyHandler struct {
	service company.Service
}

func NewCompanyHandler(service company.Service) *CompanyHandler {
	return &CompanyHandler{service: service}
}

func (h *CompanyHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
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

func (h *CompanyHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	idStr := extractResourceID(r.URL.Path, "/api/companies")
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

func (h *CompanyHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := company.ListFilter{
		Search: strings.TrimSpace(q.Get("search")),
		Limit:  parseLimit(q.Get("limit")),
		Offset: parseOffset(q.Get("offset")),
		Sort:   q.Get("sort"),
	}

	items, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		respondCompanyError(w, err)
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

func (h *CompanyHandler) create(w http.ResponseWriter, r *http.Request) {
	var payload companyPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	cmd := company.CreateCommand{
		KeyName:       payload.KeyName,
		SocialReason:  payload.SocialReason,
		RFC:           payload.RFC,
		Address:       payload.Address,
		Emails:        payload.Emails,
		Phones:        payload.Phones,
		MainContactID: payload.MainContactID,
	}

	created, err := h.service.Create(r.Context(), cmd)
	if err != nil {
		respondCompanyError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *CompanyHandler) get(w http.ResponseWriter, r *http.Request, id int) {
	c, err := h.service.Get(r.Context(), id)
	if err != nil {
		respondCompanyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CompanyHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	var payload companyPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	cmd := company.UpdateCommand{
		ID:            id,
		KeyName:       payload.KeyName,
		SocialReason:  payload.SocialReason,
		RFC:           payload.RFC,
		Address:       payload.Address,
		Emails:        payload.Emails,
		Phones:        payload.Phones,
		MainContactID: payload.MainContactID,
	}

	updated, err := h.service.Update(r.Context(), cmd)
	if err != nil {
		respondCompanyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *CompanyHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	if err := h.service.Delete(r.Context(), id); err != nil {
		respondCompanyError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func respondCompanyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, company.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

type companyPayload struct {
	KeyName       string   `json:"keyName"`
	SocialReason  string   `json:"socialReason"`
	RFC           string   `json:"rfc"`
	Address       string   `json:"address"`
	Emails        []string `json:"emails"`
	Phones        []string `json:"phones"`
	MainContactID *string  `json:"mainContactId"`
}
