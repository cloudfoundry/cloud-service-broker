package gcmencryptor

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGCMEncryptor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GCM Encryptor Suite")
}
