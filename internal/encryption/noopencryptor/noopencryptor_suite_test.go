package noopencryptor

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestNoopEncryptor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "No-op Encryptor Suite")
}
