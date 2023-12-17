package filesystem

import (
	"context"
	"testing"

	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func TestStore_RunUpMigrations(t *testing.T) {
	// We just check `NewStore` does not return an error
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We just check `RunUpMigrations` does not return an error
	err = store.RunUpMigrations(context.Background())
	require.NoError(t, err, "RunUpMigrations should not return an error")
}

func TestStore_RunDownMigrations(t *testing.T) {
	// We just check `NewStore` does not return an error
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We just check `RunDownMigrations` does not return an error
	err = store.RunDownMigrations(context.Background())
	require.NoError(t, err, "RunUpMigrations should not return an error")
}
