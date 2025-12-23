package events

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/dedup"
	"github.com/andreasstove999/ecommerce-system/services/inventory-service-go/internal/inventory"
)

func TestParseOrderCreatedEnvelopeExample(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "contracts", "examples", "order", "OrderCreated.v1.json"))
	if err != nil {
		t.Fatalf("read example: %v", err)
	}

	msg, err := parseOrderCreated(body, true)
	if err != nil {
		t.Fatalf("parse example: %v", err)
	}

	if msg.Envelope == nil {
		t.Fatalf("expected envelope")
	}
	if msg.Envelope.EventName != EventTypeOrderCreated || msg.Envelope.EventVersion != 1 {
		t.Fatalf("unexpected envelope metadata %+v", msg.Envelope)
	}
	if msg.Payload.OrderID != "f1e2d3c4-b5a6-7988-99aa-bbccddeeff00" {
		t.Fatalf("unexpected order id %s", msg.Payload.OrderID)
	}
	if len(msg.Payload.Items) != 2 {
		t.Fatalf("unexpected items len %d", len(msg.Payload.Items))
	}
}

func TestOrderCreatedHandlerDedupAndGap(t *testing.T) {
	store := newFakeStore(map[string]int{
		"p1": 5,
	})
	repo := &fakeTransactionalRepo{store: store}
	pub := &capturingPublisher{}
	dedupRepo := dedup.NewRepository(nil)

	handler := OrderCreatedHandler(repo, dedupRepo, pub, log.New(os.Stdout, "", 0), orderCreatedConsumerName, true)

	msg := makeOrderCreatedMessage("order-1", "user-1", "p1", 2, 1)
	body, _ := json.Marshal(msg)

	if err := handler(context.Background(), body); err != nil {
		t.Fatalf("first handle: %v", err)
	}
	if pub.reservedCalls != 1 {
		t.Fatalf("reserved calls=%d want=1", pub.reservedCalls)
	}
	if store.available["p1"] != 3 {
		t.Fatalf("available after first=%d want=3", store.available["p1"])
	}

	// duplicate sequence should be ignored
	if err := handler(context.Background(), body); err != nil {
		t.Fatalf("second duplicate handle: %v", err)
	}
	if pub.reservedCalls != 1 {
		t.Fatalf("reserved calls after duplicate=%d want=1", pub.reservedCalls)
	}
	if store.available["p1"] != 3 {
		t.Fatalf("available after duplicate=%d want=3", store.available["p1"])
	}

	// higher sequence with gap should process
	msgGap := makeOrderCreatedMessage("order-1", "user-1", "p1", 1, 3)
	bodyGap, _ := json.Marshal(msgGap)
	if err := handler(context.Background(), bodyGap); err != nil {
		t.Fatalf("gap handle: %v", err)
	}
	if pub.reservedCalls != 2 {
		t.Fatalf("reserved calls after gap=%d want=2", pub.reservedCalls)
	}
	if store.available["p1"] != 2 {
		t.Fatalf("available after gap=%d want=2", store.available["p1"])
	}

	lastSeq := store.checkpoints[orderCreatedConsumerName]["order-1"]
	if lastSeq != 3 {
		t.Fatalf("checkpoint=%d want=3", lastSeq)
	}
}

func TestOrderCreatedHandlerPropagatesCorrelation(t *testing.T) {
	store := newFakeStore(map[string]int{
		"p1": 2,
	})
	repo := &fakeTransactionalRepo{store: store}
	pub := &capturingPublisher{}
	dedupRepo := dedup.NewRepository(nil)

	handler := OrderCreatedHandler(repo, dedupRepo, pub, log.New(os.Stdout, "", 0), orderCreatedConsumerName, true)

	correlation := uuid.NewString()
	eventID := uuid.NewString()
	msg := EnvelopedOrderCreated{
		EventEnvelope: EventEnvelope{
			EventName:     EventTypeOrderCreated,
			EventVersion:  1,
			EventID:       eventID,
			CorrelationID: correlation,
			Producer:      "order-service",
			PartitionKey:  "order-2",
			Sequence:      10,
			OccurredAt:    time.Now().UTC(),
			Schema:        "contracts/events/order/OrderCreated.v1.payload.schema.json",
		},
		Payload: OrderCreatedPayload{
			OrderID:   "order-2",
			UserID:    "user-2",
			Items:     []OrderLineItem{{ProductID: "p1", Quantity: 1}},
			Timestamp: time.Now().UTC(),
		},
	}
	body, _ := json.Marshal(msg)

	if err := handler(context.Background(), body); err != nil {
		t.Fatalf("handle: %v", err)
	}

	if pub.lastMeta.CorrelationID != correlation {
		t.Fatalf("correlation=%s want=%s", pub.lastMeta.CorrelationID, correlation)
	}
	if pub.lastMeta.CausationID != eventID {
		t.Fatalf("causation=%s want=%s", pub.lastMeta.CausationID, eventID)
	}
	if pub.lastMeta.PartitionKey != msg.Payload.OrderID {
		t.Fatalf("partitionKey=%s want=%s", pub.lastMeta.PartitionKey, msg.Payload.OrderID)
	}
}

func makeOrderCreatedMessage(orderID, userID, productID string, quantity int, seq int64) EnvelopedOrderCreated {
	return EnvelopedOrderCreated{
		EventEnvelope: EventEnvelope{
			EventName:     EventTypeOrderCreated,
			EventVersion:  1,
			EventID:       uuid.NewString(),
			CorrelationID: uuid.NewString(),
			Producer:      "order-service",
			PartitionKey:  orderID,
			Sequence:      seq,
			OccurredAt:    time.Now().UTC(),
			Schema:        "contracts/events/order/OrderCreated.v1.payload.schema.json",
		},
		Payload: OrderCreatedPayload{
			OrderID:   orderID,
			UserID:    userID,
			Items:     []OrderLineItem{{ProductID: productID, Quantity: quantity}},
			Timestamp: time.Now().UTC(),
		},
	}
}

type capturingPublisher struct {
	reservedCalls int
	depletedCalls int
	lastMeta      EventMeta
	lastOrderID   string
	lastUserID    string
	lastReserved  []inventory.Line
	lastDepleted  []inventory.DepletedLine
}

func (f *capturingPublisher) PublishStockReserved(ctx context.Context, meta EventMeta, orderID, userID string, reserved []inventory.Line) error {
	f.reservedCalls++
	f.lastMeta = meta
	f.lastOrderID = orderID
	f.lastUserID = userID
	f.lastReserved = append([]inventory.Line(nil), reserved...)
	return nil
}

func (f *capturingPublisher) PublishStockDepleted(ctx context.Context, meta EventMeta, orderID, userID string, depleted []inventory.DepletedLine, reserved []inventory.Line) error {
	f.depletedCalls++
	f.lastMeta = meta
	f.lastOrderID = orderID
	f.lastUserID = userID
	f.lastDepleted = append([]inventory.DepletedLine(nil), depleted...)
	f.lastReserved = append([]inventory.Line(nil), reserved...)
	return nil
}

// --- fakes for transactional repo + tx ---

type fakeStore struct {
	available   map[string]int
	checkpoints map[string]map[string]int64
}

func newFakeStore(avail map[string]int) *fakeStore {
	cp := make(map[string]int, len(avail))
	for k, v := range avail {
		cp[k] = v
	}
	return &fakeStore{
		available:   cp,
		checkpoints: make(map[string]map[string]int64),
	}
}

type fakeTransactionalRepo struct {
	store      *fakeStore
	reserveErr error
}

func (r *fakeTransactionalRepo) Get(ctx context.Context, productID string) (inventory.StockItem, error) {
	return inventory.StockItem{}, nil
}

func (r *fakeTransactionalRepo) SetAvailable(ctx context.Context, productID string, available int) error {
	r.store.available[productID] = available
	return nil
}

func (r *fakeTransactionalRepo) Reserve(ctx context.Context, orderID string, lines []inventory.Line) (inventory.ReserveResult, error) {
	tx, _ := r.BeginTx(ctx, pgx.TxOptions{})
	res, err := r.ReserveWithTx(ctx, tx, orderID, lines)
	if err == nil {
		_ = tx.Commit(ctx)
	}
	return res, err
}

func (r *fakeTransactionalRepo) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return newFakeTx(r.store), nil
}

func (r *fakeTransactionalRepo) ReserveWithTx(ctx context.Context, tx pgx.Tx, orderID string, lines []inventory.Line) (inventory.ReserveResult, error) {
	if r.reserveErr != nil {
		return inventory.ReserveResult{}, r.reserveErr
	}
	fTx := tx.(*fakeTx)
	return fTx.reserve(lines), nil
}

type fakeTx struct {
	store              *fakeStore
	pendingAvailable   map[string]int
	pendingCheckpoints map[string]map[string]int64
	closed             bool
}

func newFakeTx(store *fakeStore) *fakeTx {
	return &fakeTx{
		store:              store,
		pendingAvailable:   make(map[string]int),
		pendingCheckpoints: make(map[string]map[string]int64),
	}
}

func (t *fakeTx) Begin(ctx context.Context) (pgx.Tx, error) { return t, nil }
func (t *fakeTx) Commit(ctx context.Context) error {
	if t.closed {
		return pgx.ErrTxClosed
	}
	for k, v := range t.pendingAvailable {
		t.store.available[k] = v
	}
	for consumer, parts := range t.pendingCheckpoints {
		if _, ok := t.store.checkpoints[consumer]; !ok {
			t.store.checkpoints[consumer] = make(map[string]int64)
		}
		for pk, seq := range parts {
			t.store.checkpoints[consumer][pk] = seq
		}
	}
	t.closed = true
	return nil
}
func (t *fakeTx) Rollback(ctx context.Context) error {
	t.closed = true
	return nil
}
func (t *fakeTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (t *fakeTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return &fakeBatchResults{}
}
func (t *fakeTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }
func (t *fakeTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (t *fakeTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if len(arguments) == 3 {
		consumer, _ := arguments[0].(string)
		partition, _ := arguments[1].(string)
		seq, _ := arguments[2].(int64)
		if _, ok := t.pendingCheckpoints[consumer]; !ok {
			t.pendingCheckpoints[consumer] = make(map[string]int64)
		}
		existing := t.store.checkpoints[consumer][partition]
		if seq < existing {
			seq = existing
		}
		if pending := t.pendingCheckpoints[consumer][partition]; pending > seq {
			seq = pending
		}
		t.pendingCheckpoints[consumer][partition] = seq
	}
	return pgconn.CommandTag{}, nil
}
func (t *fakeTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}
func (t *fakeTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	row := &fakeRow{}
	if len(args) == 2 {
		consumer, _ := args[0].(string)
		partition, _ := args[1].(string)
		if parts, ok := t.pendingCheckpoints[consumer]; ok {
			if seq, ok := parts[partition]; ok {
				row.val = seq
				return row
			}
		}
		if seq, ok := t.store.checkpoints[consumer][partition]; ok {
			row.val = seq
			return row
		}
		row.err = pgx.ErrNoRows
	}
	return row
}
func (t *fakeTx) Conn() *pgx.Conn { return nil }

func (t *fakeTx) reserve(lines []inventory.Line) inventory.ReserveResult {
	res := inventory.ReserveResult{}
	working := make(map[string]int)
	for k, v := range t.store.available {
		working[k] = v
	}

	for _, line := range lines {
		available := working[line.ProductID]
		if available < line.Quantity {
			res.Depleted = append(res.Depleted, inventory.DepletedLine{
				ProductID: line.ProductID,
				Requested: line.Quantity,
				Available: available,
			})
		} else {
			working[line.ProductID] = available - line.Quantity
			res.Reserved = append(res.Reserved, inventory.Line{ProductID: line.ProductID, Quantity: line.Quantity})
		}
	}

	if len(res.Depleted) == 0 {
		for k, v := range working {
			t.pendingAvailable[k] = v
		}
	}
	return res
}

type fakeRow struct {
	val int64
	err error
}

func (r *fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) == 1 {
		switch d := dest[0].(type) {
		case *int64:
			*d = r.val
		}
	}
	return nil
}

type fakeBatchResults struct{}

func (f *fakeBatchResults) Exec() (pgconn.CommandTag, error) { return pgconn.CommandTag{}, nil }
func (f *fakeBatchResults) Query() (pgx.Rows, error)         { return nil, nil }
func (f *fakeBatchResults) QueryRow() pgx.Row                { return &fakeRow{} }
func (f *fakeBatchResults) Close() error                     { return nil }
