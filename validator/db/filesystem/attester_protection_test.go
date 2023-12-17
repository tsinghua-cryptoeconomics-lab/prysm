package filesystem

import (
	"context"
	"testing"

	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	"github.com/prysmaticlabs/prysm/v4/validator/db/kv"
)

func TestStore_EIPImportBlacklistedPublicKeys(t *testing.T) {
	// Create a new store.
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err, "could not create store")

	var expected = [][fieldparams.BLSPubkeyLength]byte{}
	actual, err := store.EIPImportBlacklistedPublicKeys(context.Background())
	require.NoError(t, err, "could not get blacklisted public keys")
	require.DeepSSZEqual(t, expected, actual, "blacklisted public keys do not match")
}

func TestStore_SaveEIPImportBlacklistedPublicKeys(t *testing.T) {
	// Create a new store.
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err, "could not create store")

	// Save blacklisted public keys.
	err = store.SaveEIPImportBlacklistedPublicKeys(context.Background(), [][fieldparams.BLSPubkeyLength]byte{})
	require.NoError(t, err, "could not save blacklisted public keys")
}

func TestStore_LowestSignedTargetEpoch(t *testing.T) {
	// We define some saved source and target epoch
	savedSourceEpoch, savedTargetEpoch := 42, 43

	// We create a pubkey
	pubkey := getPubKeys(t, 1)[0]

	// We reate a new store.
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err, "could not create store")

	// We get the lowest signed target epoch.
	_, exists, err := store.LowestSignedTargetEpoch(context.Background(), [fieldparams.BLSPubkeyLength]byte{})
	require.NoError(t, err, "could not get lowest signed target epoch")
	require.Equal(t, false, exists, "lowest signed target epoch should not exist")

	// We create an attestation with both source and target epoch
	attestation := &ethpb.IndexedAttestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: primitives.Epoch(savedSourceEpoch)},
			Target: &ethpb.Checkpoint{Epoch: primitives.Epoch(savedTargetEpoch)},
		},
	}

	// We save the attestation
	err = store.SaveAttestationForPubKey(context.Background(), pubkey, [32]byte{}, attestation)
	require.NoError(t, err, "SaveAttestationForPubKey should not return an error")

	// We get the lowest signed target epoch.
	expected := primitives.Epoch(savedTargetEpoch)
	actual, exists, err := store.LowestSignedTargetEpoch(context.Background(), pubkey)
	require.NoError(t, err, "could not get lowest signed target epoch")
	require.Equal(t, true, exists, "lowest signed target epoch should not exist")
	require.Equal(t, expected, actual, "lowest signed target epoch should match")
}

func TestStore_LowestSignedSourceEpoch(t *testing.T) {
	// We create a pubkey
	pubkey := getPubKeys(t, 1)[0]

	// We reate a new store.
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err, "could not create store")

	// We get the lowest signed target epoch.
	_, exists, err := store.LowestSignedSourceEpoch(context.Background(), [fieldparams.BLSPubkeyLength]byte{})
	require.NoError(t, err, "could not get lowest signed source epoch")
	require.Equal(t, false, exists, "lowest signed source epoch should not exist")

	// We create an attestation
	savedSourceEpoch, savedTargetEpoch := 42, 43
	attestation := &ethpb.IndexedAttestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: primitives.Epoch(savedSourceEpoch)},
			Target: &ethpb.Checkpoint{Epoch: primitives.Epoch(savedTargetEpoch)},
		},
	}

	// We save the attestation
	err = store.SaveAttestationForPubKey(context.Background(), pubkey, [32]byte{}, attestation)
	require.NoError(t, err, "SaveAttestationForPubKey should not return an error")

	// We get the lowest signed target epoch.
	expected := primitives.Epoch(savedSourceEpoch)
	actual, exists, err := store.LowestSignedSourceEpoch(context.Background(), pubkey)
	require.NoError(t, err, "could not get lowest signed target epoch")
	require.Equal(t, true, exists, "lowest signed target epoch should exist")
	require.Equal(t, expected, actual, "lowest signed target epoch should match")
}

func TestStore_AttestedPublicKeys(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create some pubkeys
	pubkeys := getPubKeys(t, 5)

	// We create a new store
	s, err := NewStore(databasePath, &Config{PubKeys: pubkeys})
	require.NoError(t, err, "NewStore should not return an error")

	// We attest for some pubkeys
	attestedPubkeys := pubkeys[1:3]
	for _, pubkey := range attestedPubkeys {
		err = s.SaveAttestationForPubKey(context.Background(), pubkey, [32]byte{}, &ethpb.IndexedAttestation{
			Data: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Epoch: 42},
				Target: &ethpb.Checkpoint{Epoch: 43},
			},
		})
		require.NoError(t, err, "SaveAttestationForPubKey should not return an error")
	}

	// We check the public keys
	actual, err := s.AttestedPublicKeys(context.Background())
	require.NoError(t, err, "publicKeys should not return an error")

	// We cannot compare the slices directly because the order is not guaranteed,
	// so we compare sets instead.
	expectedSet := make(map[[fieldparams.BLSPubkeyLength]byte]bool)
	for _, pubkey := range attestedPubkeys {
		expectedSet[pubkey] = true
	}

	actualSet := make(map[[fieldparams.BLSPubkeyLength]byte]bool)
	for _, pubkey := range actual {
		actualSet[pubkey] = true
	}

	require.DeepEqual(t, expectedSet, actualSet)
}

func TestStore_SaveAttestationForPubKey(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create a public key/
	pubkey := getPubKeys(t, 1)[0]

	// We create a new store
	store, err := NewStore(databasePath, nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We save an empty list of attestations for the pubkey
	err = store.SaveAttestationsForPubKey(context.Background(), pubkey, [][]byte{}, []*ethpb.IndexedAttestation{})
	require.NoError(t, err, "SaveAttestationsForPubKey should not return an error")

	// We save multiple flavors of nil fields
	atts := []*ethpb.IndexedAttestation{
		nil,
		&ethpb.IndexedAttestation{},
		&ethpb.IndexedAttestation{Data: nil},
		&ethpb.IndexedAttestation{Data: &ethpb.AttestationData{Source: nil, Target: &ethpb.Checkpoint{}}},
		&ethpb.IndexedAttestation{Data: &ethpb.AttestationData{Source: &ethpb.Checkpoint{}, Target: nil}},
	}

	for _, att := range atts {
		err = store.SaveAttestationForPubKey(context.Background(), pubkey, [32]byte{}, att)
		require.ErrorContains(t, "incoming attestation does not contain source and/or target epoch", err)
	}

	// We create an attestation
	savedSourceEpoch, savedTargetEpoch := 42, 43
	attestation := &ethpb.IndexedAttestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: primitives.Epoch(savedSourceEpoch)},
			Target: &ethpb.Checkpoint{Epoch: primitives.Epoch(savedTargetEpoch)},
		},
	}

	// We save the attestation
	// ==> Should be accepted
	err = store.SaveAttestationForPubKey(context.Background(), pubkey, [32]byte{}, attestation)
	require.NoError(t, err, "SaveAttestationForPubKey should not return an error")

	// We check the saved attestation is correct
	validatorSlashingProtection, err := store.validatorSlashingProtection(pubkey)
	require.NoError(t, err, "validatorSlashingProtection should not return an error")
	require.Equal(t, uint64(savedSourceEpoch), validatorSlashingProtection.LastSignedAttestationSourceEpoch)
	require.Equal(t, uint64(savedTargetEpoch), *validatorSlashingProtection.LastSignedAttestationTargetEpoch)

	// We clear the database
	err = store.ClearDB()
	require.NoError(t, err, "Clear should not return an error")

	// We save an empty slashing protection file for the pubkey
	err = store.saveValidatorSlashingProtection(pubkey, &ValidatorSlashingProtection{})
	require.NoError(t, err, "saveValidatorSlashingProtection should not return an error")

	// We save the attestation
	// ==> Should be accepted
	err = store.SaveAttestationForPubKey(context.Background(), pubkey, [32]byte{}, attestation)
	require.NoError(t, err, "SaveAttestationForPubKey should not return an error")

	// We try to save the same attestation again
	// ==> Should be denied since the target epoch of the incoming attesttion is the same than
	//     the target epoch of the saved attestation
	err = store.SaveAttestationForPubKey(context.Background(), pubkey, [32]byte{}, attestation)
	require.ErrorContains(t, "could not sign attestation with target lower than or equal to recorded target epoch", err)

	// We check the saved attestation did not change
	validatorSlashingProtection, err = store.validatorSlashingProtection(pubkey)
	require.NoError(t, err, "validatorSlashingProtection should not return an error")
	require.Equal(t, uint64(savedSourceEpoch), validatorSlashingProtection.LastSignedAttestationSourceEpoch)
	require.Equal(t, uint64(savedTargetEpoch), *validatorSlashingProtection.LastSignedAttestationTargetEpoch)

	// We define an attestation with a lower source epoch
	incomingSourceEpoch, incomingTargetEpoch := 41, 43
	attestation = &ethpb.IndexedAttestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: primitives.Epoch(incomingSourceEpoch)},
			Target: &ethpb.Checkpoint{Epoch: primitives.Epoch(incomingTargetEpoch)},
		},
	}

	// We try to save the attestation
	// ==> Should be denied since the source epoch of the incoming attesttion is lower than
	//     the source epoch of the saved attestation
	err = store.SaveAttestationForPubKey(context.Background(), pubkey, [32]byte{}, attestation)
	require.ErrorContains(t, "could not sign attestation with source lower than recorded source epoch", err)

	// We check the saved attestation did not change
	validatorSlashingProtection, err = store.validatorSlashingProtection(pubkey)
	require.NoError(t, err, "validatorSlashingProtection should not return an error")
	require.Equal(t, uint64(savedSourceEpoch), validatorSlashingProtection.LastSignedAttestationSourceEpoch)
	require.Equal(t, uint64(savedTargetEpoch), *validatorSlashingProtection.LastSignedAttestationTargetEpoch)

	// We create a new, correct attestation
	savedSourceEpoch, savedTargetEpoch = 43, 44
	attestation = &ethpb.IndexedAttestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: primitives.Epoch(savedSourceEpoch)},
			Target: &ethpb.Checkpoint{Epoch: primitives.Epoch(savedTargetEpoch)},
		},
	}

	// We save the attestation
	// ==> Should be accepted
	err = store.SaveAttestationForPubKey(context.Background(), pubkey, [32]byte{}, attestation)
	require.NoError(t, err, "SaveAttestationForPubKey should not return an error")

	// We check the saved attestation is correct
	validatorSlashingProtection, err = store.validatorSlashingProtection(pubkey)
	require.NoError(t, err, "validatorSlashingProtection should not return an error")
	require.Equal(t, uint64(savedSourceEpoch), validatorSlashingProtection.LastSignedAttestationSourceEpoch)
	require.Equal(t, uint64(savedTargetEpoch), *validatorSlashingProtection.LastSignedAttestationTargetEpoch)
}

func TestStore_SaveAttestationsForPubKey(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create a public key
	pubkey := getPubKeys(t, 1)[0]

	// We create a new store
	store, err := NewStore(databasePath, nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We create some attestations
	//      30 ==========> 40
	//               40 ==========> 45   <----- Will be recorded into DB
	attestations := []*ethpb.IndexedAttestation{
		{
			Data: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Epoch: primitives.Epoch(40)},
				Target: &ethpb.Checkpoint{Epoch: primitives.Epoch(45)},
			},
		},
		{
			Data: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Epoch: primitives.Epoch(30)},
				Target: &ethpb.Checkpoint{Epoch: primitives.Epoch(40)},
			},
		},
	}

	// We save the attestations
	err = store.SaveAttestationsForPubKey(context.Background(), pubkey, [][]byte{}, attestations)
	require.NoError(t, err, "SaveAttestationsForPubKey should not return an error")

	// We check the saved attestations are correct
	validatorSlashingProtection, err := store.validatorSlashingProtection(pubkey)
	require.NoError(t, err, "validatorSlashingProtection should not return an error")
	require.Equal(t, uint64(40), validatorSlashingProtection.LastSignedAttestationSourceEpoch)
	require.Equal(t, uint64(45), *validatorSlashingProtection.LastSignedAttestationTargetEpoch)

	// We create a surrounded attestation
	//               40 ==========> 45   <----- Already recorded into DB
	//                   42 => 43        <----- Incoming attestation
	// ------------------------------------------------------------------------------------------------
	//                   42 ======> 45   <----- Will be recorded into DB (max source and target epochs)
	attestations = []*ethpb.IndexedAttestation{
		{
			Data: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Epoch: primitives.Epoch(42)},
				Target: &ethpb.Checkpoint{Epoch: primitives.Epoch(43)},
			},
		},
	}

	// We save the attestations
	err = store.SaveAttestationsForPubKey(context.Background(), pubkey, [][]byte{}, attestations)
	require.NoError(t, err, "SaveAttestationsForPubKey should not return an error")

	// We check the saved attestations are correct
	validatorSlashingProtection, err = store.validatorSlashingProtection(pubkey)
	require.NoError(t, err, "validatorSlashingProtection should not return an error")
	require.Equal(t, uint64(42), validatorSlashingProtection.LastSignedAttestationSourceEpoch)
	require.Equal(t, uint64(45), *validatorSlashingProtection.LastSignedAttestationTargetEpoch)

	// We create a surrounding attestation
	//                   42 ======> 45          <----- Already recorded into DB
	//              40 ==================> 50   <----- Incoming attestation
	// ------------------------------------------------------------------------------------------------------
	//                   42 =============> 50   <----- Will be recorded into DB (max source and target epochs)
	attestations = []*ethpb.IndexedAttestation{
		{
			Data: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Epoch: primitives.Epoch(40)},
				Target: &ethpb.Checkpoint{Epoch: primitives.Epoch(50)},
			},
		},
	}

	// We save the attestations
	err = store.SaveAttestationsForPubKey(context.Background(), pubkey, [][]byte{}, attestations)
	require.NoError(t, err, "SaveAttestationsForPubKey should not return an error")

	// We check the saved attestations are correct
	validatorSlashingProtection, err = store.validatorSlashingProtection(pubkey)
	require.NoError(t, err, "validatorSlashingProtection should not return an error")
	require.Equal(t, uint64(42), validatorSlashingProtection.LastSignedAttestationSourceEpoch)
	require.Equal(t, uint64(50), *validatorSlashingProtection.LastSignedAttestationTargetEpoch)
}

func TestStore_AttestationHistoryForPubKey(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create a public key
	pubkey := getPubKeys(t, 1)[0]

	// We create a new store
	store, err := NewStore(databasePath, nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We get the attestation history
	actual, err := store.AttestationHistoryForPubKey(context.Background(), pubkey)
	require.NoError(t, err, "AttestationHistoryForPubKey should not return an error")
	require.DeepEqual(t, []*kv.AttestationRecord{}, actual)

	// We create an attestation
	savedSourceEpoch, savedTargetEpoch := 42, 43
	attestation := &ethpb.IndexedAttestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: primitives.Epoch(savedSourceEpoch)},
			Target: &ethpb.Checkpoint{Epoch: primitives.Epoch(savedTargetEpoch)},
		},
	}

	// We save the attestation
	err = store.SaveAttestationForPubKey(context.Background(), pubkey, [32]byte{}, attestation)
	require.NoError(t, err, "SaveAttestationForPubKey should not return an error")

	// We get the attestation history
	expected := []*kv.AttestationRecord{
		{
			PubKey: pubkey,
			Source: primitives.Epoch(savedSourceEpoch),
			Target: primitives.Epoch(savedTargetEpoch),
		},
	}

	actual, err = store.AttestationHistoryForPubKey(context.Background(), pubkey)
	require.NoError(t, err, "AttestationHistoryForPubKey should not return an error")
	require.DeepEqual(t, expected, actual)
}
