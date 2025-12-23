package sequence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Store interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	store Store
}

func NewRepository(store Store) *Repository {
	return &Repository{store: store}
}

// NextSequence atomically increments and returns the next sequence for a partition.
func (r *Repository) NextSequence(ctx context.Context, partitionKey string) (int64, error) {
	var seq int64
	err := r.store.QueryRow(ctx, `
		INSERT INTO event_sequence (partition_key, last_sequence)
		VALUES ($1, 1)
		ON CONFLICT (partition_key)
		DO UPDATE SET last_sequence = event_sequence.last_sequence + 1, updated_at = now()
		RETURNING last_sequence
	`, partitionKey).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("next sequence: %w", err)
	}
	return seq, nil
}
