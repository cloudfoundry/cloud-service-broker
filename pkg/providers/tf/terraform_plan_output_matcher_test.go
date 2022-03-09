package tf

import (
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("CheckTerraformPlanOutput", func() {
	It("returns no errors if nothing is being changed", func() {
		logger := lager.NewLogger("test")
		output := CheckTerraformPlanOutput(logger, wrapper.ExecutionOutput{StdOut: `
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:

Terraform will perform the following actions:

Plan: 0 to add, 0 to change, 0 to destroy.

Changes to Outputs:
 + test = true
 - another_test = false
`})
		Expect(output).NotTo(HaveOccurred())
	})

	It("returns no errors if resources are being added", func() {
		logger := lager.NewLogger("test")
		output := CheckTerraformPlanOutput(logger, wrapper.ExecutionOutput{StdOut: `
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:

Terraform will perform the following actions:

Plan: 5 to add, 0 to change, 0 to destroy.

Changes to Outputs:
 + test = true
 - another_test = false
`})
		Expect(output).NotTo(HaveOccurred())
	})

	It("returns no errors if resources are being changed", func() {
		logger := lager.NewLogger("test")
		output := CheckTerraformPlanOutput(logger, wrapper.ExecutionOutput{StdOut: `
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:

Terraform will perform the following actions:

Plan: 0 to add, 6 to change, 0 to destroy.

Changes to Outputs:
 + test = true
 - another_test = false
`})
		Expect(output).NotTo(HaveOccurred())
	})

	It("fails if there are any deletes", func() {
		logger := lager.NewLogger("test")
		output := CheckTerraformPlanOutput(logger, wrapper.ExecutionOutput{StdOut: `
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:

Terraform will perform the following actions:

Plan: 0 to add, 0 to change, 1 to destroy.

Changes to Outputs:
 + test = true
 - another_test = false
`})
		Expect(output).To(HaveOccurred())
		Expect(output).To(MatchError("terraform plan shows that resources would be destroyed - cancelling subsume"))
	})
})
