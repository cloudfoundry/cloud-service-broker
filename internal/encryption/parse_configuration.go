package encryption

import (
	"errors"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/compoundencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/passwordparser"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/noopencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/passwordcombiner"
	"gorm.io/gorm"
)

type Configuration struct {
	Encryptor              models.Encryptor
	RotationEncryptor      models.Encryptor
	Changed                bool
	ConfiguredPrimaryLabel string
	StoredPrimaryLabel     string
	ToDeleteLabels         []string
}

func ParseConfiguration(db *gorm.DB, enabled bool, passwords string) (Configuration, error) {
	configured, err := passwordparser.Parse(passwords)
	if err != nil {
		return Configuration{}, err
	}

	stored, err := loadPasswordMetadata(db)
	if err != nil {
		return Configuration{}, err
	}

	combined, err := passwordcombiner.Combine(db, configured, stored)
	if err != nil {
		return Configuration{}, err
	}

	parsedPrimary, parsedPrimaryOK := combined.ConfiguredPrimary()
	storedPrimary, storedPrimaryOK := combined.StoredPrimary()

	changed := false
	var rotationEncyptor models.Encryptor
	if parsedPrimary.Label != storedPrimary.Label {
		changed = true
		var decryptors []compoundencryptor.Encryptor
		for _, e := range combined {
			decryptors = append(decryptors, e.Encryptor)
		}
		if !storedPrimaryOK {
			decryptors = append(decryptors, noopencryptor.New())
		}

		var encryptor compoundencryptor.Encryptor
		if !parsedPrimaryOK {
			encryptor = noopencryptor.New()
		} else {
			encryptor = parsedPrimary.Encryptor
		}

		rotationEncyptor = compoundencryptor.New(encryptor, decryptors...)
	}

	switch {
	case enabled && !parsedPrimaryOK:
		return Configuration{}, errors.New("encryption is enabled but no primary password is set")
	case !enabled && parsedPrimaryOK:
		return Configuration{}, errors.New("encryption is disabled but a primary password is set")
	}

	result := Configuration{
		ConfiguredPrimaryLabel: parsedPrimary.Label,
		StoredPrimaryLabel:     storedPrimary.Label,
		Changed:                changed,
		RotationEncryptor:      rotationEncyptor,
		ToDeleteLabels:         toDelete(stored, configured),
	}

	if !enabled && !parsedPrimaryOK {
		result.Encryptor = noopencryptor.New()
	} else {
		result.Encryptor = parsedPrimary.Encryptor
	}

	return result, nil
}

func toDelete(stored []models.PasswordMetadata, configured []passwordparser.PasswordEntry) []string {
	var toDelete []string
	for _, s := range stored {
		shouldDelete := true
		for _, c := range configured {
			if c.Label == s.Label {
				shouldDelete = false
				break
			}
		}
		if shouldDelete {
			toDelete = append(toDelete, s.Label)
		}
	}
	return toDelete
}

func loadPasswordMetadata(db *gorm.DB) ([]models.PasswordMetadata, error) {
	var stored []models.PasswordMetadata
	if err := db.Find(&stored).Error; err != nil {
		return nil, err
	}
	return stored, nil
}
