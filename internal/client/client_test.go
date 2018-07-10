package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/CanonicalLtd/dqlite/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Heartbeat(t *testing.T) {
	c, cleanup := newClient(t)
	defer cleanup()

	request := client.Message{}
	request.Init(512)
	response := client.Message{}
	response.Init(512)

	client.EncodeHeartbeat(&request, uint64(time.Now().Unix()))

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	err := c.Call(ctx, &request, &response)
	require.NoError(t, err)

	addresses, err := client.DecodeServers(&response)
	require.NoError(t, err)

	assert.Len(t, addresses, 2)
	assert.Equal(t, client.Strings{"1.2.3.4:666", "5.6.7.8:666"}, addresses)
}

/*
func TestClient_Prepare(t *testing.T) {
	client, cleanup := newClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	db, err := client.Open(ctx, "test.db", "volatile")
	require.NoError(t, err)

	stmt, err := client.Prepare(ctx, db.ID, "CREATE TABLE test (n INT)")
	require.NoError(t, err)

	assert.Equal(t, stmt.ID, uint32(0))
}

func TestClient_Exec(t *testing.T) {
	client, cleanup := newClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	db, err := client.Open(ctx, "test.db", "volatile")
	require.NoError(t, err)

	stmt, err := client.Prepare(ctx, db.ID, "CREATE TABLE test (n INT)")
	require.NoError(t, err)

	_, err = client.Exec(ctx, db.ID, stmt.ID)
	require.NoError(t, err)
}

func TestClient_Query(t *testing.T) {
	client, cleanup := newClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	db, err := client.Open(ctx, "test.db", "volatile")
	require.NoError(t, err)

	start := time.Now()

	stmt, err := client.Prepare(ctx, db.ID, "CREATE TABLE test (n INT)")
	require.NoError(t, err)

	_, err = client.Exec(ctx, db.ID, stmt.ID)
	require.NoError(t, err)

	_, err = client.Finalize(ctx, db.ID, stmt.ID)
	require.NoError(t, err)

	stmt, err = client.Prepare(ctx, db.ID, "INSERT INTO test VALUES(1)")
	require.NoError(t, err)

	_, err = client.Exec(ctx, db.ID, stmt.ID)
	require.NoError(t, err)

	_, err = client.Finalize(ctx, db.ID, stmt.ID)
	require.NoError(t, err)

	stmt, err = client.Prepare(ctx, db.ID, "SELECT n FROM test")
	require.NoError(t, err)

	_, err = client.Query(ctx, db.ID, stmt.ID)
	require.NoError(t, err)

	_, err = client.Finalize(ctx, db.ID, stmt.ID)
	require.NoError(t, err)

	fmt.Printf("time %s\n", time.Since(start))
}
*/

func newClient(t *testing.T) (*client.Client, func()) {
	t.Helper()

	listener := newListener(t)

	cluster := newTestCluster()
	cluster.leader = listener.Addr().String()

	serverCleanup := newServer(t, 0, listener, cluster)

	store := newStore(t, []string{cluster.leader})

	connector := newConnector(t, store)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	client, err := connector.Connect(ctx)

	require.NoError(t, err)

	cleanup := func() {
		client.Close()
		serverCleanup()
	}

	return client, cleanup
}
