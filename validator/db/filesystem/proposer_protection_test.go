package filesystem

import (
	"context"
	"testing"

	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	"github.com/prysmaticlabs/prysm/v4/validator/db/kv"
)

func TestStore_ProposalHistoryForPubKey(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create a public key
	pubkey := getPubKeys(t, 1)[0]

	// We create a new store
	store, err := NewStore(databasePath, nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We get the proposal history for the public key
	actual, err := store.ProposalHistoryForPubKey(context.Background(), pubkey)
	require.NoError(t, err, "ProposalHistoryForPubKey should not return an error")
	require.DeepEqual(t, []*kv.Proposal{}, actual)

	// We create a default (without proposal history) file for this pubkey
	err = store.UpdatePublicKeysBuckets([][fieldparams.BLSPubkeyLength]byte{pubkey})
	require.NoError(t, err, "UpdatePublicKeysBuckets should not return an error")

	// We get the proposal history for the public key
	actual, err = store.ProposalHistoryForPubKey(context.Background(), pubkey)
	require.NoError(t, err, "ProposalHistoryForPubKey should not return an error")
	require.DeepEqual(t, []*kv.Proposal{}, actual)

	// We save a proposal history for the public key
	err = store.SaveProposalHistoryForSlot(context.Background(), pubkey, 42, nil)
	require.NoError(t, err, "SaveProposalHistoryForSlot should not return an error")

	// We get the proposal history for the public key
	expected := []*kv.Proposal{
		{
			Slot: 42,
		},
	}

	actual, err = store.ProposalHistoryForPubKey(context.Background(), pubkey)
	require.NoError(t, err, "ProposalHistoryForPubKey should not return an error")
	require.DeepEqual(t, expected, actual)
}

func TestStore_SaveProposalHistoryForSlot(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create a public key
	pubkey := getPubKeys(t, 1)[0]

	// We create a new store
	store, err := NewStore(databasePath, nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We save a proposal history for a slot
	// ==> It should succeed
	err = store.SaveProposalHistoryForSlot(context.Background(), pubkey, 42, nil)
	require.NoError(t, err, "SaveProposalHistoryForSlot should not return an error")

	// We clear the DB
	err = store.ClearDB()
	require.NoError(t, err, "ClearDB should not return an error")

	// We create a default (without proposal history) file for this pubkey
	err = store.UpdatePublicKeysBuckets([][fieldparams.BLSPubkeyLength]byte{pubkey})
	require.NoError(t, err, "UpdatePublicKeysBuckets should not return an error")

	// We save a proposal history for a slot
	// ==> It should succeed
	err = store.SaveProposalHistoryForSlot(context.Background(), pubkey, 42, nil)
	require.NoError(t, err, "SaveProposalHistoryForSlot should not return an error")

	// We try to save a proposal history for the same slot
	// ==> It should fail
	err = store.SaveProposalHistoryForSlot(context.Background(), pubkey, 42, nil)
	require.ErrorContains(t, "could not sign proposal with slot lower than or equal to recorded slot", err)

	// We try to save a proposal history for a lower slot
	// ==> It should fail
	err = store.SaveProposalHistoryForSlot(context.Background(), pubkey, 41, nil)
	require.ErrorContains(t, "could not sign proposal with slot lower than or equal to recorded slot", err)

	// We save a proposal history for a higher slot
	// ==> It should succeed
	err = store.SaveProposalHistoryForSlot(context.Background(), pubkey, 43, nil)
	require.NoError(t, err, "SaveProposalHistoryForSlot should not return an error")
}

func TestStore_ProposedPublicKeys(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create some pubkeys
	pubkeys := getPubKeys(t, 5)

	// We create a new store
	s, err := NewStore(databasePath, &Config{PubKeys: pubkeys})
	require.NoError(t, err, "NewStore should not return an error")

	// We check the public keys
	expected := pubkeys
	actual, err := s.ProposedPublicKeys(context.Background())
	require.NoError(t, err, "publicKeys should not return an error")

	// We cannot compare the slices directly because the order is not guaranteed,
	// so we compare sets instead.

	expectedSet := make(map[[fieldparams.BLSPubkeyLength]byte]bool)
	for _, pubkey := range expected {
		expectedSet[pubkey] = true
	}

	actualSet := make(map[[fieldparams.BLSPubkeyLength]byte]bool)
	for _, pubkey := range actual {
		actualSet[pubkey] = true
	}

	require.DeepEqual(t, expectedSet, actualSet)
}
