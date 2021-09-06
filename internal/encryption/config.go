package encryption

import "gorm.io/gorm"

type Configuration struct {
}

func ParseConfiguration(db *gorm.DB, enabled bool, passwords string) (Configuration, error) {
	return Configuration{}, nil
}
