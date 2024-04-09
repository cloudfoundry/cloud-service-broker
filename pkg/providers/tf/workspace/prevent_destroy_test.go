package workspace_test

import (
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PreventDestroy", func() {
	It("can remove `prevent_destroy` directives from definition and definitions", func() {
		w := workspace.TerraformWorkspace{
			Modules: []workspace.ModuleDefinition{
				{
					Definition: `
resource "random_string" "one" {
  length = 10
  lifecycle {
    prevent_destroy = true
  }
}`,
					Definitions: map[string]string{
						"two": `
resource "random_string" "one" {
  length = 10
  lifecycle {
    prevent_destroy = true
  }
}

resource "random_string" "two" {
  length = 10
  lifecycle {
    ignore_changes  = ["stuff"]
    prevent_destroy = true
  }
}`,
					},
				},
			},
		}

		Expect(w.RemovePreventDestroy()).NotTo(HaveOccurred())

		Expect(w.Modules[0].Definition).To(Equal(`
resource "random_string" "one" {
  length = 10
  lifecycle {
    prevent_destroy = false
  }
}`,
		))
		Expect(w.Modules[0].Definitions).To(HaveKeyWithValue("two", `
resource "random_string" "one" {
  length = 10
  lifecycle {
    prevent_destroy = false
  }
}

resource "random_string" "two" {
  length = 10
  lifecycle {
    ignore_changes  = ["stuff"]
    prevent_destroy = false
  }
}`,
		))
	})

	It("ignores `prevent_destroy` fields not in a `lifecycle` block", func() {
		w := workspace.TerraformWorkspace{
			Modules: []workspace.ModuleDefinition{
				{
					Definition: `
resource "random_string" "one" {
  length          = 10
  prevent_destroy = true
}`,
				},
			},
		}

		Expect(w.RemovePreventDestroy()).NotTo(HaveOccurred())

		Expect(w.Modules[0].Definition).To(Equal(`
resource "random_string" "one" {
  length          = 10
  prevent_destroy = true
}`,
		))
	})

	It("ignores `lifecycle` blocks without a `prevent_destroy` field", func() {
		w := workspace.TerraformWorkspace{
			Modules: []workspace.ModuleDefinition{
				{
					Definition: `
resource "random_string" "fake" {
  length = 10
  lifecycle {
  }
}`,
				},
			},
		}

		Expect(w.RemovePreventDestroy()).NotTo(HaveOccurred())

		Expect(w.Modules[0].Definition).To(Equal(`
resource "random_string" "fake" {
  length = 10
  lifecycle {
  }
}`,
		))
	})

	When("there is an error parsing a single definition", func() {
		It("returns the error", func() {
			w := workspace.TerraformWorkspace{
				Modules: []workspace.ModuleDefinition{
					{
						Definition: `bad *** hcl`,
					},
				},
			}

			Expect(w.RemovePreventDestroy()).To(MatchError(ContainSubstring(
				"HCL parse error for module 0 definition: :1,1-4: Argument or block definition required",
			)))
		})
	})

	When("there is an error parsing one of the definitions", func() {
		It("returns the error", func() {
			w := workspace.TerraformWorkspace{
				Modules: []workspace.ModuleDefinition{
					{
						Definitions: map[string]string{
							"two": `bad *** hcl`,
						},
					},
				},
			}

			Expect(w.RemovePreventDestroy()).To(MatchError(ContainSubstring(
				`HCL parse error for module 0 definition "two": :1,1-4: Argument or block definition required`,
			)))
		})
	})
})
