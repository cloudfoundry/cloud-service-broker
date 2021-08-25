package encryption

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/crypto/pbkdf2"
)

const (
	encryptionEnabled = "encryption.enabled"
	encryptionKeys    = "encryption.keys"
)

func GetEncryptionKey() (string, error) {
	encryptDB := viper.GetBool(encryptionEnabled)
	rawPasswordBlocks := viper.GetString(encryptionKeys)

	var passwords PasswordConfigs

	if rawPasswordBlocks != "" {
		err := json.Unmarshal([]byte(rawPasswordBlocks), &passwords)
		if err != nil {
			return "", fmt.Errorf("error unmarshalling encryption keys: %s", err)
		}
	}

	if encryptDB {
		if err := passwords.Validate(); err != nil {
			return "", fmt.Errorf("encryption is enabled, but there was an error validating encryption keys: %s", err)
		}

		key := generateKey(passwords)
		return string(key), nil
	} else {
		for _, key := range passwords {
			if key.Primary {
				return "", fmt.Errorf("encryption is disabled, but a primary encryption key was provided")
			}
		}
	}

	return "", nil
}

func generateKey(encryptKeys PasswordConfigs) []byte {
	var currentPass PasswordConfig
	for _, key := range encryptKeys {
		if key.Primary {
			currentPass = key
			break
		}
	}

	key := pbkdf2.Key([]byte(currentPass.Password.Secret), []byte(currentPass.Label), 10000, 32, sha256.New)
	return key
}

