package encryption

import (
	"errors"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/compoundencryptor"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/noopencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/passwordcombiner"
	"gorm.io/gorm"
)

type Configuration struct {
	Encryptor          models.Encryptor
	RotationEncryptor  models.Encryptor
	Changed            bool
	ParsedPrimaryLabel string
	StoredPrimaryLabel string
}

func ParseConfiguration(db *gorm.DB, enabled bool, passwords string) (Configuration, error) {
	combined, err := passwordcombiner.CombineWithStoredMetadata(db, passwords)
	if err != nil {
		return Configuration{}, err
	}

	parsedPrimary, parsedPrimaryOK := combined.ParsedPrimary()
	storedPrimary, storedPrimaryOK := combined.StoredPrimary()

	switch {
	case enabled && !parsedPrimaryOK:
		return Configuration{}, errors.New("encryption is enabled but no primary password is set")
	case !enabled && parsedPrimaryOK:
		return Configuration{}, errors.New("encryption is disabled but a primary password is set")
	case !enabled && !parsedPrimaryOK:
		return noopEncryption(storedPrimary.Label)
	}

	changed := false
	var rotationEncyptor models.Encryptor
	if parsedPrimary.Label != storedPrimary.Label {
		changed = true
		var secondaryEncryptors []compoundencryptor.Encryptor
		for _, e := range combined {
			secondaryEncryptors = append(secondaryEncryptors, e.Encryptor)
		}
		if !storedPrimaryOK {
			secondaryEncryptors = append(secondaryEncryptors, noopencryptor.New())
		}

		rotationEncyptor = compoundencryptor.New(parsedPrimary.Encryptor, secondaryEncryptors...)
	}

	return Configuration{
		Encryptor:          parsedPrimary.Encryptor,
		RotationEncryptor:  rotationEncyptor,
		Changed:            changed,
		ParsedPrimaryLabel: labelName(parsedPrimary.Label),
		StoredPrimaryLabel: labelName(storedPrimary.Label),
	}, nil
}

func noopEncryption(storedPrimaryLabel string) (Configuration, error) {
	return Configuration{
		Encryptor:          noopencryptor.New(),
		RotationEncryptor:  nil,
		Changed:            false,
		ParsedPrimaryLabel: labelName(""),
		StoredPrimaryLabel: labelName(storedPrimaryLabel),
	}, nil
}

func labelName(label string) string {
	switch label {
	case "":
		return "none"
	default:
		return label
	}
}
