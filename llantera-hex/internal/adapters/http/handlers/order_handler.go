package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/llantera/hex/internal/adapters/http/middleware"
	notificationapp "github.com/llantera/hex/internal/application/notification"
	appOrder "github.com/llantera/hex/internal/application/order"
	"github.com/llantera/hex/internal/domain/order"
	"github.com/llantera/hex/internal/domain/user"
)

type OrderHandler struct {
	service             *appOrder.Service
	notificationService *notificationapp.Service
	userRepo            user.Repository
}

func NewOrderHandler(service *appOrder.Service, notificationService *notificationapp.Service, userRepo user.Repository) *OrderHandler {
	return &OrderHandler{
		service:             service,
		notificationService: notificationService,
		userRepo:            userRepo,
	}
}

// HandleCustomerOrders maneja las operaciones de pedidos para clientes
func (h *OrderHandler) HandleCustomerOrders(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "No autorizado")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listCustomerOrders(w, r, userID)
	case http.MethodPost:
		h.createOrder(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

// HandleCustomerOrderResource maneja operaciones sobre un pedido específico del cliente
func (h *OrderHandler) HandleCustomerOrderResource(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "No autorizado")
		return
	}

	// Extraer ID del path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		writeError(w, http.StatusBadRequest, "ID de pedido requerido")
		return
	}

	// Verificar si es una acción de status: /api/orders/{id}/status
	lastPart := parts[len(parts)-1]
	if lastPart == "status" && len(parts) >= 3 {
		orderID, err := strconv.Atoi(parts[len(parts)-2])
		if err != nil {
			writeError(w, http.StatusBadRequest, "ID de pedido inválido")
			return
		}
		h.cancelCustomerOrder(w, r, userID, orderID)
		return
	}

	orderID, err := strconv.Atoi(lastPart)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID de pedido inválido")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getCustomerOrder(w, r, userID, orderID)
	case http.MethodPatch:
		h.cancelCustomerOrder(w, r, userID, orderID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

// HandleAdminOrders maneja las operaciones de pedidos para administradores
func (h *OrderHandler) HandleAdminOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listAllOrders(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

// HandleAdminOrderResource maneja operaciones sobre un pedido específico para admin
func (h *OrderHandler) HandleAdminOrderResource(w http.ResponseWriter, r *http.Request) {
	// Extraer ID del path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeError(w, http.StatusBadRequest, "ID de pedido requerido")
		return
	}

	orderID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		// Podría ser una acción como /status
		if parts[len(parts)-1] == "status" && len(parts) >= 5 {
			orderID, err = strconv.Atoi(parts[len(parts)-2])
			if err != nil {
				writeError(w, http.StatusBadRequest, "ID de pedido inválido")
				return
			}
			h.updateOrderStatus(w, r, orderID)
			return
		}
		writeError(w, http.StatusBadRequest, "ID de pedido inválido")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getOrder(w, r, orderID)
	case http.MethodPatch:
		h.updateOrderStatus(w, r, orderID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

func (h *OrderHandler) createOrder(w http.ResponseWriter, r *http.Request, userID string) {
	var req order.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	o, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, order.ErrEmptyCart) {
			writeError(w, http.StatusBadRequest, "El carrito está vacío")
			return
		}
		if errors.Is(err, order.ErrValidation) {
			writeError(w, http.StatusBadRequest, "Datos de pedido inválidos")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al crear pedido")
		return
	}

	// Notificar al cliente y a los administradores sobre el nuevo pedido (usar contexto background)
	go h.notifyOrderCreated(o.OrderNumber, userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(o)
}

// notifyOrderCreated envía notificaciones al cliente y a todos los administradores
func (h *OrderHandler) notifyOrderCreated(orderNumber, customerUserID string) {
	if h.notificationService == nil || h.userRepo == nil {
		log.Printf("notifyOrderCreated: servicio de notificaciones o repo de usuarios es nil")
		return
	}

	// Usar contexto background para que no se cancele con la request HTTP
	ctx := context.Background()

	// Notificar al cliente que su pedido fue recibido
	if err := h.notificationService.NotifyOrderCreated(ctx, customerUserID, orderNumber); err != nil {
		log.Printf("Error notificando al cliente sobre pedido creado: %v", err)
	} else {
		log.Printf("Notificación enviada al cliente %s sobre pedido %s", customerUserID, orderNumber)
	}

	// Obtener nombre del cliente para notificar a admins
	customerName := "Cliente"
	if customer, err := h.userRepo.GetByID(ctx, customerUserID); err == nil && customer != nil {
		customerName = customer.FullName()
		if customerName == "" {
			customerName = customer.Email
		}
	}

	// Obtener todos los administradores
	adminRole := user.RoleAdmin
	admins, _, err := h.userRepo.List(ctx, user.ListFilter{Role: &adminRole, Limit: 100})
	if err != nil {
		log.Printf("Error obteniendo admins para notificación: %v", err)
		return
	}

	log.Printf("Notificando a %d administradores sobre pedido %s", len(admins), orderNumber)

	var adminIDs []string
	for _, admin := range admins {
		adminIDs = append(adminIDs, admin.ID)
	}

	if len(adminIDs) > 0 {
		if err := h.notificationService.NotifyAdminsNewOrder(ctx, adminIDs, orderNumber, customerName); err != nil {
			log.Printf("Error enviando notificaciones a admins: %v", err)
		} else {
			log.Printf("Notificaciones enviadas exitosamente a %d admins", len(adminIDs))
		}
	} else {
		log.Printf("No se encontraron administradores para notificar")
	}
}

func (h *OrderHandler) listCustomerOrders(w http.ResponseWriter, r *http.Request, userID string) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	orders, total, err := h.service.ListByUser(r.Context(), userID, limit, offset)
	if err != nil {
		log.Printf("Error listando pedidos del usuario %s: %v", userID, err)
		writeError(w, http.StatusInternalServerError, "Error al listar pedidos")
		return
	}

	response := map[string]interface{}{
		"items":  orders,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) getCustomerOrder(w http.ResponseWriter, r *http.Request, userID string, orderID int) {
	o, err := h.service.GetByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, order.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Pedido no encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener pedido")
		return
	}

	// Verificar que el pedido pertenece al usuario
	if o.UserID != userID {
		writeError(w, http.StatusNotFound, "Pedido no encontrado")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}

// cancelCustomerOrder permite a un cliente cancelar su propio pedido (solo si está en estado "solicitado")
func (h *OrderHandler) cancelCustomerOrder(w http.ResponseWriter, r *http.Request, userID string, orderID int) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	// Obtener el pedido
	o, err := h.service.GetByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, order.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Pedido no encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener pedido")
		return
	}

	// Verificar que el pedido pertenece al usuario
	if o.UserID != userID {
		writeError(w, http.StatusNotFound, "Pedido no encontrado")
		return
	}

	// Solo se puede cancelar si está en estado "solicitado"
	if o.Status != order.StatusSolicitado {
		writeError(w, http.StatusBadRequest, "Solo se pueden cancelar pedidos en estado 'solicitado'")
		return
	}

	// Leer el body para verificar que se solicita cancelación
	var req order.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	// Solo permitir cambio a "cancelado"
	if req.Status != order.StatusCancelado {
		writeError(w, http.StatusBadRequest, "Solo se permite cancelar el pedido")
		return
	}

	// Actualizar estado
	err = h.service.UpdateStatus(r.Context(), orderID, req)
	if err != nil {
		if errors.Is(err, order.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Pedido no encontrado")
			return
		}
		if errors.Is(err, order.ErrInvalidStatus) {
			writeError(w, http.StatusBadRequest, "Transición de estado no válida")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al cancelar pedido")
		return
	}

	// Obtener pedido actualizado
	updated, err := h.service.GetByID(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener pedido actualizado")
		return
	}

	// Notificar sobre la cancelación
	go h.notifyOrderStatusChange(updated.OrderNumber, updated.UserID, updated.Status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *OrderHandler) listAllOrders(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	search := r.URL.Query().Get("search")
	statusStr := r.URL.Query().Get("status")

	if limit <= 0 {
		limit = 20
	}

	filter := order.ListFilter{
		Search: search,
		Limit:  limit,
		Offset: offset,
	}

	if statusStr != "" {
		status := order.Status(statusStr)
		if status.IsValid() {
			filter.Status = &status
		}
	}

	orders, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		log.Printf("Error listando pedidos (admin): %v", err)
		writeError(w, http.StatusInternalServerError, "Error al listar pedidos")
		return
	}

	response := map[string]interface{}{
		"items":  orders,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) getOrder(w http.ResponseWriter, r *http.Request, orderID int) {
	o, err := h.service.GetByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, order.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Pedido no encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener pedido")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}

func (h *OrderHandler) updateOrderStatus(w http.ResponseWriter, r *http.Request, orderID int) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	var req order.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	err := h.service.UpdateStatus(r.Context(), orderID, req)
	if err != nil {
		if errors.Is(err, order.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Pedido no encontrado")
			return
		}
		if errors.Is(err, order.ErrInvalidStatus) {
			writeError(w, http.StatusBadRequest, "Transición de estado no válida")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al actualizar estado")
		return
	}

	// Obtener pedido actualizado
	o, err := h.service.GetByID(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener pedido actualizado")
		return
	}

	// Notificar al cliente sobre el cambio de estado
	go h.notifyOrderStatusChange(o.OrderNumber, o.UserID, o.Status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}

// notifyOrderStatusChange notifica al cliente cuando cambia el estado de su pedido
func (h *OrderHandler) notifyOrderStatusChange(orderNumber, userID string, newStatus order.Status) {
	if h.notificationService == nil {
		return
	}

	ctx := context.Background()

	switch newStatus {
	case order.StatusEnviado:
		if err := h.notificationService.NotifyOrderShipped(ctx, userID, orderNumber); err != nil {
			log.Printf("Error notificando pedido enviado: %v", err)
		}
	case order.StatusEntregado:
		if err := h.notificationService.NotifyOrderDelivered(ctx, userID, orderNumber); err != nil {
			log.Printf("Error notificando pedido entregado: %v", err)
		}
	case order.StatusCancelado:
		if err := h.notificationService.NotifyOrderCancelled(ctx, userID, orderNumber); err != nil {
			log.Printf("Error notificando pedido cancelado: %v", err)
		}
	}
}

// HandleUploadInvoice maneja la subida de archivos de factura (PDF/XML)
func (h *OrderHandler) HandleUploadInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	// Extraer ID del path: /api/admin/orders/{id}/invoice
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/orders/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "ID de pedido requerido")
		return
	}

	orderID, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID de pedido inválido")
		return
	}

	// Verificar que el pedido existe
	_, err = h.service.GetByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, order.ErrNotFound) {
			writeError(w, http.StatusNotFound, "Pedido no encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "Error al obtener pedido")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "Error al procesar el formulario")
		return
	}

	// Crear directorio de facturas si no existe
	invoicesDir := filepath.Join("uploads", "invoices", fmt.Sprintf("%d", orderID))
	if err := os.MkdirAll(invoicesDir, 0755); err != nil {
		log.Printf("Error creando directorio de facturas: %v", err)
		writeError(w, http.StatusInternalServerError, "Error al guardar archivo")
		return
	}

	var xmlPath, pdfPath string

	// Procesar archivo XML
	xmlFile, xmlHeader, err := r.FormFile("xml")
	if err == nil {
		defer xmlFile.Close()

		if !strings.HasSuffix(strings.ToLower(xmlHeader.Filename), ".xml") {
			writeError(w, http.StatusBadRequest, "El archivo XML debe tener extensión .xml")
			return
		}

		xmlFilename := fmt.Sprintf("factura_%d_%d.xml", orderID, time.Now().Unix())
		xmlPath = filepath.Join(invoicesDir, xmlFilename)

		dst, err := os.Create(xmlPath)
		if err != nil {
			log.Printf("Error creando archivo XML: %v", err)
			writeError(w, http.StatusInternalServerError, "Error al guardar archivo XML")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, xmlFile); err != nil {
			log.Printf("Error copiando archivo XML: %v", err)
			writeError(w, http.StatusInternalServerError, "Error al guardar archivo XML")
			return
		}
	}

	// Procesar archivo PDF
	pdfFile, pdfHeader, err := r.FormFile("pdf")
	if err == nil {
		defer pdfFile.Close()

		if !strings.HasSuffix(strings.ToLower(pdfHeader.Filename), ".pdf") {
			writeError(w, http.StatusBadRequest, "El archivo PDF debe tener extensión .pdf")
			return
		}

		pdfFilename := fmt.Sprintf("factura_%d_%d.pdf", orderID, time.Now().Unix())
		pdfPath = filepath.Join(invoicesDir, pdfFilename)

		dst, err := os.Create(pdfPath)
		if err != nil {
			log.Printf("Error creando archivo PDF: %v", err)
			writeError(w, http.StatusInternalServerError, "Error al guardar archivo PDF")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, pdfFile); err != nil {
			log.Printf("Error copiando archivo PDF: %v", err)
			writeError(w, http.StatusInternalServerError, "Error al guardar archivo PDF")
			return
		}
	}

	if xmlPath == "" && pdfPath == "" {
		writeError(w, http.StatusBadRequest, "Debe proporcionar al menos un archivo (XML o PDF)")
		return
	}

	// Actualizar rutas en la base de datos
	if err := h.service.UpdateInvoiceFiles(r.Context(), orderID, xmlPath, pdfPath); err != nil {
		log.Printf("Error actualizando rutas de factura: %v", err)
		writeError(w, http.StatusInternalServerError, "Error al actualizar pedido")
		return
	}

	// Obtener pedido actualizado
	updatedOrder, err := h.service.GetByID(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al obtener pedido actualizado")
		return
	}

	// Notificar al cliente que su factura está lista
	go func() {
		if h.notificationService != nil && updatedOrder.RequiresInvoice {
			ctx := context.Background()
			if err := h.notificationService.NotifyInvoiceReady(ctx, updatedOrder.UserID, updatedOrder.OrderNumber); err != nil {
				log.Printf("Error notificando factura lista: %v", err)
			}
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedOrder)
}
