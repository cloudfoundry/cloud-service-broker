package manifest_test

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TerraformResource", func() {
	Describe("GetProviderNamespace", func() {
		It("should return default namespace", func() {
			tfResource := manifest.TerraformResource{
				Name: "terraform-provider-mysql",
			}

			result := tfResource.GetProviderNamespace()

			Expect(result).To(Equal("hashicorp"))
		})

		When("provider is defined", func() {
			It("should return namespace defined in provider property", func() {
				tfResource := manifest.TerraformResource{
					Name:     "terraform-provider-postgresql",
					Provider: "cyrilgdn/postgresql",
				}

				result := tfResource.GetProviderNamespace()

				Expect(result).To(Equal("cyrilgdn"))
			})
		})
	})

	Describe("GetProviderType", func() {
		It("should return type defined in name property", func() {
			tfResource := manifest.TerraformResource{
				Name: "terraform-provider-mysql",
			}

			result := tfResource.GetProviderType()

			Expect(result).To(Equal("mysql"))
		})

		When("resource is not a terraform provider", func() {
			It("should return name as type", func() {
				tfResource := manifest.TerraformResource{
					Name: "mysql",
				}

				result := tfResource.GetProviderType()

				Expect(result).To(Equal("mysql"))
			})
		})

		When("provider is defined", func() {
			It("should return type defined in provider property", func() {
				tfResource := manifest.TerraformResource{
					Name:     "terraform-provider-postgre",
					Provider: "cyrilgdn/postgresql",
				}

				result := tfResource.GetProviderType()

				Expect(result).To(Equal("postgresql"))
			})
		})
	})
})
