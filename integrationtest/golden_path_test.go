package integrationtest_test

import (
	"encoding/json"

	"github.com/cloudfoundry/cloud-service-broker/v3/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/testdrive"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Golden Path", func() {
	const (
		serviceOfferingGUID = "f18d50e2-cbf0-11ee-a64b-f7a425623295"
		servicePlanGUID     = "fd01df6a-cbf0-11ee-ac5b-fba53664a953"
	)

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("golden-path")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	It("can create and delete a service instance and a binding", func() {
		instance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))

		binding := must(broker.CreateBinding(instance))
		var receiver struct {
			Credentials struct {
				BindOutput      string `json:"bind_output"`
				ProvisionOutput string `json:"provision_output"`
			} `json:"credentials"`
		}
		Expect(json.Unmarshal([]byte(binding.Body), &receiver)).To(Succeed())
		Expect(receiver.Credentials.ProvisionOutput).To(MatchRegexp(`^[-0-9a-f]{36}$`))
		Expect(receiver.Credentials.BindOutput).To(MatchRegexp(`^[-0-9a-f]{36}$`))

		Expect(broker.DeleteBinding(instance, binding.GUID)).To(Succeed())
		Expect(broker.Deprovision(instance)).To(Succeed())
	})
})
