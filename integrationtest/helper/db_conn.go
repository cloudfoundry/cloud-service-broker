package helper

import (
	"github.com/glebarez/sqlite"
	"github.com/onsi/gomega"
	"gorm.io/gorm"
)

func (h *TestHelper) DBConn() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(h.databaseFile), &gorm.Config{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return db
}
