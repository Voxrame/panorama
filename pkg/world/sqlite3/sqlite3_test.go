package sqlite3_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/lord-server/panorama/pkg/geom"
	"github.com/lord-server/panorama/pkg/world/sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackend_GetBlockData_NoRows(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	createSchema(t, db)

	backend, err := sqlite3.NewBackend(dbPath)
	require.NoError(t, err)
	defer func() { _ = backend.Close() }()

	pos := geom.BlockPosition{X: 999, Y: 999, Z: 999}
	result, err := backend.GetBlockData(pos)
	require.NoError(t, err)

	assert.Nil(t, result)
}

func TestBackend_GetBlockData_MultipleBlocks(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	createSchema(t, db)

	blocks := []struct {
		pos  geom.BlockPosition
		data []byte
	}{
		{pos: geom.BlockPosition{X: 0, Y: 0, Z: 0}, data: []byte{0x00}},
		{pos: geom.BlockPosition{X: 1, Y: 1, Z: 1}, data: []byte{0xde, 0xca, 0xff, 0xc0, 0xff, 0xee}},
		{pos: geom.BlockPosition{X: -1, Y: 5, Z: 10}, data: []byte{0xff, 0xfe, 0xfd}},
	}

	for _, b := range blocks {
		insertBlock(t, db, b.pos, b.data)
	}

	backend, err := sqlite3.NewBackend(dbPath)
	require.NoError(t, err)
	defer func() { _ = backend.Close() }()

	for _, b := range blocks {
		result, err := backend.GetBlockData(b.pos)
		require.NoError(t, err)

		assert.Equal(t, b.data, result)
	}
}

func createSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		CREATE TABLE blocks (
			x INTEGER,
			y INTEGER,
			z INTEGER,
			data BLOB NOT NULL,
			PRIMARY KEY (x, z, y)
		);
	`)
	require.NoError(t, err)
}

func insertBlock(t *testing.T, db *sql.DB, pos geom.BlockPosition, data []byte) {
	t.Helper()
	_, err := db.Exec(
		"INSERT INTO blocks (x, y, z, data) VALUES (?, ?, ?, ?)",
		pos.X, pos.Y, pos.Z, data,
	)
	require.NoError(t, err)
}
