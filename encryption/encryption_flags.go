package encryption

import (
	"errors"
	"flag"
	"strings"
)

type EncryptionKeys map[string]struct{}

func (EncryptionKeys) String() string {
	return ""
}

func (e EncryptionKeys) Set(key string) error {
	e[key] = struct{}{}
	return nil
}

type EncryptionFlags struct {
	activeKeyLabel string
	encryptionKeys EncryptionKeys
}

func NewEncryptionFlags() EncryptionFlags {
	return EncryptionFlags{
		encryptionKeys: make(EncryptionKeys),
	}
}

func AddEncryptionFlags(flagSet *flag.FlagSet) *EncryptionFlags {
	ef := NewEncryptionFlags()
	flagSet.Var(
		&ef.encryptionKeys,
		"encryptionKey",
		"Encryption key in label:passphrase format (may be specified multiple times)",
	)
	flagSet.StringVar(
		&ef.activeKeyLabel,
		"activeKeyLabel",
		"",
		"Label of the encryption key to be used when writing to the database",
	)
	return &ef
}

func (ef *EncryptionFlags) Parse() (Key, []Key, error) {
	if len(ef.encryptionKeys) == 0 {
		return nil, nil, errors.New("Must have at least one encryption key set")
	}

	if len(ef.activeKeyLabel) == 0 {
		return nil, nil, errors.New("Must select an active encryption key")
	}

	var encryptionKey Key
	keys := make([]Key, 0, len(ef.encryptionKeys))

	for key := range ef.encryptionKeys {
		splitKey := strings.SplitN(key, ":", 2)
		if len(splitKey) != 2 {
			return nil, nil, errors.New("Could not parse encryption keys")
		}
		label := splitKey[0]
		phrase := splitKey[1]
		key, err := NewKey(label, phrase)
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, key)

		if label == ef.activeKeyLabel {
			encryptionKey = key
		}
	}

	if encryptionKey == nil {
		return nil, nil, errors.New("The selected active key must be listed on the encryption keys flag")
	}

	return encryptionKey, keys, nil
}

// keyManager, err := NewKeyManager(encryptionKey, keys)
// if err != nil {
// 	return nil, nil, err
// }
// return keyManager, nil
