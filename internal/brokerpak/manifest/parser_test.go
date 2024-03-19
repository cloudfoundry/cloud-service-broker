package manifest_test

import (
	_ "embed"
	"fmt"

	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
	"github.com/cloudfoundry/cloud-service-broker/internal/tfproviderfqn"
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
			TerraformVersions: []manifest.TerraformVersion{
				{
					Version:     version.Must(version.NewVersion("1.1.4")),
					Source:      "https://github.com/hashicorp/terraform/archive/v1.1.4.zip",
					URLTemplate: "https://releases.hashicorp.com/${name}/${version}/${name}_${version}_${os}_${arch}.zip",
				},
			},
			TerraformProviders: []manifest.TerraformProvider{
				{
					Name:    "terraform-provider-random",
					Version: version.Must(version.NewVersion("3.1.0")),
					Source:  "https://github.com/terraform-providers/terraform-provider-random/archive/v3.1.0.zip",
					Provider: tfproviderfqn.TfProviderFQN{
						Hostname:  "registry.terraform.io",
						Namespace: "other",
						Type:      "random",
					},
					URLTemplate: "https://releases.hashicorp.com/${name}/${version}/${name}_${version}_${os}_${arch}.zip",
				},
			},
			Binaries: []manifest.Binary{
				{
					Name:        "other-random-binary",
					Version:     "latest",
					Source:      "nothing-important",
					URLTemplate: "./tools/${name}/build/${name}_${version}_${os}_${arch}.zip",
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
			TerraformStateProviderReplacements: map[string]string{"registry.terraform.io/-/random": "registry.terraform.io/hashicorp/random"},
		}))
	})

	When("there are multiple Tofu versions", func() {
		It("can parse the manifest", func() {
			m, err := manifest.Parse(fakeManifest(
				withAdditionalEntry("terraform_binaries", map[string]any{
					"name":    "tofu",
					"version": "1.1.5",
					"default": false,
				}),
				withAdditionalEntry("terraform_binaries", map[string]any{
					"name":    "tofu",
					"version": "1.1.6",
					"default": true,
				}),
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(m.TerraformVersions).To(ConsistOf(
				manifest.TerraformVersion{
					Version:     version.Must(version.NewVersion("1.1.4")),
					Source:      "https://github.com/hashicorp/terraform/archive/v1.1.4.zip",
					Default:     false,
					URLTemplate: "https://releases.hashicorp.com/${name}/${version}/${name}_${version}_${os}_${arch}.zip",
				},
				manifest.TerraformVersion{
					Version: version.Must(version.NewVersion("1.1.5")),
					Default: false,
				},
				manifest.TerraformVersion{
					Version: version.Must(version.NewVersion("1.1.6")),
					Default: true,
				},
			))
		})

		When("none are marked as default", func() {
			It("fails", func() {
				m, err := manifest.Parse(fakeManifest(withAdditionalEntry("terraform_binaries", map[string]any{
					"name":    "tofu",
					"version": "1.1.5",
					"default": false,
				})))
				Expect(err).To(MatchError("error validating manifest: multiple Tofu versions, but none marked as default: terraform_binaries"))
				Expect(m).To(BeNil())
			})
		})

		When("more than one is marked as default", func() {
			It("fails", func() {
				m, err := manifest.Parse(fakeManifest(
					withAdditionalEntry("terraform_binaries", map[string]any{
						"name":    "tofu",
						"version": "1.1.5",
						"default": true,
					}),
					withAdditionalEntry("terraform_binaries", map[string]any{
						"name":    "tofu",
						"version": "1.1.6",
						"default": true,
					}),
				))
				Expect(err).To(MatchError("error validating manifest: multiple Tofu versions, and multiple marked as default: terraform_binaries"))
				Expect(m).To(BeNil())
			})
		})

		When("the default is not the highest version", func() {
			It("fails", func() {
				m, err := manifest.Parse(fakeManifest(
					withAdditionalEntry("terraform_binaries", map[string]any{
						"name":    "tofu",
						"version": "1.1.5",
					}),
					withAdditionalEntry("terraform_binaries", map[string]any{
						"name":    "tofu",
						"version": "1.1.6",
						"default": true,
					}),
					withAdditionalEntry("terraform_binaries", map[string]any{
						"name":    "tofu",
						"version": "1.1.7",
					}),
				))
				Expect(err).To(MatchError("error validating manifest: default version of Tofu must be the highest version: terraform_binaries"))
				Expect(m).To(BeNil())
			})
		})

		When("there are duplicate versions", func() {
			It("fails", func() {
				m, err := manifest.Parse(fakeManifest(withAdditionalEntry("terraform_binaries", map[string]any{
					"name":    "tofu",
					"version": "1.1.4",
					"default": true,
				})))
				Expect(m).To(BeNil())
				Expect(err).To(MatchError("error validating manifest: duplicated value, must be unique: 1.1.4: version"))
			})
		})
	})

	Context("terraform_state_provider_replacements", func() {
		It("can parse and validate the provider replacements", func() {
			m, err := manifest.Parse(fakeManifest(with("terraform_state_provider_replacements",
				map[string]string{
					"registry.terraform.io/-/random": "registry.terraform.io/hashicorp/random",
				},
			)))

			Expect(err).NotTo(HaveOccurred())

			Expect(m.TerraformStateProviderReplacements).To(Equal(map[string]string{
				"registry.terraform.io/-/random": "registry.terraform.io/hashicorp/random",
			}))
		})
	})

	Context("terraform_upgrade_path", func() {
		It("can parse and validate the upgrade path", func() {
			m, err := manifest.Parse(fakeManifest(
				withAdditionalEntry("terraform_binaries", map[string]any{
					"name":    "tofu",
					"version": "4.5.6",
					"default": true,
				}),
				with("terraform_upgrade_path",
					[]map[string]any{
						{"version": "1.1.4"},
						{"version": "4.5.6"},
					},
				),
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(m.TerraformUpgradePath).To(Equal([]*version.Version{
				version.Must(version.NewVersion("1.1.4")),
				version.Must(version.NewVersion("4.5.6")),
			}))
		})

		It("must be semver", func() {
			m, err := manifest.Parse(fakeManifest(with("terraform_upgrade_path",
				[]map[string]any{
					{"version": "non-semver"},
				},
			)))
			Expect(err).To(MatchError(ContainSubstring("invalid value: non-semver: terraform_upgrade_path[0].version")))
			Expect(m).To(BeNil())
		})

		It("must be in order", func() {
			m, err := manifest.Parse(fakeManifest(with("terraform_upgrade_path",
				[]map[string]any{
					{"version": "1.2.3"},
					{"version": "1.2.4"},
					{"version": "1.2.1"},
				},
			)))
			Expect(err).To(MatchError(ContainSubstring(`expect versions to be in ascending order: "1.2.1" <= "1.2.4": terraform_upgrade_path[2].version`)))
			Expect(m).To(BeNil())
		})

		It("must have a corresponding terraform binary", func() {
			m, err := manifest.Parse(fakeManifest(with("terraform_upgrade_path",
				[]map[string]any{
					{"version": "1.2.3"},
				},
			)))
			Expect(err).To(MatchError(ContainSubstring(`no corresponding terrafom resource for terraform version "1.2.3": terraform_upgrade_path[0].version`)))
			Expect(m).To(BeNil())
		})

		It("must upgrade up to the default version", func() {
			m, err := manifest.Parse(fakeManifest(with("terraform_upgrade_path",
				[]map[string]any{
					{"version": "1.1.4"},
				},
			)))
			Expect(err).To(MatchError(ContainSubstring(`upgrade path does not terminate at default version: terraform_upgrade_path[0].version`)))
			Expect(m).To(BeNil())
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
		func(insert string, value map[string]any) {
			m, err := manifest.Parse(fakeManifest(withAdditionalEntry("platforms", value)))

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("missing field(s): %s", insert))))
			Expect(m).To(BeNil())
		},
		Entry("os", "platforms[2].os", map[string]any{"arch": "amd64"}),
		Entry("arch", "platforms[2].arch", map[string]any{"os": "linux"}),
	)

	DescribeTable("missing terraform binary data",
		func(insert string, value map[string]any) {
			m, err := manifest.Parse(fakeManifest(withAdditionalEntry("terraform_binaries", value)))

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("missing field(s): %s", insert))))
			Expect(m).To(BeNil())
		},
		Entry("name", "terraform_binaries[3].name", map[string]any{"version": "1.2.3", "source": "https://github.com/terraform-providers/terraform-provider-random/archive/v3.1.0.zip"}),
		Entry("version", "terraform_binaries[3].version", map[string]any{"name": "hello"}),
	)

	DescribeTable("missing parameter data",
		func(insert string, value map[string]any) {
			m, err := manifest.Parse(fakeManifest(withAdditionalEntry("parameters", value)))

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("missing field(s): %s", insert))))
			Expect(m).To(BeNil())
		},
		Entry("name", "parameters[1].name", map[string]any{"description": "something"}),
		Entry("description", "parameters[1].description", map[string]any{"name": "hello"}),
	)

	DescribeTable("terraform provider locations",
		func(provider, expected string) {
			m, err := manifest.Parse(fakeManifest(withAdditionalEntry("terraform_binaries", map[string]any{
				"name":     "terraform-provider-foo",
				"version":  "1.2.3",
				"provider": provider,
			})))
			Expect(err).NotTo(HaveOccurred())
			Expect(m.TerraformProviders[1].Provider.String()).To(Equal(expected))
		},
		Entry("empty 'provider' field", "", "registry.opentofu.org/hashicorp/foo"),
		Entry("just type", "lala", "registry.opentofu.org/hashicorp/lala"),
		Entry("type and namespace", "mycorp/lala", "registry.opentofu.org/mycorp/lala"),
		Entry("fully qualified", "mything.io/mycorp/lala", "mything.io/mycorp/lala"),
	)

	When("yaml is invalid", func() {
		It("fails", func() {
			m, err := manifest.Parse([]byte(`not yam-ls:`))

			Expect(err).To(MatchError(ContainSubstring("error parsing manifest: yaml: unmarshal errors:")))
			Expect(m).To(BeNil())
		})
	})

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

			Expect(err).To(MatchError(ContainSubstring("field foo not found in type manifest.parser")))
			Expect(m).To(BeNil())
		})
	})

	When("the default flag is applied to something that isn't Terraform", func() {
		It("fails", func() {
			m, err := manifest.Parse(fakeManifest(withAdditionalEntry("terraform_binaries", map[string]any{
				"name":    "terraform-provider-random",
				"version": "3.1.0",
				"source":  "https://github.com/terraform-providers/terraform-provider-random/archive/v3.1.0.zip",
				"default": true,
			})))
			Expect(err).To(MatchError(ContainSubstring("This field is only valid for `tofu`: terraform_binaries[3].default")))
			Expect(m).To(BeNil())
		})
	})

	When("a terraform version is not valid semver", func() {
		It("fails", func() {
			m, err := manifest.Parse(fakeManifest(withAdditionalEntry("terraform_binaries", map[string]any{
				"name":    "tofu",
				"version": "not.semver",
				"default": true,
			})))

			Expect(err).To(MatchError(ContainSubstring(`error validating manifest: Malformed version: not.semver: terraform_binaries[3].version`)))
			Expect(m).To(BeNil())
		})
	})
})

//go:embed test_manifest.yaml
var testManifest []byte

type option func(map[string]any)

func fakeManifest(opts ...option) []byte {
	var receiver map[string]any
	err := yaml.Unmarshal(testManifest, &receiver)
	Expect(err).NotTo(HaveOccurred())

	for _, o := range opts {
		o(receiver)
	}

	marshalled, err := yaml.Marshal(receiver)
	Expect(err).NotTo(HaveOccurred())
	return marshalled
}

func without(field string) option {
	return func(m map[string]any) {
		delete(m, field)
	}
}

func with(key string, value any) option {
	return func(m map[string]any) {
		m[key] = value
	}
}

func withAdditionalEntry(key string, value map[string]any) option {
	return func(m map[string]any) {
		entries := m[key].([]any)
		m[key] = append(entries, value)
	}
}
