package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/llantera/hex/internal/domain/user"
)

// UserHandler expone endpoints HTTP para gestionar usuarios.
type UserHandler struct {
	service user.Service
}

func NewUserHandler(service user.Service) *UserHandler {
	return &UserHandler{service: service}
}

// HandleCollection atiende rutas sin identificador (/api/users).
func (h *UserHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
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

// HandleResource atiende rutas con identificador (/api/users/{id}).
func (h *UserHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	id := extractResourceID(r.URL.Path, "/api/users")
	if id == "" {
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

func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := parseLimit(q.Get("limit"))
	offset := parseOffset(q.Get("offset"))

	filter := user.ListFilter{
		Search: q.Get("search"),
		Limit:  limit,
		Offset: offset,
		Sort:   q.Get("sort"),
	}

	if v := q.Get("companyId"); strings.TrimSpace(v) != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "companyId debe ser numérico")
			return
		}
		filter.CompanyID = &id
	}

	if v := q.Get("role"); strings.TrimSpace(v) != "" {
		role := user.Role(strings.TrimSpace(v))
		filter.Role = &role
	}

	if v := q.Get("active"); strings.TrimSpace(v) != "" {
		active, err := strconv.ParseBool(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "active debe ser true o false")
			return
		}
		filter.Active = &active
	}

	items, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		respondUserError(w, err)
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

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var payload userPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	cmd := user.CreateCommand{
		Email:               payload.Email,
		Password:            payload.Password,
		FirstName:           payload.FirstName,
		FirstLastName:       payload.FirstLastName,
		SecondLastName:      payload.SecondLastName,
		Phone:               payload.Phone,
		AddressStreet:       payload.AddressStreet,
		AddressNumber:       payload.AddressNumber,
		AddressNeighborhood: payload.AddressNeighborhood,
		AddressPostalCode:   payload.AddressPostalCode,
		JobTitle:            payload.JobTitle,
		Active:              payload.Active,
		CompanyID:           payload.CompanyID,
		ProfileImageURL:     payload.ProfileImageURL,
		Role:                user.Role(payload.Role),
		Level:               user.PriceLevel(payload.Level),
		PriceLevelID:        payload.PriceLevelID,
	}

	created, err := h.service.Create(r.Context(), cmd)
	if err != nil {
		respondUserError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *UserHandler) get(w http.ResponseWriter, r *http.Request, id string) {
	user, err := h.service.Get(r.Context(), id)
	if err != nil {
		respondUserError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	var payload userUpdatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	cmd := user.UpdateCommand{
		ID:                  id,
		Email:               payload.Email,
		FirstName:           payload.FirstName,
		FirstLastName:       payload.FirstLastName,
		SecondLastName:      payload.SecondLastName,
		Phone:               payload.Phone,
		AddressStreet:       payload.AddressStreet,
		AddressNumber:       payload.AddressNumber,
		AddressNeighborhood: payload.AddressNeighborhood,
		AddressPostalCode:   payload.AddressPostalCode,
		JobTitle:            payload.JobTitle,
		Active:              payload.Active,
		CompanyID:           payload.CompanyID,
		ProfileImageURL:     payload.ProfileImageURL,
		Role:                user.Role(payload.Role),
		Level:               user.PriceLevel(payload.Level),
		PriceLevelID:        payload.PriceLevelID,
	}

	updated, err := h.service.Update(r.Context(), cmd)
	if err != nil {
		respondUserError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *UserHandler) delete(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.service.Delete(r.Context(), id); err != nil {
		respondUserError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func respondUserError(w http.ResponseWriter, err error) {
	var verr *user.ValidationError

	switch {
	case errors.Is(err, user.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, user.ErrEmailAlreadyUsed):
		writeError(w, http.StatusConflict, err.Error())
	case errors.As(err, &verr):
		writeError(w, http.StatusBadRequest, verr.Error())
	default:
		writeError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

type userPayload struct {
	Email               string `json:"email"`
	Password            string `json:"password"`
	FirstName           string `json:"firstName"`
	FirstLastName       string `json:"firstLastName"`
	SecondLastName      string `json:"secondLastName"`
	Phone               string `json:"phone"`
	AddressStreet       string `json:"addressStreet"`
	AddressNumber       string `json:"addressNumber"`
	AddressNeighborhood string `json:"addressNeighborhood"`
	AddressPostalCode   string `json:"addressPostalCode"`
	JobTitle            string `json:"jobTitle"`
	Active              bool   `json:"active"`
	CompanyID           *int   `json:"companyId"`
	ProfileImageURL     string `json:"profileImageUrl"`
	Role                string `json:"role"`
	Level               string `json:"level"`
	PriceLevelID        *int   `json:"priceLevelId"`
}

type userUpdatePayload struct {
	Email               string `json:"email"`
	FirstName           string `json:"firstName"`
	FirstLastName       string `json:"firstLastName"`
	SecondLastName      string `json:"secondLastName"`
	Phone               string `json:"phone"`
	AddressStreet       string `json:"addressStreet"`
	AddressNumber       string `json:"addressNumber"`
	AddressNeighborhood string `json:"addressNeighborhood"`
	AddressPostalCode   string `json:"addressPostalCode"`
	JobTitle            string `json:"jobTitle"`
	Active              bool   `json:"active"`
	CompanyID           *int   `json:"companyId"`
	ProfileImageURL     string `json:"profileImageUrl"`
	Role                string `json:"role"`
	Level               string `json:"level"`
	PriceLevelID        *int   `json:"priceLevelId"`
}
