package compoundencryptor

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCompoundEncryptor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compound Encryptor Suite")
}
