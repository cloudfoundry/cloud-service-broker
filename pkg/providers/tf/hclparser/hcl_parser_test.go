package hclparser_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf/hclparser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HclParser", func() {
	Describe("GetSubsumedParams", func() {
		When("TF HCL contains all replace properties", func() {
			It("succeeds", func() {
				replaceVars := []hclparser.ExtractVariable{
					{
						FieldToRead:  "resource_type.resource_name.subsume_key",
						FieldToWrite: "field_to_replace",
					},
					{
						FieldToRead:  "other_resource_type.resource_name.other_subsume_key",
						FieldToWrite: "other_field_to_replace",
					},
				}
				tfHCL := "# resource_type.resource_name:\nresource \"resource_type\" \"resource_name\" {\nsubsume_key = \"subsume_value\"\n}" +
					" \n# other_resource_type.resource_name:\nresource \"other_resource_type\" \"resource_name\" {\nother_subsume_key = \"other_subsume_value\"\n} " +
					" \nOutputs:\nname = \"test-name\""

				output, err := hclparser.GetParameters(tfHCL, replaceVars)

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(map[string]interface{}{
					"field_to_replace":       "subsume_value",
					"other_field_to_replace": "other_subsume_value",
				}))
			})
		})

		When("TF HCL does not have outputs block", func() {
			It("succeeds", func() {
				replaceVars := []hclparser.ExtractVariable{
					{
						FieldToRead:  "resource_type.resource_name.subsume_key",
						FieldToWrite: "field_to_replace",
					},
				}
				tfHCL := "# resource_type.resource_name:\nresource \"resource_type\" \"resource_name\" {\nsubsume_key = \"subsume_value\"\n}"

				output, err := hclparser.GetParameters(tfHCL, replaceVars)

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(map[string]interface{}{"field_to_replace": "subsume_value"}))
			})
		})

		When("TF HCL does not contain one of the replace vars", func() {
			It("fails", func() {
				replaceVars := []hclparser.ExtractVariable{
					{
						FieldToRead:  "resource_type.resource_name.subsume_key",
						FieldToWrite: "field_to_replace",
					},
					{
						FieldToRead:  "other_resource_type.resource_name.other_subsume_key",
						FieldToWrite: "other_field_to_replace",
					},
				}
				tfHCL := "\n# other_resource_type.resource_name:\nresource \"other_resource_type\" \"resource_name\" {\nother_subsume_key = \"other_subsume_value\"\n}\nOutputs:\nname = \"test-name\""

				_, err := hclparser.GetParameters(tfHCL, replaceVars)

				Expect(err).To(MatchError("cannot find required subsumed values for fields: resource_type.resource_name.subsume_key"))
			})
		})

		When("TF HCL is empty", func() {
			It("fails", func() {
				replaceVars := []hclparser.ExtractVariable{
					{
						FieldToRead:  "resource_type.resource_name.subsume_key",
						FieldToWrite: "field_to_replace",
					},
				}

				_, err := hclparser.GetParameters("", replaceVars)

				Expect(err).To(MatchError("cannot find required subsumed values for fields: resource_type.resource_name.subsume_key"))
			})
		})

		When("TF HCL cannot be parsed", func() {
			It("fails", func() {
				_, err := hclparser.GetParameters("not valid", []hclparser.ExtractVariable{})

				Expect(err).To(MatchError(ContainSubstring("error parsing subsumed HCL file:")))
			})
		})
	})
})
