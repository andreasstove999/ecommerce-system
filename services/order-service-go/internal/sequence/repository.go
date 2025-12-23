package sequence

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository manages producer-side sequences for events.
type Repository interface {
	NextSequence(ctx context.Context, partitionKey string) (int64, error)
}

type repo struct {
	db *sql.DB
}

// NewRepository creates a new sequence repository.
func NewRepository(db *sql.DB) Repository {
	return &repo{db: db}
}

func (r *repo) NextSequence(ctx context.Context, partitionKey string) (int64, error) {
	var seq int64
	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO event_sequence (partition_key, last_sequence, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (partition_key)
		DO UPDATE SET last_sequence = event_sequence.last_sequence + 1, updated_at = NOW()
		RETURNING last_sequence
	`, partitionKey).Scan(&seq); err != nil {
		return 0, fmt.Errorf("next sequence: %w", err)
	}
	return seq, nil
}
