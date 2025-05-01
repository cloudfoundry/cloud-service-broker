package brokerchapi_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBrokerCHAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Broker CredHub API Suite")
}
