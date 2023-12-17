package filesystem

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	validatorServiceConfig "github.com/prysmaticlabs/prysm/v4/config/validator/service"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/validator"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func TestStore_ProposerSettings_SaveProposerSettings(t *testing.T) {
	// We create a new store
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err, "NewStore should not return an error")

	var expected *validatorServiceConfig.ProposerSettings

	// We check `ProposerSettings` returns on error
	actual, err := store.ProposerSettings(context.Background())
	require.ErrorContains(t, ErrNoProposerSettingsFound.Error(), err)
	require.Equal(t, expected, actual, "ProposerSettings should return nil")

	// Create empty configuration
	err = store.saveConfiguration(&Configuration{})
	require.NoError(t, err, "saveConfiguration should not return an error")

	// We check `ProposerSettings` returns on error
	actual, err = store.ProposerSettings(context.Background())
	require.ErrorContains(t, ErrNoProposerSettingsFound.Error(), err)
	require.Equal(t, expected, actual, "ProposerSettings should return nil")

	// We save some proposer settings
	feeRecipientHex := "0x1111111111111111111111111111111111111111"

	expected = &validatorServiceConfig.ProposerSettings{
		ProposeConfig: make(map[[fieldparams.BLSPubkeyLength]byte]*validatorServiceConfig.ProposerOption),
		DefaultConfig: &validatorServiceConfig.ProposerOption{
			FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
				FeeRecipient: common.HexToAddress(feeRecipientHex),
			},
		},
	}

	err = store.SaveProposerSettings(context.Background(), expected)
	require.NoError(t, err, "SaveProposerSettings should not return an error")

	// We get proposer settings
	actual, err = store.ProposerSettings(context.Background())
	require.NoError(t, err, "ProposerSettings should not return an error")
	require.DeepEqual(t, expected, actual, "ProposerSettings should return expected")
}

func TestStore_ProposerSettingsExists(t *testing.T) {
	// We create a new store
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We check `ProposerSettingsExists` returns false
	expected := false
	actual, err := store.ProposerSettingsExists(context.Background())
	require.NoError(t, err, "ProposerSettingsExists should not return an error")
	require.Equal(t, expected, actual, "ProposerSettingsExists should return false")

	// We create empty configuration
	err = store.saveConfiguration(&Configuration{})
	require.NoError(t, err, "saveConfiguration should not return an error")

	// We check `ProposerSettingsExists` returns false
	expected = false
	actual, err = store.ProposerSettingsExists(context.Background())
	require.NoError(t, err, "ProposerSettingsExists should not return an error")
	require.Equal(t, expected, actual, "ProposerSettingsExists should return false")

	// We save some proposer settings
	feeRecipientHex := "0x1111111111111111111111111111111111111111"

	proposerSettings := &validatorServiceConfig.ProposerSettings{
		ProposeConfig: make(map[[fieldparams.BLSPubkeyLength]byte]*validatorServiceConfig.ProposerOption),
		DefaultConfig: &validatorServiceConfig.ProposerOption{
			FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
				FeeRecipient: common.HexToAddress(feeRecipientHex),
			},
		},
	}

	err = store.SaveProposerSettings(context.Background(), proposerSettings)
	require.NoError(t, err, "SaveProposerSettings should not return an error")

	// We check `ProposerSettingsExists` returns true
	expected = true
	actual, err = store.ProposerSettingsExists(context.Background())
	require.NoError(t, err, "ProposerSettingsExists should not return an error")
	require.Equal(t, expected, actual, "ProposerSettingsExists should return true")
}

func TestStore_UpdateProposerSettingsDefault(t *testing.T) {
	// We create a new store
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We save nil proposer settings
	err = store.UpdateProposerSettingsDefault(context.Background(), nil)
	require.NoError(t, err, "UpdateProposerSettingsDefault should not return an error")

	// We check `ProposerSettings` returns on error
	_, err = store.ProposerSettings(context.Background())
	require.ErrorContains(t, ErrNoProposerSettingsFound.Error(), err)

	// We save some proposer settings
	feeRecipientHex := "0x1111111111111111111111111111111111111111"

	savedProposerOption := &validatorServiceConfig.ProposerOption{
		FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
			FeeRecipient: common.HexToAddress(feeRecipientHex),
		},
		BuilderConfig: &validatorServiceConfig.BuilderConfig{
			Enabled:  true,
			GasLimit: validator.Uint64(1000),
			Relays:   []string{},
		},
	}

	err = store.UpdateProposerSettingsDefault(context.Background(), savedProposerOption)
	require.NoError(t, err, "UpdateProposerSettingsDefault should not return an error")

	// We check `ProposerSettings` returns expected
	actualProposerSettings, err := store.ProposerSettings(context.Background())
	require.NoError(t, err, "ProposerSettings should not return an error")
	require.DeepEqual(t, savedProposerOption, actualProposerSettings.DefaultConfig, "ProposerSettings should return expected")

	// We clear the database
	err = store.ClearDB()
	require.NoError(t, err, "ClearDB should not return an error")

	// We create empty configuration
	err = store.saveConfiguration(&Configuration{})
	require.NoError(t, err, "saveConfiguration should not return an error")

	// We save some proposer settings
	err = store.UpdateProposerSettingsDefault(context.Background(), savedProposerOption)
	require.NoError(t, err, "UpdateProposerSettingsDefault should not return an error")

	// We check `ProposerSettings` returns expected
	actualProposerSettings, err = store.ProposerSettings(context.Background())
	require.NoError(t, err, "ProposerSettings should not return an error")
	require.DeepEqual(t, savedProposerOption, actualProposerSettings.DefaultConfig, "ProposerSettings should return expected")

	// We save some other proposer settings
	feeRecipientHex = "0x2222222222222222222222222222222222222222"

	savedProposerOption = &validatorServiceConfig.ProposerOption{
		FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
			FeeRecipient: common.HexToAddress(feeRecipientHex),
		},
		BuilderConfig: &validatorServiceConfig.BuilderConfig{
			Enabled:  true,
			GasLimit: validator.Uint64(2000),
			Relays:   []string{},
		},
	}

	err = store.UpdateProposerSettingsDefault(context.Background(), savedProposerOption)
	require.NoError(t, err, "UpdateProposerSettingsDefault should not return an error")

	// We check `ProposerSettings` returns expected
	actualProposerSettings, err = store.ProposerSettings(context.Background())
	require.NoError(t, err, "ProposerSettings should not return an error")
	require.DeepEqual(t, savedProposerOption, actualProposerSettings.DefaultConfig, "ProposerSettings should return expected")
}

func TestStore_UpdateProposerSettingsForPubkey(t *testing.T) {
	// We create a pubkey
	pubkey := getPubKeys(t, 1)[0]

	// We create a new store
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We save nil proposer settings
	err = store.UpdateProposerSettingsForPubkey(context.Background(), pubkey, nil)
	require.NoError(t, err, "UpdateProposerSettingsForPubkey should not return an error")

	// We check `ProposerSettings` returns on error
	_, err = store.ProposerSettings(context.Background())
	require.ErrorContains(t, ErrNoProposerSettingsFound.Error(), err)

	// We save some proposer settings for the given pubkey
	feeRecipientHex := "0x1111111111111111111111111111111111111111"

	savedProposerOption := &validatorServiceConfig.ProposerOption{
		FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
			FeeRecipient: common.HexToAddress(feeRecipientHex),
		},
		BuilderConfig: &validatorServiceConfig.BuilderConfig{
			Enabled:  true,
			GasLimit: validator.Uint64(1000),
			Relays:   []string{},
		},
	}

	err = store.UpdateProposerSettingsForPubkey(context.Background(), pubkey, savedProposerOption)
	require.NoError(t, err, "UpdateProposerSettingsDefault should not return an error")

	// We check `ProposerSettings` returns expected
	actualProposerSettings, err := store.ProposerSettings(context.Background())
	require.NoError(t, err, "ProposerSettings should not return an error")
	require.DeepEqual(t, savedProposerOption, actualProposerSettings.ProposeConfig[pubkey], "ProposerSettings should return expected")

	// We clear the database
	err = store.ClearDB()
	require.NoError(t, err, "ClearDB should not return an error")

	// We create empty configuration
	err = store.saveConfiguration(&Configuration{})
	require.NoError(t, err, "saveConfiguration should not return an error")

	// We save some proposer settings for the given pubkey
	err = store.UpdateProposerSettingsForPubkey(context.Background(), pubkey, savedProposerOption)
	require.NoError(t, err, "UpdateProposerSettingsDefault should not return an error")

	// We check `ProposerSettings` returns expected
	actualProposerSettings, err = store.ProposerSettings(context.Background())
	require.NoError(t, err, "ProposerSettings should not return an error")
	require.DeepEqual(t, savedProposerOption, actualProposerSettings.ProposeConfig[pubkey], "ProposerSettings should return expected")

	// We save some other proposer settings
	feeRecipientHex = "0x2222222222222222222222222222222222222222"

	savedProposerOption = &validatorServiceConfig.ProposerOption{
		FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
			FeeRecipient: common.HexToAddress(feeRecipientHex),
		},
		BuilderConfig: &validatorServiceConfig.BuilderConfig{
			Enabled:  true,
			GasLimit: validator.Uint64(2000),
			Relays:   []string{},
		},
	}

	err = store.UpdateProposerSettingsForPubkey(context.Background(), pubkey, savedProposerOption)
	require.NoError(t, err, "UpdateProposerSettingsDefault should not return an error")

	// We check `ProposerSettings` returns expected
	actualProposerSettings, err = store.ProposerSettings(context.Background())
	require.NoError(t, err, "ProposerSettings should not return an error")
	require.DeepEqual(t, savedProposerOption, actualProposerSettings.ProposeConfig[pubkey], "ProposerSettings should return expected")
}
