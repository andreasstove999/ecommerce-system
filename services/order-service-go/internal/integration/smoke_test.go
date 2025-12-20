package integration

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/andreasstove999/ecommerce-system/order-service-go/internal/testutil"
)

func TestSmoke(t *testing.T) {
	t.Parallel()

	db, _ := testutil.StartPostgres(t)
	rows, err := db.Query("SELECT 1")
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())

	conn, _ := testutil.StartRabbitMQ(t)
	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()
}
