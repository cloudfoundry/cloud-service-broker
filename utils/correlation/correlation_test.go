package correlation_test

import (
	"context"

	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/correlation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/middlewares"
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

	When("there is a request ID in the context", func() {
		It("adds it to the logger data", func() {
			const id = "6aa85874-9d04-11eb-a03b-73ee7bd59e49"
			ctx := context.WithValue(context.TODO(), middlewares.RequestIdentityKey, id)

			data := correlation.ID(ctx)

			Expect(data).To(HaveKeyWithValue("request-id", id))
			Expect(data).To(HaveLen(1))
		})
	})

	When("both correlation ID and request ID are in the context", func() {
		It("adds it to the logger data", func() {
			const cid = "417a8ca4-994b-11eb-a555-b30bdd8a2a34"
			const rid = "6aa85874-9d04-11eb-a03b-73ee7bd59e49"
			ctx := context.WithValue(context.TODO(), middlewares.CorrelationIDKey, cid)
			ctx = context.WithValue(ctx, middlewares.RequestIdentityKey, rid)

			data := correlation.ID(ctx)

			Expect(data).To(HaveKeyWithValue("correlation-id", cid))
			Expect(data).To(HaveKeyWithValue("request-id", rid))
			Expect(data).To(HaveLen(2))
		})
	})

	When("the context is empty", func() {
		It("returns an empty data object", func() {
			data := correlation.ID(context.TODO())
			Expect(data).To(BeEmpty())
		})
	})
})
