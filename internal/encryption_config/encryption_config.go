package encryption_config

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/jinzhu/gorm"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	"github.com/spf13/viper"
	"golang.org/x/crypto/pbkdf2"
)

const (
	encryptionEnabled = "encryption.enabled"
	encryptionKeys    = "encryption.keys"
	canary            = "some-test-value"
)

func GetEncryptionKey() (string, error) {
	logger := utils.NewLogger("cloud-service-broker")

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

		key, err := generateKey(passwords)
		if err != nil {
			return "", fmt.Errorf("error generating the key: %s", err)
		}

		logger.Info("encryption enabled")
		return string(key), nil
	} else {
		for _, key := range passwords {
			if key.Primary {
				return "", fmt.Errorf("encryption is disabled, but a primary encryption key was provided")
			}
		}
		logger.Info("encryption disabled")
	}

	return "", nil
}

func generateKey(encryptKeys PasswordConfigs) ([]byte, error) {
	primaryPassword := getPrimaryPassword(encryptKeys)
	var salt []byte
	newPassword := false

	encryptionDetails, err := db_service.GetEncryptionDetailByLabel(context.Background(), primaryPassword.Label)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			// TODO test
			return nil, err
		}

		newPassword = true
		salt, err = generateNewSalt()
		if err != nil {
			return nil, err
		}
	} else {
		salt = []byte(encryptionDetails.Salt)
	}

	key := pbkdf2.Key([]byte(primaryPassword.Password.Secret), salt, 10000, 32, sha256.New)

	if newPassword {
		details := models.EncryptionDetail{
			Label:   primaryPassword.Label,
			Salt:    string(salt),
			Primary: true,
		}

		models.SetEncryptor(models.ConfigureEncryption(string(key)))
		if err := details.SetCanary(canary); err != nil {
			return nil, fmt.Errorf("error setting canary value: %s", err)
		}

		err = db_service.CreateEncryptionDetail(context.Background(), &details)
		if err != nil {
			return nil, fmt.Errorf("error storing encryption details: %s", err)
		}
	}

	return key, nil
}

func getPrimaryPassword(encryptKeys PasswordConfigs) PasswordConfig {
	var currentPass PasswordConfig
	for _, key := range encryptKeys {
		if key.Primary {
			currentPass = key
			break
		}
	}
	return currentPass
}

func generateNewSalt() ([]byte, error) {
	salt := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
