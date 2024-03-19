package tfproviderfqn_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/internal/tfproviderfqn"
)

const defaultRegistrydomain = "registry.opentofu.org"

var _ = Describe("TfProviderFQN", func() {
	Context("from name", func() {
		It("can be created from a name", func() {
			n, err := tfproviderfqn.New("terraform-provider-mysql", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(n.String()).To(Equal(fmt.Sprintf("%s/hashicorp/mysql", defaultRegistrydomain)))
		})

		When("the name has the wrong prefix", func() {
			It("returns an error", func() {
				n, err := tfproviderfqn.New("mysql", "")
				Expect(err).To(MatchError("name must have prefix: terraform-provider-"))
				Expect(n).To(BeZero())
			})
		})
	})

	Context("from provider", func() {
		It("can be created from the type", func() {
			n, err := tfproviderfqn.New("", "postgresql")
			Expect(err).NotTo(HaveOccurred())
			Expect(n.String()).To(Equal(fmt.Sprintf("%s/hashicorp/postgresql", defaultRegistrydomain)))
		})

		It("can be created from the namespace and type", func() {
			n, err := tfproviderfqn.New("", "cyrilgdn/postgresql")
			Expect(err).NotTo(HaveOccurred())
			Expect(n.String()).To(Equal(fmt.Sprintf("%s/cyrilgdn/postgresql", defaultRegistrydomain)))
		})

		It("can be created from the registry, namespace and type", func() {
			n, err := tfproviderfqn.New("", "myreg.mydomain.com/cyrilgdn/postgresql")
			Expect(err).NotTo(HaveOccurred())
			Expect(n.String()).To(Equal("myreg.mydomain.com/cyrilgdn/postgresql"))
		})

		When("the format is invalid", func() {
			It("returns an error", func() {
				n, err := tfproviderfqn.New("", "myreg/mydomain.com/cyrilgdn/postgresql")
				Expect(err).To(MatchError("invalid format; valid format is [<HOSTNAME>/]<NAMESPACE>/<TYPE>"))
				Expect(n).To(BeZero())
			})
		})
	})

	Context("empty", func() {
		It("is an empty string", func() {
			Expect(tfproviderfqn.TfProviderFQN{}.String()).To(BeEmpty())
		})
	})
})
