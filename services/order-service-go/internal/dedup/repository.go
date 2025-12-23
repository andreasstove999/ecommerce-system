package dedup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Repository manages consumer-side deduplication checkpoints.
type Repository interface {
	GetLastSequence(ctx context.Context, consumerName, partitionKey string) (int64, bool, error)
	UpsertLastSequence(ctx context.Context, tx *sql.Tx, consumerName, partitionKey string, newSeq int64) error
}

type repo struct {
	db *sql.DB
}

// NewRepository creates a dedup repository.
func NewRepository(db *sql.DB) Repository {
	return &repo{db: db}
}

func (r *repo) GetLastSequence(ctx context.Context, consumerName, partitionKey string) (int64, bool, error) {
	var last int64
	err := r.db.QueryRowContext(ctx, `
		SELECT last_sequence
		FROM event_dedup_checkpoint
		WHERE consumer_name = $1 AND partition_key = $2
	`, consumerName, partitionKey).Scan(&last)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("select last_sequence: %w", err)
	}
	return last, true, nil
}

func (r *repo) UpsertLastSequence(ctx context.Context, tx *sql.Tx, consumerName, partitionKey string, newSeq int64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO event_dedup_checkpoint (consumer_name, partition_key, last_sequence, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (consumer_name, partition_key)
		DO UPDATE SET
			last_sequence = GREATEST(event_dedup_checkpoint.last_sequence, EXCLUDED.last_sequence),
			updated_at = NOW()
	`, consumerName, partitionKey, newSeq)
	if err != nil {
		return fmt.Errorf("upsert last_sequence: %w", err)
	}
	return nil
}
