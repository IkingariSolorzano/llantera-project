package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/llantera/hex/internal/domain/notification"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *notification.Notification) (*notification.Notification, error) {
	query := `
		INSERT INTO notifications (user_id, type, title, message, data, read)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	// Si Data es nil o vacío, pasamos nil a la BD (JSONB acepta NULL)
	var dataJSON interface{}
	if len(n.Data) > 0 {
		dataJSON = n.Data
	}

	err := r.db.QueryRowContext(ctx, query,
		n.UserID,
		n.Type,
		n.Title,
		n.Message,
		dataJSON,
		n.Read,
	).Scan(&n.ID, &n.CreatedAt)

	if err != nil {
		return nil, err
	}

	return n, nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id int) (*notification.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, read, created_at
		FROM notifications
		WHERE id = $1
	`

	var n notification.Notification
	var dataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID,
		&n.UserID,
		&n.Type,
		&n.Title,
		&n.Message,
		&dataJSON,
		&n.Read,
		&n.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if dataJSON != nil {
		n.Data = json.RawMessage(dataJSON)
	}

	return &n, nil
}

func (r *NotificationRepository) List(ctx context.Context, filter notification.NotificationFilter) ([]*notification.Notification, int, error) {
	baseQuery := `FROM notifications WHERE user_id = $1`
	args := []interface{}{filter.UserID}
	argIndex := 2

	if filter.Unread {
		baseQuery += ` AND read = false`
	}

	// Count total
	var total int
	countQuery := `SELECT COUNT(*) ` + baseQuery
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get items
	query := `SELECT id, user_id, type, title, message, data, read, created_at ` + baseQuery + ` ORDER BY created_at DESC`

	if filter.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(` OFFSET $%d`, argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []*notification.Notification
	for rows.Next() {
		var n notification.Notification
		var dataJSON []byte

		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Message,
			&dataJSON,
			&n.Read,
			&n.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		if dataJSON != nil {
			n.Data = json.RawMessage(dataJSON)
		}

		notifications = append(notifications, &n)
	}

	return notifications, total, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id int, userID string) error {
	query := `UPDATE notifications SET read = true WHERE id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `UPDATE notifications SET read = true WHERE user_id = $1 AND read = false`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *NotificationRepository) Delete(ctx context.Context, id int, userID string) error {
	query := `DELETE FROM notifications WHERE id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}
