package filesystem

import (
	"context"
	"testing"

	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func TestStore_GenesisValidatorsRoot_SaveGenesisValidatorsRoot(t *testing.T) {
	// Create a new store
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err)

	// // Save genesis validators root
	// genValRoot := []byte("genesis validators root")
	// require.NoError(t, store.SaveGenesisValidatorsRoot(context.Background(), genValRoot))

	var expected []byte

	// Get genesis validators root
	actual, err := store.GenesisValidatorsRoot(context.Background())
	require.NoError(t, err)
	require.DeepSSZEqual(t, expected, actual, "genesis validators root should be nil")

	// Save an empty configuration
	err = store.saveConfiguration(&Configuration{})
	require.NoError(t, err, "could not save configuration")

	// Get genesis validators root
	actual, err = store.GenesisValidatorsRoot(context.Background())
	require.NoError(t, err)
	require.DeepSSZEqual(t, expected, actual, "genesis validators root should be nil")

	// Save genesis validators root
	expected = []byte("genesis validators root")
	err = store.SaveGenesisValidatorsRoot(context.Background(), expected)
	require.NoError(t, err, "could not save genesis validators root")

	// Get genesis validators root
	actual, err = store.GenesisValidatorsRoot(context.Background())
	require.NoError(t, err, "could not get genesis validators root")
	require.DeepSSZEqual(t, expected, actual, "genesis validators root should be equal")

	// Clear database
	err = store.ClearDB()
	require.NoError(t, err, "could not clear database")

	// Re-save genesis validators root
	expected = []byte("genesis validators root")
	err = store.SaveGenesisValidatorsRoot(context.Background(), expected)
	require.NoError(t, err, "could not save genesis validators root")

	// Re-get genesis validators root
	actual, err = store.GenesisValidatorsRoot(context.Background())
	require.NoError(t, err, "could not get genesis validators root")
	require.DeepSSZEqual(t, expected, actual, "genesis validators root should be equal")
}
