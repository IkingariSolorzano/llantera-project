package notification

import (
	"context"
	"encoding/json"
	"time"
)

// NotificationType representa los tipos de notificación
type NotificationType string

const (
	TypeOrderCreated   NotificationType = "order_created"
	TypeOrderUpdated   NotificationType = "order_updated"
	TypeOrderShipped   NotificationType = "order_shipped"
	TypeOrderDelivered NotificationType = "order_delivered"
	TypeOrderCancelled NotificationType = "order_cancelled"
	TypeInvoiceReady   NotificationType = "invoice_ready"
	TypeGeneral        NotificationType = "general"
)

// Notification representa una notificación del sistema
type Notification struct {
	ID        int              `json:"id"`
	UserID    string           `json:"userId"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Data      json.RawMessage  `json:"data,omitempty"`
	Read      bool             `json:"read"`
	CreatedAt time.Time        `json:"createdAt"`
}

// NotificationFilter para filtrar notificaciones
type NotificationFilter struct {
	UserID string
	Unread bool
	Limit  int
	Offset int
}

// Repository define las operaciones de persistencia
type Repository interface {
	Create(ctx context.Context, n *Notification) (*Notification, error)
	GetByID(ctx context.Context, id int) (*Notification, error)
	List(ctx context.Context, filter NotificationFilter) ([]*Notification, int, error)
	MarkAsRead(ctx context.Context, id int, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	CountUnread(ctx context.Context, userID string) (int, error)
	Delete(ctx context.Context, id int, userID string) error
}

// Service define los casos de uso
type Service interface {
	Create(ctx context.Context, n *Notification) (*Notification, error)
	List(ctx context.Context, filter NotificationFilter) ([]*Notification, int, error)
	MarkAsRead(ctx context.Context, id int, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	CountUnread(ctx context.Context, userID string) (int, error)
	Delete(ctx context.Context, id int, userID string) error

	// Helpers para crear notificaciones específicas
	NotifyOrderCreated(ctx context.Context, userID string, orderNumber string) error
	NotifyOrderShipped(ctx context.Context, userID string, orderNumber string) error
	NotifyOrderDelivered(ctx context.Context, userID string, orderNumber string) error
	NotifyOrderCancelled(ctx context.Context, userID string, orderNumber string) error
	NotifyInvoiceReady(ctx context.Context, userID string, orderNumber string) error
}
