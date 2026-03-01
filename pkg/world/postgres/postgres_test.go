package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/lord-server/panorama/pkg/geom"
	"github.com/lord-server/panorama/pkg/world/postgres"
)

func setupPostgresContainer(ctx context.Context, t *testing.T) (string, func()) {
	t.Helper()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/postgres:18",
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "test",
			},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready to accept connections").WithStartupTimeout(30*time.Second),
				wait.ForListeningPort("5432/tcp").WithStartupTimeout(30*time.Second),
			),
		},
		Started: true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	dsn := "postgres://test:test@" + host + ":" + port.Port() + "/test?sslmode=disable"

	db, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer func() { _ = db.Close(ctx) }()

	_, err = db.Exec(ctx, `
		CREATE TABLE blocks (
			posx INTEGER,
			posy INTEGER,
			posz INTEGER,
			data BYTEA NOT NULL,
			PRIMARY KEY (posx, posz, posy)
		);
	`)
	require.NoError(t, err)

	cleanup := func() {
		_ = container.Terminate(ctx)
	}

	return dsn, cleanup
}

func TestBackend_GetBlockData_NoRows(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	dsn, cleanup := setupPostgresContainer(ctx, t)
	defer cleanup()

	db, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer func() { _ = db.Close(ctx) }()

	backend, err := postgres.NewBackend(dsn)
	require.NoError(t, err)
	defer func() { require.NoError(t, backend.Close()) }()

	pos := geom.BlockPosition{X: 999, Y: 999, Z: 999}
	result, err := backend.GetBlockData(pos)
	require.NoError(t, err)

	assert.Nil(t, result)
}

func TestBackend_GetBlockData_MultipleBlocks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	dsn, cleanup := setupPostgresContainer(ctx, t)
	defer cleanup()

	db, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer func() { _ = db.Close(ctx) }()

	blocks := []struct {
		pos  geom.BlockPosition
		data []byte
	}{
		{pos: geom.BlockPosition{X: 0, Y: 0, Z: 0}, data: []byte{0x00}},
		{pos: geom.BlockPosition{X: 1, Y: 1, Z: 1}, data: []byte{0xde, 0xca, 0xff, 0xc0, 0xff, 0xee}},
		{pos: geom.BlockPosition{X: -1, Y: 5, Z: 10}, data: []byte{0xff, 0xfe, 0xfd}},
	}

	for _, b := range blocks {
		_, err = db.Exec(ctx,
			"INSERT INTO blocks (posx, posy, posz, data) VALUES ($1, $2, $3, $4)",
			b.pos.X, b.pos.Y, b.pos.Z, b.data,
		)
		require.NoError(t, err)
	}

	backend, err := postgres.NewBackend(dsn)
	require.NoError(t, err)
	defer func() { require.NoError(t, backend.Close()) }()

	for _, b := range blocks {
		result, err := backend.GetBlockData(b.pos)
		require.NoError(t, err)

		assert.Equal(t, b.data, result)
	}
}
