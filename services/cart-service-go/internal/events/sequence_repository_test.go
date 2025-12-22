package events

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

type fakeTxStarter struct {
	sequences map[string]int64
	failBegin bool
}

func (f *fakeTxStarter) BeginTx(ctx context.Context, opts *sql.TxOptions) (txRunner, error) {
	if f.failBegin {
		return nil, errors.New("begin failed")
	}
	return &fakeTx{starter: f}, nil
}

type fakeTx struct {
	starter *fakeTxStarter
}

func (f *fakeTx) QueryRowContext(ctx context.Context, query string, args ...any) rowScanner {
	partition := args[0].(string)
	f.starter.sequences[partition]++
	return fakeRow{value: f.starter.sequences[partition]}
}

func (f *fakeTx) Commit() error {
	return nil
}

func (f *fakeTx) Rollback() error {
	return nil
}

type fakeRow struct {
	value int64
}

func (f fakeRow) Scan(dest ...any) error {
	ptr, ok := dest[0].(*int64)
	if !ok {
		return errors.New("expected *int64 destination")
	}
	*ptr = f.value
	return nil
}

func TestNextSequenceIncrementsPerPartition(t *testing.T) {
	starter := &fakeTxStarter{sequences: make(map[string]int64)}
	repo := &sequenceRepository{db: starter}

	seq1, err := repo.NextSequence(context.Background(), "cart-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seq1 != 1 {
		t.Fatalf("expected first sequence to be 1, got %d", seq1)
	}

	seq2, err := repo.NextSequence(context.Background(), "cart-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seq2 != 2 {
		t.Fatalf("expected second sequence to be 2, got %d", seq2)
	}

	seqOther, err := repo.NextSequence(context.Background(), "cart-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seqOther != 1 {
		t.Fatalf("expected new partition to start at 1, got %d", seqOther)
	}

	starter.failBegin = true
	if _, err := repo.NextSequence(context.Background(), "cart-error"); err == nil {
		t.Fatalf("expected error when begin fails")
	}

	if _, err := repo.NextSequence(context.Background(), ""); err == nil {
		t.Fatalf("expected error for empty partition key")
	}
}
