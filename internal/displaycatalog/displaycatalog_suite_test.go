package displaycatalog_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDisplayCatalog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Display Catalog Suite")
}
