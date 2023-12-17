package filesystem

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	validatorServiceConfig "github.com/prysmaticlabs/prysm/v4/config/validator/service"
	"github.com/prysmaticlabs/prysm/v4/crypto/bls"
	"github.com/prysmaticlabs/prysm/v4/io/file"
	validatorpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1/validator-client"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func getPubKeys(t *testing.T, count int) [][fieldparams.BLSPubkeyLength]byte {
	pubKeys := make([][fieldparams.BLSPubkeyLength]byte, count)

	for i := range pubKeys {
		validatorKey, err := bls.RandKey()
		require.NoError(t, err, "RandKey should not return an error")

		copy(pubKeys[i][:], validatorKey.PublicKey().Marshal())
	}

	return pubKeys
}

func TestStore_NewStore(t *testing.T) {
	// We create some pubkeys
	pubkeys := getPubKeys(t, 5)

	// We just check `NewStore` does not return an error
	_, err := NewStore(t.TempDir(), &Config{PubKeys: pubkeys})
	require.NoError(t, err, "NewStore should not return an error")
}

func TestStore_Close(t *testing.T) {
	// We create a new store
	s, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We close the DB
	require.NoError(t, s.Close(), "Close should not return an error")
}

func TestStore_Backup(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()
	originalConfigurationFilePath := path.Join(databasePath, configurationFileName)
	originalSlashingProtectionDirPath := path.Join(databasePath, SlashingProtectionDirName)

	// We get a backups directory path
	backupsPath := t.TempDir()

	// We create some pubkeys
	pubkeys := getPubKeys(t, 5)

	// We create a new store
	s, err := NewStore(databasePath, &Config{PubKeys: pubkeys})
	require.NoError(t, err, "NewStore should not return an error")

	// We update the proposer settings
	err = s.SaveProposerSettings(context.Background(), &validatorServiceConfig.ProposerSettings{
		DefaultConfig: &validatorServiceConfig.ProposerOption{
			FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
				FeeRecipient: common.Address{},
			},
		},
	})
	require.NoError(t, err, "SaveProposerSettings should not return an error")

	// We backup the DB
	require.NoError(t, s.Backup(context.Background(), backupsPath, true), "Backup should not return an error")

	// We get the directory path of the backup
	files, err := os.ReadDir(path.Join(backupsPath, backupsDirectoryName))
	require.NoError(t, err, "os.ReadDir should not return an error")
	require.Equal(t, 1, len(files), "os.ReadDir should return one file")
	backupDirEntry := files[0]
	require.Equal(t, true, backupDirEntry.IsDir(), "os.ReadDir should return a directory")
	backupDirPath := path.Join(backupsPath, backupsDirectoryName, backupDirEntry.Name())

	// We get the path of the configuration file and the slashing protection directory
	backupConfigurationFilePath := path.Join(backupDirPath, configurationFileName)
	backupSlashingProtectionDirPath := path.Join(backupDirPath, SlashingProtectionDirName)

	// We compare the content of the slashing protection directory
	require.Equal(t, true, file.DirsEqual(originalSlashingProtectionDirPath, backupSlashingProtectionDirPath))

	// We compare the content of the configuration file
	originalConfigurationFileHash, err := file.HashFile(originalConfigurationFilePath)
	require.NoError(t, err, "file.HashFile should not return an error")

	backupConfigurationFileHash, err := file.HashFile(backupConfigurationFilePath)
	require.NoError(t, err, "file.HashFile should not return an error")

	require.DeepSSZEqual(t, originalConfigurationFileHash, backupConfigurationFileHash, "file.HashFile should return the same hash")
}

func TestStore_DatabasePath(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create a new store
	s, err := NewStore(databasePath, nil)
	require.NoError(t, err, "NewStore should not return an error")

	expected := databasePath
	actual := s.DatabasePath()

	require.Equal(t, expected, actual)
}

func TestStore_ClearDB(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We compute slashing protection directory and configuration file paths
	slashingProtectionDirPath := path.Join(databasePath, SlashingProtectionDirName)
	configurationFilePath := path.Join(databasePath, configurationFileName)

	// We create some pubkeys
	pubkeys := getPubKeys(t, 5)

	// We create a new store
	s, err := NewStore(databasePath, &Config{PubKeys: pubkeys})
	require.NoError(t, err, "NewStore should not return an error")

	// We check the presence of the slashing protection directory
	_, err = os.Stat(slashingProtectionDirPath)
	require.NoError(t, err, "os.Stat should not return an error")

	// We clear the DB
	err = s.ClearDB()
	require.NoError(t, err, "ClearDB should not return an error")

	// We check the absence of the slashing protection directory
	_, err = os.Stat(slashingProtectionDirPath)
	require.ErrorIs(t, err, os.ErrNotExist, "os.Stat should return os.ErrNotExist")

	// We check the absence of the configuration file path
	_, err = os.Stat(configurationFilePath)
	require.ErrorIs(t, err, os.ErrNotExist, "os.Stat should return os.ErrNotExist")
}

func TestStore_UpdatePublickKeysBuckets(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create some pubkeys
	pubkeys := getPubKeys(t, 5)

	// We create a new store
	s, err := NewStore(databasePath, &Config{PubKeys: pubkeys})
	require.NoError(t, err, "NewStore should not return an error")

	// We update the public keys
	err = s.UpdatePublicKeysBuckets(pubkeys)
	require.NoError(t, err, "UpdatePublicKeysBuckets should not return an error")

	// We check if the public keys files have been created
	for i := range pubkeys {
		pubkeyHex := hexutil.Encode(pubkeys[i][:])
		pubkeyFile := path.Join(databasePath, SlashingProtectionDirName, fmt.Sprintf("%s.yaml", pubkeyHex))

		_, err := os.Stat(pubkeyFile)
		require.NoError(t, err, "os.Stat should not return an error")
	}
}

func TestStore_slashingProtectionDirPath(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create a new store
	s, err := NewStore(databasePath, nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We check the slashing protection directory path
	expected := path.Join(databasePath, SlashingProtectionDirName)
	actual := s.slashingProtectionDirPath()
	require.Equal(t, expected, actual)
}

func TestStore_pubkeySlashingProtectionFilePath(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create a new store
	s, err := NewStore(databasePath, nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We create a pubkey
	pubkey := getPubKeys(t, 1)[0]

	// We check the pubkey slashing protection file path
	expected := path.Join(databasePath, SlashingProtectionDirName, hexutil.Encode(pubkey[:])+".yaml")
	actual := s.pubkeySlashingProtectionFilePath(pubkey)
	require.Equal(t, path.Join(databasePath, SlashingProtectionDirName, hexutil.Encode(pubkey[:])+".yaml"), s.pubkeySlashingProtectionFilePath(pubkey))
	require.Equal(t, expected, actual)
}

func TestStore_configurationFilePath(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create a new store
	s, err := NewStore(databasePath, nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We check the configuration file path
	expected := path.Join(databasePath, configurationFileName)
	actual := s.configurationFilePath()
	require.Equal(t, expected, actual)
}

func TestStore_configuration_saveConfiguration(t *testing.T) {
	var expected *Configuration

	// We get a database path
	databasePath := t.TempDir()

	// We create a new store
	s, err := NewStore(databasePath, nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We check the configuration
	actual, err := s.configuration()
	require.NoError(t, err, "configuration should not return an error")
	require.Equal(t, expected, actual)

	// We try to save a nil configuration
	err = s.saveConfiguration(nil)
	require.NoError(t, err, "saveConfiguration should not return an error")

	// We check the configuration
	actual, err = s.configuration()
	require.NoError(t, err, "configuration should not return an error")
	require.Equal(t, expected, actual)

	// We update the proposer settings
	feeRecipientHex := "0x1111111111111111111111111111111111111111"

	proposerSettings := &validatorServiceConfig.ProposerSettings{
		DefaultConfig: &validatorServiceConfig.ProposerOption{
			FeeRecipientConfig: &validatorServiceConfig.FeeRecipientConfig{
				FeeRecipient: common.HexToAddress(feeRecipientHex),
			},
		},
	}

	err = s.SaveProposerSettings(context.Background(), proposerSettings)
	require.NoError(t, err, "SaveProposerSettings should not return an error")

	// We check the configuration
	expected = &Configuration{
		ProposerSettings: &validatorpb.ProposerSettingsPayload{
			DefaultConfig: &validatorpb.ProposerOptionPayload{
				FeeRecipient: feeRecipientHex,
			},
			ProposerConfig: map[string]*validatorpb.ProposerOptionPayload{},
		},
	}

	actual, err = s.configuration()
	require.NoError(t, err, "configuration should not return an error")
	require.DeepEqual(t, expected, actual)
}

func TestStore_validatorSlashingProtection_saveValidatorSlashingProtection(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create a new store
	s, err := NewStore(databasePath, nil)
	require.NoError(t, err, "NewStore should not return an error")

	// We create a pubkey
	pubkey := getPubKeys(t, 1)[0]

	// We save an empty validator slashing protection for the pubkey
	err = s.saveValidatorSlashingProtection(pubkey, nil)
	require.NoError(t, err, "saveValidatorSlashingProtection should not return an error")

	// We check the validator slashing protection for the pubkey
	var expected *ValidatorSlashingProtection
	actual, err := s.validatorSlashingProtection(pubkey)
	require.NoError(t, err, "validatorSlashingProtection should not return an error")
	require.Equal(t, expected, actual)

	// We update the validator slashing protection for the pubkey
	epoch := uint64(1)
	validatorSlashingProtection := &ValidatorSlashingProtection{LatestSignedBlockSlot: &epoch}
	err = s.saveValidatorSlashingProtection(pubkey, validatorSlashingProtection)
	require.NoError(t, err, "saveValidatorSlashingProtection should not return an error")

	// We check the validator slashing protection for the pubkey
	expected = &ValidatorSlashingProtection{LatestSignedBlockSlot: &epoch}
	actual, err = s.validatorSlashingProtection(pubkey)
	require.NoError(t, err, "validatorSlashingProtection should not return an error")
	require.DeepEqual(t, expected, actual)
}

func TestStore_publicKeys(t *testing.T) {
	// We get a database path
	databasePath := t.TempDir()

	// We create some pubkeys
	pubkeys := getPubKeys(t, 5)

	// We create a new store
	s, err := NewStore(databasePath, &Config{PubKeys: pubkeys})
	require.NoError(t, err, "NewStore should not return an error")

	// We check the public keys
	expected := pubkeys
	actual, err := s.publicKeys()
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
