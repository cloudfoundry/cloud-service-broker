package tf_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("HclParser", func() {
	Describe("GetSubsumedParams", func() {
		When("TF HCL is not empty", func() {
			It("succeeds", func() {
				replaceVars := []tf.ReplaceVariable{
					{
						Name:     "resource_type",
						Property: "subsume_key",
					},
					{
						Name:     "other_resource_type",
						Property: "other_subsume_key",
					},
				}
				tfHCL := "# resource_type.resource_name:\nresource \"resource_type\" \"resource_name\" {\nsubsume_key = \"subsume_value\"\n} \n# other_resource_type.resource_name:\nresource \"other_resource_type\" \"resource_name\" {\nother_subsume_key = \"other_subsume_value\"\n}\nOutputs:\nname = \"test-name\""

				output, err := tf.GetSubsumedParameters(tfHCL, replaceVars)

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(map[string]interface{}{"subsume_key": "subsume_value", "other_subsume_key": "other_subsume_value"}))
			})
		})

		When("TF HCL is empty", func() {
			It("succeeds", func() {
				replaceVars := []tf.ReplaceVariable{
					{Name: "resource_type"},
				}
				output, err := tf.GetSubsumedParameters("", replaceVars)

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(map[string]interface{}{}))
			})
		})

	})
})
