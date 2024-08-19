// Package storage implements a Database Access Object (DAO)
package storage

import "gorm.io/gorm"

type Storage struct {
	db         *gorm.DB
	encryptor  Encryptor
	InProgress map[string]bool
}

func New(db *gorm.DB, encryptor Encryptor) *Storage {
	return &Storage{
		db:         db,
		encryptor:  encryptor,
		InProgress: map[string]bool{},
	}
}

type JSONObject map[string]any
