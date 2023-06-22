package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/kkereziev/notifier/internal/entity"
)

// IdempotencyRepository is a repository for idempotency keys.
type IdempotencyRepository struct {
	conn *Connection
}

// NewIdempotencyStorage is a constructor function for IdempotencyRepository.
func NewIdempotencyStorage(conn *Connection) *IdempotencyRepository {
	return &IdempotencyRepository{conn: conn}
}

// GetByID returns an idempotency key by id.
func (r *IdempotencyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.IdempotencyKey, error) {
	var ent entity.IdempotencyKey

	err := r.conn.DB(ctx).GetContext(ctx, &ent, "SELECT * FROM idempotency_keys WHERE id = $1", id.String())
	if err != nil {
		return nil, err
	}

	return &ent, nil
}

// InsertOne inserts a new idempotency key in the database.
func (r *IdempotencyRepository) InsertOne(ctx context.Context, key *entity.IdempotencyKey) error {
	_, err := r.conn.DB(ctx).ExecContext(ctx, "INSERT INTO idempotency_keys VALUES ($1)", key.ID)

	return err
}
