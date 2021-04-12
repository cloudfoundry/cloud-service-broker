package correlation_test

import (
	"context"

	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/correlation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v7/middlewares"
)

var _ = Describe("Correlation", func() {
	When("there is a correlation ID in the context", func() {
		It("adds it to the logger data", func() {
			const id = "417a8ca4-994b-11eb-a555-a30bdd8a2a34"
			ctx := context.WithValue(context.TODO(), middlewares.CorrelationIDKey, id)

			data := correlation.ID(ctx)

			Expect(data).To(HaveKeyWithValue("correlation-id", id))
			Expect(data).To(HaveLen(1))
		})
	})

	When("the context is empty", func() {
		It("returns an empty data object", func() {
			data := correlation.ID(context.TODO())
			Expect(data).To(BeEmpty())
		})
	})
})
