package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/llantera/hex/internal/adapters/http/middleware"
	appAddress "github.com/llantera/hex/internal/application/address"
	"github.com/llantera/hex/internal/domain/address"
)

type AddressHandler struct {
	service *appAddress.Service
}

func NewAddressHandler(service *appAddress.Service) *AddressHandler {
	return &AddressHandler{service: service}
}

// HandleCollection maneja operaciones sobre la colección de direcciones
func (h *AddressHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "No autorizado")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.list(w, r, userID)
	case http.MethodPost:
		h.create(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

// HandleResource maneja operaciones sobre una dirección específica
func (h *AddressHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "No autorizado")
		return
	}

	// Extraer ID del path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		writeError(w, http.StatusBadRequest, "ID de dirección requerido")
		return
	}

	lastPart := parts[len(parts)-1]

	// Verificar si es la acción set-default
	if lastPart == "set-default" && len(parts) >= 4 {
		addressID, err := strconv.Atoi(parts[len(parts)-2])
		if err != nil {
			writeError(w, http.StatusBadRequest, "ID de dirección inválido")
			return
		}
		h.setDefault(w, r, userID, addressID)
		return
	}

	addressID, err := strconv.Atoi(lastPart)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID de dirección inválido")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, r, userID, addressID)
	case http.MethodPut, http.MethodPatch:
		h.update(w, r, userID, addressID)
	case http.MethodDelete:
		h.delete(w, r, userID, addressID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *AddressHandler) list(w http.ResponseWriter, r *http.Request, userID string) {
	addresses, err := h.service.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al listar direcciones")
		return
	}

	if addresses == nil {
		addresses = []address.Address{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addresses)
}

func (h *AddressHandler) create(w http.ResponseWriter, r *http.Request, userID string) {
	var req address.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	addr, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, address.ErrValidation) {
			writeError(w, http.StatusBadRequest, "Datos de dirección inválidos")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al crear dirección")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(addr)
}

func (h *AddressHandler) get(w http.ResponseWriter, r *http.Request, userID string, addressID int) {
	addr, err := h.service.GetByID(r.Context(), addressID)
	if err != nil {
		if errors.Is(err, address.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Dirección no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener dirección")
		return
	}

	// Verificar que la dirección pertenece al usuario
	if addr.UserID != userID {
		writeError(w, http.StatusNotFound, "Dirección no encontrada")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addr)
}

func (h *AddressHandler) update(w http.ResponseWriter, r *http.Request, userID string, addressID int) {
	var req address.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	addr, err := h.service.Update(r.Context(), addressID, userID, req)
	if err != nil {
		if errors.Is(err, address.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Dirección no encontrada")
			return
		}
		if errors.Is(err, address.ErrValidation) {
			writeError(w, http.StatusBadRequest, "Datos de dirección inválidos")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al actualizar dirección")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addr)
}

func (h *AddressHandler) delete(w http.ResponseWriter, r *http.Request, userID string, addressID int) {
	err := h.service.Delete(r.Context(), addressID, userID)
	if err != nil {
		if errors.Is(err, address.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Dirección no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al eliminar dirección")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AddressHandler) setDefault(w http.ResponseWriter, r *http.Request, userID string, addressID int) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	err := h.service.SetDefault(r.Context(), userID, addressID)
	if err != nil {
		if errors.Is(err, address.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Dirección no encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al establecer dirección predeterminada")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
