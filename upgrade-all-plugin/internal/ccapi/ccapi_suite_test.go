package ccapi_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCCapi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CCAPI Suite")
}
