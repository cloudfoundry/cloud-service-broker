package hclparser

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type ReplaceVariable struct {
	FieldToRead  string
	FieldToWrite string
}

func GetParameters(tfHCL string, parameters []ReplaceVariable) (map[string]interface{}, error) {
	splitHcl := strings.Split(tfHCL, "Outputs")
	parsedConfig, diags := hclwrite.ParseConfig([]byte(splitHcl[0]), "", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("error parsing subsumed HCL file: %v", diags.Error())
	}

	subsumedParameters := make(map[string]interface{})
	for _, block := range parsedConfig.Body().Blocks() {
		if block.Type() == "resource" {
			for _, param := range parameters {
				resourceName, attributeName := splitResource(param.FieldToRead)
				if resourceName == strings.Join(block.Labels(), ".") {
					subsumedParameters[param.FieldToWrite] = getAttribute(block, attributeName)
				}
			}
		}
	}

	var notFoundParameters []string
	if len(subsumedParameters) != len(parameters) {
		for _, rv := range parameters {
			if _, ok := subsumedParameters[rv.FieldToWrite]; !ok {
				notFoundParameters = append(notFoundParameters, rv.FieldToRead)
			}
		}
		return nil, fmt.Errorf("cannot find required subsumed values for fields: %s", strings.Join(notFoundParameters, ", "))
	}

	return subsumedParameters, nil
}

func splitResource(resource string) (string, string) {
	lastInd := strings.LastIndex(resource, ".")
	return resource[:lastInd], resource[lastInd+1:]
}

func getAttribute(block *hclwrite.Block, attribute string) string {
	return strings.Trim(strings.TrimSpace(string(block.Body().GetAttribute(attribute).Expr().BuildTokens(nil).Bytes())), "\"")
}
