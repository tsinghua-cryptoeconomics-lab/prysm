package filesystem

import (
	"context"
	"testing"

	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func TestStore_SaveGraffitiOrderedIndex(t *testing.T) {
	// Create a new store
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err)

	saved := uint64(1)
	expected := uint64(0)
	fileHash := [fieldparams.RootLength]byte{}

	// Save graffiti ordered index
	err = store.SaveGraffitiOrderedIndex(context.Background(), saved)
	require.NoError(t, err)

	// Get graffiti ordered index
	actual, err := store.GraffitiOrderedIndex(context.Background(), fileHash)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	// Clear database
	err = store.ClearDB()
	require.NoError(t, err)

	// Create empty configuration
	err = store.saveConfiguration(&Configuration{})
	require.NoError(t, err)

	// Save graffiti ordered index
	err = store.SaveGraffitiOrderedIndex(context.Background(), saved)
	require.NoError(t, err)

	// Get graffiti ordered index
	actual, err = store.GraffitiOrderedIndex(context.Background(), fileHash)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestStore_GraffitiOrderedIndex(t *testing.T) {
	// Create a new store
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err)

	expectedDefault := uint64(0)
	expectedCustom := uint64(1)
	defaultFileHash := [fieldparams.RootLength]byte{1}
	differentFileHash := [fieldparams.RootLength]byte{2}

	// Get graffiti ordered index
	actual, err := store.GraffitiOrderedIndex(context.Background(), defaultFileHash)
	require.NoError(t, err)
	require.Equal(t, expectedDefault, actual)

	// Clear database
	err = store.ClearDB()
	require.NoError(t, err)

	// Create empty configuration
	err = store.saveConfiguration(&Configuration{})
	require.NoError(t, err)

	// Get graffiti ordered index
	actual, err = store.GraffitiOrderedIndex(context.Background(), defaultFileHash)
	require.NoError(t, err)
	require.Equal(t, expectedDefault, actual)

	// Save graffiti ordered index
	err = store.SaveGraffitiOrderedIndex(context.Background(), expectedCustom)
	require.NoError(t, err)

	// Get graffiti ordered index with default file hash
	actual, err = store.GraffitiOrderedIndex(context.Background(), defaultFileHash)
	require.NoError(t, err)
	require.Equal(t, expectedCustom, actual)

	// Get graffiti ordered index with different file hash
	actual, err = store.GraffitiOrderedIndex(context.Background(), differentFileHash)
	require.NoError(t, err)
	require.Equal(t, expectedDefault, actual)
}
