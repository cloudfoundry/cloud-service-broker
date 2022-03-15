package storage

import "gorm.io/gorm"

type Storage struct {
	db        *gorm.DB
	encryptor Encryptor
}

func New(db *gorm.DB, encryptor Encryptor) *Storage {
	return &Storage{
		db:        db,
		encryptor: encryptor,
	}
}

type JSONObject map[string]interface{}
