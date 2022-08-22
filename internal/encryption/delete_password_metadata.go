package encryption

import (
	"gorm.io/gorm"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
)

func DeletePasswordMetadata(db *gorm.DB, labels []string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var passwordMetadata []models.PasswordMetadata
		if err := tx.Where("label in (?)", labels).Delete(&passwordMetadata).Error; err != nil {
			return err
		}

		return nil
	})
}
