package encryption

import (
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/encryption/compoundencryptor"
	"github.com/cloudfoundry/cloud-service-broker/internal/encryption/noopencryptor"
	"github.com/cloudfoundry/cloud-service-broker/internal/encryption/passwordcombiner"
	"github.com/cloudfoundry/cloud-service-broker/internal/encryption/passwordparser"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"gorm.io/gorm"
)

type Configuration struct {
	Encryptor              storage.Encryptor
	RotationEncryptor      storage.Encryptor
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
	var rotationEncyptor storage.Encryptor
	if parsedPrimary.Label != storedPrimary.Label {
		changed = true
		var decryptors []storage.Encryptor
		for _, e := range combined {
			decryptors = append(decryptors, e.Encryptor)
		}
		if !storedPrimaryOK {
			decryptors = append(decryptors, noopencryptor.New())
		}

		var encryptor storage.Encryptor
		if !parsedPrimaryOK {
			encryptor = noopencryptor.New()
		} else {
			encryptor = parsedPrimary.Encryptor
		}

		rotationEncyptor = compoundencryptor.New(encryptor, decryptors...)
	}

	switch {
	case enabled && !parsedPrimaryOK:
		return Configuration{}, errors.New("encryption enabled but no primary password is set; either disable encryption or to enable encryption, mark one of the passwords as primary")
	case !enabled && parsedPrimaryOK:
		return Configuration{}, errors.New("encryption disabled but a primary password is set; either enable encryption or to disable encryption, mark the existing passwords as non-primary but do not remove them")
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
	configuredMap := map[string]bool{}
	for _, c := range configured {
		configuredMap[c.Label] = true
	}

	var toDeleteList []string
	for _, s := range stored {
		if !configuredMap[s.Label] {
			toDeleteList = append(toDeleteList, s.Label)
		}
	}

	return toDeleteList
}

func loadPasswordMetadata(db *gorm.DB) ([]models.PasswordMetadata, error) {
	var stored []models.PasswordMetadata
	if err := db.Find(&stored).Error; err != nil {
		return nil, err
	}
	return stored, nil
}
