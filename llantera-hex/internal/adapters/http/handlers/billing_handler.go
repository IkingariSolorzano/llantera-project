package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/llantera/hex/internal/adapters/http/middleware"
	appBilling "github.com/llantera/hex/internal/application/billing"
	"github.com/llantera/hex/internal/domain/billing"
)

type BillingHandler struct {
	service *appBilling.Service
}

func NewBillingHandler(service *appBilling.Service) *BillingHandler {
	return &BillingHandler{service: service}
}

// HandleCollection maneja operaciones sobre la colección de datos de facturación
func (h *BillingHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
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

// HandleResource maneja operaciones sobre datos de facturación específicos
func (h *BillingHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "No autorizado")
		return
	}

	// Extraer ID del path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		writeError(w, http.StatusBadRequest, "ID de datos de facturación requerido")
		return
	}

	lastPart := parts[len(parts)-1]

	// Verificar si es la acción set-default
	if lastPart == "set-default" && len(parts) >= 4 {
		billingID, err := strconv.Atoi(parts[len(parts)-2])
		if err != nil {
			writeError(w, http.StatusBadRequest, "ID de datos de facturación inválido")
			return
		}
		h.setDefault(w, r, userID, billingID)
		return
	}

	// Verificar si es la acción default (obtener el default)
	if lastPart == "default" {
		h.getDefault(w, r, userID)
		return
	}

	billingID, err := strconv.Atoi(lastPart)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID de datos de facturación inválido")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, r, userID, billingID)
	case http.MethodPut, http.MethodPatch:
		h.update(w, r, userID, billingID)
	case http.MethodDelete:
		h.delete(w, r, userID, billingID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *BillingHandler) list(w http.ResponseWriter, r *http.Request, userID string) {
	infos, err := h.service.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al listar datos de facturación")
		return
	}

	if infos == nil {
		infos = []billing.BillingInfo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(infos)
}

func (h *BillingHandler) create(w http.ResponseWriter, r *http.Request, userID string) {
	var req billing.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	info, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, billing.ErrValidation) {
			writeError(w, http.StatusBadRequest, "Datos de facturación inválidos")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al crear datos de facturación")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(info)
}

func (h *BillingHandler) get(w http.ResponseWriter, r *http.Request, userID string, billingID int) {
	info, err := h.service.GetByID(r.Context(), billingID)
	if err != nil {
		if errors.Is(err, billing.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Datos de facturación no encontrados")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener datos de facturación")
		return
	}

	// Verificar que los datos pertenecen al usuario
	if info.UserID != userID {
		writeError(w, http.StatusNotFound, "Datos de facturación no encontrados")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (h *BillingHandler) getDefault(w http.ResponseWriter, r *http.Request, userID string) {
	info, err := h.service.GetDefaultByUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, billing.ErrNotFound) {
			// No hay datos de facturación predeterminados, devolver null
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("null"))
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener datos de facturación")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (h *BillingHandler) update(w http.ResponseWriter, r *http.Request, userID string, billingID int) {
	var req billing.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	info, err := h.service.Update(r.Context(), billingID, userID, req)
	if err != nil {
		if errors.Is(err, billing.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Datos de facturación no encontrados")
			return
		}
		if errors.Is(err, billing.ErrValidation) {
			writeError(w, http.StatusBadRequest, "Datos de facturación inválidos")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al actualizar datos de facturación")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (h *BillingHandler) delete(w http.ResponseWriter, r *http.Request, userID string, billingID int) {
	err := h.service.Delete(r.Context(), billingID, userID)
	if err != nil {
		if errors.Is(err, billing.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Datos de facturación no encontrados")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al eliminar datos de facturación")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BillingHandler) setDefault(w http.ResponseWriter, r *http.Request, userID string, billingID int) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	err := h.service.SetDefault(r.Context(), userID, billingID)
	if err != nil {
		if errors.Is(err, billing.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Datos de facturación no encontrados")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al establecer datos de facturación predeterminados")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
