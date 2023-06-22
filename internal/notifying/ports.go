package notifying

import (
	"context"

	"github.com/kkereziev/notifier/v2/internal/entity"
)

// NotificationStorage is a storage for notifications.
type NotificationStorage interface {
	Update(ctx context.Context, notification *entity.Notification) error
	GetBatchByType(
		ctx context.Context,
		t entity.NotificationType,
		limit int,
	) ([]*entity.Notification, error)
}
