package manifest_test

import (
	_ "embed"
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/platform"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/manifest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Parser", func() {
	It("can parse a manifest", func() {
		m, err := manifest.Parse(fakeManifest())

		Expect(err).NotTo(HaveOccurred())
		Expect(m).To(Equal(&manifest.Manifest{
			PackVersion: 1,
			Name:        "gcp-services",
			Version:     "0.1.0",
			Metadata:    map[string]string{"author": "VMware"},
			Platforms: []platform.Platform{
				{
					Os:   "linux",
					Arch: "amd64",
				},
				{
					Os:   "darwin",
					Arch: "amd64",
				},
			},
			TerraformResources: []manifest.TerraformResource{
				{
					Name:    "terraform",
					Version: "1.1.4",
					Source:  "https://github.com/hashicorp/terraform/archive/v1.1.4.zip",
				},
				{
					Name:    "terraform-provider-random",
					Version: "3.1.0",
					Source:  "https://github.com/terraform-providers/terraform-provider-random/archive/v3.1.0.zip",
				},
			},
			ServiceDefinitions: []string{
				"google-storage.yml",
				"google-redis.yml",
				"google-mysql.yml",
			},
			Parameters: []manifest.Parameter{
				{
					Name:        "param1",
					Description: "something about the parameter",
				},
			},
			RequiredEnvVars: []string{"FOO", "BAR"},
			EnvConfigMapping: map[string]string{
				"GOOGLE_CREDENTIALS": "gcp.credentials",
				"GOOGLE_PROJECT":     "gcp.project",
			},
		}))
	})

	When("there are multiple Terraform versions", func() {
		It("can parse the manifest", func() {
			m, err := manifest.Parse(fakeManifest(
				withAdditionalEntry("terraform_binaries", map[string]interface{}{
					"name":    "terraform",
					"version": "1.1.5",
					"default": false,
				}),
				withAdditionalEntry("terraform_binaries", map[string]interface{}{
					"name":    "terraform",
					"version": "1.1.6",
					"default": true,
				}),
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(m.TerraformResources).To(ContainElements(
				manifest.TerraformResource{
					Name:    "terraform",
					Version: "1.1.4",
					Source:  "https://github.com/hashicorp/terraform/archive/v1.1.4.zip",
					Default: false,
				},
				manifest.TerraformResource{
					Name:    "terraform",
					Version: "1.1.5",
					Default: false,
				},
				manifest.TerraformResource{
					Name:    "terraform",
					Version: "1.1.6",
					Default: true,
				},
			))
		})

		When("none are marked as default", func() {
			It("fails", func() {
				m, err := manifest.Parse(fakeManifest(withAdditionalEntry("terraform_binaries", map[string]interface{}{
					"name":    "terraform",
					"version": "1.1.5",
					"default": false,
				})))
				Expect(err).To(MatchError("error validating manifest: multiple Terraform versions, but none marked as default: terraform_binaries"))
				Expect(m).To(BeNil())
			})
		})

		When("more than one is marked as default", func() {
			It("fails", func() {
				m, err := manifest.Parse(fakeManifest(
					withAdditionalEntry("terraform_binaries", map[string]interface{}{
						"name":    "terraform",
						"version": "1.1.5",
						"default": true,
					}),
					withAdditionalEntry("terraform_binaries", map[string]interface{}{
						"name":    "terraform",
						"version": "1.1.6",
						"default": true,
					}),
				))
				Expect(err).To(MatchError("error validating manifest: multiple Terraform versions, and multiple marked as default: terraform_binaries"))
				Expect(m).To(BeNil())
			})

		})

		When("there are duplicate versions", func() {
			It("fails", func() {
				m, err := manifest.Parse(fakeManifest(withAdditionalEntry("terraform_binaries", map[string]interface{}{
					"name":    "terraform",
					"version": "1.1.4",
					"default": true,
				})))
				Expect(m).To(BeNil())
				Expect(err).To(MatchError("error validating manifest: duplicated value, must be unique: 1.1.4: version"))
			})
		})
	})

	DescribeTable(
		"missing fields",
		func(field string) {
			m, err := manifest.Parse(fakeManifest(without(field)))

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("missing field(s): %s", field))))
			Expect(m).To(BeNil())
		},
		Entry("pack version", "packversion"),
		Entry("name", "name"),
		Entry("version", "version"),
		Entry("platforms", "platforms"),
		Entry("service definitions", "service_definitions"),
	)

	DescribeTable("missing platform data",
		func(insert string, value map[string]interface{}) {
			m, err := manifest.Parse(fakeManifest(withAdditionalEntry("platforms", value)))

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("missing field(s): %s", insert))))
			Expect(m).To(BeNil())
		},
		Entry("os", "platforms[2].os", map[string]interface{}{"arch": "amd64"}),
		Entry("arch", "platforms[2].arch", map[string]interface{}{"os": "linux"}),
	)

	DescribeTable("missing terraform binary data",
		func(insert string, value map[string]interface{}) {
			m, err := manifest.Parse(fakeManifest(withAdditionalEntry("terraform_binaries", value)))

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("missing field(s): %s", insert))))
			Expect(m).To(BeNil())
		},
		Entry("name", "terraform_binaries[2].name", map[string]interface{}{"version": "1.2.3", "source": "https://github.com/terraform-providers/terraform-provider-random/archive/v3.1.0.zip"}),
		Entry("version", "terraform_binaries[2].version", map[string]interface{}{"name": "hello"}),
	)

	DescribeTable("missing parameter data",
		func(insert string, value map[string]interface{}) {
			m, err := manifest.Parse(fakeManifest(withAdditionalEntry("parameters", value)))

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("missing field(s): %s", insert))))
			Expect(m).To(BeNil())
		},
		Entry("name", "parameters[1].name", map[string]interface{}{"description": "something"}),
		Entry("description", "parameters[1].description", map[string]interface{}{"name": "hello"}),
	)

	When("the packversion is invalid", func() {
		It("fails", func() {
			m, err := manifest.Parse(fakeManifest(with("packversion", 2)))

			Expect(err).To(MatchError(ContainSubstring("invalid value: 2: packversion")))
			Expect(m).To(BeNil())
		})
	})

	When("there are unknown fields", func() {
		It("fails", func() {
			m, err := manifest.Parse(fakeManifest(with("foo", "bar")))

			Expect(err).To(MatchError(ContainSubstring("field foo not found in type manifest.Manifest")))
			Expect(m).To(BeNil())
		})
	})

	When("the default flag is applied to something that isn't Terraform", func() {
		It("fails", func() {
			m, err := manifest.Parse(fakeManifest(withAdditionalEntry("terraform_binaries", map[string]interface{}{
				"name":    "terraform-provider-random",
				"version": "3.1.0",
				"source":  "https://github.com/terraform-providers/terraform-provider-random/archive/v3.1.0.zip",
				"default": true,
			})))
			Expect(err).To(MatchError(ContainSubstring("This field is only valid for `terraform`: terraform_binaries[2].default")))
			Expect(m).To(BeNil())
		})
	})
})

//go:embed test_manifest.yaml
var testManifest string

type option func(map[string]interface{})

func fakeManifest(opts ...option) []byte {
	var receiver map[string]interface{}
	err := yaml.Unmarshal([]byte(testManifest), &receiver)
	Expect(err).NotTo(HaveOccurred())

	for _, o := range opts {
		o(receiver)
	}

	marshalled, err := yaml.Marshal(receiver)
	Expect(err).NotTo(HaveOccurred())
	return marshalled
}

func without(field string) option {
	return func(m map[string]interface{}) {
		delete(m, field)
	}
}

func with(key string, value interface{}) option {
	return func(m map[string]interface{}) {
		m[key] = value
	}
}

func withAdditionalEntry(key string, value map[string]interface{}) option {
	return func(m map[string]interface{}) {
		entries := m[key].([]interface{})
		m[key] = append(entries, value)
	}
}
