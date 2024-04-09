// Package passwordcombiner combines passwords with salt to create encryption keys
package passwordcombiner

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/encryption/gcmencryptor"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/encryption/passwordparser"
)

func Combine(db *gorm.DB, parsed []passwordparser.PasswordEntry, storedPassMetadata []models.PasswordMetadata) (CombinedPasswords, error) {
	stored, storedPrimary, err := storedWithPrimary(storedPassMetadata)
	if err != nil {
		return nil, err
	}

	labels := make(map[string]struct{})
	var result CombinedPasswords
	for _, p := range parsed {
		labels[p.Label] = struct{}{}
		combinedPassword := func() (CombinedPassword, error) {
			s, ok := stored[p.Label]
			switch ok {
			case true:
				return mergeWithStoredMetadata(s, p)
			default:
				return saveNewPasswordMetadata(db, p)
			}
		}

		entry, err := combinedPassword()
		if err != nil {
			return nil, err
		}
		result = append(result, entry)
	}

	if _, ok := labels[storedPrimary]; storedPrimary != "" && !ok {
		return nil, fmt.Errorf("the password labelled %q must be supplied to decrypt the database", storedPrimary)
	}

	return result, nil
}

func saveNewPasswordMetadata(db *gorm.DB, p passwordparser.PasswordEntry) (CombinedPassword, error) {
	salt, err := randomSalt()
	if err != nil {
		return CombinedPassword{}, err
	}

	e := encryptor(p.Secret, salt)

	canary, err := encryptCanary(e)
	if err != nil {
		return CombinedPassword{}, err
	}

	err = db.Create(&models.PasswordMetadata{
		Label:   p.Label,
		Salt:    salt,
		Canary:  canary,
		Primary: false, // Primary updated after successful rotation
	}).Error
	if err != nil {
		return CombinedPassword{}, err
	}

	return CombinedPassword{
		Label:             p.Label,
		Secret:            p.Secret,
		Salt:              salt,
		Encryptor:         e,
		configuredPrimary: p.Primary,
	}, nil
}

func mergeWithStoredMetadata(s models.PasswordMetadata, p passwordparser.PasswordEntry) (CombinedPassword, error) {
	e := encryptor(p.Secret, s.Salt)

	if err := decryptCanary(e, s.Canary, p.Label); err != nil {
		return CombinedPassword{}, err
	}

	return CombinedPassword{
		Label:             p.Label,
		Secret:            p.Secret,
		Salt:              s.Salt,
		Encryptor:         e,
		configuredPrimary: p.Primary,
		storedPrimary:     s.Primary,
	}, nil
}

func storedWithPrimary(stored []models.PasswordMetadata) (map[string]models.PasswordMetadata, string, error) {
	var primary string
	result := make(map[string]models.PasswordMetadata)
	for _, p := range stored {
		result[p.Label] = p
		if p.Primary {
			switch primary {
			case "":
				primary = p.Label
			default:
				return nil, "", errors.New("corrupt database - more than one primary found in table password_metadata; mark only one as primary but do not remove rows")
			}
		}
	}

	return result, primary, nil
}

func encryptor(secret string, salt []byte) gcmencryptor.GCMEncryptor {
	switch {
	case len(secret) < 20:
		panic("invalid secret complexity for key generation")
	case len(salt) != 32:
		panic("invalid salt complexity for key generation")
	}

	var key [32]byte
	copy(key[:], pbkdf2.Key([]byte(secret), salt, 100000, 32, sha256.New))
	return gcmencryptor.New(key)
}
