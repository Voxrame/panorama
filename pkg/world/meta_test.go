package world_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lord-server/panorama/pkg/world"
	"github.com/stretchr/testify/require"
)

func TestParseMeta_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	metaFile := filepath.Join(tmpDir, "meta.txt")

	metaContent := `enable_damage = true
creative_mode = false
mod_storage_backend = sqlite3
auth_backend = sqlite3
player_backend = sqlite3
backend = sqlite3
gameid = devtest
world_name = world
`
	require.NoError(t, os.WriteFile(metaFile, []byte(metaContent), 0o644))

	meta, err := world.ParseMeta(metaFile)
	require.NoError(t, err)

	require.Equal(t, "true", meta["enable_damage"])
	require.Equal(t, "false", meta["creative_mode"])
	require.Equal(t, "sqlite3", meta["mod_storage_backend"])
	require.Equal(t, "sqlite3", meta["auth_backend"])
	require.Equal(t, "sqlite3", meta["player_backend"])
	require.Equal(t, "sqlite3", meta["backend"])
	require.Equal(t, "devtest", meta["gameid"])
	require.Equal(t, "world", meta["world_name"])
}

func TestParseMeta_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	metaFile := filepath.Join(tmpDir, "meta.txt")

	require.NoError(t, os.WriteFile(metaFile, []byte(""), 0o644))

	meta, err := world.ParseMeta(metaFile)
	require.NoError(t, err)

	require.Empty(t, meta)
}

func TestParseMeta_InvalidPath(t *testing.T) {
	t.Parallel()

	meta, err := world.ParseMeta("/nonexistent/path/meta.txt")
	require.Error(t, err)
	require.Nil(t, meta)
}

func TestParseMeta_WhitespaceHandling(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	metaFile := filepath.Join(tmpDir, "meta.txt")

	metaContent := `  key1   =   value1
key2=value2
  key3=value3
`
	require.NoError(t, os.WriteFile(metaFile, []byte(metaContent), 0o644))

	meta, err := world.ParseMeta(metaFile)
	require.NoError(t, err)

	require.Equal(t, "value1", meta["key1"])
	require.Equal(t, "value2", meta["key2"])
	require.Equal(t, "value3", meta["key3"])
}

func TestParseMeta_NoEqualsSign(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	metaFile := filepath.Join(tmpDir, "meta.txt")

	metaContent := `valid_key = valid_value
no_equals_sign
`
	require.NoError(t, os.WriteFile(metaFile, []byte(metaContent), 0o644))

	_, err := world.ParseMeta(metaFile)
	require.Error(t, err)
}

func TestParseMeta_EmptyKey(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	metaFile := filepath.Join(tmpDir, "meta.txt")

	metaContent := `valid_key = valid_value
= a
`
	require.NoError(t, os.WriteFile(metaFile, []byte(metaContent), 0o644))

	_, err := world.ParseMeta(metaFile)
	require.Error(t, err)
}

func TestParseMeta_MultipleEquals(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	metaFile := filepath.Join(tmpDir, "meta.txt")

	metaContent := `url = http://example.com?foo=bar&baz=qux
formula = a=b+c=d
`
	require.NoError(t, os.WriteFile(metaFile, []byte(metaContent), 0o644))

	meta, err := world.ParseMeta(metaFile)
	require.NoError(t, err)

	require.Equal(t, "http://example.com?foo=bar&baz=qux", meta["url"])
	require.Equal(t, "a=b+c=d", meta["formula"])
}
