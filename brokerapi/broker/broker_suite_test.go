package broker_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Broker Suite")
}

func must[A any](input A, err error) A {
	GinkgoHelper()
	Expect(err).NotTo(HaveOccurred())
	return input
}
