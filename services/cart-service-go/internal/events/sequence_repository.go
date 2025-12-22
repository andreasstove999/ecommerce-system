package events

import (
	"context"
	"database/sql"
	"fmt"
)

type sequenceRepository struct {
	db txStarter
}

func NewSequenceRepository(db *sql.DB) SequenceRepository {
	return &sequenceRepository{db: sqlTxStarter{db: db}}
}

func (r *sequenceRepository) NextSequence(ctx context.Context, partitionKey string) (int64, error) {
	if partitionKey == "" {
		return 0, fmt.Errorf("partition key is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const query = `
INSERT INTO event_sequences (partition_key, last_sequence, updated_at)
VALUES ($1, 1, NOW())
ON CONFLICT (partition_key) DO UPDATE
SET last_sequence = event_sequences.last_sequence + 1,
    updated_at = NOW()
RETURNING last_sequence
`

	var next int64
	if err = tx.QueryRowContext(ctx, query, partitionKey).Scan(&next); err != nil {
		return 0, fmt.Errorf("increment sequence: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return next, nil
}

type txStarter interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (txRunner, error)
}

type txRunner interface {
	QueryRowContext(ctx context.Context, query string, args ...any) rowScanner
	Commit() error
	Rollback() error
}

type rowScanner interface {
	Scan(dest ...any) error
}

type sqlTxStarter struct {
	db *sql.DB
}

func (s sqlTxStarter) BeginTx(ctx context.Context, opts *sql.TxOptions) (txRunner, error) {
	tx, err := s.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return sqlTx{tx: tx}, nil
}

type sqlTx struct {
	tx *sql.Tx
}

func (s sqlTx) QueryRowContext(ctx context.Context, query string, args ...any) rowScanner {
	return s.tx.QueryRowContext(ctx, query, args...)
}

func (s sqlTx) Commit() error {
	return s.tx.Commit()
}

func (s sqlTx) Rollback() error {
	return s.tx.Rollback()
}
