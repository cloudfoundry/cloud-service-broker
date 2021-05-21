package zippy_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestZippy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zippy Suite")
}
