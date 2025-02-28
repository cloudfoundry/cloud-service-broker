package broker_test

import (
	"code.cloudfoundry.org/brokerapi/v13/domain"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"

	"github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils"

	"github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker"
)

var _ = Describe("LastBindingOperation", func() {
	It("is not implemented for async bindings", func() {
		serviceBroker, err := broker.New(&broker.BrokerConfig{}, &brokerfakes.FakeStorage{}, utils.NewLogger("brokers-test"))
		Expect(err).ToNot(HaveOccurred())

		_, err = serviceBroker.LastBindingOperation(context.TODO(), "instance-id", "binding-id", domain.PollDetails{})

		Expect(err).To(MatchError("This service plan requires client support for asynchronous service operations."))
	})
})
