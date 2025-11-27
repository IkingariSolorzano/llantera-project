package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/llantera/hex/internal/adapters/http/middleware"
	cartApp "github.com/llantera/hex/internal/application/cart"
	"github.com/llantera/hex/internal/domain/cart"
)

type CartHandler struct {
	service *cartApp.Service
}

func NewCartHandler(service *cartApp.Service) *CartHandler {
	return &CartHandler{service: service}
}

// GetCart obtiene el carrito del usuario actual
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	// Obtener nivel de precio del usuario (del contexto o query param)
	priceLevel := r.URL.Query().Get("level")
	if priceLevel == "" {
		priceLevel = "public"
	}

	cartData, err := h.service.GetCart(r.Context(), userID, priceLevel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cartData)
}

// AddItem agrega un item al carrito
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	var req cart.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	item, err := h.service.AddItem(r.Context(), userID, req)
	if err != nil {
		if err == cart.ErrInvalidQuantity {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// HandleCollection maneja GET y POST sobre /api/cart/
func (h *CartHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetCart(w, r)
	case http.MethodPost:
		h.AddItem(w, r)
	case http.MethodDelete:
		h.ClearCart(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// HandleResource maneja PUT y DELETE sobre /api/cart/items/{sku}
func (h *CartHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	// Extraer SKU del path: /api/cart/items/{sku}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "SKU requerido", http.StatusBadRequest)
		return
	}
	tireSKU := parts[len(parts)-1]
	if tireSKU == "" {
		http.Error(w, "SKU requerido", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut, http.MethodPatch:
		h.updateItem(w, r, userID, tireSKU)
	case http.MethodDelete:
		h.removeItem(w, r, userID, tireSKU)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// updateItem actualiza la cantidad de un item
func (h *CartHandler) updateItem(w http.ResponseWriter, r *http.Request, userID, tireSKU string) {

	var req cart.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	item, err := h.service.UpdateItemQuantity(r.Context(), userID, tireSKU, req.Quantity)
	if err != nil {
		if err == cart.ErrItemNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == cart.ErrInvalidQuantity {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// removeItem elimina un item del carrito
func (h *CartHandler) removeItem(w http.ResponseWriter, r *http.Request, userID, tireSKU string) {
	err := h.service.RemoveItem(r.Context(), userID, tireSKU)
	if err != nil {
		if err == cart.ErrItemNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ClearCart vacía el carrito
func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "Usuario no autenticado", http.StatusUnauthorized)
		return
	}

	err := h.service.ClearCart(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
