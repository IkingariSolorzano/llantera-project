package order

import (
	"context"
	"log"

	"github.com/llantera/hex/internal/domain/order"
	"github.com/llantera/hex/internal/domain/tire"
)

// Service define las operaciones de negocio para pedidos
type Service struct {
	repo          order.Repository
	inventoryRepo tire.InventoryRepository
}

// NewService crea una nueva instancia del servicio
func NewService(repo order.Repository, inventoryRepo tire.InventoryRepository) *Service {
	return &Service{repo: repo, inventoryRepo: inventoryRepo}
}

// Create crea un nuevo pedido
func (s *Service) Create(ctx context.Context, userID string, req order.CreateOrderRequest) (*order.Order, error) {
	if len(req.Items) == 0 {
		return nil, order.ErrEmptyCart
	}

	if !req.PaymentMethod.IsValid() {
		return nil, order.ErrValidation
	}

	// Validar modalidad de pago (usar contado por defecto)
	paymentMode := req.PaymentMode
	if paymentMode == "" || !paymentMode.IsValid() {
		paymentMode = order.PaymentModeContado
	}

	// Construir el pedido
	o := &order.Order{
		UserID:              userID,
		Status:              order.StatusSolicitado,
		ShippingAddress:     req.ShippingAddress,
		PaymentMethod:       req.PaymentMethod,
		PaymentMode:         paymentMode,
		PaymentInstallments: req.PaymentInstallments,
		PaymentNotes:        req.PaymentNotes,
		RequiresInvoice:     req.RequiresInvoice,
		BillingInfo:         req.BillingInfo,
		CustomerNotes:       req.CustomerNotes,
		Items:               make([]order.OrderItem, len(req.Items)),
	}

	// Convertir items y calcular totales
	var subtotal float64
	for i, item := range req.Items {
		itemSubtotal := float64(item.Quantity) * item.UnitPrice
		o.Items[i] = order.OrderItem{
			TireSKU:     item.TireSKU,
			TireMeasure: item.TireMeasure,
			TireBrand:   item.TireBrand,
			TireModel:   item.TireModel,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    itemSubtotal,
		}
		subtotal += itemSubtotal
	}

	// Usar totales del request si vienen (calculados en frontend con IVA)
	if req.Subtotal > 0 {
		o.Subtotal = req.Subtotal
	} else {
		o.Subtotal = subtotal
	}
	if req.IVA > 0 {
		o.IVA = req.IVA
	}
	if req.Total > 0 {
		o.Total = req.Total
	} else {
		o.Total = subtotal
	}
	o.ShippingCost = 0 // Por ahora envío gratis

	// Guardar en base de datos
	err := s.repo.Create(ctx, o)
	if err != nil {
		return nil, err
	}

	// Reservar stock para cada item
	if s.inventoryRepo != nil {
		for _, item := range o.Items {
			if err := s.inventoryRepo.ReserveStock(ctx, item.TireSKU, item.Quantity); err != nil {
				log.Printf("Error reservando stock para %s: %v", item.TireSKU, err)
			}
		}
	}

	return o, nil
}

// GetByID obtiene un pedido por su ID
func (s *Service) GetByID(ctx context.Context, id int) (*order.Order, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByOrderNumber obtiene un pedido por su número
func (s *Service) GetByOrderNumber(ctx context.Context, orderNumber string) (*order.Order, error) {
	return s.repo.GetByOrderNumber(ctx, orderNumber)
}

// List lista pedidos con filtros
func (s *Service) List(ctx context.Context, filter order.ListFilter) ([]order.Order, int, error) {
	return s.repo.List(ctx, filter)
}

// ListByUser lista los pedidos de un usuario
func (s *Service) ListByUser(ctx context.Context, userID string, limit, offset int) ([]order.Order, int, error) {
	filter := order.ListFilter{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}
	return s.repo.List(ctx, filter)
}

// UpdateStatus actualiza el estado de un pedido
func (s *Service) UpdateStatus(ctx context.Context, id int, req order.UpdateStatusRequest) error {
	if !req.Status.IsValid() {
		return order.ErrInvalidStatus
	}

	// Obtener pedido actual para validar transición
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !current.Status.CanTransitionTo(req.Status) {
		return order.ErrInvalidStatus
	}

	// Actualizar estado en BD
	if err := s.repo.UpdateStatus(ctx, id, req.Status, req.AdminNotes); err != nil {
		return err
	}

	// Manejar inventario según el nuevo estado
	if s.inventoryRepo != nil {
		switch req.Status {
		case order.StatusCancelado:
			// Liberar stock reservado
			for _, item := range current.Items {
				if err := s.inventoryRepo.ReleaseStock(ctx, item.TireSKU, item.Quantity); err != nil {
					log.Printf("Error liberando stock para %s: %v", item.TireSKU, err)
				}
			}
		case order.StatusEntregado:
			// Confirmar venta (restar de cantidad y apartadas)
			for _, item := range current.Items {
				if err := s.inventoryRepo.ConfirmSale(ctx, item.TireSKU, item.Quantity); err != nil {
					log.Printf("Error confirmando venta para %s: %v", item.TireSKU, err)
				}
			}
		}
	}

	return nil
}

// UpdateInvoiceFiles actualiza los archivos de factura
func (s *Service) UpdateInvoiceFiles(ctx context.Context, id int, xmlPath, pdfPath string) error {
	return s.repo.UpdateInvoiceFiles(ctx, id, xmlPath, pdfPath)
}
