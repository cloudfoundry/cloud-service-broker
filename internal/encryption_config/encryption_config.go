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
	"golang.org/x/crypto/pbkdf2"
)

const (
	canary = "some-test-value"
)

func GetEncryptionKey(encryptDB bool, rawPasswordBlocks string) (string, error) {
	logger := utils.NewLogger("cloud-service-broker")

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

		logger.Info("db encryption enabled")
		return string(key), nil
	} else {
		for _, key := range passwords {
			if key.Primary {
				return "", fmt.Errorf("encryption is disabled, but a primary encryption key was provided")
			}
		}
	}

	logger.Info("db encryption disabled")
	return "", nil
}

func generateKey(passwordConfigs PasswordConfigs) ([]byte, error) {
	primaryPassword := getPrimaryPassword(passwordConfigs)
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
		encryptor := models.ConfigureEncryption(string(key))
		encryptedCanary, err := encryptor.Encrypt([]byte(canary))
		if err != nil {
			return nil, fmt.Errorf("error setting canary value: %s", err)
		}

		err = db_service.CreateEncryptionDetail(
			context.Background(),
			&models.EncryptionDetail{
				Label:   primaryPassword.Label,
				Salt:    string(salt),
				Primary: true,
				Canary:  encryptedCanary,
			})
		if err != nil {
			return nil, fmt.Errorf("error storing encryption details: %s", err)
		}
	}

	return key, nil
}

func getPrimaryPassword(passwordConfigs PasswordConfigs) PasswordConfig {
	var primaryPassword PasswordConfig
	for _, p := range passwordConfigs {
		if p.Primary {
			primaryPassword = p
			break
		}
	}
	return primaryPassword
}

func generateNewSalt() ([]byte, error) {
	salt := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
