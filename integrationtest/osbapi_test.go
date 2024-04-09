package integrationtest_test

import (
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/v2/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
)

var _ = Describe("OSBAPI spec", func() {
	const (
		serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		servicePlanGUID     = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
	)

	var broker *testdrive.Broker

	BeforeEach(func() {
		brokerpak := must(packer.BuildBrokerpak(csb, fixtures("osbapi")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	When("a service instance does not exist", func() {
		Context("service instance last operation", func() {
			It("returns HTTP 410 as per OSBAPI spec", func() {
				_, err := broker.LastOperation("does-not-exist")

				Expect(err).To(matchUnexpectedStatusError(http.StatusGone, "{}"))
			})
		})

		Context("delete service instance", func() {
			It("returns HTTP 410 as per OSBAPI spec", func() {
				err := broker.Deprovision(testdrive.ServiceInstance{
					GUID:                "does-not-exist",
					ServicePlanGUID:     servicePlanGUID,
					ServiceOfferingGUID: serviceOfferingGUID,
				})

				Expect(err).To(matchUnexpectedStatusError(http.StatusGone, "{}"))
			})
		})

		Context("delete service binding", func() {
			It("returns HTTP 410 as per OSBAPI spec", func() {
				err := broker.DeleteBinding(testdrive.ServiceInstance{
					GUID:                "does-not-exist",
					ServicePlanGUID:     servicePlanGUID,
					ServiceOfferingGUID: serviceOfferingGUID,
				}, "does-not-exist")

				Expect(err).To(matchUnexpectedStatusError(http.StatusGone, "{}"))
			})
		})
	})

	When("a service instance exists", func() {
		var serviceInstance testdrive.ServiceInstance

		BeforeEach(func() {
			serviceInstance = must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
		})

		Context("provision instance with same GUID", func() {
			It("returns HTTP 409 as per OSBAPI spec", func() {
				_, err := broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionServiceInstanceGUID(serviceInstance.GUID))
				Expect(err).To(matchUnexpectedStatusError(http.StatusConflict, "{}"))
			})
		})

		When("a service binding does not exist", func() {
			Context("delete service binding", func() {
				It("returns HTTP 410 as per OSBAPI spec", func() {
					err := broker.DeleteBinding(serviceInstance, "does-not-exist")

					Expect(err).To(matchUnexpectedStatusError(http.StatusGone, "{}"))
				})
			})
		})

		When("a service binding exists", func() {
			var binding testdrive.ServiceBinding

			BeforeEach(func() {
				binding = must(broker.CreateBinding(serviceInstance))
			})

			Context("create binding with same GUID", func() {
				It("returns HTTP 409 as per OSBAPI spec", func() {
					_, err := broker.CreateBinding(serviceInstance, testdrive.WithBindingGUID(binding.GUID))

					Expect(err).To(matchUnexpectedStatusError(http.StatusConflict, `{"description":"binding already exists"}`))
				})
			})
		})
	})
})

func matchUnexpectedStatusError(code int, body string) types.GomegaMatcher {
	return gstruct.PointTo(gstruct.MatchAllFields(gstruct.Fields{
		"StatusCode":   Equal(code),
		"ResponseBody": MatchJSON(body),
	}))
}
