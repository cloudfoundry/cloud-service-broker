package passwords

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/jinzhu/gorm"
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
	result := db.Where("label = ?", label).First(&receiver)
	switch {
	case result.RecordNotFound():
		return passwordMetadata{}, false, nil
	case result.Error != nil:
		return passwordMetadata{}, false, result.Error
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
