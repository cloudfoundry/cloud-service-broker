package broker_test

import (
	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/brokerfakes"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"golang.org/x/net/context"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker"
)

var _ = Describe("GetInstance", func() {
	It("is not implemented", func() {
		serviceBroker, err := broker.New(&broker.BrokerConfig{}, utils.NewLogger("brokers-test"), &brokerfakes.FakeStorage{})
		Expect(err).ToNot(HaveOccurred())

		_, err = serviceBroker.GetInstance(context.TODO(), "instance-id", domain.FetchInstanceDetails{})

		Expect(err).To(MatchError("the service_instances endpoint is unsupported"))
	})
})
