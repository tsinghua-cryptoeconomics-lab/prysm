package filesystem

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

func (s *Store) SaveGraffitiOrderedIndex(_ context.Context, index uint64) error {
	// Get configuration
	configuration, err := s.configuration()
	if err != nil {
		return errors.Wrapf(err, "could not get configuration")
	}

	if configuration == nil {
		// Create an new configuration
		configuration = &Configuration{
			Graffiti: &Graffiti{
				OrderedIndex: index,
			},
		}

		// Save the configuration
		if err := s.saveConfiguration(configuration); err != nil {
			return errors.Wrapf(err, "could not save configuration")
		}

		return nil
	}

	if configuration.Graffiti == nil {
		// Create a new graffiti
		configuration.Graffiti = &Graffiti{
			OrderedIndex: index,
		}

		// Save the configuration
		if err := s.saveConfiguration(configuration); err != nil {
			return errors.Wrapf(err, "could not save configuration")
		}

		return nil
	}

	// Modify the value of ordered index
	configuration.Graffiti.OrderedIndex = index

	// Save the configuration
	if err := s.saveConfiguration(configuration); err != nil {
		return errors.Wrapf(err, "could not save configuration")
	}

	return nil
}

func (s *Store) GraffitiOrderedIndex(_ context.Context, fileHash [32]byte) (uint64, error) {
	// Encode file hash to string
	fileHashHex := hexutil.Encode(fileHash[:])

	// Get configuration
	configuration, err := s.configuration()
	if err != nil {
		return 0, errors.Wrapf(err, "could not get configuration")
	}

	if configuration == nil {
		// Create an ew configuration
		configuration = &Configuration{
			Graffiti: &Graffiti{
				OrderedIndex: 0,
				FileHash:     fileHashHex,
			},
		}

		// Save the configuration
		if err := s.saveConfiguration(configuration); err != nil {
			return 0, errors.Wrapf(err, "could not save configuration")
		}

		return 0, nil
	}

	if configuration.Graffiti == nil {
		// Create a new graffiti
		configuration.Graffiti = &Graffiti{
			OrderedIndex: 0,
			FileHash:     fileHashHex,
		}

		// Save the configuration
		if err := s.saveConfiguration(configuration); err != nil {
			return 0, errors.Wrapf(err, "could not save configuration")
		}

		return 0, nil
	}

	// Check if file hash is equal to the file hash in configuration
	if configuration.Graffiti.FileHash != fileHashHex {
		// Modify the value of ordered index
		configuration.Graffiti.OrderedIndex = 0

		// Modify the value of file hash
		configuration.Graffiti.FileHash = fileHashHex

		// Save the configuration
		if err := s.saveConfiguration(configuration); err != nil {
			return 0, errors.Wrapf(err, "could not save configuration")
		}

		return 0, nil
	}

	return configuration.Graffiti.OrderedIndex, nil
}
