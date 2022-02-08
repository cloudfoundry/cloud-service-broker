package hclparser

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type ReplaceVariable struct {
	Resource     string
	Property     string
	ReplaceField string
}

func GetParameters(tfHCL string, replaceVars []ReplaceVariable) (map[string]interface{}, error) {
	splitHcl := strings.Split(tfHCL, "Outputs")
	parsedConfig, diags := hclwrite.ParseConfig([]byte(splitHcl[0]), "", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("error parsing subsumed HCL file: %v", diags.Error())
	}

	subsumedParameters := make(map[string]interface{})
	for _, block := range parsedConfig.Body().Blocks() {
		if block.Type() == "resource" {
			for _, replaceVar := range replaceVars {
				if replaceVar.Resource == strings.Join(block.Labels(), ".") {
					subsumedParameters[replaceVar.ReplaceField] = getAttribute(block, replaceVar.Property)
				}
			}
		}
	}

	var notFoundParameters []string
	if len(subsumedParameters) != len(replaceVars) {
		for _, rv := range replaceVars {
			if _, ok := subsumedParameters[rv.ReplaceField]; !ok {
				notFoundParameters = append(notFoundParameters, fmt.Sprintf("%s.%s", rv.Resource, rv.Property))
			}
		}
		return nil, fmt.Errorf("cannot find required subsumed values for fields: %s", strings.Join(notFoundParameters, ", "))
	}

	return subsumedParameters, nil
}

func getAttribute(block *hclwrite.Block, attribute string) string {
	return strings.Trim(strings.TrimSpace(string(block.Body().GetAttribute(attribute).Expr().BuildTokens(nil).Bytes())), "\"")
}
