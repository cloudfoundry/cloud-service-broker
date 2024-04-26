package broker_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"golang.org/x/net/context"

	"github.com/cloudfoundry/cloud-service-broker/v3/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/v3/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils"
)

var _ = Describe("GetBinding", func() {
	It("is not implemented", func() {
		serviceBroker, err := broker.New(&broker.BrokerConfig{}, &brokerfakes.FakeStorage{}, utils.NewLogger("brokers-test"))
		Expect(err).ToNot(HaveOccurred())

		_, err = serviceBroker.GetBinding(context.TODO(), "instance-id", "binding-id", domain.FetchBindingDetails{})

		Expect(err).To(MatchError("the service_bindings endpoint is unsupported"))
	})
})
