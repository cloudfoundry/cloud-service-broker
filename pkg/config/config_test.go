// From Kibosh

package config_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/pivotal/cloud-service-broker/pkg/config"
)

var _ = Describe("Config", func() {
	Context("config parsing", func() {

		It("sets defaults", func() {
			c, err := Parse()
			Expect(err).To(BeNil())
			Expect(c).ToNot(BeNil())

			Expect(c.CredStoreConfig.HasCredHubConfig()).To(BeFalse())
		})

		Context("credstore config", func() {
			It("parses credstore config", func() {
				os.Setenv("CH_CRED_HUB_URL", "https://credhub.example.com")
				os.Setenv("CH_UAA_URL", "https://uaa.example.com")
				os.Setenv("CH_UAA_CLIENT_NAME", "my-client")
				os.Setenv("CH_UAA_CLIENT_SECRET", "my-secret")
				os.Setenv("CH_SKIP_SSL_VALIDATION", "true")

				c, err := Parse()
				Expect(err).To(BeNil())
				Expect(c).ToNot(BeNil())

				Expect(c.CredStoreConfig.HasCredHubConfig()).To(BeTrue())

				Expect(c.CredStoreConfig.CredHubURL).To(Equal("https://credhub.example.com"))
				Expect(c.CredStoreConfig.UaaURL).To(Equal("https://uaa.example.com"))
			})
		})
	})
})
