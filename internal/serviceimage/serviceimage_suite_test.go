package serviceimage_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServiceimage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ServiceImage Suite")
}
