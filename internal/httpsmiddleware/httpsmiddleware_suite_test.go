package httpsmiddleware_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHTTPSMiddleware(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTPS middleware Suite")
}
