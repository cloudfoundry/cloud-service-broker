package dbencryptor_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDbencryptor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DB Encryptor Suite")
}
