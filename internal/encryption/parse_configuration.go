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
	Encryptor              models.Encryptor
	RotationEncryptor      models.Encryptor
	Changed                bool
	ConfiguredPrimaryLabel string
	StoredPrimaryLabel     string
}

func ParseConfiguration(db *gorm.DB, enabled bool, passwords string) (Configuration, error) {
	combined, err := passwordcombiner.CombineWithStoredMetadata(db, passwords)
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
	case !enabled && !parsedPrimaryOK:
		return noopEncryption(storedPrimary.Label, changed, rotationEncyptor)
	}

	return Configuration{
		Encryptor:              parsedPrimary.Encryptor,
		RotationEncryptor:      rotationEncyptor,
		Changed:                changed,
		ConfiguredPrimaryLabel: labelName(parsedPrimary.Label),
		StoredPrimaryLabel:     labelName(storedPrimary.Label),
	}, nil
}

func noopEncryption(storedPrimaryLabel string, changed bool, rotationEncryptor compoundencryptor.Encryptor) (Configuration, error) {
	return Configuration{
		Encryptor:              noopencryptor.New(),
		RotationEncryptor:      rotationEncryptor,
		Changed:                changed,
		ConfiguredPrimaryLabel: labelName(""),
		StoredPrimaryLabel:     labelName(storedPrimaryLabel),
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
