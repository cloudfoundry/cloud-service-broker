package packer_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPacker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Packer Suite")
}
