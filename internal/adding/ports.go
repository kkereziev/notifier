package adding

import (
	"context"

	"github.com/google/uuid"
	"github.com/kkereziev/notifier/internal/entity"
)

// NotificationStorage is a storage for notifications.
type NotificationStorage interface {
	InsertOne(ctx context.Context, notification *entity.Notification) error
}

// IdempotencyStorage is a storage for idempotency keys.
type IdempotencyStorage interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.IdempotencyKey, error)
	InsertOne(ctx context.Context, key *entity.IdempotencyKey) error
}
