package encryption

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"gorm.io/gorm"
)

func UpdatePasswordMetadata(db *gorm.DB, configuredPrimaryLabel string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var passwordMetadata []models.PasswordMetadata
		if err := tx.Where("`primary` = true").Find(&passwordMetadata).Error; err != nil {
			return err
		}

		if len(passwordMetadata) == 1 && passwordMetadata[0].Label == configuredPrimaryLabel {
			return nil
		}

		for _, p := range passwordMetadata {
			if err := tx.Model(&p).Update("primary", false).Error; err != nil {
				return err
			}
		}

		if configuredPrimaryLabel != "" {
			var primaryPasswordMetadata models.PasswordMetadata
			result := tx.Where("label = ?", configuredPrimaryLabel).First(&primaryPasswordMetadata)
			switch {
			case result.RowsAffected == 0:
				return fmt.Errorf("cannot find metadata for password labelled %q", configuredPrimaryLabel)
			case result.Error != nil:
				return result.Error
			}

			if err := tx.Model(&primaryPasswordMetadata).Update("primary", true).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
