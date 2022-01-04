package compoundencryptor

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCompoundEncryptor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compound Encryptor Suite")
}
