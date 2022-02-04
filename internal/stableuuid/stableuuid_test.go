package stableuuid_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/stableuuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StableUUID", func() {
	It("generates the same UUID given the same strings", func() {
		const s1 = "foo"
		const s2 = "bar"
		u1 := stableuuid.FromStrings(s1, s2)
		u2 := stableuuid.FromStrings(s1, s2)
		u3 := stableuuid.FromStrings(s1, s2)
		Expect(u1).To(HaveLen(36))
		Expect(u1).To(Equal(u2))
		Expect(u1).To(Equal(u3))
	})

	It("generates a different UUID given different inputs strings", func() {
		u1 := stableuuid.FromStrings("foo", "bar")
		u2 := stableuuid.FromStrings("foo", "baz")
		u3 := stableuuid.FromStrings("guz", "bar")
		Expect(u1).To(HaveLen(36))
		Expect(u2).To(HaveLen(36))
		Expect(u3).To(HaveLen(36))
		Expect(u1).NotTo(Equal(u2))
		Expect(u1).NotTo(Equal(u3))
		Expect(u2).NotTo(Equal(u3))
	})
})
