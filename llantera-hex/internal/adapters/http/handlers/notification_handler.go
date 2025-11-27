package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/llantera/hex/internal/adapters/http/middleware"
	notificationapp "github.com/llantera/hex/internal/application/notification"
	"github.com/llantera/hex/internal/domain/notification"
)

type NotificationHandler struct {
	service *notificationapp.Service
}

func NewNotificationHandler(service *notificationapp.Service) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// HandleCollection maneja /api/notifications/
func (h *NotificationHandler) HandleCollection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "No autorizado")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.list(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

// HandleResource maneja /api/notifications/{id}
func (h *NotificationHandler) HandleResource(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "No autorizado")
		return
	}

	// Extraer ID
	path := strings.TrimPrefix(r.URL.Path, "/api/notifications/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "ID requerido")
		return
	}

	// Manejar /api/notifications/read-all
	if parts[0] == "read-all" && r.Method == http.MethodPost {
		h.markAllAsRead(w, r, userID)
		return
	}

	// Manejar /api/notifications/count
	if parts[0] == "count" && r.Method == http.MethodGet {
		h.countUnread(w, r, userID)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// Manejar /api/notifications/{id}/read
	if len(parts) > 1 && parts[1] == "read" && r.Method == http.MethodPost {
		h.markAsRead(w, r, id, userID)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		h.delete(w, r, id, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *NotificationHandler) list(w http.ResponseWriter, r *http.Request, userID string) {
	q := r.URL.Query()

	limit := 50
	if v := q.Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if v := q.Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	unreadOnly := q.Get("unread") == "true" || q.Get("unread") == "1"

	filter := notification.NotificationFilter{
		UserID: userID,
		Unread: unreadOnly,
		Limit:  limit,
		Offset: offset,
	}

	items, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener notificaciones")
		return
	}

	response := map[string]interface{}{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *NotificationHandler) markAsRead(w http.ResponseWriter, r *http.Request, id int, userID string) {
	if err := h.service.MarkAsRead(r.Context(), id, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "Error al marcar como leída")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) markAllAsRead(w http.ResponseWriter, r *http.Request, userID string) {
	if err := h.service.MarkAllAsRead(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "Error al marcar todas como leídas")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) countUnread(w http.ResponseWriter, r *http.Request, userID string) {
	count, err := h.service.CountUnread(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al contar notificaciones")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

func (h *NotificationHandler) delete(w http.ResponseWriter, r *http.Request, id int, userID string) {
	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "Error al eliminar notificación")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
