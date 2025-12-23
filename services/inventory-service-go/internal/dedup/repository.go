package dedup

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Executor represents the subset of pgx methods required for dedup operations.
type Executor interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Repository struct {
	executor Executor
}

func NewRepository(exec Executor) *Repository {
	return &Repository{executor: exec}
}

// WithExecutor returns a shallow copy using the provided executor (e.g., a transaction).
func (r *Repository) WithExecutor(exec Executor) *Repository {
	return &Repository{executor: exec}
}

// GetLastSequence returns the last processed sequence for a consumer/partition.
// The boolean indicates whether a checkpoint existed.
func (r *Repository) GetLastSequence(ctx context.Context, consumerName, partitionKey string) (int64, bool, error) {
	var last int64
	if err := r.executor.QueryRow(ctx, `
		SELECT last_sequence
		FROM event_dedup_checkpoint
		WHERE consumer_name=$1 AND partition_key=$2
	`, consumerName, partitionKey).Scan(&last); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("select checkpoint: %w", err)
	}
	return last, true, nil
}

// UpsertLastSequence advances the checkpoint ensuring monotonic progress even under races.
func (r *Repository) UpsertLastSequence(ctx context.Context, consumerName, partitionKey string, newSeq int64) error {
	_, err := r.executor.Exec(ctx, `
		INSERT INTO event_dedup_checkpoint (consumer_name, partition_key, last_sequence)
		VALUES ($1, $2, $3)
		ON CONFLICT (consumer_name, partition_key)
		DO UPDATE SET
			last_sequence = GREATEST(event_dedup_checkpoint.last_sequence, EXCLUDED.last_sequence),
			updated_at = now()
	`, consumerName, partitionKey, newSeq)
	if err != nil {
		return fmt.Errorf("upsert checkpoint: %w", err)
	}
	return nil
}
