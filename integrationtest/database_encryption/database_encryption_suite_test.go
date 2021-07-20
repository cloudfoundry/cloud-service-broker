package database_encryption_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDatabaseEncryption(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Database Encryption Integration Suite")
}
