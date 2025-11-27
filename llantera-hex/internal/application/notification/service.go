package notificationapp

import (
	"context"
	"fmt"

	"github.com/llantera/hex/internal/domain/notification"
)

type Service struct {
	repo notification.Repository
}

func NewService(repo notification.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, n *notification.Notification) (*notification.Notification, error) {
	return s.repo.Create(ctx, n)
}

func (s *Service) List(ctx context.Context, filter notification.NotificationFilter) ([]*notification.Notification, int, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) MarkAsRead(ctx context.Context, id int, userID string) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}

func (s *Service) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *Service) CountUnread(ctx context.Context, userID string) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, id int, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

// NotifyOrderCreated crea una notificación cuando se crea un pedido
func (s *Service) NotifyOrderCreated(ctx context.Context, userID string, orderNumber string) error {
	n := &notification.Notification{
		UserID:  userID,
		Type:    notification.TypeOrderCreated,
		Title:   "Pedido recibido",
		Message: fmt.Sprintf("Tu pedido %s ha sido recibido y está siendo procesado.", orderNumber),
	}
	_, err := s.repo.Create(ctx, n)
	return err
}

// NotifyOrderShipped crea una notificación cuando se envía un pedido
func (s *Service) NotifyOrderShipped(ctx context.Context, userID string, orderNumber string) error {
	n := &notification.Notification{
		UserID:  userID,
		Type:    notification.TypeOrderShipped,
		Title:   "Pedido enviado",
		Message: fmt.Sprintf("Tu pedido %s ha sido enviado.", orderNumber),
	}
	_, err := s.repo.Create(ctx, n)
	return err
}

// NotifyOrderDelivered crea una notificación cuando se entrega un pedido
func (s *Service) NotifyOrderDelivered(ctx context.Context, userID string, orderNumber string) error {
	n := &notification.Notification{
		UserID:  userID,
		Type:    notification.TypeOrderDelivered,
		Title:   "Pedido entregado",
		Message: fmt.Sprintf("Tu pedido %s ha sido entregado.", orderNumber),
	}
	_, err := s.repo.Create(ctx, n)
	return err
}

// NotifyInvoiceReady crea una notificación cuando la factura está lista
func (s *Service) NotifyInvoiceReady(ctx context.Context, userID string, orderNumber string) error {
	n := &notification.Notification{
		UserID:  userID,
		Type:    notification.TypeInvoiceReady,
		Title:   "Factura disponible",
		Message: fmt.Sprintf("La factura de tu pedido %s ya está disponible para descargar.", orderNumber),
	}
	_, err := s.repo.Create(ctx, n)
	return err
}

// NotifyOrderCancelled crea una notificación cuando se cancela un pedido
func (s *Service) NotifyOrderCancelled(ctx context.Context, userID string, orderNumber string) error {
	n := &notification.Notification{
		UserID:  userID,
		Type:    notification.TypeOrderCancelled,
		Title:   "Pedido cancelado",
		Message: fmt.Sprintf("Tu pedido %s ha sido cancelado.", orderNumber),
	}
	_, err := s.repo.Create(ctx, n)
	return err
}

// NotifyAdminsNewOrder notifica a los admins sobre un nuevo pedido
func (s *Service) NotifyAdminsNewOrder(ctx context.Context, adminUserIDs []string, orderNumber string, customerName string) error {
	for _, adminID := range adminUserIDs {
		n := &notification.Notification{
			UserID:  adminID,
			Type:    notification.TypeOrderCreated,
			Title:   "Nuevo pedido",
			Message: fmt.Sprintf("Nuevo pedido %s de %s", orderNumber, customerName),
		}
		if _, err := s.repo.Create(ctx, n); err != nil {
			return err
		}
	}
	return nil
}
