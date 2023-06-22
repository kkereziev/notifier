package entity

import "github.com/google/uuid"

// IdempotencyKey is a struct that holds the idempotency key.
type IdempotencyKey struct {
	ID uuid.UUID `json:"id"`
}
