package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kkereziev/notifier/v2/internal/entity"
)

// NotificationRepository is a repository for notifications.
type NotificationRepository struct {
	conn *Connection
}

// NewNotificationRepository is a constructor function for NotificationRepository.
func NewNotificationRepository(conn *Connection) *NotificationRepository {
	return &NotificationRepository{conn: conn}
}

// GetBatchByType returns a batch of notifications by type.
func (r *NotificationRepository) GetBatchByType(
	ctx context.Context,
	t entity.NotificationType,
	limit int,
) ([]*entity.Notification, error) {
	var ent []*entity.Notification

	err := r.conn.DB(ctx).
		SelectContext(ctx,
			&ent,
			"SELECT * FROM notifications WHERE type = $1 AND status != 'SENT' LIMIT $2 FOR UPDATE SKIP LOCKED",
			t,
			limit,
		)
	if err != nil {
		return nil, err
	}

	return ent, nil
}

// Update updates the notification in the database.
func (r *NotificationRepository) Update(ctx context.Context, notification *entity.Notification) error {
	_, err := r.conn.DB(ctx).
		NamedExecContext(
			ctx, "UPDATE notifications SET status = :status, updated_at = :updated_at WHERE id = :id", notification,
		)

	return err
}

// InsertOne inserts a new notification in the database.
func (r *NotificationRepository) InsertOne(ctx context.Context, notification *entity.Notification) error {
	notification.CreatedAt = time.Now().UTC()
	notification.UpdatedAt = time.Now().UTC()
	notification.ID = uuid.New()

	_, err := r.conn.DB(ctx).ExecContext(
		ctx, "INSERT INTO notifications(id,type,status,data,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6)",
		notification.ID,
		notification.Type,
		notification.Status,
		notification.Data,
		notification.CreatedAt,
		notification.UpdatedAt,
	)

	return err
}
