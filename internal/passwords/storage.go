package passwords

import (
	"errors"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"gorm.io/gorm"
)

type passwordMetadata struct {
	Label   string
	Salt    [32]byte
	Canary  string
	Primary bool
}

func savePasswordMetadata(db *gorm.DB, p passwordMetadata) error {
	return db.Create(&models.PasswordMetadata{
		Label:   p.Label,
		Salt:    p.Salt[:],
		Canary:  p.Canary,
		Primary: p.Primary,
	}).Error
}

func findPasswordMetadata(db *gorm.DB, label string) (passwordMetadata, bool, error) {
	var receiver models.PasswordMetadata
	err := db.Where("label = ?", label).First(&receiver).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return passwordMetadata{}, false, nil
	case err != nil:
		return passwordMetadata{}, false, err
	}

	var salt [32]byte
	copy(salt[:], receiver.Salt)

	return passwordMetadata{
		Label:   receiver.Label,
		Salt:    salt,
		Canary:  receiver.Canary,
		Primary: receiver.Primary,
	}, true, nil
}
