package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/llantera/hex/internal/adapters/http/middleware"
	"github.com/llantera/hex/internal/domain/customerrequest"
)

// CustomerRequestHandler expone CRUD HTTP para solicitudes de "Quiero ser cliente".
type CustomerRequestHandler struct {
	service customerrequest.Service
}

func NewCustomerRequestHandler(service customerrequest.Service) *CustomerRequestHandler {
	return &CustomerRequestHandler{service: service}
}

func (h *CustomerRequestHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
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

func (h *CustomerRequestHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	id := extractResourceID(r.URL.Path, "/api/customer-requests")
	if strings.TrimSpace(id) == "" {
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

func (h *CustomerRequestHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := customerrequest.ListFilter{
		Search: strings.TrimSpace(q.Get("search")),
		Limit:  parseLimit(q.Get("limit")),
		Offset: parseOffset(q.Get("offset")),
		Sort:   q.Get("sort"),
	}

	if v := strings.TrimSpace(q.Get("status")); v != "" {
		status := customerrequest.Status(v)
		filter.Status = &status
	}
	if v := strings.TrimSpace(q.Get("employeeId")); v != "" {
		filter.EmployeeID = &v
	}

	// Si el request viene autenticado como empleado, forzamos a que solo
	// pueda ver sus propias solicitudes, ignorando cualquier employeeId
	// que venga en el query string.
	if claims := middleware.FromContext(r.Context()); claims != nil {
		role := strings.ToLower(strings.TrimSpace(claims.Role))
		if role == "employee" {
			id := claims.UserID
			if strings.TrimSpace(id) != "" {
				filter.EmployeeID = &id
			}
		}
	}

	items, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		respondCustomerRequestError(w, err)
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

func (h *CustomerRequestHandler) create(w http.ResponseWriter, r *http.Request) {
	var payload customerRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	cmd := customerrequest.CreateCommand{
		FullName:          payload.FullName,
		RequestType:       payload.RequestType,
		Message:           payload.Message,
		Phone:             payload.Phone,
		ContactPreference: payload.ContactPreference,
		Email:             payload.Email,
	}

	created, err := h.service.Create(r.Context(), cmd)
	if err != nil {
		respondCustomerRequestError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *CustomerRequestHandler) get(w http.ResponseWriter, r *http.Request, id string) {
	cr, err := h.service.Get(r.Context(), id)
	if err != nil {
		respondCustomerRequestError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cr)
}

func (h *CustomerRequestHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	var payload customerRequestUpdatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	var statusPtr *customerrequest.Status
	if strings.TrimSpace(payload.Status) != "" {
		s := customerrequest.Status(strings.TrimSpace(payload.Status))
		statusPtr = &s
	}

	cmd := customerrequest.UpdateCommand{
		ID:         id,
		Message:    payload.Message,
		Status:     statusPtr,
		EmployeeID: payload.EmployeeID,
		Agreement:  payload.Agreement,
	}

	updated, err := h.service.Update(r.Context(), cmd)
	if err != nil {
		respondCustomerRequestError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *CustomerRequestHandler) delete(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.service.Delete(r.Context(), id); err != nil {
		respondCustomerRequestError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func respondCustomerRequestError(w http.ResponseWriter, err error) {
	var verr *customerrequest.ValidationError

	switch {
	case errors.Is(err, customerrequest.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.As(err, &verr):
		writeError(w, http.StatusBadRequest, verr.Error())
	default:
		writeError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

// Payloads HTTP

type customerRequestPayload struct {
	FullName          string `json:"fullName"`
	RequestType       string `json:"requestType"`
	Message           string `json:"message"`
	Phone             string `json:"phone"`
	ContactPreference string `json:"contactPreference"`
	Email             string `json:"email"`
}

type customerRequestUpdatePayload struct {
	Message    string  `json:"message"`
	Status     string  `json:"status"`
	EmployeeID *string `json:"employeeId"`
	Agreement  string  `json:"agreement"`
}
